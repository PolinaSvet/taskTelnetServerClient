package main

import (
	"GoTelnet/pkg/api"
	"GoTelnet/pkg/chatclient"
	"log"

	"flag"
	"fmt"
	"net/http"
)

type client struct {
	api *api.API
}

func main() {

	// Обрабатываем флаги при запуске программы
	// go run client.go -host "localhost:12345"
	var host string

	flag.StringVar(&host, "host", "localhost:12345", "connect ip:port")
	flag.Parse()

	fmt.Println("flags: host->", host)

	//=========================================================================
	// Канал для отправки сообщений
	sendChannel := make(chan string)
	dataChannel := make(chan string)

	// Создаем клиента telnet
	c, err := chatclient.New(host)
	if err != nil {
		log.Fatalf("Ошибка при создании клиента: %v", err)
	}

	// Создаем горутину для чтения сообщений от сервера
	go func() {
		for {
			msg := c.ReceiveMessage()
			dataChannel <- msg
		}
	}()

	// Горутина для отправки сообщений
	go func() {
		for {
			msg := <-sendChannel
			c.SendMessage(msg)
		}
	}()

	//web интерфейс для клиента telnet
	var cln client
	cln.api = api.New(sendChannel, dataChannel)
	fmt.Println("Запуск веб-сервера на http://127.0.0.1:8080 ...")
	http.ListenAndServe(":8080", cln.api.Router())
}
