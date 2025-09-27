/*
Copyright © 2025 Adharsh Manikandan <debugslayer@gmail.com>
*/
package main

import (
	"log"
	"go_starter_api/cmd"
	_ "go_starter_api/docs"

	"github.com/joho/godotenv"
)

// @title			go_starter API
// @version		1.0
// @description	This is the API for the go_starter platform.
// @host			localhost:8080
// @BasePath		/
// @schemes		http
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not loaded... ignoring")
	}

	cmd.Execute()
}
