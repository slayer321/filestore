package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"
)

const (
	fileDir = "./filestoredir"
)

var (
	KeyValueSlice []KeyValue
	freqWord      = make(map[string]int)
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
		fmt.Fprintf(w, fmt.Sprintf(" %d , %s", len(wordSlice), fileHeader.Filename))
	}
}

type KeyValue struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

func freqWords(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(fileDir)
	if err != nil {
		http.Error(w, "Unable to read the Dir", http.StatusInternalServerError)
	}

	//freqWord = make(map[string]int)

	for _, file := range files {
		fmt.Printf("name of the file %s\n", file.Name())
		if file.IsDir() {
			// Skip directories
			continue
		}
		filePath := filepath.Join(fileDir, file.Name())

		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusInternalServerError)
		}
		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanWords)
		for fileScanner.Scan() {
			if freqWord[fileScanner.Text()] > 0 {
				freqWord[fileScanner.Text()] += 1
			} else {
				freqWord[fileScanner.Text()] = 1
			}
		}
	}
	result := sortByValue(freqWord)
	jsonResponse, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "unable to marshal result", http.StatusInternalServerError)
	}
	w.Write(jsonResponse)
}

func sortByValue(valuesMap map[string]int) []KeyValue {

	for key, value := range valuesMap {

		found := false
		for _, kv := range KeyValueSlice {
			if kv.Key == key {
				found = true
				break
			}
		}
		if !found {
			KeyValueSlice = append(KeyValueSlice, KeyValue{Key: key, Value: value})
		}
	}

	sort.Slice(KeyValueSlice, func(i, j int) bool {
		return KeyValueSlice[i].Value > KeyValueSlice[j].Value
	})
	return KeyValueSlice
}

func startHttpServer(ctx context.Context, wg *sync.WaitGroup, handler http.Handler) {
	defer wg.Done()
	server := http.Server{
		Addr:    ":8090",
		Handler: handler,
	}

	go func() {

		fmt.Println("Starting the server at port 8090")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	select {
	case <-ctx.Done():
		timeCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := server.Shutdown(timeCtx)
		if err != nil {
			log.Fatalf("Error in shutting down the server %v", err)
		}

		os.RemoveAll(fileDir)
		log.Println("Shutdown Completed")
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
	mux.HandleFunc("/freqwords", freqWords)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go startHttpServer(ctx, &wg, mux)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)

	<-signalCh

	fmt.Println("\nGracefully shutting down HTTP server...")

	cancel()
	wg.Wait()

	log.Println("Stopped the Server")

}
