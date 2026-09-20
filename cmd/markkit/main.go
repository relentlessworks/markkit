package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/relentlessworks/markkit/internal/api"
	"github.com/relentlessworks/markkit/internal/config"
)

func main() {
	cfg := config.Load()
	handler := api.New(cfg.Secret)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: handler.Routes(),
	}

	fmt.Fprintf(log.Default().Writer(), "markkit listening on %s\n", cfg.Addr)
	log.Fatal(server.ListenAndServe())
}
