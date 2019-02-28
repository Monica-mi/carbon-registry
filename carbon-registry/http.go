package carbon_registry

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type CarbonHTTP struct {
	Cache        *CarbonCache
	Host         string
	Port         uint16
	InstanceName string
	HostName     string
	HTTPRequests uint64
	HTTPErrors   uint64
}

type CarbonHTTPStatus struct {
	Status         string `json:"status"`
	MetricReceived uint64 `json:"metric_received"`
	MetricCount    uint64 `json:"metric_count"`
	InstanceName   string `json:"instance_name"`
	HostName       string `json:"host_name"`
	FlushCount     uint64 `json:"flush_count"`
	FlushErrors    uint64 `json:"flush_errors"`
	HTTPRequests   uint64 `json:"http_requests"`
	HTTPErrors     uint64 `json:"http_errors"`
}

func (c *CarbonHTTPStatus) Dump() (error, string) {
	jsonDump, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err, ""
	}
	return nil, string(jsonDump)
}

func (c *CarbonHTTP) CacheHandler(writer http.ResponseWriter, request *http.Request) {
	c.HTTPRequests++

	var err error
	var text string

	err, text = c.Cache.Dump()
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	_, err = fmt.Fprintln(writer, text)
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *CarbonHTTP) GetStatus() *CarbonHTTPStatus {
	return &CarbonHTTPStatus{
		Status:         "OK",
		MetricCount:    uint64(len(c.Cache.Data)),
		MetricReceived: uint64(c.Cache.MetricsReceived),
		InstanceName:   c.InstanceName,
		HostName:       c.HostName,
		FlushCount:     c.Cache.FlushCount,
		FlushErrors:    c.Cache.FlushErrors,
		HTTPRequests:   c.HTTPRequests,
		HTTPErrors:     c.HTTPErrors,
	}
}

func (c *CarbonHTTP) StatusHandler(writer http.ResponseWriter, request *http.Request) {
	c.HTTPRequests++

	var err error
	var text string

	status := c.GetStatus()
	err, text = status.Dump()
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	_, err = fmt.Fprintln(writer, text)
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *CarbonHTTP) MetricsHandler(writer http.ResponseWriter, request *http.Request) {
	c.HTTPRequests++

	var err error
	var text string

	status := c.GetStatus()
	text = fmt.Sprintf(`# Prometheus metrics for carbon-registry #
carbon_registry_status{instance="%s",host="%s"} %d
carbon_registry_metric_count{instance="%s",host="%s"} %d
carbon_registry_metric_received{instance="%s",host="%s"} %d
carbon_registry_flush_count{instance="%s",host="%s"} %d
carbon_registry_flush_errors{instance="%s",host="%s"} %d
carbon_registry_http_requests{instance="%s",host="%s"} %d
carbon_registry_http_errors{instance="%s",host="%s"} %d
`,
		status.HostName, status.InstanceName, 1,
		status.HostName, status.InstanceName, status.MetricCount,
		status.HostName, status.InstanceName, status.MetricReceived,
		status.HostName, status.InstanceName, status.FlushCount,
		status.HostName, status.InstanceName, status.FlushErrors,
		status.HostName, status.InstanceName, status.HTTPRequests,
		status.HostName, status.InstanceName, status.HTTPErrors,
	)

	_, err = fmt.Fprintln(writer, text)
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *CarbonHTTP) Start() {
	log.Printf("Start HTTP on %s:%d with instance: %s and host: %s\n", c.Host, c.Port, c.InstanceName, c.HostName)
	var err error

	http.HandleFunc("/", c.StatusHandler)
	http.HandleFunc("/cache", c.CacheHandler)
	http.HandleFunc("/metrics", c.MetricsHandler)

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", c.Host, c.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func NewCarbonHTTP(cache *CarbonCache) *CarbonHTTP {
	var err error

	carbonHTTP := &CarbonHTTP{
		Cache: cache,
		Host:  "0.0.0.0",
		Port:  8084,
	}

	if carbonHTTP.HostName == "" {
		carbonHTTP.HostName, err = os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
	}

	if carbonHTTP.InstanceName == "" {
		carbonHTTP.InstanceName = "main"
	}

	return carbonHTTP
}
