package main

import (
    "net/http"
	"github.com/gorilla/mux"
)

func unpinHandler(w http.ResponseWriter, r *http.Request) {
	req := mux.Vars(r) 
	hash := req["hash"]
	if r.Method != "PUT" {
		sendResponse(w, 400, "")
		return
	}
    sendResponse(w, 200, hash)
}