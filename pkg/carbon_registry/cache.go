package carbon_registry

import (
	"encoding/json"
	"gopkg.in/mcuadros/go-syslog.v2"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CarbonCache struct {
	Data            map[string]*CarbonMetric
	MetricsReceived uint64
	MetricsCount    uint64
	FlushCount      uint64
	FlushErrors     uint64
}

type CarbonMetric struct {
	Source    string  `json:"source"`
	Date      string  `json:"date"`
	Value     float64 `json:"value"`
	Timestamp uint64  `json:"timestamp"`
	Metric    string  `json:"metric"`
	Count     uint64  `json:"count"`
}

func (c *CarbonCache) Listen(channel syslog.LogPartsChannel) {
	log.Println("Start cache listen")
	var message string
	var messageFields []string

	var metric string
	var value float64
	var timestamp uint64

	var source string
	var date string

	var err error

	err = c.Purge()
	if err != nil {
		if err != nil {
			log.Fatal(err)
		}
	}

	for line := range channel {
		c.MetricsReceived++

		message = line["message"].(string)
		date = line["timestamp"].(time.Time).String()
		source = line["hostname"].(string)

		messageFields = strings.Fields(message)
		metric = string(messageFields[0])
		value, err = strconv.ParseFloat(messageFields[1], 64)
		if err != nil {
			log.Println(err)
			continue
		}
		timestamp, err = strconv.ParseUint(messageFields[2], 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}

		c.Receive(metric, source, date, value, timestamp)
	}
}

func (c *CarbonCache) Receive(metric string, source string, date string, value float64, timestamp uint64) {
	//log.Printf("Receive: %s %f %d from: %s at: %s\n", metric, value, timestamp, source, date)

	var record *CarbonMetric
	var found bool

	record, found = c.Data[metric]
	if found {
		record.Source = source
		record.Date = date
		record.Value = value
		record.Timestamp = timestamp
		record.Count++
	} else {
		record = &CarbonMetric{
			Source:    source,
			Date:      date,
			Value:     value,
			Timestamp: timestamp,
			Metric:    metric,
			Count:     1,
		}
		c.Data[metric] = record
		c.MetricsCount++
	}
}

func (c *CarbonCache) Dump() (error, string) {
	metrics := make([]*CarbonMetric, 0, len(c.Data))

	for _, value := range c.Data {
		metrics = append(metrics, value)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Metric < metrics[j].Metric
	})

	jsonDump, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		return err, ""
	}
	return nil, string(jsonDump)
}

func (c *CarbonCache) Purge() error {
	c.Data = make(map[string]*CarbonMetric)
	c.MetricsReceived = 0
	c.MetricsCount = 0
	return nil
}

func NewCarbonCache() *CarbonCache {
	return &CarbonCache{}
}
