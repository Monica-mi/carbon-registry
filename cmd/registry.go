package main

import (
	"../pkg/carbon_registry"
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"time"
)

func main() {
	flag.Parse()

	args := flag.Args()

	var configFilePath string
	if len(args) >= 1 {
		configFilePath = args[0]
	} else {
		configFilePath = "config.ini"
	}
	cfg, err := ini.Load(configFilePath)
	if err != nil {
		log.Fatalf("Could not read the config file: '%s' - %s", configFilePath, err)
	}
	logLevelName := cfg.Section("log").Key("level").MustString("info")
	logLevel, err := log.ParseLevel(logLevelName)
	if err != nil {
		log.Fatalf("Could not parse the log level: '%s' - %s", logLevelName, err)
	}
	log.SetLevel(logLevel)
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)

	log.Infof("Start carbon-registry with the config file: '%s'", configFilePath)

	carbonSyslog := carbon_registry.NewCarbonSyslog()
	carbonSyslog.Host = cfg.Section("syslog").Key("host").MustString("0.0.0.0")
	carbonSyslog.Port = uint16(cfg.Section("syslog").Key("port").MustInt(2033))
	carbonSyslog.Start()

	carbonCache := carbon_registry.NewCarbonCache()
	go carbonCache.Listen(carbonSyslog.Channel)

	carbonHTTP := carbon_registry.NewCarbonHTTP(carbonCache)
	carbonHTTP.Host = cfg.Section("http").Key("host").MustString("0.0.0.0")
	carbonHTTP.Port = uint16(cfg.Section("http").Key("port").MustInt(8084))
	carbonHTTP.InstanceName = cfg.Section("http").Key("instance").MustString("main")
	carbonHTTP.HostName = cfg.Section("http").Key("hostname").MustString("")
	carbonHTTP.IndexFile = cfg.Section("http").Key("index").MustString("status")
	carbonHTTP.Prefix = cfg.Section("http").Key("prefix").MustString("/")
	carbonHTTP.SearchParameter = cfg.Section("http").Key("search_parameter").MustString("search[value]")
	go carbonHTTP.Start()

	carbonFlush := carbon_registry.NewCarbonFlush(carbonCache)
	carbonFlush.Interval, err = time.ParseDuration(cfg.Section("flush").Key("interval").MustString("24h"))
	if err != nil {
		log.Fatalf("Could not parse the flush interval - %s", err)
	}
	carbonFlush.FileEnabled = cfg.Section("flush").Key("file").MustBool(true)
	carbonFlush.FilePath = cfg.Section("flush").Key("path").MustString("graphite-metrics-2006-01-02_15-04-05.json")
	carbonFlush.LogEnabled = cfg.Section("flush").Key("log").MustBool(false)
	carbonFlush.PurgeEnabled = cfg.Section("flush").Key("purge").MustBool(false)
	go carbonFlush.Start()

	carbonSyslog.Wait()
}
