package utils

import (
	"log"
	"mime/multipart"
)

func CloseFile(file multipart.File) {
	err := file.Close()
	if err != nil {
		log.Println("error:", err.Error())
	}
}
