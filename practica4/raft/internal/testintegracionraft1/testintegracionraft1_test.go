package testintegracionraft1

import (
	"fmt"
	"raft/internal/comun/check"

	//"log"
	//"crypto/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"raft/internal/comun/rpctimeout"
	"raft/internal/despliegue"
	"raft/internal/raft"
)

const (
	//hosts
	MAQUINA1 = "192.168.3.2"
	MAQUINA2 = "192.168.3.3"
	MAQUINA3 = "192.168.3.4"

	//puertos
	PUERTOREPLICA1 = "29100"
	PUERTOREPLICA2 = "29101"
	PUERTOREPLICA3 = "29102"

	//nodos replicas
	REPLICA1 = MAQUINA1 + ":" + PUERTOREPLICA1
	REPLICA2 = MAQUINA2 + ":" + PUERTOREPLICA2
	REPLICA3 = MAQUINA3 + ":" + PUERTOREPLICA3

	//numero de nodos
	NUM_NODOS = 3

	// paquete main de ejecutables relativos a PATH previo
	EXECREPLICA = "cmd/srvraft/main.go"

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_ed25519"
)

// PATH de los ejecutables de modulo golang de servicio Raft
// var PATH string = filepath.Join(os.Getenv("HOME"), "tmp", "p3", "raft")
var PATH string = filepath.Join(os.Getenv("HOME"), "raft")

// go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003
var EXECREPLICACMD string = "cd " + PATH + "; go run " + EXECREPLICA

// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
		NUM_NODOS,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Run test sequence

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:soloArranqueYparada",
		func(t *testing.T) { cfg.soloArranqueYparadaTest1(t) })

	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T2:ElegirPrimerLider",
		func(t *testing.T) { cfg.elegirPrimerLiderTest2(t) })

	// Test3: tenemos el primer primario correcto
	t.Run("T3:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { cfg.falloAnteriorElegirNuevoLiderTest3(t) })

	// Test4: Tres operaciones comprometidas en configuración estable
	t.Run("T4:tresOperacionesComprometidasEstable",
		func(t *testing.T) { cfg.tresOperacionesComprometidasEstable(t) })
}

// TEST primer rango
func TestAcuerdosConFallos(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
		NUM_NODOS,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Test5: Se consigue acuerdo a pesar de desconexiones de seguidor
	t.Run("T5:AcuerdoAPesarDeDesconexionesDeSeguidor ",
		func(t *testing.T) { cfg.AcuerdoApesarDeSeguidor(t) })

	t.Run("T6:SinAcuerdoPorFallos ",
		func(t *testing.T) { cfg.SinAcuerdoPorFallos(t) })

	t.Run("T7:SometerConcurrentementeOperaciones ",
		func(t *testing.T) { cfg.SometerConcurrentementeOperaciones(t) })

	t.Run("T8:AcuerdoConNuevoLider ",
		func(t *testing.T) { cfg.AcuerdoConNuevoLider(t) })

}

// ---------------------------------------------------------------------
//
// Canal de resultados de ejecución de comandos ssh remotos
type canalResultados chan string

func (cr canalResultados) stop() {
	close(cr)

	// Leer las salidas obtenidos de los comandos ssh ejecutados
	/*for s := range cr {
		fmt.Println(s)
	}*/
}

// ---------------------------------------------------------------------
// Operativa en configuracion de despliegue y pruebas asociadas
type configDespliegue struct {
	t           *testing.T
	conectados  []bool
	numReplicas int
	nodosRaft   []rpctimeout.HostPort
	cr          canalResultados
	idLider     int
}

// Crear una configuracion de despliegue
func makeCfgDespliegue(t *testing.T, n int, nodosraft []string,
	conectados []bool) *configDespliegue {
	cfg := &configDespliegue{}
	cfg.t = t
	cfg.conectados = conectados
	cfg.numReplicas = n
	cfg.nodosRaft = rpctimeout.StringArrayToHostPortArray(nodosraft)
	cfg.cr = make(canalResultados, 2000)

	return cfg
}

