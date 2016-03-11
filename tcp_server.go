package streamcast

import (
	"net"
)

type TcpServer struct {
	clients  []*Client
	joins    chan net.Conn
	incoming chan []byte
	outgoing chan []byte
}

func (tcpServer *TcpServer) Broadcast(data []byte) {
	for _, client := range tcpServer.clients {
		client.outgoing <- data
	}
}

func (tcpServer *TcpServer) Join(connection net.Conn) {
	client := NewClient(connection)
	tcpServer.clients = append(tcpServer.clients, client)
	go func() {
		for {
			tcpServer.incoming <- <-client.incoming
		}
	}()
}

func (tcpServer *TcpServer) Listen() {
	go func() {
		for {
			select {
			case data := <-tcpServer.incoming:
				tcpServer.Broadcast(data)
			case conn := <-tcpServer.joins:
				tcpServer.Join(conn)
			}
		}
	}()
}

func (tcpServer *TcpServer) Accept(listener net.Listener) {
	for {
		conn, _ := listener.Accept()
		tcpServer.joins <- conn
	}
}

func (tcpServer *TcpServer) Close() {
	for _, client := range tcpServer.clients {
		client.Close()
	}
}

func NewTcpServer() *TcpServer {
	tcpServer := &TcpServer{
		clients:  make([]*Client, 0),
		joins:    make(chan net.Conn),
		incoming: make(chan []byte),
		outgoing: make(chan []byte),
	}

	tcpServer.Listen()
	return tcpServer
}

func StartTcpServer(addr string) (error, *TcpServer) {
	tcpServer := NewTcpServer()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err, nil
	}
	go tcpServer.Accept(listener)
	return err, tcpServer
}
