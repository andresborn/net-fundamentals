package main

import (
	"log"
	"net"
	"strings"
)

// TCP Server listens and accepts incoming connections

// handleClientConnections
// goroutine per connection, loop to collect buffer message
// append message to channel

// broadcastMessages
// goroutine with loop over range of message channel and sends to all

var HOST = "127.0.0.1"
var PORT = "8080"

func main() {
	addr := net.JoinHostPort(HOST, PORT)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("Error starting server: ", err)
		return
	}

	defer listener.Close()

	log.Printf("Listening on %s", addr)

	messages := make(chan string)

	go func() {
		for message := range messages {
			log.Printf("New message: %s", message)
		}
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		go handleConnection(conn, messages)
	}

}

func handleConnection(conn net.Conn, messages chan string) {
	log.Printf("Client connection: %s\n", conn.RemoteAddr().String())

	defer func() {
		log.Printf("Closing connection with %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()

	buffer := make([]byte, 1024)
	var accumulated []byte

	for {
		n, err := conn.Read(buffer)

		if err != nil {
			log.Println("Error reading buffer: ", err)
			return
		}

		accumulated = append(accumulated, buffer[:n]...)

		if strings.Contains(string(accumulated), "\n") {
			message := string(accumulated)
			messages <- message
			accumulated = nil
		}
	}

}
