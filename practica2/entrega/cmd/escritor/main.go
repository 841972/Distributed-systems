package main

import (
	"codigo/com"
	"codigo/ra"
	"strconv"
	"sync"
)

const (
	barrierEndpoint = "192.168.3.1:29100"
	endpoint        = ":29103" // endpoint for start.go to send work order
)

func main() {
	// Clear ports
	com.CleanPorts()

	// Get id
	id := com.GetArgs()

	// Initialize ra
	sharedRA := ra.New(id, "users.txt", "write")

	// Start listening for work order
	var wg sync.WaitGroup
	wg.Add(1)
	go com.WaitWorkOrder(&wg, endpoint)

	// Send ready message
	com.SendReady(id, barrierEndpoint)

	// Wait work order
	wg.Wait()

	go com.SetEnd()

	// Start writing
	i := 1
	for i <= 100 {
		sharedRA.PreProtocol()
		text := "I am " + strconv.Itoa(id) + " and this is my contribution number " + strconv.Itoa(i) + "\n"
		sharedRA.File.Write(text)
		sharedRA.SendUpdate(text)
		sharedRA.PostProtocol()
		i++
	}

	// Infinite loop which will be break by the com.SetEnd() function
	for {
	}
}
