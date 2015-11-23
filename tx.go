package streamcast

import (
	"fmt"
	"net"
	"time"
)

type Tx struct {
	conn         net.PacketConn
	addr         net.Addr
	copiesToSend int
	currentId    uint32
	timeout      time.Duration
}

func NewBroadcastTx(network string, port int, copiesToSend int) (err error, s *Tx) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", network, port))
	if err != nil {
		return err, nil
	}
	return NewTx(addr, copiesToSend)
}

func NewTx(addr net.Addr, copiesToSend int) (err error, s *Tx) {
	s = new(Tx)
	s.addr = addr
	s.conn, err = net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return err, nil
	}
	s.currentId = 1
	s.copiesToSend = copiesToSend
	s.timeout = 1 * time.Second

	return
}

func (s *Tx) WriteFrame(f *Frame) (err error) {
	var b [MAX_FRAME_LENGTH]byte
	s.conn.SetDeadline(time.Now().Add(s.timeout))

	n, err := f.Write(b[:])
	if err != nil {
		return err
	}
	for i := 0; i < s.copiesToSend; i++ {
		written, err := s.conn.WriteTo(b[:n], s.addr)
		if err != nil {
			return err
		}
		if n != written {
			return fmt.Errorf("Could not write full chunk %d/%d", written, n)
		}
	}
	return
}

func (s *Tx) Write(metadata []byte, data []byte) (err error) {
	var f Frame
	f.Data = data
	f.Metadata = metadata
	f.FrameId = s.currentId
	s.currentId += 1
	s.WriteFrame(&f)
	return
}

func (s *Tx) SetTimeout(t time.Duration) {
	s.timeout = t
}

func (t *Tx) Close() {
	t.conn.Close()
}
