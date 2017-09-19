package main

import (
    "net/http"
	//"github.com/gorilla/mux"
	//"fmt"
	"io/ioutil"
	"log"
	//"crypto/sha1"
	//"bytes"
)

// calculate the hash with SHA-1

func storeHandler(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != "POST" {
		sendResponse(w, http.StatusBadRequest, "400 - Not a POST request")
		return
	}
	
	data, err:= ioutil.ReadAll(r.Body) // here, get the data in byte array
	
	//kademlia.Store(data)
	
	//hash := sha1.Sum(data)
	
	if err != nil {
		log.Fatal(err)
		sendResponse(w, http.StatusInternalServerError, "500 - Couldn't read body")
	}
	
	    
	sendResponse(w, http.StatusOK, string(data))
	
}