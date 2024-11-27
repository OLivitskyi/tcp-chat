package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("[USAGE]: ./TCPClient $IP $port")
		return
	}

	serverIP := os.Args[1]
	port := os.Args[2]

	conn, err := net.Dial("tcp", serverIP+":"+port)
	if err != nil {
		log.Fatalf("Error connecting to server: %v\n", err)
	}
	defer conn.Close()

	fmt.Println("Connected to the server.")

	go readMessages(conn)
	writeMessages(conn)
}

func readMessages(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Disconnected from server.")
			return
		}
		fmt.Print(message)
	}
}

func writeMessages(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "" {
			continue
		}

		_, err := conn.Write([]byte(message + "\n"))
		if err != nil {
			log.Println("Error writing to server:", err)
			return
		}
	}
}
