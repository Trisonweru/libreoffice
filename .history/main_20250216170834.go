package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func convertHandler(c *gin.Context) {

	log.Println("I'm Here")
	// Parse file upload
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		log.Println("File upload error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error"})
		return
	}
	defer file.Close()

	log.Println("I'm Her2")

	// Save uploaded file
	inputPath := filepath.Join("/tmp", handler.Filename)
	outFile, err := os.Create(inputPath)
	if err != nil {
		log.Println("Failed to save file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer outFile.Close()
	io.Copy(outFile, file)

	log.Println("I'm Here")

	// Determine output file extension
	outputExt := ".pdf"
	if filepath.Ext(inputPath) == ".pdf" {
		outputExt = ".docx"
	}
	outputPath := inputPath + outputExt

	// Convert using LibreOffice
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", outputExt[1:], "--outdir", "/tmp", inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Conversion failed:", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Conversion failed", "details": string(output)})
		return
	}

	// Ensure output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		log.Println("Output file not found:", outputPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Output file not generated"})
		return
	}

	// Serve converted file
	c.File(outputPath)

	// Clean up temporary files
	go func() {
		time.Sleep(10 * time.Second) // Delay to ensure file is served
		os.Remove(inputPath)
		os.Remove(outputPath)
	}()
}

func main() {
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	// Debugging route
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Define routes
	r.POST("/convert", convertHandler)

	fmt.Println("Server running on port 8050...")
	if err := r.Run(":8050"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
