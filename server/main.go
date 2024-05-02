package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	fileDir = "./filestoredir"
)

func addAndUpdateFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form %v", err), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			log.Fatalf("Failed to read file %v", err)
		}
		defer file.Close()

		dstFile, err := os.Create(filepath.Join(fileDir, filepath.Base(fileHeader.Filename)))
		if err != nil {
			http.Error(w, "Failed to store file", http.StatusInternalServerError)
			log.Fatalf("Failed to create the file on server %v", err)
		}

		defer dstFile.Close()

		_, err = io.Copy(dstFile, file)
		if err != nil {
			http.Error(w, "Failed to Copy file to dst", http.StatusInternalServerError)
			log.Fatalf("Failed to Copy file to dst %v", err)
		}
		log.Printf("Received the file %s and worked on it.\n", fileHeader.Filename)
		fmt.Fprint(w, fileHeader.Filename)

	}
}

func listFiles(w http.ResponseWriter, _ *http.Request) {
	files, err := os.ReadDir(fileDir)
	if err != nil {
		http.Error(w, "Unable to read the Dir", http.StatusInternalServerError)
	}

	for _, file := range files {
		fmt.Fprintln(w, "", file)
	}
}

func removeFile(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read the request body", http.StatusInternalServerError)
	}
	fileName := string(body)
	rootPath := fileDir + "/" + fileName

	err = os.Remove(rootPath)
	if err != nil {
		log.Printf("Not able to delete the file from the server %v", err)
		http.Error(w, "Not able to delete the file from the server", http.StatusInternalServerError)
	}
	log.Printf("Removed file %s from the server", fileName)
}

func wc(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form %v", err), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
		}
		defer file.Close()

		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanWords)

		var wordSlice []string

		for fileScanner.Scan() {
			wordSlice = append(wordSlice, fileScanner.Text())
		}

		log.Printf("Number of word count in the file name %s is %d", fileHeader.Filename, len(wordSlice))
		fmt.Fprintln(w, fmt.Sprintf(" %d , %s", len(wordSlice), fileHeader.Filename))
	}
}

func main() {

	err := os.Mkdir(fileDir, os.ModePerm)

	if err != nil {
		log.Fatalf("Failed to create the filestoredir %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/store", addAndUpdateFile)
	mux.HandleFunc("/list", listFiles)
	mux.HandleFunc("/rm", removeFile)
	mux.HandleFunc("/update", addAndUpdateFile)
	mux.HandleFunc("/wc", wc)

	fmt.Println("Starting the server at port 8090")
	http.ListenAndServe(":8090", mux)
}
