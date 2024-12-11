package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

var out io.Writer // Variable global para escribir la traza

func main() {

	// Abrir el archivo de traza
	f, err := os.Create("../output/trace.txt")
	if err != nil {
		fmt.Println("Error al crear archivo de traza:", err)
		return
	}
	defer f.Close()

	// Crear un MultiWriter para escribir a consola y al archivo
	out = io.MultiWriter(os.Stdout, f)

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
			fmt.Fprintln(out, "PUSH LISTO ->", nuevosProcesos[i].Nombre, "CREADO EN ->", pc.Tiempo, "ms")
			nuevosProcesos[i].cargarInstrucciones(nuevosProcesos[i].Nombre)
			fmt.Fprintln(out, "INSTRUCCIONES ->", nuevosProcesos[i].Instrucciones)
			d.PushProcessListos(&nuevosProcesos[i])
		}
	}

	d.gestionarProcesos()
	time.Sleep(3 * time.Second)
}
