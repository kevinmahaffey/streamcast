package main

import (
	"fmt"
	"github.com/kevinmahaffey/streamcast"
	"os"
	"time"
)

func main() {
	rx, err := streamcast.NewRxIsochronous(os.Args[1], os.Args[2], 1337, 100 * time.Millisecond, 5 * time.Millisecond)
	if err != nil {
		panic(err)
	}
	for {
		n, err := rx.Read()
		if err != nil {
			panic(err)
		}
		fmt.Print("Read ", n)
	}
}
