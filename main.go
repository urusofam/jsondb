package main

import (
	"log"

	"github.com/urusofam/jsondb/config"
)

func main() {
	// Создать конфигурацию
	cfg := config.NewFileStorageConfig("./data", true)
	
	// Запустить интерактивный CLI
	cli := NewCLI(cfg)
	cli.Run()
}
