package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const (
	serverURL = "http://localhost:8090"
)

var (
	rm string
	n  int
)

func uploadFile(filePath string, uploadType string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}
	var req *http.Request

	switch uploadType {
	case "add":
		req, err = http.NewRequest("POST", serverURL+"/store", body)
	case "update":
		req, err = http.NewRequest("POST", serverURL+"/update", body)
	case "wc":
		req, err = http.NewRequest("POST", serverURL+"/wc", body)
	case "freq-words":
		req, err = http.NewRequest("POST", serverURL+"/freqwords", body)
	}
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}

	switch uploadType {
	case "add":
		fmt.Printf("File %s uploaded successfully\n", string(respBody))
	case "update":
		fmt.Printf("File %s updated successfully\n", string(respBody))
	case "wc":
		fmt.Println(string(respBody))
	}

	return nil

}

func listFiles() ([]string, error) {
	req, err := http.NewRequest("GET", serverURL+"/list", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("unable to list the files %v", err)
	}
	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	files := strings.Split(string(readBytes), "-")
	return files, nil

}

func removeFile(fileName []string) error {
	fileNameBytes, err := json.Marshal(fileName)
	if err != nil {
		log.Fatalf("failed to marshal the fileName bytes %v\n", fileNameBytes)
	}

	req, err := http.NewRequest("POST", serverURL+"/rm", bytes.NewBuffer(fileNameBytes))
	if err != nil {
		return err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Unable to remove the file from server")
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Unable to read the response body %v\n", err)
	}
	fmt.Println(string(respBody))

	return nil

}

type KeyValue struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

func freqWordsCount(n int) error {
	req, err := http.NewRequest("POST", serverURL+"/freqwords", nil)
	if err != nil {
		return fmt.Errorf("unable to create the newrequest %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}
	var kv []KeyValue
	err = json.Unmarshal(respBody, &kv)
	if err != nil {
		fmt.Printf("Unable to unmarshal %v\n", err)
	}
	for i, val := range kv {
		if i == n {
			break
		}
		fmt.Println(val.Key, val.Value)

	}
	return nil
}

func checkDup(fileName string) (bool, error) {
	files, err := listFiles()
	if err != nil {
		return false, err
	}
	for _, oldfile := range files {
		trimOldfile := strings.TrimSpace(oldfile)
		if trimOldfile == fileName {
			log.Printf("file name %s already exists in the server\n", fileName)
			return true, nil
		}
	}
	return false, nil

}

func main() {
	uploadCommand := flag.NewFlagSet("add", flag.ExitOnError)
	removeCommand := flag.NewFlagSet("rm", flag.ExitOnError)
	updateCommand := flag.NewFlagSet("update", flag.ExitOnError)
	wordCountCommand := flag.NewFlagSet("wc", flag.ExitOnError)
	freqWords := flag.NewFlagSet("freq-words", flag.ExitOnError)
	freqWords.IntVar(&n, "n", 10, "numbers of frequent words you want")

	switch os.Args[1] {
	case "add":
		uploadCommand.Parse(os.Args[2:])

		if len(uploadCommand.Args()) < 1 {
			log.Fatal("Please provide at least single file")
		}
		for _, file := range uploadCommand.Args() {
			isDupe, _ := checkDup(file)
			if isDupe {
				continue
			}
			err := uploadFile(file, "add")
			if err != nil {
				log.Fatal(err)
			}
		}
	case "ls":

		files, err := listFiles()
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			fmt.Printf("%v\n", strings.TrimSpace(file))
		}
	case "rm":
		removeCommand.Parse(os.Args[2:])
		if len(removeCommand.Args()) < 1 {
			log.Fatal("Please provide at least single file")
		}
		err := removeFile(removeCommand.Args())
		if err != nil {
			fmt.Printf("Not able to delete the file %s due to error %v", rm, err)
		}

	case "update":
		updateCommand.Parse(os.Args[2:])
		if len(updateCommand.Args()) < 1 {
			log.Fatal("Please provide at least single file")
		}
		for _, file := range updateCommand.Args() {
			err := uploadFile(file, "update")
			if err != nil {
				log.Fatal(err)
			}
		}
	case "wc":
		wordCountCommand.Parse(os.Args[2:])
		if len(wordCountCommand.Args()) < 1 {
			log.Fatal("Please provide at least single file")
		}
		for _, file := range wordCountCommand.Args() {
			err := uploadFile(file, "wc")
			if err != nil {
				log.Fatal(err)
			}
		}
	case "freq-words":
		freqWords.Parse(os.Args[2:])
		err := freqWordsCount(n)
		if err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Println("Please provide the expected command Invalid command")
		uploadCommand.Usage()
		updateCommand.Usage()
		wordCountCommand.Usage()
		removeCommand.Usage()
		os.Exit(1)
	}

}
