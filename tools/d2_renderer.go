package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RenderRequest defines the structure for render request with options
type RenderRequest struct {
	Content string                 `json:"content"`
	Options map[string]interface{} `json:"options"`
}

func handleRenderRequest(w http.ResponseWriter, r *http.Request) {
	var requestData RenderRequest

	// Parse JSON body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	output, err := renderText(requestData.Content, requestData.Options)
	if err != nil {
		log.Println("Error rendering diagram:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the content type to SVG
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func renderText(content string, options map[string]interface{}) ([]byte, error) {
	// Use temporary file for bundling the output
	outputFile, err := os.CreateTemp("", "d2_output_*.svg")
	if err != nil {
		return nil, fmt.Errorf("Failed to create temporary output file: %w", err)
	}
	outputFile.Close()                 // Close the file to ensure it's ready for writing
	defer os.Remove(outputFile.Name()) // Clean up the temp file after use

	// Start with base command
	args := []string{}

	// Add all options from the map
	for key, value := range options {
		if value == "" {
			// For flags without values
			args = append(args, "--"+key)
		} else {
			args = append(args, "--"+key+"="+fmt.Sprintf("%v", value))
		}
	}

	// Add output file path (input will come from stdin)
	args = append(args, "-", outputFile.Name())

	log.Println("Start D2 render", args)

	command := exec.Command("d2", args...)
	command.Stdin = bytes.NewBuffer([]byte(content))
	var stderr strings.Builder
	command.Stderr = &stderr
	err = command.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute d2 command: %s", stderr.String())
	}
	output, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("Failed to read output file: %w", err)
	}
	return output, nil
}

func getPortFromEnv() string {
	const defaultPort = "8080"
	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		return defaultPort // Default port if not specified
	}

	// Optional: validate that the port is a valid number
	_, err := strconv.Atoi(portEnv)
	if err != nil {
		log.Printf("Invalid PORT environment variable: %s, using default 8080\n", portEnv)
		return defaultPort
	}

	return portEnv
}

func main() {
	http.HandleFunc("/render", handleRenderRequest)
	// Move to use d2 icons
	wd, _ := os.Getwd()
	port := getPortFromEnv()
	log.Printf("D2 rendering service started on %s at %s", port, wd)
	http.ListenAndServe("localhost:"+port, nil)
}