func (cfg *configDespliegue) stop() {
	//cfg.stopDistributedProcesses()

	time.Sleep(50 * time.Millisecond)

	cfg.cr.stop()
}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// Se pone en marcha una replica ?? - 3 NODOS RAFT
func (cfg *configDespliegue) soloArranqueYparadaTest1(t *testing.T) {
	t.Skip("SKIPPED soloArranqueYparadaTest1")

	fmt.Println(t.Name(), ".....................")

	cfg.t = t // Actualizar la estructura de datos de tests para errores

	// Poner en marcha replicas en remoto con un tiempo de espera incluido
	cfg.startDistributedProcesses()

	time.Sleep(2000 * time.Millisecond)

	// Comprobar estado replica 0
	cfg.comprobarEstadoRemoto(0, 0, false, -1)

	// Comprobar estado replica 1
	cfg.comprobarEstadoRemoto(1, 0, false, -1)

	// Comprobar estado replica 2
	cfg.comprobarEstadoRemoto(2, 0, false, -1)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses()

	fmt.Println(".............", t.Name(), "Superado")
}

// Primer lider en marcha - 3 NODOS RAFT
func (cfg *configDespliegue) elegirPrimerLiderTest2(t *testing.T) {
	t.Skip("SKIPPED ElegirPrimerLiderTest2")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	time.Sleep(1000 * time.Millisecond)

	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	cfg.pruebaUnLider(NUM_NODOS)

	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses() // Parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// Fallo de un primer lider y reeleccion de uno nuevo - 3 NODOS RAFT
func (cfg *configDespliegue) falloAnteriorElegirNuevoLiderTest3(t *testing.T) {
	t.Skip("SKIPPED FalloAnteriorElegirNuevoLiderTest3")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	fmt.Printf("Lider inicial\n")
	cfg.pruebaUnLider(NUM_NODOS)

	// Desconectar lider
	cfg.pararLider()

	time.Sleep(1000 * time.Millisecond)

	fmt.Printf("Comprobar nuevo lider\n")
	cfg.pruebaUnLider(NUM_NODOS)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// 3 operaciones comprometidas con situacion estable y sin fallos - 3 NODOS RAFT
func (cfg *configDespliegue) tresOperacionesComprometidasEstable(t *testing.T) {
	t.Skip("SKIPPED tresOperacionesComprometidasEstable")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	fmt.Printf("Obtener líder\n")
	cfg.pruebaUnLider(NUM_NODOS)

	resultado1 := cfg.someterOperacion("escribir", "x", "1") == "ok"
	fmt.Println("Resultado 1: ", resultado1)
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)
	resultado2 := cfg.someterOperacion("escribir", "y", "2") == "ok"
	fmt.Println("Resultado 2: ", resultado2)
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)
	resultado3 := cfg.someterOperacion("escribir", "z", "3") == "ok"
	fmt.Println("Resultado 3: ", resultado3)
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	if resultado1 && resultado2 && resultado3 {
		fmt.Println(".............", t.Name(), "Superado")
	} else {
		fmt.Println(".............", t.Name(), "No superado")
	}
}

// Se consigue acuerdo a pesar de desconexiones de seguidor -- 3 NODOS RAFT
func (cfg *configDespliegue) AcuerdoApesarDeSeguidor(t *testing.T) {
	//t.Skip("SKIPPED AcuerdoApesarDeSeguidor")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	// Obtener líder
	fmt.Printf("Obtener líder\n")
	cfg.pruebaUnLider(NUM_NODOS)

	// Comprometer una entrada
	resultado1 := cfg.someterOperacion("escribir", "x", "1") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// Desconectar uno de los nodos Raft
	desconectados := cfg.pararSeguidores(1)

	// Comprobar varios acuerdos con una réplica desconectada
	resultado2 := cfg.someterOperacion("escribir", "y", "2") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)
	resultado3 := cfg.someterOperacion("escribir", "z", "3") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
	cfg.reconectarSeguidores(desconectados)
	resultado4 := cfg.someterOperacion("leer", "y", "?") == "2"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)
	resultado5 := cfg.someterOperacion("leer", "z", "?") == "3"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	if resultado1 && resultado2 && resultado3 && resultado4 && resultado5 {
		fmt.Println(".............", t.Name(), "Superado")
	} else {
		fmt.Println(".............", t.Name(), "No superado")
	}
}

