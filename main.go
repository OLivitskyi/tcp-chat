package main

import (
	"log"
	"os"
	"tcp-chat/server"
)

func main() {
	port := "8989"

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	if len(os.Args) > 2 {
		log.Fatalf("[USAGE]: ./TCPChat $port")
	}
	stopCh := make(chan bool)
	server.StartServer(port, stopCh)
}
