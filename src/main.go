package main

import (
	"fmt"
	"os"
)

func main() {
	//Por ahora solo son pruebas
	//../input/process_1.txt
	// abrir archivo
	arch, err := os.Open("../input/process_1.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer arch.Close()

	//crear dispatcher
	d := Dispatcher{}
	//crear proceso
	p := Process{
		ID:              1,
		Estado:          "listo",
		Program_counter: 0,
		ESduracion:      0,
	}
	//ejecutar proceso
	p.ejecutarProceso(arch.Name(), d)

}
