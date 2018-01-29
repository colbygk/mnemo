package controllers

import (
	"encoding/json"
	"net/http"
	"okea/services/models"
)

func StatusController(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	status := models.OkeaStatus{Okea: "ok", DNS: "ok"}

	bytes, e := json.Marshal(status)
	if e != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
