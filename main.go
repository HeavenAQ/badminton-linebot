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

	app.InfoLogger.Println("\n\tServer started on port 3000: http://localhost:3000")
	if err := http.ListenAndServe(":"+"3000", nil); err != nil {
		log.Fatal(err)
	}
}
