package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Process struct {
	Nombre          string   // nombre del proceso
	Estado          string   // listo, bloqueado, ejecutando
	Program_counter int      // contador de programa
	Instrucciones   []string // estado del proceso (instrucciones)
	Tiempo_ES       int      // tiempo de espera antes de ser desbloqueado
}

// Estructura para la creación de procesos
type ProcessCreation struct {
	Procesos []string
	Tiempo   int
}

// arrancar el proceso
func (p *Process) arrancar(cmdns <-chan string, statusCanal chan<- string, probabilidadCierre int) {

	for {
		comandos, ok := <-cmdns
		if !ok {
			// Si el canal se cerró desde el dispatcher, el proceso termina
			return
		}
		if comandos == "EXECUTE" {
			instruccion := p.ejecutarInstrucciones()
			if rand.Intn(probabilidadCierre) == 0 {
				fmt.Fprintln(out, p.Nombre, "Cerrado por causa desconocida")
				statusCanal <- "FINISHED"
				return
			}
			fmt.Fprintln(out, p.Nombre, instruccion, "Numero de instruccion ->", p.Program_counter)
			if p.Program_counter == len(p.Instrucciones) || instruccion == "F" {
				statusCanal <- "FINISHED"
				return
			}

			re := regexp.MustCompile(`ES\s+(\d+)`)
			match := re.FindStringSubmatch(instruccion)

			if match != nil {
				n, _ := strconv.Atoi(match[1])
				statusCanal <- "BLOCKED:" + strconv.Itoa(n)
				return
			} else {
				statusCanal <- "EXECUTING"
			}
		}
	}
}

// Carga las instrucciones de un proceso desde su archivo
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
		if err == io.EOF {
			break
		}
	}
}

func (p *Process) RestaurarEstado(estado *BCP) {
	p.Program_counter = estado.Program_counter
	p.Estado = estado.Estado
	p.Tiempo_ES = estado.Tiempo_ES
	p.Instrucciones = estado.Instrucciones
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
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				break
			}
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
