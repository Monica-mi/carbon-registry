package carbon_registry

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type CarbonHTTP struct {
	Cache        *CarbonCache
	Host         string
	Port         uint16
	InstanceName string
	HostName     string
	HTTPRequests uint64
	HTTPErrors   uint64
	Prefix       string
	IndexFile    string
	SearchParameter string
}

type CarbonHTTPSearchResponse struct {
	Draw            int            `json:"draw"`
	RecordsTotal    uint64         `json:"recordsTotal"`
	RecordsFiltered uint64         `json:"recordsFiltered"`
	Data            []CarbonMetric `json:"data"`
	Error           string         `json:"error"`
}

func (c *CarbonHTTPSearchResponse) Dump() (error, string) {
	jsonDump, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err, ""
	}
	return nil, string(jsonDump)
}

type CarbonHTTPStatus struct {
	Status         string `json:"status"`
	MetricReceived uint64 `json:"metric_received"`
	MetricsErrors  uint64 `json:"metrics_errors"`
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

	err, text = c.Cache.DumpPretty()
	if err != nil {
		c.HTTPErrors++
		log.Errorf("Could not dump pretty JSON - %s", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	_, err = fmt.Fprintln(writer, text)
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *CarbonHTTP) SearchHandler(writer http.ResponseWriter, request *http.Request) {
	c.HTTPRequests++

	var err error
	var search string
	var draw int
	var responseText string

	keys := request.URL.Query()
	search = keys.Get(c.SearchParameter)
	draw, err = strconv.Atoi(keys.Get("draw"))
	if err != nil {
		draw = 0
	}

	log.Debug(keys)

	searchResponse := CarbonHTTPSearchResponse{
		Draw: draw,
		Data: make([]CarbonMetric, 0, 0),
		RecordsFiltered: 0,
		RecordsTotal: c.Cache.MetricsCount,
		Error: "",
	}

	if len(search) > 1 {
		for _, record := range c.Cache.Data {
			if strings.Contains(record.Metric, search) || strings.Contains(record.Source, search) {
				searchResponse.Data = append(searchResponse.Data, *record)
			}
			searchResponse.RecordsFiltered++
		}
	}

	err, responseText = searchResponse.Dump()
	if err != nil {
		c.HTTPErrors++
		log.Errorf("Could not dump pretty JSON - %s", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	_, err = fmt.Fprintln(writer, responseText)
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *CarbonHTTP) GetStatus() *CarbonHTTPStatus {
	return &CarbonHTTPStatus{
		Status:         "OK",
		MetricReceived: c.Cache.MetricsReceived,
		MetricsErrors:  c.Cache.MetricsErrors,
		MetricCount:    c.Cache.MetricsCount,
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
		log.Errorf("Could not dump status - %s", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	_, err = fmt.Fprintln(writer, text)
	if err != nil {
		c.HTTPErrors++
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *CarbonHTTP) IndexHandler(writer http.ResponseWriter, request *http.Request) {
	if c.IndexFile == "status" {
		c.StatusHandler(writer, request)
		return
	}
	c.HTTPRequests++
	http.ServeFile(writer, request, c.IndexFile)
}

func (c *CarbonHTTP) MetricsHandler(writer http.ResponseWriter, request *http.Request) {
	c.HTTPRequests++

	var err error
	var text string

	status := c.GetStatus()
	text = fmt.Sprintf(`# Prometheus metrics for carbon-registry #
carbon_registry_status{instance="%s",host="%s"} %d
carbon_registry_metric_received{instance="%s",host="%s"} %d
carbon_registry_metric_errors{instance="%s",host="%s"} %d
carbon_registry_metric_count{instance="%s",host="%s"} %d
carbon_registry_flush_count{instance="%s",host="%s"} %d
carbon_registry_flush_errors{instance="%s",host="%s"} %d
carbon_registry_http_requests{instance="%s",host="%s"} %d
carbon_registry_http_errors{instance="%s",host="%s"} %d
`,
		status.HostName, status.InstanceName, 1,
		status.HostName, status.InstanceName, status.MetricReceived,
		status.HostName, status.InstanceName, status.MetricsErrors,
		status.HostName, status.InstanceName, status.MetricCount,
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
	log.Infof("Start HTTP on %s:%d with instance: '%s' and host: '%s'", c.Host, c.Port, c.InstanceName, c.HostName)
	var err error

	router := mux.NewRouter()
	router.HandleFunc(c.Prefix, c.IndexHandler)

	subrouter := router.PathPrefix(c.Prefix).Subrouter()
	subrouter.HandleFunc("/", c.IndexHandler)
	subrouter.HandleFunc("/cache", c.CacheHandler)
	subrouter.HandleFunc("/status", c.StatusHandler)
	subrouter.HandleFunc("/metrics", c.MetricsHandler)
	subrouter.HandleFunc("/search", c.SearchHandler)

	http.Handle("/", router)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", c.Host, c.Port), nil)
	if err != nil {
		log.Fatalf("Could not start HTTP server - %s", err)
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
			log.Fatalf("Could not get system hostname - %s", err)
		}
	}

	if carbonHTTP.InstanceName == "" {
		carbonHTTP.InstanceName = "main"
	}

	if carbonHTTP.IndexFile == "" {
		carbonHTTP.IndexFile = "status"
	}

	return carbonHTTP
}
