package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"groundops/internal/console"
	"groundops/internal/sla"
	"groundops/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	data := flag.String("data", "./data", "data directory")
	flag.Parse()

	state := store.NewState(*data)
	handler := console.NewAPI(state)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if _, err := sla.Tick(state, time.Now().UTC()); err != nil {
				log.Printf("sla tick: %v", err)
			}
		}
	}()

	log.Printf("GroundOps listening on %s with data dir %s", *addr, *data)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal(err)
	}
}
