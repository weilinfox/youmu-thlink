package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	client "github.com/weilinfox/youmu-thlink/client/lib"
	"sort"
)

var logger = logrus.WithField("client", "main")

func main() {

	localPort := flag.Int("p", client.DefaultLocalPort, "local port will connect to")
	server := flag.String("s", client.DefaultServerHost, "hostname of server")
	tunnelType := flag.String("t", client.DefaultTunnelType, "tunnel type, support tcp and quic")
	autoSelect := flag.Bool("a", true, "auto select broker in network with lowest latency")
	noAutoSelect := flag.Bool("na", false, "DO NOT auto select broker in network with lowest latency (override -a)")
	plugin := flag.Int("l", 0, "enable plugin, 123 for hisoutensoku spectacle support")
	debug := flag.Bool("d", false, "debug mode")

	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	chooseBroker := *server
	if *autoSelect && !*noAutoSelect {

		// ping delays
		serverDelays, err := client.NetBrokerDelay(*server)
		if err != nil {
			logger.WithError(err).Fatal("Get broker delay in network failed")
		}

		// int=>string
		delayServers := make(map[int]string)
		sortDelay := make([]int, len(serverDelays))
		i := 0
		for k, v := range serverDelays {
			delayServers[v] = k
			sortDelay[i] = v
			i++
		}

		// sort ping delays
		sort.Ints(sortDelay)

		// print 5 of low latency brokers
		for i = 0; i < 5 && i < len(delayServers); i++ {
			// drop >200ms
			if sortDelay[i] >= 1000000*200 {
				break
			}
			logger.Infof("%.3fms %s", float64(sortDelay[i])/1000000, delayServers[sortDelay[i]])
		}
		chooseBroker = delayServers[sortDelay[0]]
	}

	c, err := client.New(*localPort, chooseBroker, *tunnelType)
	if err != nil {
		logger.WithError(err).Fatal("Start client error")
	}
	defer c.Close()

	tunnelVersion, version := c.Version()
	logger.Info("Client v", version, " with tunnel version ", tunnelVersion)
	brokerTVersion, brokerVersion := c.BrokerVersion()
	logger.Info("Broker v", brokerVersion, " with tunnel version ", brokerTVersion)
	if tunnelVersion != brokerTVersion {
		logger.Warn("Broker tunnel version code not match, there may have compatible issue")
	}

	err = c.Connect()
	if err != nil {
		logger.WithError(err).Fatal("Client connect error")
	}

	switch *plugin {
	case 123:
		logger.Info("Append th12.3 hisoutensoku plugin")
		h := client.NewHisoutensoku()
		err = c.Serve(h.ReadFunc, h.WriteFunc)
	default:
		err = c.Serve(nil, nil)
	}
	if err != nil {
		logger.WithError(err).Fatal("Serve client error")
	}

	// fmt.Println("Enter to quit")
	// _, _ = fmt.Scanln()
}
