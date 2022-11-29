package main

import (
	"fmt"
	"net"
	"strconv"

	client "github.com/weilinfox/youmu-thlink/client/lib"

	"github.com/sirupsen/logrus"
)

func main() {

	port := "10800"
	serverHost := "thlink.inuyasha.love"
	sPort := "4646"
	var localPort, serverPort int
	var tunnelType string

	// local port 花 17723/10800 则 10800
	for {
		fmt.Println()
		fmt.Println("Input local port (default: 10800)")
		_, _ = fmt.Scanln(&port)

		localPort64, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			fmt.Println("Invalid input port")
			continue
		}
		localPort = int(localPort64)
		if localPort <= 0 || localPort > 65535 {
			fmt.Println("Invalid port ", localPort)
			continue
		}

		break
	}

	// broker address
	for {
		fmt.Println()
		fmt.Println("Input broker address (default: thlink.inuyasha.love)")
		_, _ = fmt.Scanln(&serverHost)
		_, err := net.ResolveUDPAddr("udp", serverHost+":0")
		if err != nil {
			fmt.Println("Cannot resolve host: ", serverHost)
			continue
		}

		break
	}

	// broker port
	for {
		fmt.Println()
		fmt.Println("Input broker port (default: 4646)")
		_, _ = fmt.Scanln(&sPort)

		serverPort64, err := strconv.ParseInt(sPort, 10, 32)
		if err != nil {
			fmt.Println("Invalid input port")
			continue
		}
		serverPort = int(serverPort64)
		if serverPort <= 0 || serverPort > 65535 {
			fmt.Println("Invalid port ", serverPort)
			continue
		}

		break
	}

	// tunnel type
	fmt.Println()
	fmt.Println("Input tunnel type tcp/quic (default: tcp)")
	_, _ = fmt.Scanln(&tunnelType)

	if tunnelType == "" {
		tunnelType = "tcp"
	} else {

		switch tunnelType[0] {
		case 'q' | 'Q':
			fmt.Println("Use QUIC tunnel")
			tunnelType = "quic"
		case 't' | 'T':
			fmt.Println("Use TCP tunnel")
			tunnelType = "tcp"
		default:
			fmt.Println("No such tunnel type, fallback to TCP")
			tunnelType = "tcp"
		}

	}

	var debug string
	fmt.Println()
	fmt.Println("Show debug info? (Y/n)")
	fmt.Scanln(&debug)
	if debug == "" {
		debug = "Y"
	}
	switch debug[0] {
	case 'N' | 'n':
		fmt.Println("Set logger level to info")
		logrus.SetLevel(logrus.InfoLevel)
	default:
		fmt.Println("Set logger level to debug")
		logrus.SetLevel(logrus.DebugLevel)
	}

	client.Main(localPort, serverHost, serverPort, tunnelType[0])

	fmt.Println("Enter to quit")
	fmt.Scanln()
}
