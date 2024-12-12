package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

func main() {
	instruccionesMaximas, probabilidadCierre, archivoOrden, archivoSalida, err := recibirParametros()
	if err != nil {
		fmt.Println("Error en parámetros:", err)
		os.Exit(1)
	}

	err = os.MkdirAll("output", 0755)
	if err != nil {
		fmt.Println("Error creando directorio output:", err)
		os.Exit(1)
	}

	f, err := os.Create("output/" + archivoSalida + ".txt")
	if err != nil {
		fmt.Println("Error al crear archivo de traza:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Crear un MultiWriter para escribir a consola y al archivo
	logger := io.MultiWriter(os.Stdout, f)

	// Crear dispatcher con mutex y logger
	d := &Dispatcher{
		maxInstructions: instruccionesMaximas,
		tick:            50e6, // 50ms
		wg:              &sync.WaitGroup{},
		logger:          logger,
	}

	// Crear proceso "p" para leer las ordenes de creación
	p := &Process{}

	ch := make(chan ProcessCreation)

	// Leer las ordenes de creación
	go func() {
		if err := p.OrdenProcesos(archivoOrden, ch, logger); err != nil {
			fmt.Fprintln(logger, "Error leyendo órdenes:", err)
		}
		close(ch)
	}()

	// Crear procesos a partir de las ordenes
	for pc := range ch {
		nuevosProcesos := p.IniciarProceso(&pc)
		for i := range nuevosProcesos {
			fmt.Fprintln(logger, "PUSH LISTO ->", nuevosProcesos[i].Nombre, "CREADO EN ->", pc.Tiempo, "ms")
			if err := nuevosProcesos[i].cargarInstrucciones(logger); err != nil {
				fmt.Fprintln(logger, "Error al cargar instrucciones:", err)
				continue
			}
			fmt.Fprintln(logger, "INSTRUCCIONES ->", nuevosProcesos[i].Instrucciones)

			// Crear el BCP asociado al proceso
			nuevoBCP := &BCP{
				Nombre:          nuevosProcesos[i].Nombre,
				Estado:          "Listo",
				Program_counter: 0,
				Tiempo_ES:       0,
			}
			d.PushProcessListos(&nuevosProcesos[i], nuevoBCP)
		}
	}

	// Iniciar la gestión de procesos por el Dispatcher
	d.gestionarProcesos(probabilidadCierre)
	d.wg.Wait()
	fmt.Fprintln(logger, "Todos los procesos terminaron.")
}
