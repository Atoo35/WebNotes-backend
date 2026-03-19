package notion

import "time"

type BaseNotionClient struct {
	RootURL string
}

type Highlights struct {
	ID           string `json:"id"`
	Text         string `json:"text"`
	URL          string `json:"url"`
	Title        string `json:"title"`
	Domain       string `json:"domain"`
	Selector     string `json:"selector"`
	Timestamp    int64  `json:"timestamp"`
	NotionPageID string `json:"notion_page_id"`
}

type CreateHighlightRequest struct {
	DatabaseID  string    `json:"database_id"`
	Title       string    `json:"title"`
	PageTitle   string    `json:"page_title"`
	URL         string    `json:"url"`
	Domain      string    `json:"domain"`
	Date        time.Time `json:"date"`
	HighlightID string    `json:"highlight_id"`
	Selector    string    `json:"selector"`
}

type EnsurePropertiesRequest struct {
	DatabaseID string `json:"database_id"`
}

type CreateDatabaseRequest struct {
	ParentPageID string `json:"parent_page_id"`
}
