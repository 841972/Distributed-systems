package test

import (
	"codigo/fileManager"
	"fmt"
)

func main() {
	f := fileManager.Create("fichero.txt")
	f.Write("Hola esto va dentro del fichero")
	f.Write("Otra cosa")
	fmt.Println(f.Read())
}
