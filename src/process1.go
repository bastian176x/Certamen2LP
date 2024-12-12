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
	Nombre        string
	Instrucciones []string
}

type ProcessCreation struct {
	Procesos []string
	Tiempo   int
}

// arrancar el proceso con su BCP
func (p *Process) arrancar(cmdns <-chan string, statusCanal chan<- string, probabilidadCierre int, bcp *BCP, logger io.Writer) {
	rand.Seed(time.Now().UnixNano())
	reES := regexp.MustCompile(`ES\s+(\d+)`)

	for {
		comando, ok := <-cmdns
		if !ok {
			// Canal cerrado desde dispatcher
			return
		}
		if comando == "EXECUTE" {
			instruccion := p.ejecutarInstrucciones(bcp)
			if probabilidadCierre > 0 && rand.Intn(probabilidadCierre) == 0 {
				fmt.Fprintln(logger, p.Nombre, "Cerrado por causa desconocida")
				statusCanal <- "FINISHED"
				return
			}
			fmt.Fprintln(logger, p.Nombre, instruccion, "Numero de instruccion ->", bcp.Program_counter)
			if bcp.Program_counter == len(p.Instrucciones) || instruccion == "F" {
				statusCanal <- "FINISHED"
				return
			}

			match := reES.FindStringSubmatch(instruccion)
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

// cargarInstrucciones carga las instrucciones de un archivo asociado al proceso
func (p *Process) cargarInstrucciones(logger io.Writer) error {
	filePath := fmt.Sprintf("input/%s.txt", p.Nombre)
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error abriendo archivo de instrucciones %s: %v", p.Nombre, err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	reES := regexp.MustCompile(`ES\s+\d+`)

	for {
		line, err := reader.ReadString('\n')
		// Si llega EOF y line no está vacía, aún así debemos procesarla
		if err == io.EOF {
			if len(line) > 0 {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "#") {
					// Es una línea de comentario, no hacemos nada
				} else {
					// Procesamos la línea como si fuera una línea normal
					if strings.Contains(line, "I") {
						p.Instrucciones = append(p.Instrucciones, "I")
					}
					if match := reES.FindString(line); match != "" {
						p.Instrucciones = append(p.Instrucciones, match)
					}
					if strings.Contains(line, "F") {
						p.Instrucciones = append(p.Instrucciones, "F")
					}
				}
			}
			// Salimos del for una vez procesada la última línea
			break
		}
		if err != nil {
			return fmt.Errorf("error leyendo instrucciones de %s: %v", p.Nombre, err)
		}

		line = strings.TrimSpace(line)
		// Ignorar comentarios
		if strings.Contains(line, "#") {
			continue
		}

		if strings.Contains(line, "I") {
			p.Instrucciones = append(p.Instrucciones, "I")
		}
		if match := reES.FindString(line); match != "" {
			p.Instrucciones = append(p.Instrucciones, match)
		}
		if strings.Contains(line, "F") {
			p.Instrucciones = append(p.Instrucciones, "F")
			// Si encontramos F, el proceso termina aquí
			break
		}
	}

	return nil
}

// Ejecutar la siguiente instrucción usando el Program_counter del BCP
func (p *Process) ejecutarInstrucciones(bcp *BCP) string {
	if bcp.Program_counter < len(p.Instrucciones) {
		instruccion := p.Instrucciones[bcp.Program_counter]
		bcp.Program_counter++
		return instruccion
	}
	return "F"
}

// OrdenProcesos lee las órdenes de creación de procesos
func (p *Process) OrdenProcesos(archivo string, canal_procesos chan ProcessCreation, logger io.Writer) error {
	filePath := fmt.Sprintf("input/%s.txt", archivo)
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error abriendo archivo de órdenes %s: %v", archivo, err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	re := regexp.MustCompile(`^\d+`)         // Coincidir solo números al inicio de línea
	re2 := regexp.MustCompile(`process_\d+`) // Coincidir con nombres de procesos

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line) // Eliminar espacios extra
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			fmt.Fprintln(logger, "DEBUG: Línea ignorada:", line)
			if err == io.EOF {
				break
			}
			continue
		}

		// Buscar el tiempo (número al inicio)
		matchNum := re.FindString(line)
		if matchNum == "" {
			fmt.Fprintln(logger, "DEBUG: Línea sin tiempo válido:", line)
			continue
		}
		num, _ := strconv.Atoi(matchNum)

		// Buscar nombres de procesos
		procesosNombres := re2.FindAllString(line, -1)
		if len(procesosNombres) > 0 {
			fmt.Fprintln(logger, "DEBUG: Procesos encontrados en línea:", procesosNombres)
			canal_procesos <- ProcessCreation{Procesos: procesosNombres, Tiempo: num}
		} else {
			fmt.Fprintln(logger, "DEBUG: Sin procesos válidos en línea:", line)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error leyendo líneas de %s: %v", archivo, err)
		}
	}

	return nil
}

// IniciarProceso simula el inicio de procesos tras el tiempo indicado
func (p *Process) IniciarProceso(pc *ProcessCreation) []Process {
	time.Sleep(time.Duration(pc.Tiempo) * time.Millisecond)
	var nuevosProcesos []Process
	for _, nombreProceso := range pc.Procesos {
		nuevoProceso := Process{
			Nombre: nombreProceso,
		}
		nuevosProcesos = append(nuevosProcesos, nuevoProceso)
	}
	return nuevosProcesos
}
