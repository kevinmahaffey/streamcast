package streamcast

import (
	"fmt"
	"net"
)

type Rx struct {
	conn *net.UDPConn
	addr net.Addr
}

func NewBroadcastRx(network string, port int) (err error, r *Rx) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", network, port))
	if err != nil {
		return err, nil
	}
	return NewRx(addr)
}

func NewRx(addr *net.UDPAddr) (err error, r *Rx) {
	r = new(Rx)
	r.addr = addr
	r.conn, err = net.ListenUDP("udp4", addr)
	if err != nil {
		return err, nil
	}
	return
}

func (r *Rx) Read(b []byte) (n int, err error) {
	n, _, err = r.conn.ReadFromUDP(b)
	if err != nil {
		return 0, err
	}
	if n >= len(b) {
		return 0, fmt.Errorf("Read overflow")
	}

	return
}
