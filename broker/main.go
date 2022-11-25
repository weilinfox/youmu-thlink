package main

import (
	"fmt"

	broker "github.com/weilinfox/youmu-thlink/broker/lib"
)

func main() {

	listenHost := "0.0.0.0:4646"

	broker.Main(listenHost)

	fmt.Println("Enter to quit")
	fmt.Scanln()
}
