package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	var senderWG sync.WaitGroup
	var readConnWG sync.WaitGroup
	var eventLoopWG sync.WaitGroup

	defer stop()

	log.Printf("Listening on %s", addr)

	messagesChan := make(chan Message, 256)
	pool := Pool{clients: make(map[string]Client)}

	eventLoopWG.Go(func() {
		for {
			select {
			case <-ctx.Done():
				{

					log.Println("Shutting down server...")
					// 1. Stop accepting new connections
					listener.Close()
					// 2. Drain remaining messages and wait to finish
					pool.mu.Lock()
					for _, c := range pool.clients {
						close(c.outgoing)
					}
					pool.mu.Unlock()

					senderWG.Wait()

					// 3. Close connections
					pool.mu.Lock()
					for _, c := range pool.clients {
						c.conn.Close()
					}
					// 4. Delete clients and exit
					pool.clients = nil
					pool.mu.Unlock()
					return
				}

			case message := <-messagesChan:
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
		}
	})

	readConnWG.Go(func() {
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
			readConnWG.Go(func() { handleConnection(ctx, conn, messagesChan, &pool, &senderWG) })
		}
	})

	<-ctx.Done()

	senderWG.Wait()
	readConnWG.Wait()
	eventLoopWG.Wait()

}

func handleConnection(ctx context.Context, conn net.Conn, messagesChan chan Message, pool *Pool, senderWG *sync.WaitGroup) {

	id := getId(conn)
	client := Client{conn: conn, id: id, outgoing: make(chan Message, 16)}

	defer func() {
		log.Printf("Closing connection with %s\n", conn.RemoteAddr().String())
		conn.Close()
		pool.unsubscribe(id)
	}()

	go func() {
		<-ctx.Done()
		conn.SetReadDeadline(time.Now()) // Stop receiving new messages once server shutdown starts
	}()

	pool.subscribe(id, client)
	log.Printf("Client connection: %s\n", conn.RemoteAddr().String())

	senderWG.Go(func() {
		for msg := range client.outgoing {
			err := sendMessage(client.conn, msg)
			if err != nil {
				return
			}
		}
	})

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		message := Message{from: id, text: scanner.Text()}
		select {
		case messagesChan <- message:
			// Message sent to channel
		case <-ctx.Done():
			{
				log.Println("Stopping read because server is shutting down.")
				return
			}

		}
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
