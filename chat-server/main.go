package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"
)

type Message struct {
	from string
	text string
}

type Client struct {
	conn     net.Conn
	id       string
	outgoing chan Message // Client inbox. Server flushes out these messages to the client.
}

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

	messagesChan := make(chan Message)
	subscribeChan := make(chan Client)
	unsubscribeChan := make(chan Client)
	var clients = make(map[string]Client)

	go func() {
		for {
			select {
			case client := <-subscribeChan:
				{
					// Add it to the list to keep track of connected clients
					clients[client.id] = client

					// Create goroutine to send messages to other clients concurrently
					go func() {
						for msg := range client.outgoing {
							sendMessage(client.conn, msg)
						}
					}()
				}
			case client := <-unsubscribeChan:
				{
					close(client.outgoing)
					delete(clients, client.id)
				}
			case message := <-messagesChan:
				{
					// Pass the message from the global inbox "messages" channel to all other clients' outgoing channels
					for _, c := range clients {
						if message.from == c.id { // Don't add to client inbox if message comes from them
							continue
						}
						c.outgoing <- message
					}
				}
			default:
				{
					// Channel full - message dropped. Ensures server is not blocked because of slow client
				}
			}
		}
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		go handleConnection(conn, messagesChan, subscribeChan, unsubscribeChan)
	}

}

func handleConnection(conn net.Conn, messagesChan chan Message, subscribeChan chan Client,
	unsubscribeChan chan Client) {

	id := getId(conn)
	client := Client{conn: conn, id: id, outgoing: make(chan Message, 16)}

	defer func() {
		log.Printf("Closing connection with %s\n", conn.RemoteAddr().String())
		conn.Close()
		unsubscribeChan <- client
	}()

	subscribeChan <- client
	log.Printf("Client connection: %s\n", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		message := Message{from: id, text: scanner.Text()}
		messagesChan <- message
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error from %s: %v ", conn.RemoteAddr().String(), err)
		return
	}

}

func getId(conn net.Conn) string {
	id := strings.Split(conn.RemoteAddr().String(), ":")[1]
	return id
}

func sendMessage(conn net.Conn, message Message) {
	conn.Write([]byte(message.from + ": " + message.text + " at " + time.Now().String() + "\n"))
}
