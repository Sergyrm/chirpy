package main

import "net/http"

func main() {
	myServer := http.NewServeMux()
	myServer.Handle("/", http.FileServer(http.Dir(".")))

	server := &http.Server{
		Handler: myServer,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}