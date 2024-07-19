package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func getFileDrive(id int, idFileDrive string) ([]byte, error) {
	log.Println("func : getFileDrive()")
	downloadURL := "https://drive.google.com/uc?export=download&id=" + idFileDrive
	filePath := "lessons/" + strconv.Itoa(id)
	data, err := os.ReadFile(filePath)
	// have file
	if err == nil {
		log.Println("info : The file id already exists, no recording is performed. Use another function to correct data if corrupted")
		return data, nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		panic(err)
	}
	println(downloadURL)
	response, err := http.Get(downloadURL)
	if err != nil {
		log.Println("Error downloading file:", err)
		panic(err)
	}
	println(response.Status)
	defer response.Body.Close()
	if response.Status != "200 OK" {
		return []byte{}, errors.New("file drive not found")
	}

	sizeFile, err := io.Copy(file, response.Body)
	if err != nil {
		log.Println("Error copying file:", err)
		panic(err)
	}
	log.Printf("log  : Downloaded %d bytes to %s\n", sizeFile, filePath)

	data, err = os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return data, nil
}
