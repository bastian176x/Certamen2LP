package main

import "fmt"

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
		if len(d.colaListos) == 0 {
			break
		}
		proceso := d.colaListos[0]

		var instruccionFinal string
		//bucle interno para ejecutar las instrucciones del proceso
		for {
			instruccion := proceso.ejecutarInstrucciones()

			if instruccion == "F" {
				fmt.Println("Proceso finalizado ->", proceso.Nombre)
				instruccionFinal = "F"
				break
			}

			if proceso.Program_counter%d.maxInstructions == 0 {
				fmt.Println("CAMBIO DE PROCESO")
				instruccionFinal = "CICLO"
				break
			}
		}

		//quitar el proceso de la cabeza de la cola
		d.colaListos = d.colaListos[1:]
		if instruccionFinal != "F" {
			d.PushProcessListos(&proceso)
		}
	}
}
