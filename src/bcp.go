package main

// BCP: Bloque de Control de Proceso
type BCP struct {
	Nombre          string
	Estado          string
	Program_counter int
	Instrucciones   []string
	Tiempo_ES       int
}

// Convierte un Process en un BCP
func (p *Process) ToBCP() *BCP {
	return &BCP{
		Nombre:          p.Nombre,
		Estado:          p.Estado,
		Program_counter: p.Program_counter,
		Instrucciones:   p.Instrucciones,
		Tiempo_ES:       p.Tiempo_ES,
	}
}

// Restaura un Process desde un BCP
func FromBCP(b *BCP) *Process {
	return &Process{
		Nombre:          b.Nombre,
		Estado:          b.Estado,
		Program_counter: b.Program_counter,
		Instrucciones:   b.Instrucciones,
		Tiempo_ES:       b.Tiempo_ES,
	}
}
