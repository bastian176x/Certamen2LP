package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Process struct {
	Nombre          string   // nombre del proceso
	Estado          string   // listo, bloqueado, ejecutando.
	Program_counter int      // contador de programa
	Instrucciones   []string //ayuda a guardar el estado del proceso y reanudarlo con el program counter
	Tiempo_ES       int      // tiempo de espera antes de ser desbloqueado
}

// Estructura para la creación de procesos, se lee el archivo order y el tiempo de creación
type ProcessCreation struct {
	Procesos []string
	Tiempo   int
}

func (p *Process) finalizarProceso() {
	p.Estado = "finalizado"
}

// para leer las instrucciones de un proceso usando el archivo process_n.txt
func (p *Process) cargarInstrucciones(archivo string) {
	arch, err := os.Open(fmt.Sprintf("../input/%s.txt", archivo))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer arch.Close()

	reader := bufio.NewReader(arch)
	re := regexp.MustCompile(`ES\s+\d+`)
	for {
		line, err := reader.ReadString('\n')
		// Si el error es distinto de nil y no es EOF, se imprime y se sale
		if err != nil && err != io.EOF {
			fmt.Println(err)
			break
		}

		if strings.Contains(line, "#") {
			if err == io.EOF {
				break
			}
			continue
		}
		if strings.Contains(line, "I") {

			p.Instrucciones = append(p.Instrucciones, "I")
		}
		if match := re.FindString(line); match != "" {
			p.Instrucciones = append(p.Instrucciones, match)
		}
		if strings.Contains(line, "F") {
			p.Instrucciones = append(p.Instrucciones, "F")
			break
		}
		//Si el error es EOF y no se encontró 'F', se termina el bucle
		if err == io.EOF {
			break
		}
	}
}

func (p *Process) ejecutarInstrucciones() string {
	instruccion := p.Instrucciones[p.Program_counter]
	p.Program_counter++
	return instruccion
}

func (p *Process) OrdenProcesos(archivo string, canal_procesos chan ProcessCreation) {
	re := regexp.MustCompile(`\d+`)
	re2 := regexp.MustCompile(`process_\d+`)

	orden, err := os.Open(fmt.Sprintf("../input/%s.txt", archivo))
	if err != nil {
		fmt.Println(err)
	}
	defer orden.Close()

	reader := bufio.NewReader(orden)
	for {
		line, err := reader.ReadString('\n')

		match := re.FindString(line)
		if match == "" {
			continue
		}

		num, _ := strconv.Atoi(match)
		procesos_nombres := re2.FindAllString(line, -1)

		canal_procesos <- ProcessCreation{Procesos: procesos_nombres, Tiempo: num}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			break
		}

	}
}

func (p *Process) IniciarProceso(pc *ProcessCreation) []Process {
	time.Sleep(time.Duration(pc.Tiempo) * time.Millisecond)
	var nuevosProcesos []Process
	for _, nombreProceso := range pc.Procesos {
		nuevoProceso := Process{
			Nombre:          nombreProceso,
			Estado:          "Listo",
			Program_counter: 0,
		}
		nuevosProcesos = append(nuevosProcesos, nuevoProceso)
	}
	return nuevosProcesos
}
