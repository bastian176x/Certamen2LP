package main

import (
	"fmt"
	"regexp"
	"strconv"
)

type Dispatcher struct {
	colaListos      []Process
	colaBloqueados  []Process
	maxInstructions int
}

// Agrergar un proceso a la cola de listos
func (d *Dispatcher) PushProcessListos(p *Process) {
	d.colaListos = append(d.colaListos, *p)
}

// Agregar un proceso a la cola de bloqueados
func (d *Dispatcher) addProcessBloqueados(p *Process) {
	d.colaBloqueados = append(d.colaBloqueados, *p)
}

func (d *Dispatcher) gestionarProcesos(p *Process) {
	for {

		if len(d.colaListos) == 0 && len(d.colaBloqueados) > 0 {
			d.descontarTiempoBloqueados()
			continue
		} else if len(d.colaListos) == 0 && len(d.colaBloqueados) == 0 {
			break
		}
		proceso := d.colaListos[0]

		var instruccionFinal string
		//bucle interno para ejecutar las instrucciones del proceso
		fmt.Println("PULL Dispatcher")
		d.descontarTiempoBloqueados()
		fmt.Println("LOAD ->", proceso.Nombre)
		d.descontarTiempoBloqueados()
		fmt.Println("EXECUTE ->", proceso.Nombre)
		d.descontarTiempoBloqueados()
		for {
			instruccion := proceso.ejecutarInstrucciones()
			fmt.Println(proceso.Nombre, instruccion, "Numero de instruccion ->", proceso.Program_counter)
			d.descontarTiempoBloqueados()
			if instruccion == "F" {
				fmt.Println("Proceso finalizado ->", proceso.Nombre)
				instruccionFinal = "F"
				break
			}

			re := regexp.MustCompile(`ES\s+(\d+)`) // Capturar "ES n" y extraer n
			match := re.FindStringSubmatch(instruccion)

			if match != nil {
				fmt.Println("STORE ->", proceso.Nombre)
				fmt.Println("PUSH BLOQUEADO ->", proceso.Nombre)
				n, _ := strconv.Atoi(match[1])
				proceso.Tiempo_ES = n
				d.addProcessBloqueados(&proceso)
				break

			}

			if proceso.Program_counter%d.maxInstructions == 0 && len(d.colaListos) > 1 {
				fmt.Println("CAMBIO DE PROCESO")
				instruccionFinal = "CICLO"
				break
			}
		}

		//quitar el proceso de la cabeza de la cola
		d.colaListos = d.colaListos[1:]

		if instruccionFinal != "F" {
			if !d.estaBloqueado(&proceso) {
				fmt.Println("STORE ->", proceso.Nombre)
				d.descontarTiempoBloqueados()
				fmt.Println("PUSH LISTO ->", proceso.Nombre)
				d.descontarTiempoBloqueados()
				d.PushProcessListos(&proceso)
			}

		}
	}
}

func (d *Dispatcher) estaBloqueado(p *Process) bool {
	for _, bloqueado := range d.colaBloqueados {
		if bloqueado.Nombre == p.Nombre {
			return true
		}
	}
	return false
}

func (d *Dispatcher) descontarTiempoBloqueados() {
	//si no hay bloqueados, no hace nada
	if len(d.colaBloqueados) == 0 {
		return
	}

	//descontar el tiempo de todos los bloqueados

	for i := 0; i < len(d.colaBloqueados); i++ {
		d.colaBloqueados[i].Tiempo_ES--
		if d.colaBloqueados[i].Tiempo_ES <= 0 {
			fmt.Println("DESBLOQUEADO")
			d.colaBloqueados[i].Estado = "Listo"
			d.PushProcessListos(&d.colaBloqueados[i])
			d.colaBloqueados = append(d.colaBloqueados[:i], d.colaBloqueados[i+1:]...)
			i--
		}
	}

}
