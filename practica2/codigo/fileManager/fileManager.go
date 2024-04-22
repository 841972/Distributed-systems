/*
* AUTORES: Pablo Moreno (841972) y Andrés Yubero (842236)
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: octubre de 2023
* DESCRIPCIÓN: Gestión de un fichero distribuido en el problema de los lectores y escritores
 */

package fileManager

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type File struct {
	filename string
	Mutex    sync.Mutex
}

func Create(filename string) *File {
	_, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	f := File{filename, sync.Mutex{}}
	return &f
}

func (f *File) Read() string {
	f.Mutex.Lock()
	// Attempt to open the file for reading
	file, err := os.Open(f.filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return ""
	}
	defer file.Close()

	// Create a scanner to read lines from the file
	scanner := bufio.NewScanner(file)

	// Read each line
	contents := ""
	for scanner.Scan() {
		contents += scanner.Text()
	}

	// Check errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading the file:", err)
	}
	f.Mutex.Unlock()

	return contents
}

func (f *File) Write(data string) {
	f.Mutex.Lock()
	// Attempt to open the file for writing
	file, err := os.OpenFile(f.filename, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Write data in the file
	_, err = file.WriteString(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	f.Mutex.Unlock()
}
