package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

// RenderRequest defines the structure for render request with options
type RenderRequest struct {
	Content string            `json:"content"`
	Options map[string]string `json:"options"`
}

func handleRenderRequest(w http.ResponseWriter, r *http.Request) {
	var requestData RenderRequest

	// Parse JSON body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	output, err := renderText(requestData.Content, requestData.Options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, output)
}

func renderText(content string, options map[string]string) (string, error) {
	// Start with base command
	args := []string{}
	// Add all options from the map
	for key, value := range options {
		if value == "" {
			// For flags without values
			args = append(args, "--"+key)
		} else {
			args = append(args, "--"+key+"="+value)
		}
	}

	// Add output file path (input will come from stdin)
	args = append(args, "-")

	fmt.Println("D2 render", args)
	fmt.Println(content)

	command := exec.Command("d2", args...)
	command.Stdin = bytes.NewBuffer([]byte(content))

	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute d2 command: %w", err)
	}
	res := string(output)
	// fmt.Println("D2 render result", res)
	return res, nil
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
		fmt.Printf("Invalid PORT environment variable: %s, using default 8080\n", portEnv)
		return defaultPort
	}

	return portEnv
}

func main() {
	http.HandleFunc("POST /render", handleRenderRequest)
	// Move to use d2 icons
	wd, _ := os.Getwd()
	port := getPortFromEnv()
	fmt.Printf("D2 rendering service started on %s at %s", port, wd)
	http.ListenAndServe("localhost:"+port, nil)
}
