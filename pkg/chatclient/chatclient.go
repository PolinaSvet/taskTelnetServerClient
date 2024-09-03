package chatclient

import (
	"bufio"
	"log"
	"net"
	"os"
)

// Client представляет собой клиента telnet.
type Client struct {
	conn          net.Conn
	MsgToServer   chan string
	MsgFromServer chan string
}

// New создаёт новый экземпляр клиента.
func New(address string) (*Client, error) {
	conn, err := net.Dial("tcp4", address)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:          conn,
		MsgToServer:   make(chan string),
		MsgFromServer: make(chan string),
	}

	go client.readFromServer()
	go client.writeToServer()

	return client, nil
}

// readFromServer читает данные от сервера и отправляет их в канал.
func (c *Client) readFromServer() {
	reader := bufio.NewReader(c.conn)
	for {
		b, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println("Ошибка при чтении от сервера:", err)
			os.Exit(1)
		}
		c.MsgFromServer <- string(b)
	}
}

// writeToServer записывает данные в сервер из канала.
func (c *Client) writeToServer() {
	for msg := range c.MsgToServer {
		_, err := c.conn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Println("Ошибка при отправке на сервер:", err)
			os.Exit(1)
		}
	}
}

// SendMessage отправляет сообщение на сервер.
func (c *Client) SendMessage(msg string) {
	c.MsgToServer <- msg
}

// ReceiveMessage получает сообщение от сервера.
func (c *Client) ReceiveMessage() string {
	return <-c.MsgFromServer
}
