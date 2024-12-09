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
	Nombre          string // nombre del proceso
	Estado          string // listo, bloqueado, ejecutando.
	Program_counter int    // contador de programa
	ESduracion      int    // duracion de la operacion de E/S
}

type ProcessCreation struct {
	Procesos []string
	Tiempo   int
}

func (p *Process) finalizarProceso() {
	p.Estado = "finalizado"
}

func (p *Process) ejecutarProceso(archivo string, d *Dispatcher) {
	arch, err := os.Open(fmt.Sprintf("../input/%s.txt", archivo))
	if err != nil {
		fmt.Println(err)
	}
	defer arch.Close()

	reader := bufio.NewReader(arch)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {

			if err == io.EOF {
				break
			}

			fmt.Println(err)
			break
		}
		/*
			if p.Program_counter < d.maxInstructions {
				d.addProcessListos(*p)
			}*/
		if strings.Contains(line, "#") {
			continue
		}
		if strings.Contains(line, "I") {
			fmt.Println("Instrucción ->", archivo)
			p.Program_counter++
		}
		if strings.Contains(line, "ES") {
			fmt.Println("E/S ->", archivo)
			p.Program_counter++
			d.addProcessBloqueados(*p)
		}
		if strings.Contains(line, "F") {
			fmt.Println("Fin ->", archivo)
			break
		}

	}
}

func (p *Process) OrdenProcesos(archivo string, d *Dispatcher, canal_procesos chan ProcessCreation) {
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

func (p *Process) IniciarProceso(pc *ProcessCreation, d *Dispatcher) {
	time.Sleep(time.Duration(pc.Tiempo) * time.Millisecond)
	fmt.Println("Iniciando proceso", pc.Procesos, "en", pc.Tiempo, "milisegundos")

	for _, nombreProceso := range pc.Procesos {
		nuevoProceso := Process{
			Nombre:          nombreProceso,
			Estado:          "listo",
			Program_counter: 0,
		}
		fmt.Println("PUSH LISTO ->", nuevoProceso.Nombre)
		d.addProcessListos(nuevoProceso)
	}
}
