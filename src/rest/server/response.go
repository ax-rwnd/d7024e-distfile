package main

import (
    "net/http"
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


func sendResponse(w http.ResponseWriter, statusCode int, message string) {

	w.WriteHeader(statusCode)
    w.Write([]byte(message))

}
