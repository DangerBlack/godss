package chart

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type ChartConfig struct {
	Type string `json:"type"`
	Data struct {
		Labels   []string `json:"labels"`
		Datasets []struct {
			Data []int64 `json:"data"`
		} `json:"datasets"`
	} `json:"data"`
	Options struct {
		Title struct {
			Display bool   `json:"display"`
			Text    string `json:"text"`
		} `json:"title"`
		Plugins struct {
			Datalabels struct {
				Display         bool   `json:"display"`
				Formatter       string `json:"formatter"`
				Color           string `json:"color"`
				BackgroundColor string `json:"backgroundColor"`
				BorderRadius    int    `json:"borderRadius"`
			} `json:"datalabels"`
		} `json:"plugins"`
	} `json:"options"`
}

// CreatePieChart sends the pie chart configuration to QuickChart API and returns the chart image as binary data
func CreatePieChart(headLabel string, labels []string, data []int64) ([]byte, error) {
	// Create the chart configuration
	var chartConfig ChartConfig
	chartConfig.Type = "pie"
	chartConfig.Data.Labels = labels
	chartConfig.Data.Datasets = []struct {
		Data []int64 `json:"data"`
	}{
		{Data: data},
	}

	// Add the options for the title and plugins
	chartConfig.Options.Title.Display = true
	chartConfig.Options.Title.Text = headLabel
	chartConfig.Options.Plugins.Datalabels.Display = true
	chartConfig.Options.Plugins.Datalabels.Formatter = `
		function(value) {
			if (value > 1000) {
				return (value / 1000) + "k";
			}
			return value;
		}`
	chartConfig.Options.Plugins.Datalabels.Color = "#000000"
	chartConfig.Options.Plugins.Datalabels.BackgroundColor = "#FFFFFF"
	chartConfig.Options.Plugins.Datalabels.BorderRadius = 3

	// Marshal the chart configuration into JSON
	jsonData, err := json.Marshal(chartConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshaling chart config: %v", err)
	}

	// Send the HTTP request to QuickChart API
	url := "https://quickchart.io/chart?c=" + url.QueryEscape(string(jsonData))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error sending request to QuickChart API: %v", err)
	}
	defer resp.Body.Close()

	// Read the response (image binary data)
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return imageData, nil
}

func main() {
	// Example usage of CreatePieChart
	headLabel := "Sample Pie Chart"
	labels := []string{"Red", "Blue", "Yellow", "Green"}
	data := []int64{10, 20, 30, 40}

	// Get the pie chart image as binary data
	image, err := CreatePieChart(headLabel, labels, data)
	if err != nil {
		fmt.Printf("Error generating chart: %v\n", err)
		return
	}

	// Save the image to a file
	err = os.WriteFile("pie_chart.png", image, 0644)
	if err != nil {
		fmt.Printf("Error saving chart image: %v\n", err)
	} else {
		fmt.Println("Chart image saved as pie_chart.png")
	}
}
