package chatserver

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"GoTelnet/pkg/proverbs"
)

const (
	CmdList              = "list"
	CmdHelp              = "help"
	CmdClose             = "close"
	CmdSendTo            = "<-"
	CmdSend              = "send"
	CmdBroadcastNewsSend = "broadcastNewsSend"
	CmdBroadcastNewsOn   = "newsOn"
	CmdBroadcastNewsOff  = "newsOff"
)

var CmdHelpMenu = []string{
	"List of commands:",
	"help:          display help;",
	"close:         leave chat;",
	"list:          list of connected participants;",
	"name<-content: send a message to a specific member;",
	"newsOn:        turns on broadcast news;",
	"newsOff:       turns off broadcast news;",
}

type Client struct {
	Conn            net.Conn
	Name            string
	BroadcastNewsOn bool
}

type Mess struct {
	Name    string
	Content string
	From    string
	To      string
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		Conn:            conn,
		BroadcastNewsOn: false,
	}
}

type ChatServer struct {
	clients map[string]*Client
}

func getHelpMenu() string {
	return strings.Join(CmdHelpMenu, "\r\n")
}

func (c *Client) SendMessage(message Mess) {
	if !c.BroadcastNewsOn && message.Name == CmdBroadcastNewsSend {
		return
	}

	writer := bufio.NewWriter(c.Conn)
	writer.WriteString(message.Content + "\r\n")
	writer.Flush()
}

func (c *Client) ReadName() error {
	reader := bufio.NewReader(c.Conn)

	writer := bufio.NewWriter(c.Conn)
	writer.WriteString("Enter your name: ")
	writer.Flush()

	name, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	c.Name = strings.TrimSpace(name)
	c.SendMessage(Mess{From: c.Conn.RemoteAddr().String(), To: "", Name: "", Content: fmt.Sprintf("%v Welcome to the chat!", c.Name)})

	return nil
}

func (c *Client) HandleClient(removeCh chan<- net.Conn, messageCh chan<- Mess) {
	defer c.Conn.Close()

	if err := c.ReadName(); err != nil {
		messageCh <- Mess{From: c.Conn.RemoteAddr().String(), To: "", Name: "Error", Content: fmt.Sprintf("Error reading client name: %v", err)}
		return
	}

	messageCh <- Mess{From: c.Conn.RemoteAddr().String(), To: "", Name: "", Content: fmt.Sprintf("[%s] Joined the chat", c.Name)}
	reader := bufio.NewReader(c.Conn)
	for {
		message, err := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if err != nil {
			break
		}

		messageSlice := strings.Split(message, CmdSendTo)
		messageFrom := c.Conn.RemoteAddr().String()
		messageTo := ""
		messageCmd := message

		if len(messageSlice) == 2 {
			messageTo = fmt.Sprintf("%v", messageSlice[0])
			message = fmt.Sprintf("%v", messageSlice[1])
			messageCmd = CmdSendTo
		}

		//fmt.Printf("[%v]:[%v]:[%v]:[%v]\n", len(messageSlice), messageSlice, messageTo, message)
		switch messageCmd {
		case CmdHelp:
			messageCh <- Mess{From: messageFrom, To: messageTo, Name: CmdHelp, Content: fmt.Sprintf("[%s]: %s", c.Name, CmdHelp)}
		case CmdList:
			messageCh <- Mess{From: messageFrom, To: messageTo, Name: CmdList, Content: fmt.Sprintf("[%s]: %s", c.Name, CmdList)}
		case CmdBroadcastNewsOn:
			c.BroadcastNewsOn = true
			messageCh <- Mess{From: messageFrom, To: c.Name, Name: CmdSendTo, Content: fmt.Sprintf("[%s] Broadcast news on", c.Name)}
		case CmdBroadcastNewsOff:
			c.BroadcastNewsOn = false
			messageCh <- Mess{From: messageFrom, To: c.Name, Name: CmdSendTo, Content: fmt.Sprintf("[%s] Broadcast news off", c.Name)}
		case CmdSendTo:
			messageCh <- Mess{From: messageFrom, To: messageTo, Name: CmdSendTo, Content: fmt.Sprintf("[%s]: %s", c.Name, message)}
		case CmdClose:
			messageCh <- Mess{From: messageFrom, To: messageTo, Name: CmdSend, Content: fmt.Sprintf("[%s] Left the chat", c.Name)}
			removeCh <- c.Conn
		default:
			messageCh <- Mess{From: messageFrom, To: messageTo, Name: CmdSend, Content: fmt.Sprintf("[%s]: %s", c.Name, message)}
		}

	}
	removeCh <- c.Conn
}

// Запуск сетевой службы и обработка подключений
func Start(proto, addr string) {
	listener, err := net.Listen(proto, addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	// Каналы для управления клиентами
	addCh := make(chan *Client)
	removeCh := make(chan net.Conn)
	messageCh := make(chan Mess)

	server := &ChatServer{
		clients: make(map[string]*Client, 0),
	}

	go func() {

		for {
			select {
			case clientCurr := <-addCh:
				server.clients[clientCurr.Conn.RemoteAddr().String()] = clientCurr
			case conn := <-removeCh:
				for _, clientCurr := range server.clients {
					if clientCurr.Conn == conn {
						conn.Close()
						delete(server.clients, clientCurr.Conn.RemoteAddr().String())
						break
					}
				}
			case message := <-messageCh:
				//fmt.Printf("[%#v]:[%#v]\n", message, server.clients)
				if clientFrom, ok := server.clients[message.From]; ok {
					switch message.Name {
					case CmdHelp:
						clientFrom.SendMessage(Mess{From: message.From, To: "", Name: "", Content: getHelpMenu()})
					case CmdList:
						strMess := "List of participants:\r\n"
						for key, clientCurr := range server.clients {
							strMess += fmt.Sprintf("id>%v name>%v\r\n", key, clientCurr.Name)
						}
						clientFrom.SendMessage(Mess{From: message.From, To: "", Name: "", Content: strMess})
					case CmdSendTo:
						findClient := false
						for _, clientCurr := range server.clients {
							if clientCurr.Name == message.To {
								clientCurr.SendMessage(message)
								findClient = true
								break
							}
						}
						if !findClient {
							clientFrom.SendMessage(Mess{From: message.From, To: message.From, Name: CmdSendTo, Content: fmt.Sprintf("Members [%v] not founded.", message.To)})
						}

					default:
						for _, clientCurr := range server.clients {
							clientCurr.SendMessage(message)
						}
					}
				}
			case <-time.After(10 * time.Second):
				proverb := proverbs.GetRandomProverb()
				fmt.Println(proverb)
				for _, clientCurr := range server.clients {
					clientCurr.SendMessage(Mess{From: "", To: "", Name: CmdBroadcastNewsSend, Content: fmt.Sprintf("Broadcast News: %v", proverb)})
				}

			}
		}
	}()

	fmt.Println("Server started. Waiting for connections...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error connecting client:", err)
			continue
		}

		clientCurr := NewClient(conn)
		addCh <- clientCurr

		go clientCurr.HandleClient(removeCh, messageCh)
	}
}
