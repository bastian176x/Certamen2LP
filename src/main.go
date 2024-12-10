package main

import (
	"fmt"
	"time"
)

func main() {

	instruccionesMaximas, _ := recibir_parametros()

	//crear dispatcher
	d := Dispatcher{maxInstructions: instruccionesMaximas}

	//crear proceso
	p := Process{}

	ch := make(chan ProcessCreation)

	go func() {
		p.OrdenProcesos("order", ch)
		close(ch)
	}()

	for pc := range ch {
		go p.IniciarProceso(&pc)
		nuevosProcesos := p.IniciarProceso(&pc)
		for i := range nuevosProcesos {
			fmt.Println("PUSH LISTO ->", nuevosProcesos[i].Nombre, "CREADO EN ->", pc.Tiempo, "ms")
			nuevosProcesos[i].cargarInstrucciones(nuevosProcesos[i].Nombre)
			fmt.Println("INSTRUCCIONES ->", nuevosProcesos[i].Instrucciones)
			d.PushProcessListos(&nuevosProcesos[i])
		}
	}

	d.gestionarProcesos()
	time.Sleep(3 * time.Second)

}
