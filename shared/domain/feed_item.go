package domain

type FeedItem struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Categories  []string `json:"categories"`
}
