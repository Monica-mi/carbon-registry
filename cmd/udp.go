package main

import (
	"fmt"
	"log"
	"net"
)

const message1 = "<30>1 2019-02-28T12:28:59.000Z 172.31.101.80 carbon-c-relay - - - n.lg3.tbg.serv.lg3tbg.ac6tbg1.net_in 62012080.0 1551356400"
const message2 = "<30>1 2019-02-28T12:28:59.000Z 172.31.101.80 carbon-c-relay - - - n.lg3.tbg.serv.lg3tbg.ac6tbg1.net_out 32012080.0 1551356400"

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:2033")
	if err != nil {
		log.Fatal(err)
	}
	_, err = fmt.Fprintln(conn, message1)
	_, err = fmt.Fprintln(conn, message2)
	if err != nil {
		log.Fatal(err)
	}
	err = conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}
