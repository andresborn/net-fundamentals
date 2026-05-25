package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	start := time.Now()
	for port := range 81 {
		address := fmt.Sprintf("scanme.nmap.org:%d", port)
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Port %d in scanme.nmap.org is open.\n", port)
			conn.Close()
		}

	}
	elapsed := time.Since(start)
	fmt.Printf("Scan took %v to run.", elapsed)

}
