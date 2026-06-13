package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"
)

// https://en.wikipedia.org/wiki/Domain_Name_System#DNS_message_format
type DNSHeader struct {
	ID    uint16
	Flags uint16
	NQ    uint16 // Number of questions
	NA    uint16 // Number of answers
	NARR  uint16 // Number of Authority RRs
	NADRR uint16 // Number of Additional RRs
}

func buildHeader(id uint16) []byte {
	header := make([]byte, 12)

	binary.BigEndian.PutUint16(header[0:2], id)

	// Quick notes: << operator shifts bits left and |= changes bits without affecting other bits
	var flags uint16 // init 00000000 00000000
	flags |= 0 << 15 // QR 0: Message is query. Already set but added for learning purposes
	flags |= 1 << 8  // RD 1: Recursive query

	binary.BigEndian.PutUint16(header[2:4], flags)

	numberOfQuestions := uint16(1)
	binary.BigEndian.PutUint16(header[4:6], numberOfQuestions)

	numberOfAnswers := uint16(0)
	binary.BigEndian.PutUint16(header[6:8], numberOfAnswers)

	// Leave 0 for Number of Authority RRs and Number of Additional RRs

	return header
}

func buildBody(domain string) []byte {
	split := strings.Split(domain, ".")

	var question []byte
	for n := range split {
		question = append(question, uint8(len(split[n])))
		question = append(question, split[n]...)
	}
	question = append(question, 0x00) // Marks end of question

	// QTYPE
	question = binary.BigEndian.AppendUint16(question, 1) // A record

	// QCLASS
	question = binary.BigEndian.AppendUint16(question, 1) // Internet (IN)

	return question
}

func parseHeader(header []byte) (uint16, uint16, uint16) {
	id := binary.BigEndian.Uint16(header[0:2])
	flags := binary.BigEndian.Uint16(header[2:4])

	qr := flags >> 15 // QR: query/response
	rCode := flags & 0x000F

	return id, qr, rCode
}

func main() {

	conn, err := net.Dial("udp", "8.8.8.8:53")

	if err != nil {
		log.Fatal("Error: ", err)
		return
	}

	defer conn.Close()

	// Send message
	id := uint16(rand.Intn(math.MaxUint16 + 1))
	header := buildHeader(id)
	question := buildBody("google.com")

	sndBuffer := append(header, question...)

	fmt.Printf("Sending %d bytes: %x\n", len(sndBuffer), sndBuffer)
	conn.Write(sndBuffer)

	// Receive message
	seconds := time.Second * 3
	now := time.Now()

	conn.SetReadDeadline(now.Add(seconds))

	rcvBuffer := make([]byte, 512)
	n, err := conn.Read(rcvBuffer)

	if err != nil {
		log.Println("Error reading buffer: ", err)
		return
	}

	fmt.Printf("Received %d bytes: %x\n", n, rcvBuffer[:n])

	resId, qr, rCode := parseHeader(rcvBuffer)

	if id == resId {
		fmt.Println("We got a response!")
		fmt.Println("ID: ", resId)
		fmt.Println("Is response: ", qr)
		fmt.Println("DNS error code: ", rCode)
	}

}
