package controllers

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func SiteInfo(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	vars := mux.Vars(r)
	w.Write([]byte(fmt.Sprintf("Info, %s", vars["name"])))
}
