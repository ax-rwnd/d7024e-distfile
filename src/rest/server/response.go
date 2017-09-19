package main

import (
    "net/http"
	"encoding/json"
)

/*
HTTP standard status codes
200 : OK
201 : Created
202 : Request accepted but not completed, may or may not use
204 : Request successfull but no content is returning
400 : Bad request, client send wrong syntax ex
404 : Not found
*/

type Response struct {
	Code int 	`json: message` // standard HTTP status code, dont' know it should use message like OK or something instead
	Data	string	`json: "data"`
}

func sendResponse(w http.ResponseWriter, code int, data string) {
	response := Response{Code: code, Data: data}	
	json.NewEncoder(w).Encode(response)	
}
