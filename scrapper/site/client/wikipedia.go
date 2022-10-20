package client

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func GetLinks(webPage string) ([]string, []string, error) {
	resp, err := http.Get(webPage)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("error status %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	categories := getLinksFromDocument(doc, true)
	articles := getLinksFromDocument(doc, false)
	return categories, articles, nil
}

func getLinksFromDocument(doc *goquery.Document, isCategory bool) []string {
	f := func(i int, s *goquery.Selection) bool {
		link, _ := s.Attr("href")
		if isCategory {
			return strings.HasPrefix(link, "/wiki/Category:")
		} else {
			return !strings.HasPrefix(link, "/wiki/Category:")
		}
	}
	links := make([]string, 0)
	doc.Find(".contentsPage__section a").FilterFunction(f).Each(func(_ int, tag *goquery.Selection) {
		link, _ := tag.Attr("href")
		lk := fmt.Sprintf("https://en.wikipedia.org%s", link)
		links = append(links, lk)
		return
	})
	return links
}
