package main

import (
	"flag"
	"encoding/json"
	"os"
	"fmt"
	"log"
	"net/http"
)

func getRoutes() {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/healthcheck", healthHandler)
	http.HandleFunc("/api", apiHandler)
}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func defaultHandler(w http.ResponseWriter, req *http.Request) {
    log.Printf("[%s] - %s %s\n", req.RemoteAddr, req.Method, req.URL)


	_, err := fmt.Fprintf(w, "Default handler, hello there")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fprintf: %v\n", err)
	}
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
    log.Printf("[%s] - %s %s\n", req.RemoteAddr, req.Method, req.URL)

	w.Header().Add("Content-Type", "application/json")

	resp, err := json.Marshal(map[string]string{
	        "status": "ok",
    })
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fprintf: %v\n", err)
	}

    w.Write(resp)
}
func apiHandler(w http.ResponseWriter, req *http.Request) {
    log.Printf("[%s] - %s %s\n", req.RemoteAddr, req.Method, req.URL)

	w.Header().Add("Content-Type", "application/json")

	resp, err := json.Marshal(map[string]string{
	        "status": "ok",
    		"ip": GetIP(req),
    })
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fprintf: %v\n", err)
	}

    w.Write(resp)
}

func main() {
	var listen = flag.String("listen", "localhost:8080", "a webserver listen to address")
	flag.Parse()

	getRoutes()


	fmt.Printf("Starting web server at %s\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}