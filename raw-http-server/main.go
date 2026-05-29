package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		log.Fatal("Error listening: ", err)
	}

	defer listener.Close()

	log.Println("Listening...")

	for {

		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error accepting connection: ", err)
		}

		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)

	_, err := conn.Read(buffer)

	if err != nil {
		log.Println("Error reading connection: ", err)
		return
	}

	message := string(buffer[:])
	// Split by new lines
	splitLines := strings.Split(message, "\r\n")

	// Split start line by empty spaces
	startLine := strings.Split(splitLines[0], " ")

	method := startLine[0]
	uri := startLine[1]
	protocolVersion := startLine[2]

	log.Printf("Method %s, URI %s, Protocol Version %s", method, uri, protocolVersion)

	headers := splitLines[1 : len(splitLines)-2] // Remove body (last) and \r\n\r\n (second to last)
	body := splitLines[len(splitLines)-1]        // Select last from slice

	// Todo: Format headers into JSON and return everything to client?

	fmt.Println(headers)
	fmt.Println(body)

}

func readMsg(conn net.Conn) {
	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')

	if err != nil {
		log.Println("Error reading message: ", err)
		return
	}

	log.Println("Message: ", message)
}
