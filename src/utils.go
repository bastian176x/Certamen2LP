package main

import (
	"fmt"
	"os"
	"strconv"
)

// recibirParametrosMejorado: valida y devuelve los parámetros
func recibirParametros() (int, int, string, string, error) {
	args := os.Args
	if len(args) < 5 {
		return 0, 0, "", "", fmt.Errorf("faltan argumentos. Uso: <ejecutable> <instruccionesMaximas> <probCierre> <archivoOrden> <archivoSalida>")
	}
	instruccionesMaximas, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("error al convertir el argumento 1 a entero: %v", err)
	}
	probabilidadCierre, err := strconv.Atoi(args[2])
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("error al convertir el argumento 2 a entero: %v", err)
	}
	archivoOrden := args[3]
	archivoSalida := args[4]
	return instruccionesMaximas, probabilidadCierre, archivoOrden, archivoSalida, nil
}
