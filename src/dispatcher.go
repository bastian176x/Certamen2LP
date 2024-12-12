package main

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type Dispatcher struct {
	colaListos      []*Process
	maxInstructions int
	tick            time.Duration
	canalesProceso  map[string]*CanalProcesos
	bcpTable        map[string]*BCP
	cpu             *CPU
	wg              *sync.WaitGroup
	logger          io.Writer

	mu sync.Mutex
}

type CanalProcesos struct {
	proceso      *Process
	bcp          *BCP
	comandosChan chan string
	statusChan   chan string
}

type CPU struct {
	procesoActual *Process
	bcpActual     *BCP
}

var regexBlocked = regexp.MustCompile(`BLOCKED:(\d+)`)

// Crear el BCP si no existe la tabla
func (d *Dispatcher) initBCPTable() {
	if d.bcpTable == nil {
		d.bcpTable = make(map[string]*BCP)
	}
}

// Actualiza el BCP en la tabla (con mutex)
func (d *Dispatcher) updateBCP(bcp *BCP) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.initBCPTable()
	d.bcpTable[bcp.Nombre] = bcp
}

// Agregar un proceso a la cola de listos con su BCP (con mutex)
func (d *Dispatcher) PushProcessListos(p *Process, bcp *BCP) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.colaListos = append(d.colaListos, p)
	d.initBCPTable()
	d.bcpTable[bcp.Nombre] = bcp
}

// Cargar un proceso en la CPU (con mutex)
func (d *Dispatcher) cargarProceso(proceso *Process) {
	d.mu.Lock()
	defer d.mu.Unlock()
	bcp, ok := d.bcpTable[proceso.Nombre]
	if !ok {
		fmt.Fprintln(d.logger, "ERROR: BCP no encontrado para", proceso.Nombre)
		return
	}
	fmt.Fprintln(d.logger, "LOADING ->", proceso.Nombre)
	d.cpu.procesoActual = proceso
	d.cpu.bcpActual = bcp
}

// Lanzar un proceso en una goroutine (con mutex)
func (d *Dispatcher) lanzarProceso(p *Process, probabilidadCierre int) {
	d.mu.Lock()
	bcp, ok := d.bcpTable[p.Nombre]
	if !ok {
		d.mu.Unlock()
		fmt.Fprintln(d.logger, "ERROR: No se encontró BCP para el proceso:", p.Nombre)
		return
	}
	proch := &CanalProcesos{
		proceso:      p,
		bcp:          bcp,
		comandosChan: make(chan string),
		statusChan:   make(chan string),
	}
	if d.canalesProceso == nil {
		d.canalesProceso = make(map[string]*CanalProcesos)
	}
	d.canalesProceso[p.Nombre] = proch
	d.mu.Unlock()

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		p.arrancar(proch.comandosChan, proch.statusChan, probabilidadCierre, proch.bcp, d.logger)
	}()
}

// Cerrar canales de un proceso (con mutex)
func (d *Dispatcher) cerrarCanalesProceso(nombre string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if ch, existe := d.canalesProceso[nombre]; existe {
		close(ch.comandosChan)
		close(ch.statusChan)
		delete(d.canalesProceso, nombre)
	}
}

// Descontar tiempo de los procesos bloqueados dentro de colaListos (con mutex)
func (d *Dispatcher) descontarTiempoBloqueados(probabilidadCierre int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, p := range d.colaListos {
		bcp := d.bcpTable[p.Nombre]
		if bcp.Estado == "Bloqueado" && bcp.Tiempo_ES > 0 {
			bcp.Tiempo_ES--
			if bcp.Tiempo_ES <= 0 {
				fmt.Fprintln(d.logger, "DESBLOQUEADO ->", bcp.Nombre)
				bcp.Estado = "Listo"
				d.bcpTable[bcp.Nombre] = bcp
				// Lanzar proceso desbloqueado fuera del candado
				go d.lanzarProceso(p, probabilidadCierre)
			}
		}
	}
}

