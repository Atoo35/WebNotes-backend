package notion

import "time"

type GetTokenResponse struct {
	AccessToken          string      `json:"access_token"`
	TokenType            string      `json:"token_type"`
	RefreshToken         string      `json:"refresh_token"`
	BotID                string      `json:"bot_id"`
	WorkspaceName        string      `json:"workspace_name"`
	WorkspaceIcon        string      `json:"workspace_icon"`
	WorkspaceID          string      `json:"workspace_id"`
	Owner                Owner       `json:"owner"`
	DuplicatedTemplateID interface{} `json:"duplicated_template_id"`
	RequestID            string      `json:"request_id"`
}

type Owner struct {
	Type string `json:"type"`
	User User   `json:"user"`
}

type User struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
	Person    struct {
		Email string `json:"email"`
	} `json:"person"`
}

type HighlightResponse struct {
	Object         string      `json:"object"`
	Results        []Result    `json:"results"`
	NextCursor     interface{} `json:"next_cursor"`
	HasMore        bool        `json:"has_more"`
	Type           string      `json:"type"`
	PageOrDatabase struct {
	} `json:"page_or_database"`
	RequestID string `json:"request_id"`
}

type Result struct {
	Object         string    `json:"object"`
	ID             string    `json:"id"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_time"`
	CreatedBy      struct {
		Object string `json:"object"`
		ID     string `json:"id"`
	} `json:"created_by"`
	LastEditedBy struct {
		Object string `json:"object"`
		ID     string `json:"id"`
	} `json:"last_edited_by"`
	Cover  interface{} `json:"cover"`
	Icon   interface{} `json:"icon"`
	Parent struct {
		Type       string `json:"type"`
		DatabaseID string `json:"database_id"`
	} `json:"parent"`
	Archived   bool        `json:"archived"`
	InTrash    bool        `json:"in_trash"`
	IsLocked   bool        `json:"is_locked"`
	Properties Properties  `json:"properties"`
	URL        string      `json:"url"`
	PublicURL  interface{} `json:"public_url"`
}

type Properties struct {
	Domain    TextProperty `json:"Domain"`
	PageTitle TextProperty `json:"Page Title"`
	Date      struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Date struct {
			Start    time.Time   `json:"start"`
			End      interface{} `json:"end"`
			TimeZone interface{} `json:"time_zone"`
		} `json:"date"`
	} `json:"Date"`
	URL         URLProperty  `json:"URL"`
	Selector    TextProperty `json:"Selector"`
	HighlightID TextProperty `json:"Highlight ID"`
	Title       TextProperty `json:"Title"`
}

type TextProperty struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	RichText []struct {
		Type string `json:"type"`
		Text struct {
			Content string      `json:"content"`
			Link    interface{} `json:"link"`
		} `json:"text"`
		Annotations struct {
			Bold          bool   `json:"bold"`
			Italic        bool   `json:"italic"`
			Strikethrough bool   `json:"strikethrough"`
			Underline     bool   `json:"underline"`
			Code          bool   `json:"code"`
			Color         string `json:"color"`
		} `json:"annotations"`
		PlainText string      `json:"plain_text"`
		Href      interface{} `json:"href"`
	} `json:"rich_text"`
}

type URLProperty struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

type CreateHighlight struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	URL       string    `json:"url"`
	Domain    string    `json:"domain"`
	Date      time.Time `json:"date"`
	HilightID string    `json:"highlight_id"`
	Selector  string    `json:"selector"`
}

type DatabaseResponse struct {
	Object         string      `json:"object"`
	Results        []interface{} `json:"results"`
	NextCursor     interface{} `json:"next_cursor"`
	HasMore        bool        `json:"has_more"`
}

type EnsureDatabasePropertiesRequest struct {
	DatabaseID string `json:"database_id"`
}
