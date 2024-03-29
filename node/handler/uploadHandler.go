package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"node/database"
	"node/tcpserver"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	uploadType := r.FormValue("uploadType")

	if uploadType == "single" {
		SingleUploadHandler(w, r)
	} else if uploadType == "entire" {
		EntireUploadHandler(w, r)
	} else {
		http.Error(w, "Invalid upload type", http.StatusBadRequest)
		return
	}
}

func SingleUploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("singleupload 호출")

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	//_, port, _ := net.SplitHostPort(r.Host)

	uploadDir := "./uploads/" + database.DbName + "/"
	os.MkdirAll(uploadDir, os.ModePerm)
	filePath := uploadDir + header.Filename
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	err = database.SaveFileToDB(header.Filename, filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func EntireUploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("entireupload 호출")

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads/" + database.DbName + "/"
	os.MkdirAll(uploadDir, os.ModePerm)
	filePath := uploadDir + header.Filename
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	err = database.SaveFileToDB(header.Filename, filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 다른 노드들에게 파일 전송
	_, err = file.Seek(0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tcpserver.SendFileToOtherNodes(file, header.Filename, filePath)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}
