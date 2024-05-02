package main

import (
	"bytes"
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
	store string
	list  string
	rm    string
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

func removeFile(fileName string) error {

	req, err := http.NewRequest("POST", serverURL+"/rm", bytes.NewBufferString(fileName))
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

	switch os.Args[1] {
	case "add":
		uploadCommand.Parse(os.Args[2:])

		if len(uploadCommand.Args()) < 0 {
			log.Fatal("Please provide at lease single file")
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
		fmt.Printf("Remove file name is %s\n", removeCommand.Args()[0])
		err := removeFile(removeCommand.Args()[0])
		if err != nil {
			fmt.Printf("Not able to delete the file %s due to error %v", rm, err)
		}

	case "update":
		updateCommand.Parse(os.Args[2:])
		if len(updateCommand.Args()) < 0 {
			log.Fatal("Please provide at lease single file")
		}
		for _, file := range updateCommand.Args() {
			err := uploadFile(file, "update")
			if err != nil {
				log.Fatal(err)
			}
		}
	case "wc":
		wordCountCommand.Parse(os.Args[2:])
		if len(wordCountCommand.Args()) < 0 {
			log.Fatal("Please provide at lease single file")
		}
		for _, file := range wordCountCommand.Args() {
			err := uploadFile(file, "wc")
			if err != nil {
				log.Fatal(err)
			}
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
