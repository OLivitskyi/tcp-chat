package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"tcp-chat/utils"
	"time"
)

var clients = make(map[net.Conn]string)
var messageHistory []string
var mutex = sync.Mutex{}

const maxMessageHistory = 100 // Limit message history to prevent memory leak

func StartServer(port string, stopCh chan bool) {
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
	utils.LogMessage(fmt.Sprintf("Server is running on port %s...", port))

	go func() {
		<-stopCh
		fmt.Println("Shutting down server...")
		utils.LogMessage("Shutting down server...")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			fmt.Printf("Error accepting connection: %v\n", err)
			utils.LogMessage(fmt.Sprintf("Error accepting connection: %v", err))
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		if conn != nil {
			conn.Close()
		}
		mutex.Lock()
		name := clients[conn]
		delete(clients, conn)
		mutex.Unlock()
		broadcast(fmt.Sprintf("%s has left the chat.\n", name), nil)
	}()

	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Minute))

	mutex.Lock()
	if len(clients) >= 10 {
		mutex.Unlock()
		conn.Write([]byte("Server is full. Please try again later.\n"))
		return
	}
	clients[conn] = ""
	mutex.Unlock()

	sendWelcomeMessage(conn)

	name, _ := bufio.NewReader(conn).ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" || len(name) > 20 {
		conn.Write([]byte("Invalid name. Connection closed.\n"))
		return
	}

	mutex.Lock()
	clients[conn] = name
	mutex.Unlock()

	broadcast(fmt.Sprintf("%s has joined the chat!\n", name), conn)

	conn.Write([]byte("Loading chat history...\n"))
	sendHistory(conn)

	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") || strings.Contains(err.Error(), "connection reset by peer") {
				return
			}
			return
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		message = filterMessage(message)
		formattedMessage := fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, message)
		saveMessage(formattedMessage)
		broadcast(fmt.Sprintf("%s\n", formattedMessage), conn)
	}
}

func sendWelcomeMessage(conn net.Conn) {
	message := loadWelcomeMessage()

	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}

	conn.Write([]byte(message))
	utils.LogMessage("Sent welcome message to client")
}

func loadWelcomeMessage() string {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "welcome.txt")

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading welcome message: %v\n", err)
		utils.LogMessage(fmt.Sprintf("Error reading welcome message: %v", err))
		return "Welcome to TCP-Chat!"
	}

	return strings.TrimSpace(string(data))
}

func sendHistory(conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	log.Printf("Sending chat history to client: %v", conn.RemoteAddr())

	for _, msg := range messageHistory {
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Printf("Error sending history to client %v: %v", conn.RemoteAddr(), err)
			return
		}
	}
}

func broadcast(message string, sender net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	var disconnectedClients []net.Conn

	for client := range clients {
		if client == sender {
			continue
		}
		_, err := client.Write([]byte(message + "\n"))
		if err != nil {
			if strings.Contains(err.Error(), "broken pipe") {
				disconnectedClients = append(disconnectedClients, client)
				continue
			}
		}
	}

	for _, client := range disconnectedClients {
		client.Close()
		delete(clients, client)
	}
}

func saveMessage(message string) {
	mutex.Lock()
	defer mutex.Unlock()

	messageHistory = append(messageHistory, message)
	// Trim history if it exceeds the limit to prevent memory leak
	if len(messageHistory) > maxMessageHistory {
		messageHistory = messageHistory[len(messageHistory)-maxMessageHistory:]
	}
}

func filterMessage(message string) string {
	bannedWords := []string{"ban", "badword"}
	for _, word := range bannedWords {
		message = strings.ReplaceAll(message, word, "***")
	}
	return message
}
