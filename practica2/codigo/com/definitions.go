/*
* BASADO EN: "definitions.go" de Rafael Tolosana Calasanz y Unai Arronategui
 */
package com

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BarrierMessage struct {
	Id     int
	Status string
}

func KillProcessUsingPort(endpoint string) {
	// Use the lsof command to find the process using the specified port.
	cmd := exec.Command("lsof", "-i", endpoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}

	// Parse the output of lsof to get the process ID.
	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) > 1 {
			pid := fields[1]

			// Use the kill command to terminate the process.
			killCmd := exec.Command("kill", "-9", pid)
			if err := killCmd.Run(); err != nil {
				fmt.Println("Error killing process: ", err)
				os.Exit(1)
			}

			fmt.Println("Killed process with ID\n", pid, " using port ", endpoint)
		} else {
			fmt.Println("No process found using port ", endpoint)
		}
	} else {
		fmt.Println("No process found using port ", endpoint)
	}
}

func CleanPorts() {
	i := 29100
	for i <= 29103 {
		KillProcessUsingPort(":" + strconv.Itoa(i))
		i++
	}
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		//os.Exit(1)
	}
}

func GetArgs() int {
	// Check if at least one command-line argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Please provide an integer as a command-line argument.")
		os.Exit(1)
	}

	// Get the first argument (id)
	id, err := strconv.Atoi(os.Args[1])
	CheckError(err)

	fmt.Printf("The id is: %d\n", id)
	return id
}

func SendReady(id int, barrierEndpoint string) {
	// Attempt to establish a TCP connection
	conn, err := net.Dial("tcp", barrierEndpoint)
	CheckError(err)
	defer conn.Close()

	// Communicate that this process is ready
	msg := BarrierMessage{
		Id:     id,
		Status: "ready",
	}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(msg)
	CheckError(err)
}

func WaitWorkOrder(wg *sync.WaitGroup, endpoint string) {
	defer wg.Done()
	listener, err := net.Listen("tcp", endpoint)
	CheckError(err)
	defer listener.Close()

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		CheckError(err)
		defer conn.Close()

		var msg BarrierMessage
		decoder := gob.NewDecoder(conn)
		err = decoder.Decode(&msg)
		CheckError(err)

		if msg.Status == "work" {
			break
		}
	}
}

func SetEnd() {
	timer := time.NewTimer(5 * time.Second)
	<-timer.C
	fmt.Println("It's time to die")
	os.Exit(0)
}
