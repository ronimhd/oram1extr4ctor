package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

func main() {
	fName := "oramiproducts.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"Category", "Name", "Price"})

	// Instantiate main collector
	c := colly.NewCollector(
		colly.AllowedDomains("orami.co.id", "www.orami.co.id"),
		colly.CacheDir("./orami_cache"),
	)

	homeProductsURL := "https://www.orami.co.id/c/fashion-dan-aksesoris"
	extractProductsPerCategory(homeProductsURL, c, writer)
	homeProductsURL = "https://www.orami.co.id/c/popok-krim-dan-tisu-bayi"
	extractProductsPerCategory(homeProductsURL, c, writer)

}

func extractProductsPerCategory(homeProductsURL string, c *colly.Collector, writer *csv.Writer) {
	// Create another collector to scrape product information details in category pages
	detailCollector := c.Clone()
	detailCollectorPages := c.Clone()

	// for debugging
	detailCollector.OnRequest(func(r *colly.Request) {
		log.Println("visiting product ", r.URL.String())
	})

	detailCollectorPages.OnRequest(func(r *colly.Request) {
		log.Println("visiting product pages ", r.URL.String())
	})

	// Extract details of the course
	detailCollector.OnHTML(`div[class=wrap-widget-detail]`, extractProductsFunc(writer))

	detailCollectorPages.OnHTML(`div[class=wrap-widget-detail]`, extractProductsFunc(writer))

	var lastPage int

	detailCollector.OnHTML(`ul[class=pagination]`, func(e *colly.HTMLElement) {
		log.Println("Paging found", e.Request.URL)
		var lastPageURL string

		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			lastPageURL = el.ChildAttr("a", "href")

		})

		log.Println("lastPage URL found", lastPageURL)
		lastPage, _ = strconv.Atoi(string([]rune(lastPageURL)[(strings.Index(lastPageURL, "?page=") + 6):]))

		log.Println("lastPage found", lastPage)

	})

	homeProductsURLPOne := homeProductsURL + "?page=1"

	log.Println(homeProductsURLPOne)

	detailCollector.Visit(fmt.Sprintf(homeProductsURLPOne))

	for i := 2; i <= lastPage; i++ {
		detailCollectorPages.Visit(fmt.Sprintf((homeProductsURL + "?page=%d"), i))
	}
}

func extractProductsFunc(writer *csv.Writer) func(*colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		log.Println("Product found", e.Request.URL)
		cat := e.ChildText(".prod-cat > label > a")
		if cat == "" {
			log.Println("No title found", e.Request.URL)
		}
		name := e.ChildText(".prod-name >  a")
		log.Println("name ", name)
		if name == "" {
			log.Println("No name found", e.Request.URL)
		}
		price := e.ChildText(".widget-price > p")
		if price == "" {
			log.Println("No price found", e.Request.URL)
		}

		writer.Write([]string{
			cat,
			name,
			price,
		})
	}
}