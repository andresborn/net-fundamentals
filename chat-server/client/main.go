package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

var GREETINGS = []string{
	"¡Hola, mundo!",      // Spanish
	"Bonjour le monde !", // French
	"你好，世界！",             // Mandarin Chinese
	"नमस्ते दुनिया!",     // Hindi
	"مرحباً بالعالم!",    // Arabic
	"Olá, mundo!",        // Portuguese
	"Привет, мир!",       // Russian
	"こんにちは、世界！",          // Japanese
	"Hallo, Welt!",       // German
	"Ciao, mondo!",       // Italian
}

func main() {
	var wg sync.WaitGroup

	for range 100 {
		time.Sleep(1 * time.Millisecond)
		wg.Go(func() {
			createClient()
		})
	}

	wg.Wait()
	fmt.Println("All clients disconnected.")
}

func createClient() {
	var (
		Host = "127.0.0.1"
		Port = "8080"
	)
	addr := net.JoinHostPort(Host, Port)
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Println("Error connecting: ", err)
		return
	}

	defer conn.Close()

	n := rand.Intn(len(GREETINGS))
	randomGreeting := GREETINGS[n]

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			seconds := time.Duration(rand.Intn(50)) + 10
			time.Sleep(seconds * time.Second)

			greet := randomGreeting + "\n"
			fmt.Println("Sending: " + greet)

			conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
			_, err := conn.Write([]byte(greet))
			if err != nil {
				log.Println("Error sending message: ", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			fmt.Println("Received: " + scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error from %s: %v ", conn.RemoteAddr().String(), err)
			return
		}
	}()

	wg.Wait()

}
