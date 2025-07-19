package main

import (
	"SomeTask/internal/app"
	"log"
)

func main() {
	log.Println("creating application")
	application := app.NewApp()
	log.Println("starting application")
	application.Run()
}
