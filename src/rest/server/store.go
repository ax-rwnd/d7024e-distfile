package main

import (
    "net/http"
	"github.com/gorilla/mux"
)

func storeHandler(w http.ResponseWriter, r *http.Request) {
	req := mux.Vars(r) 
	file := req["file"]
	if r.Method != "POST" {
		sendResponse(w, 400, "")
		return
	} 
    sendResponse(w, 200, file)
}