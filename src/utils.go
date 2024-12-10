package main

import (
	"fmt"
	"os"
	"strconv"
)

func recibir_parametros() (int, int) {
	args := os.Args
	if len(args) < 3 {
		fmt.Println("Faltan argumentos")
		return 0, 0
	}

	instruccionesMaximas, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Error al convertir el argumento 1")
		return 0, 0
	}
	probabilidadCierre, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("Error al convertir el argumento 2")
		return 0, 0
	}
	return instruccionesMaximas, probabilidadCierre
}
