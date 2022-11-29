package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	client "github.com/weilinfox/youmu-thlink/client/lib"

	"github.com/sirupsen/logrus"
)

func main() {

	localPort := flag.Int("p", 10080, "local port will connect to")
	server := flag.String("s", "thlink.inuyasha.love:4646", "hostname of server")
	tunnelType := flag.String("t", "tcp", "tunnel type, support tcp and quic")
	debug := flag.Bool("d", false, "debug mode")

	flag.Parse()

	if *localPort <= 0 || *localPort > 65535 {
		fmt.Println("Invalid port ", localPort)
		os.Exit(1)
	}

	host, port, err := net.SplitHostPort(*server)
	if err != nil {
		fmt.Println("Invalid hostname ", server)
	}
	port64, err := strconv.ParseInt(port, 10, 32)
	if port64 <= 0 || port64 > 65535 {
		fmt.Println("Invalid port ", port64)
		os.Exit(1)
	}

	if strings.ToLower(*tunnelType) != "tcp" && strings.ToLower(*tunnelType) != "quic" {
		fmt.Println("Invalid tunnel type ", *tunnelType)
		os.Exit(1)
	}

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	client.Main(*localPort, host, int(port64), (*tunnelType)[0])

	// fmt.Println("Enter to quit")
	// _, _ = fmt.Scanln()
}
