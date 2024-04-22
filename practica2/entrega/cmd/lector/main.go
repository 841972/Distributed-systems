package main

import (
	"codigo/com"
	"codigo/ra"
	"sync"
)

const (
	barrierEndpoint = "192.168.3.1:29100"
	endpoint        = ":29102" // endpoint for start.go to send work order
)

func main() {
	// Clear ports
	com.CleanPorts()

	// Get id
	id := com.GetArgs()

	// Initialize ra
	sharedRA := ra.New(id, "users.txt", "read")

	// Start listening for work order
	var wg sync.WaitGroup
	wg.Add(1)
	go com.WaitWorkOrder(&wg, endpoint)

	// Send ready message
	com.SendReady(id, barrierEndpoint)

	// Wait work order
	wg.Wait()

	go com.SetEnd()

	// Start reading
	i := 1
	for i <= 100 {
		sharedRA.PreProtocol()
		sharedRA.File.Read()
		sharedRA.PostProtocol()
		i++
	}

	// Infinite loop which will be break by the com.SetEnd() function
	for {
	}
}
