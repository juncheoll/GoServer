package handler

import (
	"net/http"
	"text/template"

	"node/database"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	files, err := database.GetFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("C:/Users/th6re8e/OneDrive - 계명대학교/GoServer/node/templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 파일 목록 전달
	tmpl.Execute(w, files)
}
