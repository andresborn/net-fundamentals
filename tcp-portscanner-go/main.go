package main

import (
	"fmt"
	"net"
	"sort"
	"time"
)

func scanPort(port int, hostname string, ch chan int) {
	address := fmt.Sprintf("%v:%d", hostname, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)

	if err != nil {
		ch <- 0
	} else {
		fmt.Printf("Port %d in %s is open.\n", port, hostname)
		ch <- port
		conn.Close()
	}

}

func main() {

	const (
		portsToScan = 100
		host        = "" // scanme.nmap.org (don't hammer the server) or home server IP
	)

	start := time.Now()
	ch := make(chan int)

	for port := range portsToScan {
		go scanPort(port, host, ch)
	}

	var open []int
	for range portsToScan {
		port := <-ch
		if port != 0 {
			open = append(open, port)
		}
	}
	close(ch) // I don't think this is neccessary but it is safe.

	sort.Ints(open)
	fmt.Printf("Open channels: %v\n", open)

	elapsed := time.Since(start)
	fmt.Printf("Scan took %v to run.", elapsed)

}
