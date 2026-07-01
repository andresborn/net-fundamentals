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

type Pool struct {
	mu      sync.RWMutex
	clients map[string]Client
}

func (p *Pool) subscribe(id string, client Client) {
	defer p.mu.Unlock()
	p.mu.Lock()
	p.clients[id] = client
}

func (p *Pool) unsubscribe(id string) {
	defer p.mu.Unlock()
	p.mu.Lock()
	delete(p.clients, id)
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

	messagesChan := make(chan Message, 256)
	pool := Pool{clients: make(map[string]Client)}

	go func() {
		for message := range messagesChan {
			{
				log.Printf("New message: '%s' from %s", message.text, message.from)
				// Pass the message from the global inbox "messages" channel to all other clients' outgoing channels
				pool.mu.RLock()
				for _, c := range pool.clients {
					if message.from == c.id { // Don't add to client inbox if message comes from them
						continue
					}
					select {

					case c.outgoing <- message:
					default:
						// Slow client. Dropped. Ensures server is not blocked because of slow client
					}
				}
				pool.mu.RUnlock()
			}
		}
	}()

	func() {
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
			go func() { handleConnection(conn, messagesChan, &pool) }()
		}
	}()

}

func handleConnection(conn net.Conn, messagesChan chan Message, pool *Pool) {

	id := getId(conn)
	client := Client{conn: conn, id: id, outgoing: make(chan Message, 16)}

	defer func() {
		log.Printf("Closing connection with %s\n", conn.RemoteAddr().String())
		conn.Close()
		pool.unsubscribe(id)
	}()

	pool.subscribe(id, client)
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

func sendMessage(conn net.Conn, message Message) error {
	err := conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
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
