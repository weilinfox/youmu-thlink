package main

import (
	"fmt"
	"net"
	"strconv"

	client "github.com/weilinfox/youmu-thlink/client/lib"
)

func main() {

	port := "10800"
	serverHost := "thlink.inuyasha.love"
	sPort := "4646"
	var localPort, serverPort int

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

	client.Main(localPort, serverHost, serverPort)

	fmt.Println("Enter to quit")
	fmt.Scanln()
}
