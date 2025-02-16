package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func convertHandler(c *gin.Context) {
	log.Println("Received conversion request...")

	// Parse file upload
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		log.Println("File upload error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error"})
		return
	}
	defer file.Close()

	// Ensure safe filename
	safeFilename := strings.ReplaceAll(handler.Filename, " ", "_")
	inputPath := filepath.Join("/tmp", safeFilename)

	// Save uploaded file
	outFile, err := os.Create(inputPath)
	if err != nil {
		log.Println("Failed to save file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer outFile.Close()
	io.Copy(outFile, file)
	log.Println("File saved:", inputPath)

	// Determine conversion tool & format
	var outputFormat, tool string
	if filepath.Ext(inputPath) == ".pdf" {
		outputFormat = "docx"
		tool = "python3 convert.py" // Use Python script for PDF-to-DOCX
	} else {
		outputFormat = "pdf"
		tool = "libreoffice" // Use LibreOffice for other conversions
	}

	// Define output file path
	outputDir := "/tmp"
	outputFilename := strings.TrimSuffix(safeFilename, filepath.Ext(safeFilename)) + "." + outputFormat
	outputPath := filepath.Join(outputDir, outputFilename)

	// Ensure the required tool is installed
	if tool == "python3 convert.py" {
		_, err = exec.LookPath("python3")
	} else {
		_, err = exec.LookPath("libreoffice")
	}

	if err != nil {
		log.Println(tool, "not found:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": tool + " is not installed"})
		return
	}

	// Run the appropriate conversion command
	var cmd *exec.Cmd
	if tool == "python3 convert.py" {
		cmd = exec.Command("python3", "convert.py", inputPath, outputPath)
	} else {
		cmd = exec.Command("libreoffice", "--headless", "--convert-to", outputFormat, "--outdir", outputDir, inputPath)
	}

	// Execute the conversion
	output, err := cmd.CombinedOutput()
	log.Println(tool, "Output:", string(output))

	if err != nil {
		log.Println("Conversion failed:", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Conversion failed",
			"details": string(output),
		})
		return
	}

	// Ensure the converted file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		log.Println("Output file not found:", outputPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Output file not generated", "details": string(output)})
		return
	}
	log.Println("Conversion successful:", outputPath)

	// Serve converted file
	c.File(outputPath)

	// Clean up files after serving
	go func() {
		time.Sleep(10 * time.Second) // Delay cleanup to ensure file is downloaded
		os.Remove(inputPath)
		os.Remove(outputPath)
		log.Println("Cleaned up:", inputPath, outputPath)
	}()
}

func main() {
	r := gin.Default()

	// CORS Middleware
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
	if err := r.Run(":8052"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
