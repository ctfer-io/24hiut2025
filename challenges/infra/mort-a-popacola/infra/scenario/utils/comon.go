package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type RangeRequest struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type RangeResponse struct {
	Value int `json:"value"`
}

func GetAvailableId(min, max int) (int, error) {
	url := "http://tools.ctfer-io.lab:8080/next"

	// Customize the range here
	requestBody := RangeRequest{
		Min: min,
		Max: max,
	}

	// Encode request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return 0, err
	}

	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error: %s\n", resp.Status)
		return 0, err
	}

	// Decode response
	var response RangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println("Error decoding response:", err)
		return 0, err
	}

	return response.Value, nil
}
