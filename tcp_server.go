package streamcast

import (
  "bufio"
  "net"
)

type Client struct {
  incoming  chan []byte
  outgoing  chan []byte
  reader   *bufio.Reader
  writer   *bufio.Writer
  conn      net.Conn
}

func (client *Client) Read() {
  for {
    // todo we could use this to have any client broadcast to all its peers.
    line, _ := client.reader.ReadBytes(100)
    client.incoming <- line
  }
}

func (client *Client) Write() {
  for data := range client.outgoing {
    client.writer.Write(data)
    client.writer.Flush()
  }
}

func (client *Client) Listen() {
  go client.Read()
  go client.Write()
}

func (client *Client) Close() {
  client.conn.Close()
}

func NewClient(connection net.Conn) *Client {
  writer := bufio.NewWriter(connection)
  reader := bufio.NewReader(connection)

  client := &Client{
    incoming: make(chan []byte),
    outgoing: make(chan []byte),
    reader: reader,
    writer: writer,
  }

  client.Listen()
  return client
}

type TcpServer struct {
  clients   []*Client
  joins     chan net.Conn
  incoming  chan []byte
  outgoing  chan []byte
}

func (tcpServer *TcpServer) Broadcast(data []byte) {
  for _, client := range tcpServer.clients {
    client.outgoing <- data
  }
}

func (tcpServer *TcpServer) Join(connection net.Conn) {
  client := NewClient(connection)
  tcpServer.clients = append(tcpServer.clients, client)
  go func() { for { tcpServer.incoming <- <-client.incoming } }()
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

func  (tcpServer *TcpServer) Close() {
  for _, client := range tcpServer.clients {
    client.Close()
  }}

func NewTcpServer() *TcpServer {
  tcpServer := &TcpServer{
    clients: make([]*Client, 0),
    joins: make(chan net.Conn),
    incoming: make(chan []byte),
    outgoing: make(chan []byte),
  }

  tcpServer.Listen()
  return tcpServer
}

func StartTcpServer(addr string) (error, *TcpServer) {
  tcpServer := NewTcpServer()
  listener, err := net.Listen("tcp", addr)
  if(err!=nil) {
    return err, nil
  }
  go tcpServer.Accept(listener)
  return err, tcpServer
}