package streamcast

import (
	"errors"
	"fmt"
	"time"
)

type Tx interface {
	Write(metadata []byte, data []byte) (err error)
	SetTimeout(t time.Duration)
	Close()
}

func NewTx(protocol string, network string, port int) (t Tx, err error) {
	switch protocol {
	case "tcp":
		t, err = NewTcpTx(network, port)
	case "udp":
		t, err = NewUdpTx(network, port, 1)
	default:
		err = errors.New(fmt.Sprintf("Unsupported Protocol: %s.", protocol))
	}
	if err != nil {
		return nil, err
	}
	return
}
