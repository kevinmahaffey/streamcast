package streamcast

import (
	"bufio"
	"net"
)

type Client struct {
	incoming chan []byte
	outgoing chan []byte
	reader   *bufio.Reader
	writer   *bufio.Writer
	conn     net.Conn
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		incoming: make(chan []byte),
		outgoing: make(chan []byte),
		reader:   reader,
		writer:   writer,
	}

	client.Listen()
	return client
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
