package broker

type SiteRequestEvent struct {
	Url string `json:"url"`
}

type SiteResponseEvent struct {
	ParentUrl  string   `json:"parentUrl"`
	Categories []string `json:"categories"`
	Articles   []string `json:"articles"`
}
