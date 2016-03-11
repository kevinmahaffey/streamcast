package streamcast

import (
	"net"
	"time"
)

/* Generic Receiver Connection interface */
type RxConn interface {
	Close()
	Reset() (err error)
	Read(b []byte) (int, error)
	SetDeadline(t time.Time) error
}

/* UDP Receiver Connection: mapping the generic methods above to UDP specific methods. */
type UdpRxConn struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func (udpRxConn *UdpRxConn) Reset() (err error) {
	udpRxConn.Close()
	udpRxConn.conn, err = net.ListenUDP("udp4", udpRxConn.addr)
	if err != nil {
		return err
	}
	return
}

func (udpRxConn *UdpRxConn) SetDeadline(t time.Time) error {
	return udpRxConn.conn.SetDeadline(t)
}

func (udpRxConn *UdpRxConn) Close() {
	if udpRxConn.conn != nil {
		udpRxConn.conn.Close()
	}
}

func (udpRxConn *UdpRxConn) Read(b []byte) (int, error) {
	n, _, err := udpRxConn.conn.ReadFromUDP(b)
	return n, err
}

/* TCP Receiver Connection: mapping the generic methods above to TCP specific methods. */
type TcpRxConn struct {
	conn *net.TCPConn
	addr *net.TCPAddr
}

func (tcpRxConn *TcpRxConn) Reset() (err error) {
	tcpRxConn.Close()
	tcpRxConn.conn, err = net.DialTCP("tcp4", nil, tcpRxConn.addr)
	if err != nil {
		return err
	}
	return
}

func (tcpRxConn *TcpRxConn) SetDeadline(t time.Time) error {
	return tcpRxConn.conn.SetDeadline(t)
}

func (tcpRxConn *TcpRxConn) Close() {
	if tcpRxConn.conn != nil {
		tcpRxConn.conn.Close()
	}
}

func (tcpRxConn *TcpRxConn) Read(b []byte) (int, error) {
	n, err := tcpRxConn.conn.Read(b)
	return n, err
}
