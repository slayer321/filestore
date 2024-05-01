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
)

const (
	serverURL = "http://localhost:8090"
)

var (
	store string
	list  string
)

func uploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Printf("Name of the file is %s\n", file.Name())

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

	req, err := http.NewRequest("POST", serverURL+"/store", body)
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

	// resp, err := http.Post(serverURL+"/store", "multipart/form-data", file)
	// if err != nil {
	// 	return err
	// }

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Printing the respBody %v\n", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}

	fmt.Println("File uploaded successfully")
	return nil

}

func listFiles() error {

	req, err := http.NewRequest("GET", serverURL+"/list", nil)
	if err != nil {
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("unable to list the files %v", err)
	}

	//fmt.Printf("resp.Body: %v\n", resp.Body)
	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%v", string(readBytes))

	return nil

}

func main() {

	uploadCommand := flag.NewFlagSet("add", flag.ExitOnError)
	uploadCommand.Parse(os.Args[2:])
	//uploadFilePath := uploadCommand.String("add", "", "Path of the file to upload")

	switch os.Args[1] {
	case "add":

		// fmt.Printf("file path given by user %v\n", &uploadFilePath)
		// if *uploadFilePath == "" {
		// 	log.Fatal("Please provide a file path using the -add flag")
		// }
		if len(uploadCommand.Args()) < 0 {
			log.Fatal("Please provide at lease single file")
		}
		for _, file := range uploadCommand.Args() {
			err := uploadFile(file)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "ls":

		err := listFiles()
		if err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Invalid command")
		os.Exit(1)
	}

}
