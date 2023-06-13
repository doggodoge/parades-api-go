package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"garymoore.ie/parades-api/lru"
	"garymoore.ie/parades-api/parsing"
	"github.com/gin-gonic/gin"
)

var detailsCache *lru.Cache

// TODO: Add a hashmap of street -> []*ParadeDetails mappings. Searching is
//  horrendously slow right now due to search complexity. Best get it indexed.

func main() {
	router := gin.Default()
	detailsCache = lru.New(1000)

	router.GET("/parades", getParades)
	router.GET("/parades_by_street_belfast/:street", getParadesByStreetBelfast)

	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}

type ParadeQueryParams struct {
	Location string `form:"location"`
}

func getParades(c *gin.Context) {
	var queryParams ParadeQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	fmt.Println("queryParams", queryParams)

	parades := parsing.AllParades()

	if queryParams.Location != "" {
		parades = filterByLocation(queryParams.Location, parades)
	}

	c.JSON(http.StatusOK, parades)
}

func filterByLocation(location string, parades []*parsing.ParadeSummary) []*parsing.ParadeSummary {
	filteredParades := make([]*parsing.ParadeSummary, 0)
	for _, p := range parades {
		if containsIgnoreCase(p.Town, location) {
			filteredParades = append(filteredParades, p)
		}
	}
	return filteredParades
}

func getParadesByStreetBelfast(c *gin.Context) {
	street := c.Param("street")

	decodedStreet, err := url.QueryUnescape(street)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parades := parsing.AllParades()

	paradeURLs := make([]string, 0)
	for _, p := range parades {
		if strings.ToLower(p.Town) == "belfast" {
			paradeURLs = append(paradeURLs, p.DetailsURL)
		}
	}

	parsedParadeDetails := make([]parsing.ParadeDetails, len(paradeURLs))
	var wg sync.WaitGroup
	for i, url := range paradeURLs {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()

			cachedEntry, ok := detailsCache.Get(url)
			if ok {
				fmt.Println("cache hit")
				parsedParadeDetails[i] = cachedEntry.(parsing.ParadeDetails)
				return
			}
			fmt.Println("cache miss")

			response, err := http.Get(url)
			if err != nil {
				// Handle error...
				return
			}
			defer response.Body.Close()

			html, err := ioutil.ReadAll(response.Body)
			if err != nil {
				// Handle error...
				return
			}

			parsedParadeDetails[i] = parsing.ParseParadesDetailsHTML(string(html))
			detailsCache.Add(url, parsedParadeDetails[i])
		}(i, url)
	}

	wg.Wait()

	paradesByStreet := make([]parsing.ParadeDetails, 0)
	for _, parade := range parsedParadeDetails {
		if parade.ProposedOutwardRoute != "" &&
			strings.Contains(strings.ToLower(parade.ProposedOutwardRoute), strings.ToLower(decodedStreet)) {
			paradesByStreet = append(paradesByStreet, parade)
		}
	}

	c.JSON(http.StatusOK, paradesByStreet)
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
