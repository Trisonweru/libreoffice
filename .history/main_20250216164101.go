package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func convertHandler(c *gin.Context) {
	// Parse file upload
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error"})
		return
	}
	defer file.Close()

	// Save uploaded file
	inputPath := filepath.Join("/tmp", handler.Filename)
	outFile, err := os.Create(inputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Conversion failed"})
		return
	}

	// Send back the converted file
	c.File(outputPath)
}

func main() {
	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		cors.Default().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		}))(c.Writer, c.Request)
	})

	// Define routes
	r.POST("/convert", convertHandler)

	fmt.Println("Server running on port 8080...")
	r.Run(":8080") // Start Gin server
}
