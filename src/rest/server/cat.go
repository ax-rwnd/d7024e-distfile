package main

import (
    "net/http"
	"github.com/gorilla/mux"
	"fmt"
)

func catHandler(w http.ResponseWriter, r *http.Request) {
	req := mux.Vars(r)
	hash := req["hash"]
	
	if r.Method != "GET" {
		sendResponse(w, http.StatusBadRequest, "400 - Not a GET request")
		return
	}
	
	fmt.Println(hash)
	
    sendResponse(w, http.StatusOK, "200 - OK ")
}