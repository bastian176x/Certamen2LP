package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Process struct {
	ID              int    // id del proceso
	estado          string // listo, bloqueado, ejecutando.
	program_counter int    // contador de programa
	ESduracion      int    // duracion de la operacion de E/S
}

func (p *Process) finalizarProceso() {
	p.estado = "finalizado"
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
			p.program_counter++
		}
		if strings.Contains(line, "ES") {
			p.program_counter++
			d.addProcessBloqueados(*p)
		}
		if strings.Contains(line, "F") {

		}
	}
}
