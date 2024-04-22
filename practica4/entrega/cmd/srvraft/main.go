package main

import (
	//"errors"

	//"log"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"raft/internal/comun/check"
	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"
	"strconv"
	//"time"
)

func main() {
	// obtener entero de indice de este nodo
	me, err := strconv.Atoi(os.Args[1])
	check.CheckError(err, "Main, mal numero entero de indice de nodo:")

	var nodos []rpctimeout.HostPort
	// Resto de argumento son los end points como strings
	// De todas la replicas-> pasarlos a HostPort
	for _, endPoint := range os.Args[2:] {
		nodos = append(nodos, rpctimeout.HostPort(endPoint))
	}

	database := make(map[string]string)
	canalAplicarOperacion := make(chan raft.AplicaOperacion, 1000)
	go applyOperation(database, canalAplicarOperacion)

	// Parte Servidor
	nr := raft.NuevoNodo(nodos, me, canalAplicarOperacion)
	rpc.Register(nr)

	fmt.Println("Replica escucha en :", me, " de ", os.Args[2:])

	l, err := net.Listen("tcp", os.Args[2:][me])
	check.CheckError(err, "Main listen error:")

	rpc.Accept(l)
}

func applyOperation(database map[string]string, canalAplicar chan raft.AplicaOperacion) {
	for {
		solicitud := <-canalAplicar
		if solicitud.Operacion.Operacion == "leer" {
			solicitud.CanalRespuesta <- database[solicitud.Operacion.Clave]
		} else if solicitud.Operacion.Operacion == "escribir" {
			database[solicitud.Operacion.Clave] = solicitud.Operacion.Valor
			solicitud.CanalRespuesta <- "ok"
		}
	}
}
