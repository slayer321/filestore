package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// type Server struct{}

// func (s *Server) fooHandler(w http.ResponseWriter, r *http.Request) {

// }
const (
	fileDir = "filestoredir"
)

func addFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside the addfiles")
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form %v", err), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]
	fmt.Println("just before the file")
	fmt.Printf("length is %v\n", len(files))

	for _, fileHeader := range files {
		fmt.Println("Inside the loop")
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
		}
		defer file.Close()

		dstFile, err := os.Create(filepath.Join(fileDir, filepath.Base(fileHeader.Filename)))
		if err != nil {
			http.Error(w, "Failed to store file", http.StatusInternalServerError)
			log.Fatalf("Failed to create the file on server %v", err)
		}

		defer dstFile.Close()

		fmt.Printf("dstFile %v", dstFile)
		_, err = io.Copy(dstFile, file)
		if err != nil {
			http.Error(w, "Failed to Copy file to dst", http.StatusInternalServerError)
			log.Fatalf("Failed to Copy file to dst %v", err)
		}

		fmt.Printf("Filename %s", fileHeader.Filename)

		fmt.Fprintf(w, "File %s uploaded successfully\n", fileHeader.Filename)

	}
}

func listFiles(w http.ResponseWriter, _ *http.Request) {
	log.Println("Inside the listFiles")
	files, err := os.ReadDir("./filestoredir")
	if err != nil {
		http.Error(w, "Unable to read the Dir", http.StatusInternalServerError)
	}

	for _, file := range files {
		fmt.Fprintln(w, "", file)
		fmt.Printf("Printing the file name %v\n", file)

	}
}

func main() {

	err := os.Mkdir(fileDir, os.ModePerm)

	if err != nil {
		log.Fatalf("Failed to create the filestoredir %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/store", addFile)
	mux.HandleFunc("/list", listFiles)

	fmt.Println("Starting the server at port 8090")
	http.ListenAndServe(":8090", mux)
}
