package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func getFileDrive(id int, idFileDrive string) []byte {
	log.Println("func : getFileDrive()")
	downloadURL := "https://drive.google.com/uc?export=download&id=" + idFileDrive
	filePath := "lessons/" + strconv.Itoa(id)
	data, err := os.ReadFile(filePath)
	// have file
	if err == nil {
		log.Println("info : The file id already exists, no recording is performed. Use another function to correct data if corrupted")
		return data
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		panic(err)
	}

	response, err := http.Get(downloadURL)
	if err != nil {
		log.Println("Error downloading file:", err)
		panic(err)
	}
	defer response.Body.Close()

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

	return data
}
