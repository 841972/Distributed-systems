/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
 */
package ra

import (
	"codigo/fileManager"
	"codigo/ms"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/DistributedClocks/GoVector/govec/vclock"
)

type Request struct {
	Pid      int
	Clock    vclock.VClock
	Function string
	// Function puede ser "read" o "write", esto es usado
	// para hacer comprobaciones con el mapa de exclusión
}

type Reply struct {
}

type Update struct {
	Pid  int
	Text string
}

type UpdateReply struct {
}

type Exclusion struct {
	function1 string
	function2 string
}

type RASharedDB struct {
	OurSeqNum   vclock.VClock
	HigSeqNum   vclock.VClock
	OutRepCnt   int
	ReqCS       bool
	RepDefd     []bool
	ms          *ms.MessageSystem
	done        chan bool
	chrep       chan bool
	chlog       chan string
	Mutex       sync.Mutex // Mutex para proteger concurrencia sobre las variables
	exclude     map[Exclusion]bool
	request     chan Request
	reply       chan Reply
	updateReply chan UpdateReply
	Function    string
	File        *fileManager.File
	meAsString  string
}

const numProc = 6 // Número de procesos lectores y escritores en el sistema

func New(me int, usersFile string, function string) *RASharedDB {
	// Initialize message system
	messageTypes := []ms.Message{Request{}, Reply{}, Update{}, UpdateReply{}}
	msgs := ms.New(me, usersFile, messageTypes)

	meAsString := strconv.Itoa(me)

	// Initialize vectorial clocks
	vc1 := vclock.New()
	vc1.Set(meAsString, 0)
	vc2 := vclock.New()
	vc2.Set(meAsString, 0)

	// Create file
	filename := "file_" + meAsString + ".txt"
	file := fileManager.Create(filename)

	ra := RASharedDB{vc1, vc2, 0, false, make([]bool, numProc), &msgs, make(chan bool), make(chan bool),
		make(chan string), sync.Mutex{}, make(map[Exclusion]bool), make(chan Request), make(chan Reply),
		make(chan UpdateReply), function, file, meAsString}

	ra.exclude[Exclusion{"read", "read"}] = false
	ra.exclude[Exclusion{"read", "write"}] = true
	ra.exclude[Exclusion{"write", "read"}] = true
	ra.exclude[Exclusion{"write", "write"}] = true

	go handleLogs(&ra)
	go receiveMessages(&ra)
	go handleRequests(&ra)
	go handleReplies(&ra)

	return &ra
}

// Pre: Verdad
// Post: Realiza  el  PreProtocol  para el  algoritmo de
//
//	Ricart-Agrawala Generalizado
func (ra *RASharedDB) PreProtocol() {
	// Get mutex to protect critic section
	ra.Mutex.Lock()

	// Set request indicator to true
	ra.ReqCS = true

	// Increment our sequence number
	ra.OurSeqNum[ra.meAsString] = ra.HigSeqNum[ra.meAsString] + 1
	ra.OurSeqNum.Merge(ra.HigSeqNum) // Get logic clock of rest of processes

	// Release mutex
	ra.Mutex.Unlock()

	// Outstanding reply count
	ra.OutRepCnt = numProc - 1

	// Send requests to the other processes
	for pid := 1; pid <= numProc; pid++ {
		if pid != ra.ms.Me {
			request := Request{
				Pid:      ra.ms.Me,
				Clock:    ra.OurSeqNum,
				Function: ra.Function,
			}
			ra.ms.Send(pid, request)
			log.Println("Request sent to " + strconv.Itoa(pid))
		}
	}

	// Wait to all replies
	<-ra.chrep
	ra.chlog <- "Critic section entered"
	ra.chlog <- "My vectorial clock is " + ra.OurSeqNum.ReturnVCString()
}

// Pre: Verdad
// Post: Realiza el PostProtocol para el algoritmo de
//
//	Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol() {
	ra.Mutex.Lock()
	ra.ReqCS = false
	ra.Mutex.Unlock()
	for pid := 1; pid <= numProc; pid++ {
		if ra.RepDefd[pid-1] {
			ra.RepDefd[pid-1] = false
			ra.Mutex.Lock()
			ra.ms.Send(pid, Reply{})
			ra.Mutex.Unlock()
		}
	}
}

func (ra *RASharedDB) Stop() {
	ra.ms.Stop()
	ra.done <- true
}

func handleLogs(ra *RASharedDB) {
	// Initialize logs
	logFile, err := os.OpenFile(ra.meAsString+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
	}
	defer logFile.Close()
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	// Set the output of the logger to the opened file.
	log.SetOutput(logFile)
	log.Println("Log of Ricart-Agrawala Algorithm")

	// Attend each new log
	for {
		text := <-ra.chlog
		log.Println(text)
	}
}

func isPriorityMine(vc1 vclock.VClock, vc2 vclock.VClock, me int, other int) bool {
	/*if vc1[strconv.Itoa(me)] < vc2[strconv.Itoa(other)] {
		return true
	} else if vc1[strconv.Itoa(me)] == vc2[strconv.Itoa(other)] {
		return me < other
	} else {
		return false
	}*/
	if vc1.Compare(vc2, vclock.Descendant) {
		return true
	} else if vc1.Compare(vc2, vclock.Concurrent) {
		return me < other
	} else {
		return false
	}
}

func receiveMessages(ra *RASharedDB) {
	for {
		message := ra.ms.Receive()
		switch messageType := message.(type) {
		case Request:
			ra.request <- messageType
		case Reply:
			ra.reply <- messageType
		case Update:
			ra.File.Write(messageType.Text)
			ra.ms.Send(messageType.Pid, UpdateReply{})
		case UpdateReply:
			ra.updateReply <- messageType
		}
	}
}

func handleRequests(ra *RASharedDB) {
	for {
		request := <-ra.request
		ra.chlog <- "Request received from " + strconv.Itoa(request.Pid)
		requestVC := request.Clock
		ra.Mutex.Lock()
		ra.HigSeqNum[ra.meAsString] = max(ra.HigSeqNum[ra.meAsString], requestVC[strconv.Itoa(request.Pid)])
		ra.HigSeqNum.Merge(requestVC) // Merge with new clock
		deferIt := ra.ReqCS && isPriorityMine(ra.OurSeqNum, requestVC, ra.ms.Me, request.Pid) && ra.exclude[Exclusion{ra.Function, request.Function}]
		ra.Mutex.Unlock()
		if deferIt {
			ra.RepDefd[request.Pid-1] = true
		} else {
			ra.Mutex.Lock()
			ra.ms.Send(request.Pid, Reply{})
			ra.Mutex.Unlock()
		}
	}
}

func handleReplies(ra *RASharedDB) {
	for {
		<-ra.reply
		ra.OutRepCnt--
		if ra.OutRepCnt == 0 {
			ra.chrep <- true
		}
	}
}

func (ra *RASharedDB) SendUpdate(text string) {
	// Send update to the rest of processes
	for pid := 1; pid <= numProc; pid++ {
		if pid != ra.ms.Me {
			ra.ms.Send(pid, Update{ra.ms.Me, text})
		}
	}
	// Wait in the loop until all replies for this update has been received
	updateRepCnt := numProc - 1
	for updateRepCnt > 0 {
		<-ra.updateReply
		updateRepCnt--
	}
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
