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
		log.Println("Error loading .env file:", err)
		log.Println("Using system environment variables.")
	}

	app := app.NewApp()
	http.HandleFunc("/callback", app.HandleCallback)

	app.InfoLogger.Println("\n\tServer started on port 3000")
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
