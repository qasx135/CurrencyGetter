package main

import (
	"SomeTask/internal/app"
	"fmt"
	"log"
)

func main() {
	log.Println("creating application")
	application := app.NewApp()
	log.Println("starting application")
	application.Run()
	fmt.Println("Программа завершена. Нажмите Enter для выхода...")
	fmt.Scanln()
}
