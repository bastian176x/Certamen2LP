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
	canalesProceso  map[string]*CanalProcesos // NUEVO: guardamos aquí el mapa de canales
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

	d.canalesProceso = make(map[string]*CanalProcesos) // NUEVO: inicializamos aquí

	// Lanzar goroutines para los procesos iniciales
	for _, proceso := range d.colaListos {
		proch := &CanalProcesos{
			proceso:      proceso,
			comandosChan: make(chan string),
			statusChan:   make(chan string),
		}
		d.canalesProceso[proceso.Nombre] = proch
		go proceso.arrancar(proch.comandosChan, proch.statusChan)
	}

	for {
		if len(d.colaListos) == 0 && len(d.colaBloqueados) > 0 {
			d.descontarTiempoBloqueados()
			continue
		} else if len(d.colaListos) == 0 && len(d.colaBloqueados) == 0 {
			break
		}
		proceso := d.colaListos[0]

		fmt.Println("PULL Dispatcher")
		d.descontarTiempoBloqueados()
		fmt.Println("LOAD ->", proceso.Nombre)
		d.descontarTiempoBloqueados()
		fmt.Println("EXECUTE ->", proceso.Nombre)
		d.descontarTiempoBloqueados()

		for {
			d.canalesProceso[proceso.Nombre].comandosChan <- "EXECUTE"
			status := <-d.canalesProceso[proceso.Nombre].statusChan
			if status == "FINISHED" {
				fmt.Println("FINISHED ->", proceso.Nombre)
				d.colaListos = d.colaListos[1:]
				// El proceso terminó. Podríamos cerrar sus canales aquí si queremos.
				close(d.canalesProceso[proceso.Nombre].comandosChan)
				close(d.canalesProceso[proceso.Nombre].statusChan)
				delete(d.canalesProceso, proceso.Nombre)
				break
			}
			if matched := regexp.MustCompile(`BLOCKED:(\d+)`).FindStringSubmatch(status); matched != nil {
				fmt.Println("STORING ->", proceso.Nombre)
				fmt.Println("PUSH BLOQUEADO ->", proceso.Nombre)

				tiempo_bloq, _ := strconv.Atoi(matched[1])
				proceso.Tiempo_ES = tiempo_bloq

				// El proceso se bloquea. Cerramos sus canales actuales ya que la goroutine terminará.
				close(d.canalesProceso[proceso.Nombre].comandosChan)
				close(d.canalesProceso[proceso.Nombre].statusChan)
				delete(d.canalesProceso, proceso.Nombre)

				d.addProcessBloqueados(proceso)
				d.colaListos = d.colaListos[1:]
				break
			}
			if status == "EXECUTING" {
				d.descontarTiempoBloqueados()
				if proceso.Program_counter%d.maxInstructions == 0 && len(d.colaListos) > 1 {
					fmt.Println("CAMBIO DE CONTEXTO -> " + proceso.Nombre)
					d.colaListos = d.colaListos[1:]
					fmt.Println("STORING ->", proceso.Nombre)
					fmt.Println("PUSH LISTO->", proceso.Nombre)
					d.PushProcessListos(proceso)
					break
				}
			}
			time.Sleep(d.tick)
		}
	}
}

/*
func (d *Dispatcher) estaBloqueado(p *Process) bool {
	for _, bloqueado := range d.colaBloqueados {
		if bloqueado.Nombre == p.Nombre {
			return true
		}
	}
	return false
}*/

func (d *Dispatcher) descontarTiempoBloqueados() {
	if len(d.colaBloqueados) == 0 {
		return
	}

	for i := 0; i < len(d.colaBloqueados); i++ {
		d.colaBloqueados[i].Tiempo_ES--
		if d.colaBloqueados[i].Tiempo_ES <= 0 {
			fmt.Println("DESBLOQUEADO ->", d.colaBloqueados[i].Nombre)
			d.colaBloqueados[i].Estado = "Listo"

			// Cuando desbloqueamos el proceso, lo volvemos a poner en cola de listos
			p := d.colaBloqueados[i]
			d.PushProcessListos(p)
			d.colaBloqueados = append(d.colaBloqueados[:i], d.colaBloqueados[i+1:]...)
			i--

			// NUEVO: Crear nuevos canales y relanzar la goroutine del proceso desbloqueado
			proch := &CanalProcesos{
				proceso:      p,
				comandosChan: make(chan string),
				statusChan:   make(chan string),
			}
			d.canalesProceso[p.Nombre] = proch
			go p.arrancar(proch.comandosChan, proch.statusChan)
		}
	}
}