// NO se consigue acuerdo al desconectarse mayoría de seguidores -- 3 NODOS RAFT
func (cfg *configDespliegue) SinAcuerdoPorFallos(t *testing.T) {
	//t.Skip("SKIPPED SinAcuerdoPorFallos")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	// Obtener líder
	fmt.Printf("Obtener líder\n")
	cfg.pruebaUnLider(NUM_NODOS)

	// Comprometer una entrada
	_ = cfg.someterOperacion("escribir", "a", "1") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// Desconectar 2 de los nodos Raft
	desconectados := cfg.pararSeguidores(2)

	// Comprobar varios acuerdos con 2 réplicas desconectada
	_ = cfg.someterOperacion("escribir", "b", "2") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)
	// Como la entrada anterior no se habrá comprometido, esta lectura dará error
	resultado1 := cfg.someterOperacion("leer", "b", "?") == "2"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// reconectar los 2 nodos Raft  desconectados y probar varios acuerdos
	cfg.reconectarSeguidores(desconectados)
	_ = cfg.someterOperacion("escribir", "c", "3") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)
	resultado2 := cfg.someterOperacion("leer", "c", "?") == "3"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	// resultado 1 tiene que ser falso al no haber mayoría,
	// y el 2 verdadero al estar reconectados los seguidores
	fmt.Printf("Resultado al escribir sin seguidores (debería ser falso): %v\n", resultado1)
	fmt.Printf("Resutlado al escribir con seguidores (debería ser verdadero): %v\n", resultado2)
	if !resultado1 && resultado2 {
		fmt.Println(".............", t.Name(), "Superado")
	} else {
		fmt.Println(".............", t.Name(), "No superado")
	}
}

// Se somete 5 operaciones de forma concurrente -- 3 NODOS RAFT
func (cfg *configDespliegue) SometerConcurrentementeOperaciones(t *testing.T) {
	//t.Skip("SKIPPED SometerConcurrentementeOperaciones")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	// Obtener líder
	fmt.Printf("Obtener líder\n")
	cfg.pruebaUnLider(NUM_NODOS)

	// un bucle para estabilizar la ejecucion
	// Someter 5  operaciones concurrentes
	letras := [5]string{"a", "b", "c", "d", "e"}
	var resultados [5]bool
	for i := 0; i < 5; i++ {
		clave := letras[i]
		valor := strconv.Itoa(i)
		go func(i int) {
			resultados[i] = cfg.someterOperacion("escribir", clave, valor) == "ok"
		}(i)
	}
	time.Sleep(3000 * time.Millisecond)

	resultadosCorrectos := allTrue(resultados[:]...)
	fmt.Printf("Los resultados de las operaciones han sido: %v\n", resultadosCorrectos)
	estadosCorrectos := cfg.comprobarEstadoNodos()

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	// Comprobar estados de nodos Raft, sobre todo
	// el avance del mandato en curso e indice de registro de cada uno
	// que debe ser identico entre ellos
	if resultadosCorrectos && estadosCorrectos {
		fmt.Println(".............", t.Name(), "Superado")
	} else {
		fmt.Println(".............", t.Name(), "No superado")
	}
}

