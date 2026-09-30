// Package main es el punto de entrada de api-go (Go + Fiber v3).
//
// Carga y valida la configuración, crea el adaptador hacia api-node, construye la app
// (internal/server) y levanta el servidor con apagado ordenado ante SIGINT/SIGTERM.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/elfab/retotecnico/api-go/internal/client"
	"github.com/elfab/retotecnico/api-go/internal/config"
	"github.com/elfab/retotecnico/api-go/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	// Desarrollo local: carga .env si existe. No sobrescribe variables ya definidas,
	// así que en Docker o en la nube mandan las variables del entorno.
	_ = godotenv.Load()

	cfg, err := config.FromEnv()
	if err != nil {
		slog.Error("configuración inválida", "error", err)
		os.Exit(1)
	}

	stats := client.NewStatisticsHTTPClient(cfg.NodeAPIURL, cfg.NodeAPITimeout)
	app := server.New(cfg, stats, server.Options{})

	// Apagado ordenado: termina las peticiones en curso antes de salir (p. ej. en un redeploy).
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Info("apagando api-go")
		_ = app.Shutdown()
	}()

	slog.Info("api-go iniciado", "port", cfg.Port, "nodeApiUrl", cfg.NodeAPIURL)
	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("el servidor se detuvo con error", "error", err)
		os.Exit(1)
	}
}
