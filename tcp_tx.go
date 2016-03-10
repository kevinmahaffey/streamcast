package streamcast

import (
	"fmt"
	"time"
)

type TcpTx struct {
  addr         string
  tcpServer   *TcpServer
	currentId    uint32
	timeout      time.Duration
}

func NewTcpTx(network string, port int) (err error, s *TcpTx) {
  addr := fmt.Sprintf("%s:%d", network, port)
	s = new(TcpTx)
	s.addr = addr
	err, s.tcpServer = StartTcpServer(addr)
	if err != nil {
		return err, nil
	}
	s.currentId = 1
	s.timeout 	= 1 * time.Second

	return
}

func (s *TcpTx) WriteFrame(f *Frame) (err error) {
	var b [MAX_FRAME_LENGTH]byte

	n, err := f.Write(b[:])
	if err != nil {
		return err
	}
  s.tcpServer.Broadcast(b[:n])
	return
}

func (s *TcpTx) Write(metadata []byte, data []byte) (err error) {
	var f Frame
	f.Data = data
	f.Metadata = metadata
	f.FrameId = s.currentId
	s.currentId += 1
	s.WriteFrame(&f)
	return
}

func (s *TcpTx) SetTimeout(t time.Duration) {
	s.timeout = t
}

func (s *TcpTx) Close() {
	s.tcpServer.Close()
}
