package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var userAgents = [...]string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

const baseUrl = "https://www.google.fr/search?q="
const languageCode = "fr"

type SearchResult struct {
	Rank  int
	URL   string
	Title string
}

func randomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

func buildGoogleUrls(searchTerm string, pages, count int) ([]string, error) {
	toScrape := []string{}
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	for i := 0; i < pages; i++ {
		start := i * count
		scrapURL := fmt.Sprintf("%s%s&num=%d&hl=%s&start=%d&filter=0", baseUrl, searchTerm, count, languageCode, start)
		toScrape = append(toScrape, scrapURL)
	}
	return toScrape, nil
}

func scrapeClientRequest(page string) (*http.Response, error) {
	baseClient := &http.Client{}

	req, _ := http.NewRequest("GET", page, nil)
	req.Header.Set("User-Agent", randomUserAgent())
	res, err := baseClient.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		err := fmt.Errorf("Non 200 status")
		return nil, err
	}
	return res, nil
}

func googleReslutParsing(response *http.Response, rank int) ([]SearchResult, error) {
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return nil, err
	}
	results := []SearchResult{}
	sel := doc.Find("div.g")

	for i := range sel.Nodes {
		item := sel.Eq(i)

		linkTag := item.Find("a")
		titleTag := linkTag.Find("h3")

		link, _ := linkTag.Attr("href")

		title := titleTag.Text()
		link = strings.Trim(link, " ")
		if link != "" && link != "#" && !strings.HasPrefix(link, "/") {
			results = append(results, SearchResult{
				rank,
				link,
				title,
			})
			rank++
		}
	}
	return results, nil
}

func GoogleScraper(searchTerm string, pages, count int) ([]SearchResult, error) {
	results := []SearchResult{}
	resultsCounter := 0
	googlepages, err := buildGoogleUrls(searchTerm, pages, count)
	if err != nil {
		return nil, err
	}
	for _, page := range googlepages {
		res, err := scrapeClientRequest(page)
		if err != nil {
			return nil, err
		}

		data, err := googleReslutParsing(res, resultsCounter)
		if err != nil {
			return nil, err
		}

		resultsCounter += len(data)
		for _, result := range data {
			results = append(results, result)
		}
	}

	return results, nil
}

func main() {
	if res, err := GoogleScraper("global packaging service", 2, 30); err == nil {
		j, err := json.Marshal(res)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		fmt.Println(string(j))
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}
