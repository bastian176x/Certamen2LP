package main

import (
	"time"
)

func main() {
	//Por ahora solo son pruebas
	//../input/process_1.txt

	//crear dispatcher
	d := Dispatcher{maxInstructions: 2}

	//crear proceso
	p := Process{}

	ch := make(chan ProcessCreation)

	go func() {
		p.OrdenProcesos("order", &d, ch)
		close(ch)
	}()

	for pc := range ch {
		go p.IniciarProceso(&pc, &d)
	}
	d.gestionarProcesos()
	time.Sleep(3 * time.Second)

}
