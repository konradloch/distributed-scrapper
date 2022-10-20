package client

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type Wiki struct {
	logger *zap.SugaredLogger
}

func NewWiki(logger *zap.SugaredLogger) *Wiki {
	return &Wiki{
		logger: logger,
	}
}

func (w *Wiki) GetLinks(webPage string) ([]string, []string, error) {
	w.logger.Infow("fetching content", "content", webPage)
	resp, err := http.Get(webPage)
	if err != nil {
		w.logger.Errorf("wiki http get failed")
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		w.logger.Errorf("wiki http get failed cause difrent code")
		return nil, nil, fmt.Errorf("error status %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		w.logger.Errorf("goquery fails")
		return nil, nil, err
	}
	root := w.getLinksFromDocument(doc, true, ".contentsPage__section a")
	if len(root) > 100 {
		w.logger.Infow("content done", "content", webPage, "categories", len(root), "articles", 0)
		return root, []string{}, nil
	}
	categories := w.getLinksFromDocument(doc, true, ".mw-category-group a")
	articles := w.getLinksFromDocument(doc, false, ".mw-category-group a")
	fmt.Println(articles)
	w.logger.Infow("content done", "content", webPage, "categories", len(categories), "articles", len(articles))
	return categories, articles, nil
}

func (w *Wiki) getLinksFromDocument(doc *goquery.Document, isCategory bool, section string) []string {
	f := func(i int, s *goquery.Selection) bool {
		link, _ := s.Attr("href")
		if isCategory {
			return strings.HasPrefix(link, "/wiki/Category:")
		} else {
			return !strings.HasPrefix(link, "/wiki/Category:") &&
				!strings.HasPrefix(link, "/w/") &&
				!strings.HasPrefix(link, "/wiki/Special:") &&
				!strings.HasPrefix(link, "/wiki/Wikipedia:") &&
				!strings.HasPrefix(link, "/wiki/Help:") &&
				!strings.HasPrefix(link, "/wiki/File") &&
				!strings.HasPrefix(link, "#")
		}
	}
	links := make([]string, 0)
	doc.Find(section).FilterFunction(f).Each(func(_ int, tag *goquery.Selection) {
		link, _ := tag.Attr("href")
		lk := fmt.Sprintf("https://en.wikipedia.org%s", link)
		links = append(links, lk)
		return
	})
	return links
}
