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
