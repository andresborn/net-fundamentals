package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
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
	// Loop and create clients
	for range 10 {
		go createClient()
	}

	select {}

	// Randomly kill client connections

}

func createClient() {
	var (
		Host = "127.0.0.1"
		Port = "8080"
	)
	addr := net.JoinHostPort(Host, Port)
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatal("Error connecting: ", err)
		return
	}

	defer conn.Close()

	n := rand.Intn(len(GREETINGS))
	randomGreeting := GREETINGS[n]

	go func() {
		for {
			conn.Write([]byte(randomGreeting + "\n"))
			seconds := time.Duration(rand.Intn(30))
			time.Sleep(seconds * time.Second)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error from %s: %v ", conn.RemoteAddr().String(), err)
			return
		}
	}()

	select {}

}
