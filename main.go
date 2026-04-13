package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/EnriqueBravo115/crawler/handler"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/{vin}", handler.GetEngines)

	fmt.Println("We are up and running. localhost:8000/{vin}")
	http.ListenAndServe(":8000", router)
}
