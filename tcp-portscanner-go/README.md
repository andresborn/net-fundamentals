# TCP Port Scanner

Objective: Understand practical applications of the networking concepts we've been studying so far.

This project has three steps:
1. Scan port 80 in `scanme.nmap.org` to understand Go's `net` package basics.
    - Establish connection, handle error, close connection if successful.
2. Scan ports 1 to 1024 in a blocking and sequential way.
    - For loop previous function. Use `net.DialTimeout` to avoid hanging indefinitely.
3. Scan ports 1 to 1024 in a non-blocking, concurrent way using goroutines.
    - Transform previous iteration to goroutines

Steps 1 and 2 are pretty straightforward. On the other hand, I've encountered multiple ways to 
approach step 3. I'm new to Go so concepts like goroutines and channels are fairly novel.

The `scanPort` function remained pretty much the same during all the experiments.

## Looping over results channel

This didn't work. Channels became deadlocked.
```go
	ch := make(chan int)

	for port := range portsToScan {
		go scanPort(port, host, ch)
	}

	var open []int
	for range ch {
		port := <-ch
		if port != 0 {
			open = append(open, port)
		}
	}
	close(ch) // Never reached
```
The channel is never closed (which would prevent the deadlock) because the function is never 
reached. The for loop keeps waiting for new values to come to `ch`.

## WaitGroup
Create `sync.WaitGroup` and add `wg.Add()` and `wg.Done()` inside the goroutines. We still need a 
channel to pass values between goroutines.

## Producer, Worker, Feeder pattern which using channels
I came to this solution reading [this blogpost](https://gomonk.substack.com/p/building-a-tiny-tcp-port-scanner) 
(code can be found [here](https://github.com/go-monk/tcp-scanner/blob/main/tcpscanner.go)).

Really, my code is just a dumbed down version of this.

In this solution the author creates scan goroutines (workers), then creates a goroutine with a loop 
that feeds a channel with the ports to be scanned and finally, loops over the ports range that were 
meant to be scanned.

> It's important to notice that the worker/scan function always adds a value to `ch` even if the 
port is not open. This keeps the channel in sync with the value extractor loop, else we would have 
less values in the `ch` than in the loop causing a deadlock.

Reading this program helped me understand how looping over channels works. The loop will wait 
indefinitely for potential incoming values.

### Current solution
My current solution is essentially the same except that we're using one goroutine per port. This is 
simpler but definitely not resource efficient.

```go
	const (
		portsToScan = 100
		host        = "scanme.nmap.org"
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
```

## Context Timeout
// TODO: Implement with Context Timeout

## Networking notes
I allowed several ports in the firewall of my home server but they appeared as closed in the 
scan tool. A process needs to be listening to the port in order to respond back.