package main

import (
	"GoTelnet/pkg/chatserver"
	"flag"
	"fmt"
)

func main() {

	// Обрабатываем флаги при запуске программы
	// go run server.go -host "localhost:12345"
	var host string

	flag.StringVar(&host, "host", "localhost:12345", "connect ip:port")
	flag.Parse()

	fmt.Println("flags: host->", host)

	chatserver.Start("tcp", host)
}
