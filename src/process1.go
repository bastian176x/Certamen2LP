package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Process struct {
	ID              int    // id del proceso
	Estado          string // listo, bloqueado, ejecutando.
	Program_counter int    // contador de programa
	ESduracion      int    // duracion de la operacion de E/S
}

func (p *Process) finalizarProceso() {
	p.Estado = "finalizado"
}

func (p *Process) ejecutarProceso(archivo string, d Dispatcher) {
	arch, err := os.Open(archivo)
	if err != nil {
		fmt.Println(err)
	}
	defer arch.Close()

	reader := bufio.NewReader(arch)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		if strings.Contains(line, "#") {
			continue
		}
		if strings.Contains(line, "I") {
			fmt.Println("Instrucción")
			p.Program_counter++
		}
		if strings.Contains(line, "ES") {
			fmt.Println("E/S")
			p.Program_counter++
			d.addProcessBloqueados(*p)
		}
		if strings.Contains(line, "F") {

		}
	}
}
