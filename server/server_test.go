package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func startTestServer(t *testing.T, port string, stopCh chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	StartServer(port, stopCh)
}

func connectToServer(t *testing.T, address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	return conn
}

func TestClientConnection(t *testing.T) {
	const port = "8990"
	serverAddr := ":" + port
	stopCh := make(chan bool)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go startTestServer(t, port, stopCh, wg)
	time.Sleep(500 * time.Millisecond)

	client := connectToServer(t, serverAddr)
	defer client.Close()

	var welcomeMessage string
	reader := bufio.NewReader(client)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read welcome message: %v", err)
		}
		welcomeMessage += line
		if strings.Contains(line, "[ENTER YOUR NAME]:") {
			break
		}
	}

	if !strings.Contains(welcomeMessage, "Welcome to TCP-Chat!") {
		t.Errorf("Unexpected welcome message: %s", welcomeMessage)
	}

	stopCh <- true
	wg.Wait()
}

func TestClientMessageBroadcast(t *testing.T) {
	const port = "8991"
	serverAddr := ":" + port
	stopCh := make(chan bool)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go startTestServer(t, port, stopCh, wg)
	time.Sleep(500 * time.Millisecond)

	client1 := connectToServer(t, serverAddr)
	defer client1.Close()
	client2 := connectToServer(t, serverAddr)
	defer client2.Close()

	client1.Write([]byte("Client1\n"))
	client2.Write([]byte("Client2\n"))

	time.Sleep(1 * time.Second) // Додаткова пауза для синхронізації

	testMessage := "Hello, this is Client1!"
	client1.Write([]byte(testMessage + "\n"))

	done := make(chan bool, 1)
	go func() {
		reader := bufio.NewReader(client2)
		for {
			response, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Error reading from client2: %v", err)
				done <- false
				return
			}
			log.Printf("Message received by client2: %s", strings.TrimSpace(response))
			if strings.Contains(response, testMessage) {
				done <- true
				return
			}
		}
	}()

	select {
	case success := <-done:
		if !success {
			t.Errorf("Broadcast failed. Expected message: %s", testMessage)
		}
	case <-time.After(5 * time.Second): // Зменшений таймаут для ефективності
		t.Errorf("Broadcast timed out. Expected message: %s", testMessage)
	}

	stopCh <- true
	wg.Wait()
}

func TestMaxConnections(t *testing.T) {
	const port = "8992"
	serverAddr := ":" + port
	stopCh := make(chan bool)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go startTestServer(t, port, stopCh, wg)
	time.Sleep(500 * time.Millisecond)

	var clients []net.Conn
	for i := 0; i < 10; i++ {
		client := connectToServer(t, serverAddr)
		clients = append(clients, client)
		client.Write([]byte(fmt.Sprintf("Client%d\n", i)))
	}

	lastClient, err := net.Dial("tcp", serverAddr)
	if err == nil {
		defer lastClient.Close()
		response, _ := bufio.NewReader(lastClient).ReadString('\n')
		if !strings.Contains(response, "Server is full") {
			t.Errorf("Expected server to reject connection, but it was accepted: %s", response)
		}
	} else {
		log.Printf("Server correctly rejected the connection: %v", err)
	}

	for _, client := range clients {
		client.Close()
	}

	stopCh <- true
	wg.Wait()
}
