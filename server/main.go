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

// type Server struct{}

// func (s *Server) fooHandler(w http.ResponseWriter, r *http.Request) {

// }
const (
	fileDir = "filestoredir"
)

func addAndUpdateFile(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Inside the addfiles")
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form %v", err), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]
	//fmt.Println("just before the file")
	// fmt.Printf("length is %v\n", len(files))

	for _, fileHeader := range files {
		//fmt.Println("Inside the loop")
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

		// fmt.Printf("dstFile %v", dstFile)
		_, err = io.Copy(dstFile, file)
		if err != nil {
			http.Error(w, "Failed to Copy file to dst", http.StatusInternalServerError)
			log.Fatalf("Failed to Copy file to dst %v", err)
		}

		//fmt.Printf("Filename %s", fileHeader.Filename)
		log.Printf("File %s uploaded successfully \n", fileHeader.Filename)
		fmt.Fprint(w, fileHeader.Filename)

	}
}

func listFiles(w http.ResponseWriter, _ *http.Request) {
	//log.Println("Inside the listFiles")
	//var fileList []fs.DirEntry
	files, err := os.ReadDir("./filestoredir")
	if err != nil {
		http.Error(w, "Unable to read the Dir", http.StatusInternalServerError)
	}

	for _, file := range files {
		//fileList = append(fileList, file)
		fmt.Fprintln(w, "", file)

		//fmt.Printf("Printing the file name %v\n", file)
	}
}

func removeFile(w http.ResponseWriter, r *http.Request) {
	// files, err := os.ReadDir("./filestoredir")
	// if err != nil {
	// 	http.Error(w, "Unable to read the Dir", http.StatusInternalServerError)
	// }

	//var f string

	//err := json.NewDecoder(r.Body).Decode(&f)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to decode the request string", http.StatusInternalServerError)
	}
	mainBody := string(body)
	log.Printf("Decoded filename is %s\n", mainBody)
	rootPath := "./filestoredir/" + mainBody
	//filePath := filepath.Join(rootPath, mainBody)
	log.Printf("File path is %s", rootPath)
	err = os.Remove(rootPath)
	if err != nil {
		log.Printf("Not able to delete the file")
	}
}

func wc(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form %v", err), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]
	//fmt.Println("just before the file")
	// fmt.Printf("length is %v\n", len(files))

	for _, fileHeader := range files {
		//fmt.Println("Inside the loop")
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
		fmt.Fprintln(w, fmt.Sprintf("Number of word count in the file name %s is %d", fileHeader.Filename, len(wordSlice)))
		// dstFile, err := os.Create(filepath.Join(fileDir, filepath.Base(fileHeader.Filename)))
		// if err != nil {
		// 	http.Error(w, "Failed to store file", http.StatusInternalServerError)
		// 	log.Fatalf("Failed to create the file on server %v", err)
		// }

		// defer dstFile.Close()

		// // fmt.Printf("dstFile %v", dstFile)
		// _, err = io.Copy(dstFile, file)
		// if err != nil {
		// 	http.Error(w, "Failed to Copy file to dst", http.StatusInternalServerError)
		// 	log.Fatalf("Failed to Copy file to dst %v", err)
		// }

		// //fmt.Printf("Filename %s", fileHeader.Filename)
		// log.Printf("File %s uploaded successfully \n", fileHeader.Filename)
		// fmt.Fprint(w, fileHeader.Filename)

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