// Encontrar el primer proceso listo en colaListos (con mutex)
func (d *Dispatcher) encontrarProcesoListo() *Process {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, p := range d.colaListos {
		bcp := d.bcpTable[p.Nombre]
		if bcp.Estado == "Listo" {
			return p
		}
	}
	return nil
}

// Eliminar un proceso de la cola de listos (con mutex)
func (d *Dispatcher) eliminarProcesoDeListos(p *Process) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, proc := range d.colaListos {
		if proc == p {
			d.colaListos = append(d.colaListos[:i], d.colaListos[i+1:]...)
			return
		}
	}
}

// Gestión principal de procesos
func (d *Dispatcher) gestionarProcesos(probabilidadCierre int) {
	d.canalesProceso = make(map[string]*CanalProcesos)
	d.cpu = &CPU{}

	// Lanzar todos los procesos iniciales
	d.mu.Lock()
	for _, proceso := range d.colaListos {
		go d.lanzarProceso(proceso, probabilidadCierre)
	}
	d.mu.Unlock()

	for {
		d.mu.Lock()
		if len(d.colaListos) == 0 {
			// No hay procesos
			d.mu.Unlock()
			break
		}
		d.mu.Unlock()

		proceso := d.encontrarProcesoListo()
		if proceso == nil {
			d.descontarTiempoBloqueados(probabilidadCierre)
			time.Sleep(d.tick)
			continue
		}

		d.descontarTiempoBloqueados(probabilidadCierre)
		fmt.Fprintln(d.logger, "PULL ->", proceso.Nombre)
		d.cargarProceso(proceso)
		bcp := d.cpu.bcpActual

		fmt.Fprintln(d.logger, "EXECUTE ->", proceso.Nombre)
		d.descontarTiempoBloqueados(probabilidadCierre)

		for {
			d.mu.Lock()
			canal, existe := d.canalesProceso[proceso.Nombre]
			d.mu.Unlock()

			if !existe {
				// El canal no existe: el proceso finalizó o fue cerrado
				break
			}

			select {
			case canal.comandosChan <- "EXECUTE":
			default:
				// No se pudo mandar el comando, canal cerrado
				break
			}

			d.descontarTiempoBloqueados(probabilidadCierre)

			status, ok := <-canal.statusChan
			if !ok {
				// Canal cerrado, el proceso terminó
				break
			}

			if status == "FINISHED" {
				fmt.Fprintln(d.logger, "FINISHED ->", proceso.Nombre)
				d.eliminarProcesoDeListos(proceso)
				d.cerrarCanalesProceso(proceso.Nombre)
				break
			}

			if matched := regexBlocked.FindStringSubmatch(status); matched != nil {
				fmt.Fprintln(d.logger, "STORE ->", proceso.Nombre)
				fmt.Fprintln(d.logger, "PROCESO BLOQUEADO ->", proceso.Nombre)
				fmt.Fprintln(d.logger, "PUSH BLOQUEADO ->", proceso.Nombre)
				tiempoBloq, _ := strconv.Atoi(matched[1])
				bcp.Tiempo_ES = tiempoBloq
				bcp.Estado = "Bloqueado"
				d.updateBCP(bcp)

				d.cerrarCanalesProceso(proceso.Nombre)
				// permanecerá en colaListos como bloqueado
				break
			}

			if status == "EXECUTING" {
				// actualizar BCP
				d.updateBCP(bcp)

				// hacer cambio de contexto si corresponde
				if bcp.Program_counter%d.maxInstructions == 0 {
					d.mu.Lock()
					if len(d.colaListos) > 1 {
						fmt.Fprintln(d.logger, "STORE ->", proceso.Nombre)
						fmt.Fprintln(d.logger, "CAMBIO DE CONTEXTO ->", proceso.Nombre)
						fmt.Fprintln(d.logger, "PUSH LISTO ->", proceso.Nombre)
						// mover el proceso al final de la cola
						d.colaListos = append(d.colaListos[1:], proceso)
						d.mu.Unlock()
						break
					}
					d.mu.Unlock()
				}
			}
			time.Sleep(d.tick)
		}
	}
}
