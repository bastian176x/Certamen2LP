package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Dispatcher struct {
	colaListos      []*Process
	colaBloqueados  []*Process
	maxInstructions int
	tick            time.Duration
	canalesProceso  map[string]*CanalProcesos
}

type CanalProcesos struct {
	proceso      *Process
	comandosChan chan string
	statusChan   chan string
}

// Agrergar un proceso a la cola de listos
func (d *Dispatcher) PushProcessListos(p *Process) {
	d.colaListos = append(d.colaListos, p)
}

// Agregar un proceso a la cola de bloqueados
func (d *Dispatcher) addProcessBloqueados(p *Process) {
	d.colaBloqueados = append(d.colaBloqueados, p)
}

func (d *Dispatcher) gestionarProcesos() {
	d.canalesProceso = make(map[string]*CanalProcesos)

	// Lanzar goroutines para los procesos iniciales
	for _, proceso := range d.colaListos {
		d.lanzarProceso(proceso)
	}

	for {
		// Si no hay listos y sí bloqueados, seguir descontando tiempo
		if len(d.colaListos) == 0 && len(d.colaBloqueados) > 0 {
			d.descontarTiempoBloqueados()
			continue
		}
		// Si no hay ni listos ni bloqueados, terminar
		if len(d.colaListos) == 0 && len(d.colaBloqueados) == 0 {
			break
		}

		proceso := d.colaListos[0]

		fmt.Fprintln(out, "PULL Dispatcher")
		d.descontarTiempoBloqueados()

		fmt.Fprintln(out, "LOADING ->", proceso.Nombre)
		d.descontarTiempoBloqueados()

		fmt.Fprintln(out, "EXECUTE ->", proceso.Nombre)
		d.descontarTiempoBloqueados()

		for {
			d.canalesProceso[proceso.Nombre].comandosChan <- "EXECUTE"
			d.descontarTiempoBloqueados()

			status := <-d.canalesProceso[proceso.Nombre].statusChan
			if status == "FINISHED" {
				fmt.Fprintln(out, "FINISHED ->", proceso.Nombre)
				d.descontarTiempoBloqueados()

				d.colaListos = d.colaListos[1:]
				d.cerrarCanalesProceso(proceso.Nombre)
				break
			}

			if matched := regexp.MustCompile(`BLOCKED:(\d+)`).FindStringSubmatch(status); matched != nil {
				fmt.Fprintln(out, "STORING ->", proceso.Nombre)
				d.descontarTiempoBloqueados()

				fmt.Fprintln(out, "PUSH BLOQUEADO ->", proceso.Nombre)
				d.descontarTiempoBloqueados()

				tiempo_bloq, _ := strconv.Atoi(matched[1])
				proceso.Tiempo_ES = tiempo_bloq

				d.cerrarCanalesProceso(proceso.Nombre)

				d.addProcessBloqueados(proceso)
				d.colaListos = d.colaListos[1:]
				break
			}

			if status == "EXECUTING" {
				//se verifica si el proceso ha llegado a la cantidad maxima de instrucciones por ciclo
				if proceso.Program_counter%d.maxInstructions == 0 && len(d.colaListos) > 1 {
					fmt.Fprintln(out, "CAMBIO DE CONTEXTO -> "+proceso.Nombre)
					d.descontarTiempoBloqueados()

					d.colaListos = d.colaListos[1:]

					fmt.Fprintln(out, "STORING ->", proceso.Nombre)
					d.descontarTiempoBloqueados()

					fmt.Fprintln(out, "PUSH LISTO->", proceso.Nombre)
					d.descontarTiempoBloqueados()

					d.PushProcessListos(proceso)
					break
				}
			}
			time.Sleep(d.tick)
		}
	}
}

// Lanza la goroutine para un proceso, creando sus canales
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

// Cierra los canales de un proceso y lo elimina del mapa
func (d *Dispatcher) cerrarCanalesProceso(nombre string) {
	if ch, existe := d.canalesProceso[nombre]; existe {
		close(ch.comandosChan)
		close(ch.statusChan)
		delete(d.canalesProceso, nombre)
	}
}

// Descuenta el tiempo de los procesos bloqueados y desbloquea si corresponde
func (d *Dispatcher) descontarTiempoBloqueados() {
	if len(d.colaBloqueados) == 0 {
		return
	}

	for i := 0; i < len(d.colaBloqueados); i++ {
		d.colaBloqueados[i].Tiempo_ES--
		if d.colaBloqueados[i].Tiempo_ES <= 0 {
			fmt.Fprintln(out, "DESBLOQUEADO ->", d.colaBloqueados[i].Nombre)
			d.colaBloqueados[i].Estado = "Listo"

			p := d.colaBloqueados[i]
			// quitar de bloqueados
			d.colaBloqueados = append(d.colaBloqueados[:i], d.colaBloqueados[i+1:]...)
			i--

			d.PushProcessListos(p)
			d.lanzarProceso(p)
		}
	}
}
