package server

import (
	"net/http"
	"time"

	"512b.it/godss/src/dss"
	"512b.it/godss/src/models"
	"github.com/gin-gonic/gin"

	"bufio"
	"os"
)

var counter = dss.NewDss(1000, 24*time.Hour)
var firstDay time.Time = time.Date(2024, time.December, 19, 0, 0, 0, 0, time.UTC)

var words []string

func loadWordListFromFile(file string) ([]string, error) {
	var words []string

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return words, nil
}

func HandleChallenge(c *gin.Context) {
	var numberOfDays int
	if words == nil {
		var err error
		if words, err = loadWordListFromFile("word_list.txt"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load word list"})
			return
		}
	}

	numberOfDays = int(time.Since(firstDay).Hours() / 24)

	if numberOfDays >= len(words) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No more challenges available"})
		return
	}

	response := models.ChallengeResponse{
		Challenge: words[numberOfDays],
	}

	c.JSON(http.StatusOK, response)
}

func HandlePopularity(c *gin.Context) {
	var err error
	var req models.PopularityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response := models.PopularityResponse{
		Options: make(map[string]int64),
	}

	var results map[string]int64
	if results, err = counter.CountEvents("", req.Options, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch popularity scores"})
		return
	}

	for option, results := range results {

		response.Options[option] = results
	}

	c.JSON(http.StatusOK, response)
}
