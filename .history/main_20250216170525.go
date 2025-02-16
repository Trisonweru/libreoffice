package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
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

	// Enable CORS using Gin middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Define routes
	r.POST("/convert", convertHandler)

	fmt.Println("Server running on port 8080...")
	r.Run(":8050") // Start Gin server
	fmt.Println("Server running on port 8080...")

}
