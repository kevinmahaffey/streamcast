package streamcast

import (
	"fmt"
	"net"
	"time"
)

type UdpTx struct {
	conn         net.PacketConn
	addr         net.Addr
	copiesToSend int
	currentId    uint32
	timeout      time.Duration
}

func NewUdpTx(network string, port int, copiesToSend int) (s *UdpTx, err error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", network, port))
	if err != nil {
		return nil, err
	}
	s = new(UdpTx)
	s.addr = addr
	s.conn, err = net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	s.currentId = 1
	s.copiesToSend = copiesToSend
	s.timeout = 1 * time.Second

	return
}

func (s *UdpTx) WriteFrame(f *Frame) (err error) {
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

func (s *UdpTx) Write(metadata []byte, data []byte) (err error) {
	var f Frame
	f.Data = data
	f.Metadata = metadata
	f.FrameId = s.currentId
	s.currentId += 1
	s.WriteFrame(&f)
	return
}

func (s *UdpTx) SetTimeout(t time.Duration) {
	s.timeout = t
}

func (t *UdpTx) Close() {
	t.conn.Close()
}
