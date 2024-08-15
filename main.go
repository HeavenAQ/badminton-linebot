package main

import (
	"log"
	"net/http"
	"os"

	"github.com/HeavenAQ/app"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		log.Println("Trying to load from system environment variables")
	}

	app := app.NewApp()
	http.HandleFunc("/callback", app.HandleCallback)

	app.InfoLogger.Println("\n\tServer started on port: http://localhost:" + os.Getenv("PORT"))
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
