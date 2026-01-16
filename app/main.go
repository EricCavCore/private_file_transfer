package main

import (
	"log"
	"net/http"
	"os"
)

func health_check(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Alive!"))
}

func main() {

	// Ensure enc folder exists
	os.Mkdir(folder, 0700)

	// Create the HTTP server
	r := http.NewServeMux()

	r.Handle("/", http.FileServer(http.Dir("./www")))

	r.HandleFunc("POST /", post_file)
	r.HandleFunc("GET /", get_file)
	r.HandleFunc("DELETE /", delete_file)

	r.HandleFunc("GET /api/health", health_check)

	addr := ":8080"
	log.Printf("Started listening on %s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Web server crashed: %s\n", err.Error())
	}
}
