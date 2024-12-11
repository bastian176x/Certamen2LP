package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Dispatcher struct {
	colaListos      []*Process
	maxInstructions int
	tick            time.Duration
	canalesProceso  map[string]*CanalProcesos
	bcpTable        map[string]*BCP
	cpu             *CPU
}

type CanalProcesos struct {
	proceso      *Process
	comandosChan chan string
	statusChan   chan string
}

type CPU struct {
	procesoActual *Process
}

// Crear el BCP si no existe la tabla
func (d *Dispatcher) initBCPTable() {
	if d.bcpTable == nil {
		d.bcpTable = make(map[string]*BCP)
	}
}

// Actualiza el BCP de un proceso
func (d *Dispatcher) updateBCP(p *Process) {
	d.initBCPTable()
	d.bcpTable[p.Nombre] = p.ToBCP()
}

// Agregar un proceso a la cola de listos
func (d *Dispatcher) PushProcessListos(p *Process) {
	d.colaListos = append(d.colaListos, p)
	d.updateBCP(p)
}

// Cargar un proceso en la CPU
func (d *Dispatcher) cargarProceso(proceso *Process) {
	fmt.Fprintln(out, "LOADING ->", proceso.Nombre)
	estado := d.bcpTable[proceso.Nombre]
	proceso.RestaurarEstado(estado)
	d.cpu.procesoActual = proceso
}

// Lanzar un proceso en una goroutine
func (d *Dispatcher) lanzarProceso(p *Process) {
	proch := &CanalProcesos{
		proceso:      p,
		comandosChan: make(chan string),
		statusChan:   make(chan string),
	}
	d.canalesProceso[p.Nombre] = proch
	_, probabilidadCierre := recibir_parametros()
	go p.arrancar(proch.comandosChan, proch.statusChan, probabilidadCierre)
}

// Cerrar canales de un proceso
func (d *Dispatcher) cerrarCanalesProceso(nombre string) {
	if ch, existe := d.canalesProceso[nombre]; existe {
		close(ch.comandosChan)
		close(ch.statusChan)
		delete(d.canalesProceso, nombre)
	}
}

// Descontar tiempo de los procesos bloqueados dentro de colaListos
func (d *Dispatcher) descontarTiempoBloqueados() {
	for _, p := range d.colaListos {
		if p.Estado == "Bloqueado" && p.Tiempo_ES > 0 {
			p.Tiempo_ES--
			if p.Tiempo_ES <= 0 {
				fmt.Fprintln(out, "DESBLOQUEADO ->", p.Nombre)
				p.Estado = "Listo"
				d.updateBCP(p)
				// Se relanza el proceso ya que se cerraron canales cuando se bloqueó
				d.lanzarProceso(p)
			}
		}
	}
}

// Encontrar el primer proceso listo en colaListos
func (d *Dispatcher) encontrarProcesoListo() *Process {
	for _, p := range d.colaListos {
		if p.Estado == "Listo" {
			return p
		}
	}
	return nil
}

// Eliminar un proceso de la cola de listos
func (d *Dispatcher) eliminarProcesoDeListos(p *Process) {
	for i, proc := range d.colaListos {
		if proc == p {
			d.colaListos = append(d.colaListos[:i], d.colaListos[i+1:]...)
			return
		}
	}
}

// Gestión principal de procesos
func (d *Dispatcher) gestionarProcesos() {
	d.canalesProceso = make(map[string]*CanalProcesos)
	d.cpu = &CPU{}

	for _, proceso := range d.colaListos {
		d.lanzarProceso(proceso)
		d.updateBCP(proceso)
	}

	for {
		if len(d.colaListos) == 0 {
			// No hay procesos en colaListos, break
			break
		}

		// Encontrar el primer proceso que esté "Listo"
		proceso := d.encontrarProcesoListo()
		if proceso == nil {
			// si ningun proceso está listo, descontar tiempos de bloqueo
			d.descontarTiempoBloqueados()
			continue
		}

		d.descontarTiempoBloqueados()
		fmt.Fprintln(out, "PULL ->", proceso.Nombre)
		d.cargarProceso(proceso)

		fmt.Fprintln(out, "EXECUTE ->", proceso.Nombre)
		d.descontarTiempoBloqueados()

		for {
			d.canalesProceso[proceso.Nombre].comandosChan <- "EXECUTE"
			d.descontarTiempoBloqueados()

			status := <-d.canalesProceso[proceso.Nombre].statusChan
			if status == "FINISHED" {
				fmt.Fprintln(out, "FINISHED ->", proceso.Nombre)
				d.descontarTiempoBloqueados()

				d.eliminarProcesoDeListos(proceso)
				d.cerrarCanalesProceso(proceso.Nombre)
				break
			}

			if matched := regexp.MustCompile(`BLOCKED:(\d+)`).FindStringSubmatch(status); matched != nil {
				fmt.Fprintln(out, "STORE ->", proceso.Nombre)
				fmt.Fprintln(out, "PROCESO BLOQUEADO ->", proceso.Nombre)
				fmt.Fprintln(out, "PUSH BLOQUEADO ->", proceso.Nombre)
				tiempoBloq, _ := strconv.Atoi(matched[1])
				proceso.Tiempo_ES = tiempoBloq
				proceso.Estado = "Bloqueado"
				d.updateBCP(proceso)

				d.cerrarCanalesProceso(proceso.Nombre)
				// permanecerá en la colaListos en estado bloqueado
				break
			}

			if status == "EXECUTING" {
				//  actualizar BCP después de ejecutar instrucción
				d.updateBCP(proceso)

				// hacer cambio de contexto
				if proceso.Program_counter%d.maxInstructions == 0 && len(d.colaListos) > 1 {
					fmt.Fprintln(out, "STORE ->", proceso.Nombre)
					fmt.Fprintln(out, "CAMBIO DE CONTEXTO ->", proceso.Nombre)
					//sacar el proceso ejecutado y ponerlo al final de la cola
					fmt.Fprintln(out, "PUSH LISTO ->", proceso.Nombre)
					d.colaListos = append(d.colaListos[1:], proceso)
					break
				}
			}
			time.Sleep(d.tick)
		}
	}
}
