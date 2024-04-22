/*
* AUTOR: Rafael Tolosana Calasanz y Unai Arronategui
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2022
* FICHERO: server-draft.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes a la práctica 1
 */
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os/exec"
	"codigo/com"
	"time"
)

const CONN_TYPE = "tcp"
const endpoint = "127.0.0.1:29100"
const maxRoutines = 10 // For fixed goroutine pool architecture
var mode = 0
var clientJobs chan net.Conn

// PRE: verdad = !foundDivisor
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func isPrime(n int) (foundDivisor bool) {
	foundDivisor = false
	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
//
//	intervalo [interval.A, interval.B]
func findPrimes(interval com.TPInterval) (primes []int) {
	for i := interval.Min; i <= interval.Max; i++ {
		if isPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}

func routine(jobs <-chan net.Conn) {
	for conn := range jobs {
		handleClient(conn)
	}
}

func startWorker(address string) {
	err := exec.Command("ssh", address, "/home/a842236/worker").Run()
	com.CheckError(err)
}

func handleWorker(address string, jobs <-chan net.Conn) {
	endpoint := fmt.Sprintf("%s:29101", address)

	for clientConn := range jobs {
		defer clientConn.Close()

		clientDecoder := gob.NewDecoder(clientConn)
		clientEncoder := gob.NewEncoder(clientConn)

		// Read client request
		var request com.Request
		err := clientDecoder.Decode(&request)
		com.CheckError(err)

		if (request.Id == -1) {
			fmt.Println("Received end request from client.")
		} else {
			workerConn, err := net.Dial("tcp", endpoint)
			com.CheckError(err)
			defer workerConn.Close()

			workerDecoder := gob.NewDecoder(workerConn)
			workerEncoder := gob.NewEncoder(workerConn)

			// Send request to worker
			err = workerEncoder.Encode(request)
			com.CheckError(err)
			log.Println("Send request", request.Id, "to worker")

			// Get reply from worker
			var reply com.Reply
			err = workerDecoder.Decode(&reply)
			com.CheckError(err)

			// Send reply to client
			err = clientEncoder.Encode(reply)
			com.CheckError(err)
			log.Println("Send reply", reply.Id, "to client")
		}
	}
}

// handleClient handles incoming client requests
func handleClient(conn net.Conn) {
	defer conn.Close()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	var request com.Request
	err := decoder.Decode(&request)
	com.CheckError(err)

	if (request.Id == -1) {
		fmt.Println("Received end request from client.")
		return
	}

	// Process the request
	start := time.Now()
	reply := com.Reply {
		Id: request.Id,
		Primes: findPrimes(request.Interval),
	}
	end := time.Now()
	fmt.Println("Processing time: ", end.Sub(start))

	// Send the reply back to the client
	err = encoder.Encode(reply)
	com.CheckError(err)
}

func selectMode() {
	validMode := false
	for validMode == false {
		fmt.Println("Choose a mode:")
		fmt.Println("[0] Client - Secuential server")
		fmt.Println("[1] Client - Concurrent server with goroutines")
		fmt.Println("[2] Client - Concurrent server with fixed goroutine pool")
		fmt.Println("[3] Master - Worker")
		fmt.Scanf("%d\n", &mode)
		if mode >= 0 && mode <= 3 {
			validMode = true
		}
	}
}

func main() {
	selectMode()

	listener, err := net.Listen(CONN_TYPE, endpoint)
	com.CheckError(err)

	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	defer listener.Close()
	
	log.Println("***** Listening for new connection in endpoint ", endpoint)

	if (mode == 2) {
		// Create a channel for incoming client connections
		clientJobs = make(chan net.Conn)

		// Create worker goroutines
		for i := 0; i < maxRoutines; i++ {
			go routine(clientJobs)
		}
	}

	if (mode == 3) {
		// Create a channel for incoming client connections
		clientJobs = make(chan net.Conn)

		// There are three workers
		for i := 2; i <= 4; i++ {
			// Set up worker
			address := fmt.Sprintf("%s%d", "192.168.3.", i)

			go startWorker(address)
			log.Println("Worker is working on", address)

			// Give the worker enough time to start listening
			time.Sleep(2 * time.Second)

			// Create routine to handle worker
			go handleWorker(address, clientJobs)
		}
	}
	
	for {
		// Accept incoming client connections
		conn, err := listener.Accept()
		com.CheckError(err)

		// Handle client
		switch mode {
			case 0: // Client - Secuential server
				handleClient(conn)
			case 1: // Client - Concurrent server with goroutines
				go handleClient(conn)
			case 2: // Client - Concurrent server with fixed goroutine pool
				clientJobs <- conn // Add the connection to the jobs channel for a worker to handle
			case 3: // Master - Worker
				clientJobs <- conn
		}
	}
}
