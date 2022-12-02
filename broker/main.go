package main

import (
	"flag"
	"fmt"

	broker "github.com/weilinfox/youmu-thlink/broker/lib"
)

func main() {

	listenHost := flag.String("s", "0.0.0.0:4646", "listen hostname")
	upperHost := flag.String("u", "", "upper broker hostname")

	flag.Parse()

	broker.Main(*listenHost, *upperHost)

	fmt.Println("Enter to quit")
	_, _ = fmt.Scanln()

}
