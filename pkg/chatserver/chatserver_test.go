package chatserver

import (
	"bufio"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_SendMessage_BroadcastNewsOn(t *testing.T) {
	srv, cl := net.Pipe()
	defer srv.Close()
	defer cl.Close()

	serverConn := bufio.NewReader(srv)

	client := &Client{
		Conn:            cl,
		Name:            "TestUser",
		BroadcastNewsOn: true,
	}

	testMessage := Mess{
		Name:    "broadcastNewsSend",
		Content: "Hello, World!",
		From:    "TestUser",
		To:      "",
	}

	go func() {
		client.SendMessage(testMessage)
	}()

	message, err := serverConn.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read message from server: %v", err)
	}

	message = strings.TrimSuffix(message, "\r\n")
	t.Log(message)

	assert.Equal(t, "Hello, World!", message, "The message content does not match the expected value")
}