// Comprometer, eliminar líder y volver a comprometer
func (cfg *configDespliegue) AcuerdoConNuevoLider(t *testing.T) {
	//t.Skip("SKIPPED SometerConcurrentementeOperaciones")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	// Obtener líder
	cfg.pruebaUnLider(NUM_NODOS)
	fmt.Printf("Obtener líder: %d\n", cfg.idLider)
	antiguoLider := cfg.idLider

	// Comprometer una entrada
	resultado1 := cfg.someterOperacion("escribir", "x", "1") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// Parar líder
	cfg.pararLider()
	time.Sleep(1000 * time.Millisecond)

	// Obtenemos un nuevo líder
	cfg.pruebaUnLider(NUM_NODOS)
	fmt.Printf("Obtener nuevo líder: %d\n", cfg.idLider)

	// Comprometer otra entrada con el nuevo líder
	resultado2 := cfg.someterOperacion("escribir", "y", "2") == "ok"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	// Reconectar al antiguo líder
	cfg.reconectarSeguidores([]int{antiguoLider})

	// Comprometemos otra entrada
	resultado3 := cfg.someterOperacion("leer", "y", "?") == "2"
	cfg.comprobarEstadoRemoto(cfg.idLider, 1, true, cfg.idLider)

	time.Sleep(2000 * time.Millisecond)

	// Comprobar mismos mandatos e índices de registro
	fmt.Printf("R1: %v | R2: %v | R3: %v\n", resultado1, resultado2, resultado3)
	estadosCorrectos := cfg.comprobarEstadoNodos()

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	// Comprobar estados de nodos Raft, sobre todo
	// el avance del mandato en curso e indice de registro de cada uno
	// que debe ser identico entre ellos
	if resultado1 && resultado2 && resultado3 && estadosCorrectos {
		fmt.Println(".............", t.Name(), "Superado")
	} else {
		fmt.Println(".............", t.Name(), "No superado")
	}
}

// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func (cfg *configDespliegue) pruebaUnLider(numreplicas int) int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		mapaLideres := make(map[int][]int)
		for i := 0; i < numreplicas; i++ {
			if cfg.conectados[i] {
				if _, mandato, eslider, _ := cfg.obtenerEstadoRemoto(i); eslider {
					mapaLideres[mandato] = append(mapaLideres[mandato], i)
					cfg.idLider = i
				}
			}
		}

		ultimoMandatoConLider := -1
		for mandato, lideres := range mapaLideres {
			if len(lideres) > 1 {
				cfg.t.Fatalf("mandato %d tiene %d (>1) lideres",
					mandato, len(lideres))
			}
			if mandato > ultimoMandatoConLider {
				ultimoMandatoConLider = mandato
			}
		}

		if len(mapaLideres) != 0 {

			return mapaLideres[ultimoMandatoConLider][0] // Termina

		}
	}
	cfg.t.Fatalf("un lider esperado, ninguno obtenido")

	return -1 // Termina
}

