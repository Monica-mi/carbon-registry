package carbon_registry

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mcuadros/go-syslog.v2"
)

type CarbonSyslog struct {
	Host    string
	Port    uint16
	Channel syslog.LogPartsChannel
	Server  *syslog.Server
}

func (c *CarbonSyslog) Start() {
	log.Infof("Start syslog on %s:%d", c.Host, c.Port)

	c.Channel = make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(c.Channel)

	c.Server = syslog.NewServer()
	c.Server.SetFormat(syslog.RFC5424)
	c.Server.SetHandler(handler)

	err := c.Server.ListenUDP(fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		log.Fatalf("Could not listen UDP - %s", err)
	}

	err = c.Server.Boot()
	if err != nil {
		log.Fatalf("Could not start syslog server - %s", err)
	}
}

func (c *CarbonSyslog) Wait() {
	c.Server.Wait()
}

func NewCarbonSyslog() *CarbonSyslog {
	return &CarbonSyslog{
		Host: "0.0.0.0",
		Port: 2033,
	}
}
