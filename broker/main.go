package main

import (
	"flag"
	"fmt"

	broker "github.com/weilinfox/youmu-thlink/broker/lib"
)

func main() {

	listenHost := flag.String("s", "0.0.0.0:4646", "listen hostname")

	flag.Parse()

	broker.Main(*listenHost)

	fmt.Println("Enter to quit")
	_, _ = fmt.Scanln()

}