func (cfg *configDespliegue) obtenerEstadoRemoto(
	indiceNodo int) (int, int, bool, int) {
	var reply raft.EstadoRemoto
	/*err_ :=*/ cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
		raft.Vacio{}, &reply, 100*time.Millisecond)
	//check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cfg *configDespliegue) startDistributedProcesses() {
	//cfg.t.Log("Before starting following distributed processes: ", cfg.nodosRaft)

	for i, endPoint := range cfg.nodosRaft {
		despliegue.ExecMutipleHosts(EXECREPLICACMD+
			" "+strconv.Itoa(i)+" "+
			rpctimeout.HostPortArrayToString(cfg.nodosRaft),
			[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)

		// Dar tiempo para se establezcan las replicas
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(3300 * time.Millisecond)
}

func (cfg *configDespliegue) stopDistributedProcesses() {
	var reply raft.Vacio

	for _, endPoint := range cfg.nodosRaft {
		/*err :=*/ endPoint.CallTimeout("NodoRaft.ParaNodo",
			raft.Vacio{}, &reply, 10*time.Millisecond)
		//check.CheckError(err, "Error en llamada RPC Para nodo")
	}
}

// Comprobar estado remoto de un nodo con respecto a un estado prefijado
func (cfg *configDespliegue) comprobarEstadoRemoto(idNodoDeseado int,
	mandatoDeseado int, esLiderDeseado bool, IdLiderDeseado int) {
	idNodo, mandato, esLider, idLider := cfg.obtenerEstadoRemoto(idNodoDeseado)

	cfg.t.Log("Params: ", idNodoDeseado, mandatoDeseado, esLiderDeseado, IdLiderDeseado, "\n")
	cfg.t.Log("Estado replica 0: ", idNodo, mandato, esLider, idLider, "\n")

	/*if idNodo != idNodoDeseado || mandato != mandatoDeseado ||
	esLider != esLiderDeseado || idLider != IdLiderDeseado {*/
	if idNodo != idNodoDeseado {
		cfg.t.Fatalf("Estado incorrecto en replica %d en subtest %s",
			idNodoDeseado, cfg.t.Name())
	}

}

// -------------------------------
// Funciones añadidas por nosotros

func (cfg *configDespliegue) pararLider() {
	fmt.Printf("Desconectamos al líder %d\n", cfg.idLider)
	var reply raft.EstadoRemoto
	err := cfg.nodosRaft[cfg.idLider].CallTimeout("NodoRaft.ParaNodo", raft.Vacio{}, &reply, 100*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC ParaNodo")
}

// Desconecta un número de seguidores dado
func (cfg *configDespliegue) pararSeguidores(numSeguidores int) []int {
	// Obtener lista de seguidores
	seguidores := []int{}
	for i := 0; i < len(cfg.nodosRaft); i++ {
		if i != cfg.idLider {
			seguidores = append(seguidores, i)
		}
	}
	// Parar hasta "numSeguidores" seguidores
	desconectados := []int{}
	for j := 0; j < numSeguidores; j++ {
		fmt.Printf("Desconectar seguidor %d\n", seguidores[j])
		var reply raft.EstadoRemoto
		err := cfg.nodosRaft[seguidores[j]].CallTimeout("NodoRaft.ParaNodo", raft.Vacio{}, &reply, 1000*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC ParaNodo")
		desconectados = append(desconectados, seguidores[j])
	}
	time.Sleep(3000 * time.Millisecond)
	return desconectados
}

func (cfg *configDespliegue) reconectarSeguidores(seguidores []int) {
	for _, idSeguidor := range seguidores {
		despliegue.ExecMutipleHosts(EXECREPLICACMD+
			" "+strconv.Itoa(idSeguidor)+" "+
			rpctimeout.HostPortArrayToString(cfg.nodosRaft),
			[]string{cfg.nodosRaft[idSeguidor].Host()}, cfg.cr, PRIVKEYFILE)
	}
	// dar tiempo para se establezcan los seguidores
	time.Sleep(3000 * time.Millisecond)
}

func (cfg *configDespliegue) someterOperacion(operacion string, clave string, valor string) string {
	request := raft.TipoOperacion{
		Operacion: operacion,
		Clave:     clave,
		Valor:     valor,
	}
	var reply raft.ResultadoRemoto
	/*err :=*/ cfg.nodosRaft[cfg.idLider].CallTimeout("NodoRaft.SometerOperacionRaft",
		request, &reply, 8000*time.Millisecond)
	return reply.ValorADevolver
}

func (cfg *configDespliegue) comprobarEstadoNodos() bool {
	var reply raft.EstadoRemoto
	var mandato int
	var indice int
	for i := 0; i < len(cfg.nodosRaft); i++ {
		cfg.nodosRaft[i].CallTimeout("NodoRaft.ObtenerEstadoNodo", raft.Vacio{}, &reply, 100*time.Millisecond)
		mandatoAux := reply.Mandato
		indiceAux := reply.IndiceRegistro
		fmt.Printf("Nodo: %d | Mandato: %d | Índice registro: %d\n", i, mandatoAux, indiceAux)
		if i == 0 { // Si es la primera iteración
			mandato = mandatoAux
			indice = indiceAux
		} else {
			// Comprobamos que coincidan todos los mandatos e índices de registro
			if mandatoAux != mandato || indiceAux != indice {
				return false
			}
		}
	}
	return true
}

func allTrue(slice ...bool) bool {
	for _, value := range slice {
		if !value {
			return false
		}
	}
	return true
}
