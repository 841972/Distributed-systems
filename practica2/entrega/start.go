package main

import (
	"bufio"
	"codigo/com"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

const (
	numProc  = 6
	endpoint = ":29100"
)

func runProcess(program string, address string, id int) {
	fmt.Printf("Process started at %s with id %d\n", address, id)
	err := exec.Command("ssh", address, "/home/a842236/"+program, strconv.Itoa(id), "> output.log 2> error.log").Run()
	com.CheckError(err)
}

func startProcesses() {
	id := 1
	for i := 2; i <= 4; i++ {
		address := fmt.Sprintf("%s%d", "192.168.3.", i)
		// Start reader
		go runProcess("lector", address, id)
		id++
		// Start writer
		go runProcess("escritor", address, id)
		id++
	}
}

func waitProcesses(wg *sync.WaitGroup) {
	defer wg.Done()
	listener, err := net.Listen("tcp", endpoint)
	com.CheckError(err)
	defer listener.Close()
	fmt.Printf("Listening in %s\n", endpoint)

	i := 1
	for i <= numProc {
		// Accept incoming connection
		conn, err := listener.Accept()
		com.CheckError(err)
		defer conn.Close()

		var msg com.BarrierMessage
		decoder := gob.NewDecoder(conn)
		err = decoder.Decode(&msg)
		com.CheckError(err)

		if msg.Status == "ready" {
			fmt.Printf("Process with id %d is ready\n", msg.Id)
			i++
		}
	}
}

func putProcessesToWork() {
	// Open the file for reading
	file, err := os.Open("ms/barrierUsers.txt")
	com.CheckError(err)
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Read and print each line
	for scanner.Scan() {
		endpoint := scanner.Text()

		// Attempt to establish a TCP connection
		conn, err := net.Dial("tcp", endpoint)
		com.CheckError(err)
		defer conn.Close()

		// Order process to work
		msg := com.BarrierMessage{
			Id:     0,
			Status: "work",
		}
		encoder := gob.NewEncoder(conn)
		encoder.Encode(msg)
		com.CheckError(err)
		fmt.Printf("%s has been ordered to start working\n", endpoint)
	}

	fmt.Printf("All processes are working\n")
}

func main() {
	// Wait for all processes to be ready
	var wg sync.WaitGroup
	wg.Add(1)
	go waitProcesses(&wg)

	// Start readers and writers
	startProcesses()

	// All the processes are ready, order them to work
	wg.Wait()
	fmt.Printf("All processes are ready\n")
	putProcessesToWork()

	time.Sleep(6 * time.Second)
}
