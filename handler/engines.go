package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/EnriqueBravo115/crawler/scraper"
	"github.com/gorilla/mux"
)

// NOTE:
// Orchestrates the three scrapers in cascade and returns the list of available engines for the given VIN.
func GetEngines(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vin := vars["vin"]

	if vin == "" {
		http.Error(w, "vin is required", http.StatusBadRequest)
		return
	}

	// NOTE: Step 1: Get the URL of the engines page for this VIN
	link, err := scraper.GetEngineLink(vin)
	if err != nil {
		log.Printf("error getting engine link for VIN %s: %v", vin, err)
		http.Error(w, "no engines found for this VIN", http.StatusNotFound)
		return
	}

	// NOTE: Step 2: Get the list of available engines
	engines, err := scraper.GetEngineList(link)
	if err != nil {
		log.Printf("error getting engine list: %v", err)
		http.Error(w, "error fetching engines", http.StatusInternalServerError)
		return
	}

	if len(engines) == 0 {
		http.Error(w, "no engines found", http.StatusNotFound)
		return
	}

	// NOTE: Step 3: Get detailed data for each engine in parallel
	engineData, err := scraper.GetEngineDetails(engines)
	if err != nil {
		log.Printf("error getting individual engine details: %v", err)
		http.Error(w, "error fetching engine details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(engineData); err != nil {
		log.Printf("error serializing response: %v", err)
	}
}
