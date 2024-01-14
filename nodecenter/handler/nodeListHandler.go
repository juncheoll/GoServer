package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func NodeListHandler(w http.ResponseWriter, r *http.Request, nodeList []string) string {
	nodePort := r.URL.Query().Get("port")
	fmt.Printf("노드(%s) 참가\n", nodePort)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeList)

	return nodePort
}
