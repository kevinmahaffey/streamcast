package main

import (
	"github.com/kevinmahaffey/streamcast"
	"os"
)

func main() {
	err, src := streamcast.NewUdpTx(os.Args[1], 1337, 3)
	if err != nil {
		panic(err)
	}
	src.Write([]byte{4,5,6}, []byte{1, 2, 3})
}
