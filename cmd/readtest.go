package main

import (
	"fmt"
	"github.com/kevinmahaffey/streamcast"
	"os"
)

func main() {
	err, rx := streamcast.NewBroadcastRx(os.Args[1], 1337)
	if err != nil {
		panic(err)
	}
	var data [10000]byte
	for {
		n, err := rx.Read(data[:])
		if err != nil {
			panic(err)
		}
		fmt.Print("Read ", n)
	}
}
