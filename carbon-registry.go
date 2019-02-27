package main

import (
	"fmt"
	"gopkg.in/mcuadros/go-syslog.v2"
	"strings"
)

// map[msg_id:- structured_data:- message:n.lg1.serv.lg1mnc.ac4mnc1.net_out 6557544000.0 1551290400 client:127.0.0.1:37072 tls_peer: priority:30 facility:3 timestamp:2019-02-27 18:05:16 +0000 UTC app_name:carbon-c-relay proc_id:- severity:6 version:1 hostname:172.31.101.80]

type CarbonMetric struct {
	Source string `json:"source"`
	Value float64 `json:"value"`
	Date string `json:"date"`
}

type CarbonMetrics struct {
	Cache map[string]CarbonMetric{}
}

func (cm *CarbonMetrics) listen(channel syslog.LogPartsChannel) {
	cm.Cache = make(map[string]CarbonMetric)

	var message string
	var messageFields []string

	var metric string
	var value float64
	var timestamp int64

	var source string
	var date string
	var found bool

	for line := range channel {
		// fmt.Println(metric)
		message = line["message"].(string)
		messageFields = strings.Fields(message)
        metric = messageFields[0]
        value = messageFields[1]
        timestamp = messageFields[2]

		//_, found = cm.Cache[metric]
		//if ok {
		//	//do something here
		//}

	}
}


func main() {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)
	server.SetHandler(handler)
	err := server.ListenUDP("0.0.0.0:2033")
	if err != nil {
		panic(err)
	}

	err = server.Boot()
	if err != nil {
		panic(err)
	}

	go processMessages(channel)

	server.Wait()
}
