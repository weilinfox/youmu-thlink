package main

import (
	"flag"
	"fmt"

	broker "github.com/weilinfox/youmu-thlink/broker/lib"

	"github.com/sirupsen/logrus"
)

func main() {

	listenHost := flag.String("s", "0.0.0.0:4646", "listen hostname")
	upperHost := flag.String("u", "", "upper broker hostname")
	debug := flag.Bool("d", false, "debug mode")

	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	broker.Main(*listenHost, *upperHost)

	fmt.Println("Enter to quit")
	_, _ = fmt.Scanln()

}
