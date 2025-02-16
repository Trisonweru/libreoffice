package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func convertHandler(w http.ResponseWriter, r *http.Request) {
	// Parse file upload
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File upload error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save uploaded file
	inputPath := filepath.Join("/tmp", handler.Filename)
	outFile, err := os.Create(inputPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	io.Copy(outFile, file)

	// Define output file
	outputExt := ".pdf"
	if filepath.Ext(inputPath) == ".pdf" {
		outputExt = ".docx"
	}
	outputPath := inputPath + outputExt

	// Convert using LibreOffice
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", outputExt[1:], "--outdir", "/tmp", inputPath)
	err = cmd.Run()
	if err != nil {
		http.Error(w, "Conversion failed", http.StatusInternalServerError)
		return
	}

	// Send back the converted file
	http.ServeFile(w, r, outputPath)
}

func main() {
	http.HandleFunc("/convert", convertHandler)
	fmt.Println("Server running on port 8080...")
	http.ListenAndServe(":8080", nil)
}

//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