package server

import (
	"net"
	"testing"
	"time"
)

func TestServer_StartServer(t *testing.T) {
	go StartServer("9999") // Запускаємо сервер у фоновому режимі
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Тестуємо відправку і отримання повідомлення
	_, err = conn.Write([]byte("Hello\n"))
	if err != nil {
		t.Fatalf("Failed to send message to server: %v", err)
	}

	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read message from server: %v", err)
	}
}

func TestServer_ConnectionLimit(t *testing.T) {
	go StartServer("9998")
	time.Sleep(1 * time.Second)

	clients := make([]net.Conn, 0)
	for i := 0; i < 12; i++ {
		conn, err := net.Dial("tcp", "localhost:9998")
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		clients = append(clients, conn)
	}

	if len(clients) > 10 {
		t.Fatalf("Server allowed more than 10 connections")
	}
}
