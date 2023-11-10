package main

import (
	"log"
	"net/http"

	"github.com/HeavenAQ/app"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	app := app.NewApp()
	http.HandleFunc("/callback", app.HandleCallback)

	if err := http.ListenAndServe(":"+"3000", nil); err != nil {
		log.Fatal(err)
	}
}
