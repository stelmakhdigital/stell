package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8088", "listen address")
	results := flag.String("results", "eval/results/results.json", "path to latest results.json")
	static := flag.String("static", "dashboard", "dashboard static directory")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*static)))
	mux.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(*results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"file": filepath.Base(*results),
			"at":   time.Now().UTC(),
		})
	})
	log.Printf("dashboard on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
