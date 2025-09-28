/*
Copyright © 2025 Adharsh Manikandan <debugslayer@gmail.com>
*/
package main

import (
	"log"
	"spsyncpro_api/cmd"
	_ "spsyncpro_api/docs"

	"github.com/joho/godotenv"
)

// @title			spsyncpro API
// @version		1.0
// @description	This is the API for the spsyncpro platform.
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
