package main

import (
	"log"

	"github.com/kraconst/GO_server/internal/config"
	"github.com/kraconst/GO_server/internal/handler"
)

func main() {
	cfg := config.Must()
	h := handler.New()
	log.Fatal(h.Run(cfg.Addr))
}
