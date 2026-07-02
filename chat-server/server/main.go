package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
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

type Chatroom struct {
	// Channels
	subscribe   chan *Client
	unsubscribe chan *Client
	broadcast   chan Message

	// State
	clients map[string]*Client
	mu      sync.Mutex
}

func (cr *Chatroom) handleRead(client *Client) {
	scanner := bufio.NewScanner(client.conn)

	for scanner.Scan() {
		message := Message{from: client.id, text: scanner.Text()}
		cr.broadcast <- message
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error from %s: %v ", client.conn.RemoteAddr().String(), err)
		return
	}
}

func (cr *Chatroom) handleWrite(client *Client) {

	for message := range client.outgoing {
		err := sendMessage(client.conn, message)
		if err != nil {
			log.Println("Error sending message: ", err)
			return
		}
	}
}

func sendMessage(conn net.Conn, message Message) error {
	err := conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = conn.Write([]byte(message.from + ": " + message.text + "\n"))
	if err != nil {
		log.Println("Error sending message: ", err)
		return err
	}
	return nil
}

func (cr *Chatroom) handleBroadcast(message Message) {
	clients := make([]*Client, 0) // Local copy
	cr.mu.Lock()
	for _, client := range cr.clients {
		if client.id == message.from {
			continue
		}
		clients = append(clients, client)

	}
	cr.mu.Unlock()

	log.Printf("Broadcasting %s from %s\n", message.text, message.from)

	for _, client := range clients {
		select {
		case client.outgoing <- message:
		default:
			log.Println("Message dropped for slow client: ", client.id)
		}
	}
}

func (cr *Chatroom) handleSub(client *Client) {
	cr.mu.Lock()
	cr.clients[client.id] = client
	cr.mu.Unlock()
}

func (cr *Chatroom) handleUnsub(client *Client) {
	cr.mu.Lock()
	delete(cr.clients, client.id)
	cr.mu.Unlock()

	// Close channel safely
	select {
	case <-client.outgoing:
		// Already closed
	default:
		close(client.outgoing)
	}
}

var (
	Host = "127.0.0.1"
	Port = "8080"
)

func main() {
	addr := net.JoinHostPort(Host, Port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("Error starting server: ", err)
		return
	}

	defer listener.Close()

	log.Printf("Listening on %s", addr)

	cr := Chatroom{
		subscribe:   make(chan *Client),
		unsubscribe: make(chan *Client),
		broadcast:   make(chan Message),
		clients:     map[string]*Client{},
	}

	// Event broker
	go func() {
		for {
			select {
			case client := <-cr.subscribe:
				cr.handleSub(client)
			case client := <-cr.unsubscribe:
				cr.handleUnsub(client)
			case message := <-cr.broadcast:
				cr.handleBroadcast(message)
			}
		}
	}()

	for {
		conn, err := listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			log.Println("Listener connection closed: ", err)
			return
		}

		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}
		go handleConnection(conn, &cr)
	}

}

func handleConnection(conn net.Conn, cr *Chatroom) {

	id := getId(conn)
	client := &Client{conn: conn, id: id, outgoing: make(chan Message, 16)}

	defer func() {
		log.Printf("Closing connection with %s\n", conn.RemoteAddr().String())
		client.conn.Close()
	}()

	log.Printf("Client connection: %s\n", conn.RemoteAddr().String())

	cr.subscribe <- client

	go cr.handleWrite(client)

	cr.handleRead(client) // Blocks until client disconnect returns scanner error

	// handleUnsub closes channel. cr.handleWrite finishes sending queued messages (which will most
	// likely return an error), exits, unsubscribes, and finally connection closes on defer.
	cr.unsubscribe <- client

}

func getId(conn net.Conn) string {
	id := strings.Split(conn.RemoteAddr().String(), ":")[1]
	return id
}

// Graceful shutdown, not implemented
func shutdown(listener net.Listener, cr *Chatroom, writeWG *sync.WaitGroup) {

	log.Println("Shutting down gracefully...")

	listener.Close()

	// Close outgoing channels
	cr.mu.Lock()
	for _, client := range cr.clients {
		// Safe close
		select {
		case <-client.outgoing:
		default:
			close(client.outgoing)
		}
	}
	cr.mu.Unlock()

	// Wait for messages to be sent/drained
	writeWG.Wait()

	// close connections
	cr.mu.Lock()
	for _, client := range cr.clients {
		client.conn.Close()
	}
	cr.mu.Unlock()

	// Delete clients
	cr.mu.Lock()
	cr.clients = nil
	cr.mu.Unlock()

	log.Println("Shutdown complete.")
}
