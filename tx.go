package streamcast

import (
	"fmt"
	"net"
)

type Tx struct {
	conn net.PacketConn
	addr net.Addr
}

func NewBroadcastTx(network string, port int) (err error, s *Tx) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", network, port))
	if err != nil {
		return err, nil
	}
	return NewTx(addr)
}

func NewTx(addr net.Addr) (err error, s *Tx) {
	s = new(Tx)
	s.addr = addr
	s.conn, err = net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return err, nil
	}

	return
}

func (s *Tx) Write(b []byte) (err error) {
	n, err := s.conn.WriteTo(b, s.addr)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("Could not write full chunk %d/%d", n, len(b))
	}
	return
}
