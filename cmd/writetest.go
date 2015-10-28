package main

import (
	"github.com/kevinmahaffey/streamcast"
	"os"
)

func main() {
	err, src := streamcast.NewBroadcastTx(os.Args[1], 1337)
	if err != nil {
		panic(err)
	}
	src.Write([]byte{1, 2, 3})
}
