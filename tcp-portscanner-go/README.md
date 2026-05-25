# TCP Port Scanner

Objective: Understand practical applications of the networking concepts we've been studying so far.

This project has three steps:
1. Scan port 80 in `scanme.nmap.org` to understand Go's `net` package basics.
  - Establish connection, handle error, close connection if successful.
2. Scan ports 1 to 1024 in a blocking and sequential way.
  - For loop previous function. Use `net.DialTimeout` to avoid hanging indefinitely.
3. Scan ports 1 to 1024 in a non-blocking, concurrent way using goroutines.
  - Transform previous iteration to goroutines