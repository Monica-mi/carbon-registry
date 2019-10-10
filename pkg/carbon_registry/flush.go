package carbon_registry

import (
	"compress/gzip"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type CarbonFlush struct {
	Cache        *CarbonCache
	Interval     time.Duration
	FileEnabled  bool
	FilePath     string
	LogEnabled   bool
	PurgeEnabled bool
}

func (c *CarbonFlush) Start() {
	log.Infof("Start flush with interval: %s", c.Interval.String())
	var err error

	for {
		time.Sleep(c.Interval)
		c.Cache.FlushCount++

		c.OutputLog()
		c.OutputFile()
		if c.PurgeEnabled {
			err = c.Cache.Purge()
			if err != nil {
				log.Println(err)
				c.Cache.FlushErrors++
				continue
			}
		}
	}
}

func (c *CarbonFlush) OutputLog() {
	if c.LogEnabled {
		err, text := c.Cache.DumpPlain()
		if err != nil {
			log.Errorf("Could not dump plain JSON - %s", err)
			c.Cache.FlushErrors++
			return
		}
		log.Printf("%s", text)
	}
}

func (c *CarbonFlush) OutputFile() {
	if c.FileEnabled {
		err, text := c.Cache.DumpPretty()
		if err != nil {
			log.Errorf("Could not dump pretty JSON - %s", err)
			c.Cache.FlushErrors++
			return
		}
		filePath := time.Now().Format(c.FilePath)
		log.Debugf("Dump cache to file: '%s'", filePath)

		fileWriter, err := os.Create(filePath)
		if err != nil {
			log.Errorf("Could not create file: '%s' - %s", filePath, err)
			return
		}
		defer fileWriter.Close()

		gzipWriter := gzip.NewWriter(fileWriter)
		defer gzipWriter.Close()

		_, err = gzipWriter.Write([]byte(text + "\n"))
		if err != nil {
			log.Errorf("Could not write to file: '%s' - %s", filePath, err)
			c.Cache.FlushErrors++
			return
		}
		err = gzipWriter.Flush()
		if err != nil {
			log.Errorf("Could not flush to file: '%s' - %s", filePath, err)
			c.Cache.FlushErrors++
			return
		}
	}
}

func NewCarbonFlush(cache *CarbonCache) *CarbonFlush {
	return &CarbonFlush{
		Cache:        cache,
		Interval:     time.Hour * 24,
		FileEnabled:  true,
		FilePath:     "graphite-metrics-2006-01-02_15-04-05.json.gz",
		LogEnabled:   true,
		PurgeEnabled: false,
	}
}
