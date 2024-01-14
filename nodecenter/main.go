package main

import (
	"fmt"
	"net/http"

	"nodecenter/handler"

	"github.com/gorilla/mux"
)

var nodeList []string

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newNodePort := handler.NodeListHandler(w, r, nodeList)
		nodeList = append(nodeList, newNodePort)
	}).Methods("GET")

	fmt.Printf("HTTP Server runnig : %s\n", "8080")
	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("HTTP server Error : %s\n", err)
		return
	}
}
