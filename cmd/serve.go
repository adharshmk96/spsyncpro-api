/*
Copyright © 2025 Adharsh Manikandan <debugslayer@gmail.com>
*/
package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"go_starter_api/infra"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the go_starter api",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.Fatalf("error getting port: %v", err)
			return
		}

		logger := logrus.New()

		shutdown, err := infra.SetupOtelSDK(context.Background())
		if err != nil {
			log.Printf("error setting up otel sdk: %v", err)
			return
		}
		defer shutdown(context.Background())

		config := infra.Config{
			Port: port,
		}

		db := infra.InitGormDB()

		srv := infra.NewServer(db, logger, config)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("error serving the api: %v", err)
			}
		}()

		log.Println("api running on port", port)

		// block until the signal is received
		<-ch
		log.Println("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("error shutting down server: %v", err)
		}

		log.Println("server shutdown...")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// flag to set the port
	serveCmd.Flags().IntP("port", "p", 8080, "port to serve the api")
}
