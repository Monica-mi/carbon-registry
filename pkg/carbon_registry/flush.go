package carbon_registry

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
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
	var text string

	for {
		time.Sleep(c.Interval)
		c.Cache.FlushCount ++

		c.OutputLog()
		c.OutputFile(text)
		if c.PurgeEnabled {
			err = c.Cache.Purge()
			if err != nil {
				log.Println(err)
				c.Cache.FlushErrors ++
				continue
			}
		}
	}
}

func (c *CarbonFlush) OutputLog() {
	if c.LogEnabled {
		err, text := c.Cache.DumpPlain()
		if err != nil {
			log.Println(err)
			c.Cache.FlushErrors ++
			return
		}
		log.Infof("%s", text)
	}
}

func (c *CarbonFlush) OutputFile(text string) {
	if c.FileEnabled {
		filePath := time.Now().Format(c.FilePath)
		log.Infof("Dump cache to file: '%s'", filePath)

		data := []byte(text)
		err := ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			log.Println(err)
			c.Cache.FlushErrors ++
		}
	}
}

func NewCarbonFlush(cache *CarbonCache) *CarbonFlush {
	return &CarbonFlush{
		Cache:        cache,
		Interval:     time.Hour * 24,
		FileEnabled:  true,
		FilePath:     "graphite-metrics-2006-01-02_15-04-05.json",
		LogEnabled:   true,
		PurgeEnabled: false,
	}
}
