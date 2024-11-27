package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"tcp-chat/utils"
	"time"
)

var clients = make(map[net.Conn]string)
var messageHistory []string
var mutex = sync.Mutex{}

func StartServer(port string) {
	utils.InitLogger()
	defer utils.CloseLogger()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		utils.LogMessage(fmt.Sprintf("Error starting server: %v", err))
		return
	}
	defer listener.Close()

	fmt.Printf("Server is running on port %s...\n", port)
	utils.LogMessage(fmt.Sprintf("Server started on port %s", port))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			utils.LogMessage(fmt.Sprintf("Error accepting connection: %v", err))
			continue
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		mutex.Lock()
		name := clients[conn]
		delete(clients, conn)
		mutex.Unlock()

		broadcast(fmt.Sprintf("%s has left the chat.\n", name), nil)
		utils.LogMessage(fmt.Sprintf("Client disconnected: %s", name))
		conn.Close()
	}()

	mutex.Lock()
	if len(clients) >= 10 {
		mutex.Unlock()
		conn.Write([]byte("Server is full. Please try again later.\n"))
		utils.LogMessage("Connection rejected: chat is full.")
		conn.Close()
		return
	}
	mutex.Unlock()

	sendWelcomeMessage(conn)

	name, _ := bufio.NewReader(conn).ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" || len(name) > 20 || strings.Contains(name, " ") {
		conn.Write([]byte("Invalid name. Connection closed.\n"))
		conn.Close()
		return
	}

	mutex.Lock()
	clients[conn] = name
	mutex.Unlock()

	utils.LogMessage(fmt.Sprintf("Client connected: %s", name))

	conn.Write([]byte("Loading chat history...\n"))
	sendHistory(conn)

	broadcast(fmt.Sprintf("%s has joined the chat!\n", name), conn)

	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			break
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		message = filterMessage(message)

		saveMessage(fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, message))
		broadcast(fmt.Sprintf("[%s][%s]: %s\n", time.Now().Format("2006-01-02 15:04:05"), name, message), conn)
		utils.LogMessage(fmt.Sprintf("Message from %s: %s", name, message))
	}
}

func sendWelcomeMessage(conn net.Conn) {
	conn.Write([]byte(loadWelcomeMessage()))
}

func loadWelcomeMessage() string {
	data, err := os.ReadFile("welcome.txt")
	if err != nil {
		log.Printf("Error reading welcome message: %v\n", err)
		return "Welcome to TCP-Chat!"
	}
	return string(data)
}

func sendHistory(conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	for _, msg := range messageHistory {
		conn.Write([]byte(msg + "\n"))
	}
}

func broadcast(message string, sender net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range clients {
		if client != sender {
			_, err := client.Write([]byte(message))
			if err != nil {
				log.Printf("Error sending message to client %s: %v\n", clients[client], err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func saveMessage(message string) {
	mutex.Lock()
	defer mutex.Unlock()

	messageHistory = append(messageHistory, message)
}

func filterMessage(message string) string {
	bannedWords := []string{"ban", "badword"}
	for _, word := range bannedWords {
		message = strings.ReplaceAll(message, word, "***")
	}
	return message
}
