package main

import (
	"flag"

	client "github.com/weilinfox/youmu-thlink/client/lib"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client", "main")

func main() {

	localPort := flag.Int("p", client.DefaultLocalPort, "local port will connect to")
	server := flag.String("s", client.DefaultServerHost, "hostname of server")
	tunnelType := flag.String("t", client.DefaultTunnelType, "tunnel type, support tcp and quic")
	debug := flag.Bool("d", false, "debug mode")

	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	c, err := client.New(*localPort, *server, *tunnelType)
	if err != nil {
		logger.WithError(err).Fatal("Start client error")
	}
	defer c.Close()
	err = c.Serve()
	if err != nil {
		logger.WithError(err).Fatal("Serve client error")
	}

	// fmt.Println("Enter to quit")
	// _, _ = fmt.Scanln()
}
