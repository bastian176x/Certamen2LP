package main

type Dispatcher struct {
	colaListos      []Process
	colaBloqueados  []Process
	maxInstructions int
}

// Agrergar un proceso a la cola de listos
func (d *Dispatcher) addProcessListos(p Process) {

	d.colaListos = append(d.colaListos, p)
}

// Agregar un proceso a la cola de bloqueados
func (d *Dispatcher) addProcessBloqueados(p Process) {
	d.colaBloqueados = append(d.colaBloqueados, p)
}

func (d *Dispatcher) gestionarProcesos() {
	for {
		if len(d.colaListos) > 0 {
			proceso := d.colaListos[0]
			d.colaListos = d.colaListos[1:]
			proceso.ejecutarProceso(proceso.Nombre, d)
		}
	}
}
