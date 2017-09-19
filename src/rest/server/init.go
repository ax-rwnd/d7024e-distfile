package main

import (
	"net/http"
	"github.com/gorilla/mux"
)

func initialize() {
	router := mux.NewRouter()	
	router.HandleFunc("/cat/{hash}", catHandler) // cat.go
	router.HandleFunc("/store", storeHandler) // store.go
	router.HandleFunc("/pin/{hash}", pinHandler) // pin.go
	router.HandleFunc("/unpin/{hash}", unpinHandler) // unpin.go	
    http.ListenAndServe(":8080", router) // fix so take port from config file
	// could use log.Fatal here, prints the error but then uses os.exit
}