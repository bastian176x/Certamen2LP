# Nombre del binario que se generará
BINARY := programa
SRC_DIR := src
GO_FILES := dispatcher.go process1.go utils.go bcp.go main.go

.PHONY: all build run clean

all: build run

build:
	cd $(SRC_DIR) && go build -o ../$(BINARY) $(GO_FILES)

run:
	./$(BINARY) 2 50 order trace

clean:
	rm -f $(BINARY)
