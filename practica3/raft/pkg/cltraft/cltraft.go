// package cltraft
package main

import (
	"fmt"
	"log"
	"net/rpc"
	"raft/internal/raft"
	"time"
)

func main() {
	hosts := []string{
		":29100",
		":29101",
		":29102",
		":29103",
		":29104",
		":29105",
		":29106",
		":29107",
		":29108",
		":29109"}
	idLider := 2 // Inicialmente suponemos un líder
	i := 0
	for i < 3 {
		endpoint := hosts[idLider]
		cliente, err := rpc.Dial("tcp", endpoint)
		if err != nil {
			log.Fatal("Conexion:", err)
		}
		defer cliente.Close()

		operacion := raft.TipoOperacion{Operacion: "escribir", Clave: "pi", Valor: "3.1415"}

		var reply raft.ResultadoRemoto
		err = cliente.Call("NodoRaft.SometerOperacionRaft", operacion, &reply)
		if err != nil {
			log.Fatal("rpc :", err)
		}
		fmt.Println(reply)
		// Actualizar líder si es necesario
		//idLider = reply.IdLider

		// Esperar algo de tiempo antes de enviar la siguiente operación al líder
		time.Sleep(1 * time.Second)
		i++
	}
}
