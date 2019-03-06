package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name": "anuchito zz"}`))
	})

	log.Println("starting zz ...")
	log.Fatal(http.ListenAndServe(":1234", nil))
}
