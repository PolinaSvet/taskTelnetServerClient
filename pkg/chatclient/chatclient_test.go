package chatclient

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_writeToServer(t *testing.T) {
	srv, cl := net.Pipe()

	client := &Client{
		conn:          cl,
		MsgToServer:   make(chan string),
		MsgFromServer: make(chan string),
	}

	go client.writeToServer()

	go func() {
		reader := bufio.NewReader(srv)
		msg, err := reader.ReadString('\n')
		assert.NoError(t, err)
		assert.Equal(t, msg, "Test Message\n", "Error message")
	}()

	client.SendMessage("Test Message")
	time.Sleep(100 * time.Millisecond)
}

func TestClientReceiveMessage(t *testing.T) {
	srv, cl := net.Pipe()

	client := &Client{
		conn:          cl,
		MsgToServer:   make(chan string),
		MsgFromServer: make(chan string, 1),
	}

	go client.readFromServer()

	go func() {
		_, err := srv.Write([]byte("Test Response\n"))
		assert.NoError(t, err)
	}()

	// Ожидаем получения сообщения
	select {
	case msg := <-client.MsgFromServer:
		assert.Equal(t, msg, "Test Response\n", "Error message")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout while waiting for message from client")
	}
}
