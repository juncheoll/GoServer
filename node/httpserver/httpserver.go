package httpserver

import (
	"fmt"
	"net/http"

	"node/handler"

	"github.com/gorilla/mux"
)

var Port string

func Run() {

	r := mux.NewRouter()

	r.HandleFunc("/", handler.IndexHandler).Methods("GET")
	r.HandleFunc("/upload/", handler.UploadHandler).Methods("POST")
	r.HandleFunc("/download/{fileName}", handler.DownloadHandler).Methods("GET")

	fmt.Printf("HTTP Server running : %s\n", Port)
	http.Handle("/", r)
	err := http.ListenAndServe(":"+Port, nil)
	if err != nil {
		fmt.Printf("HTTP server Error : %s\n", err)
		return
	}
}
