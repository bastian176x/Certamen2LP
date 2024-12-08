package main

import "fmt"

func main() {
	//Por ahora solo son pruebas
	//../input/process_1.txt

	//crear dispatcher
	d := Dispatcher{maxInstructions: 10}

	//crear proceso
	p := Process{
		ID:              1,
		Program_counter: 0,
		Estado:          "listo",
	}

	//crear canal
	ch := make(chan ProcessCreation)

	go p.OrdenProcesos("order", &d, ch)

	for valor := range ch {
		fmt.Println(valor)
	}

}
