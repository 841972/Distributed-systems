package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"codigo/com"
	"strings"
	"time"
)

const CONN_TYPE = "tcp"
const PORT = "29101"

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

func killProcessUsingPort() {
	// Use the lsof command to find the process using the specified port.
	cmd := exec.Command("lsof", "-i", ":" + PORT)
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

					fmt.Println("Killed process with ID\n", pid, " using port ", PORT)
			} else {
					fmt.Println("No process found using port ", PORT)
			}
	} else {
			fmt.Println("No process found using port ", PORT)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	var request com.Request
	err := decoder.Decode(&request)
	com.CheckError(err)

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

func main() {
	killProcessUsingPort()

	listener, err := net.Listen(CONN_TYPE, ":" + PORT)
	com.CheckError(err)

	defer listener.Close()

	fmt.Println("***** Listening for new connection in port", PORT)
	
	for {
		conn, err:= listener.Accept()
		com.CheckError(err)
		
		fmt.Print("New connection accepted\n")

		handleRequest(conn)
	}
}
