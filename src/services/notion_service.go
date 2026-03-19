package services

import (
	"strings"

	notion_client "github.com/Atoo35/WebNotes-backend/src/clients/notion"
	"github.com/Atoo35/WebNotes-backend/src/domains/notion"
	"github.com/Atoo35/WebNotes-backend/src/utils/database"
	zlog "github.com/Atoo35/WebNotes-backend/src/utils/logger"
	"github.com/Atoo35/WebNotes-backend/src/utils/resterror"
	"go.uber.org/zap"
)

type NotionServiceI interface {
	StartOAuthFlow() (notion.BaseNotionClient, *resterror.RestError)
	GetToken(code string) (string, *resterror.RestError)
	LoadHighlightsByDomain(domain string, databaseId string) ([]notion.Highlights, *resterror.RestError)
	SaveHighlight(databaseID string, highlight *notion.CreateHighlightRequest) *resterror.RestError
	GetHighlightByID(databaseID string, highlightID string) (*notion.Highlights, *resterror.RestError)
	DeleteHighlight(notionPageID string) *resterror.RestError
	ListDatabases() ([]interface{}, *resterror.RestError)
	EnsureDatabaseProperties(databaseID string) *resterror.RestError
	CreateDatabaseAuto() (map[string]interface{}, *resterror.RestError)
	CreateDatabase(parentPageID string) (map[string]interface{}, *resterror.RestError)
	LoadAllHighlights(databaseID string) ([]notion.Highlights, *resterror.RestError)
}

type notionService struct{}

var NotionService NotionServiceI = &notionService{}

func (ns *notionService) StartOAuthFlow() (notion.BaseNotionClient, *resterror.RestError) {
	oauthUrl := notion_client.NotionClient.StartOAuthFlow()

	// Convert client (from clients/notion) to domain-level BaseNotionClient
	// which contains only the RootURL.
	return notion.BaseNotionClient{RootURL: oauthUrl}, nil
}

func (ns *notionService) GetToken(code string) (string, *resterror.RestError) {
	token, err := notion_client.NotionClient.GetToken(code)
	if err != nil {
		return "", resterror.InternalServerError("Failed to get token from Notion")
	}

	user := &database.User{
		Email:        token.Owner.User.Person.Email,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	err = database.SQLClient.Set(user)
	if err != nil {
		return "", resterror.InternalServerError("Failed to store tokens in database")
	}

	zlog.Logger.Info("Tokens stored in cache successfully")

	return token.AccessToken, nil
}

func (ns *notionService) LoadHighlightsByDomain(domain string, databaseId string) ([]notion.Highlights, *resterror.RestError) {
	res, err := notion_client.NotionClient.LoadHighlightsByDomain(domain, databaseId)
	if err != nil {
		return nil, resterror.InternalServerError("Failed to load highlights from Notion")
	}

	zlog.Logger.Info("Highlights loaded from Notion successfully", zap.Any("response", res))
	response := ns.parseHighlightResponse(res)
	return response, nil
}

func (ns *notionService) SaveHighlight(databaseID string, highlight *notion.CreateHighlightRequest) *resterror.RestError {
	data := &notion_client.CreateHighlight{
		ID:       highlight.HighlightID,
		Title:    highlight.PageTitle,
		Text:     highlight.Title,
		URL:      highlight.URL,
		Domain:   highlight.Domain,
		Date:     highlight.Date,
		Selector: highlight.Selector,
	}
	err := notion_client.NotionClient.SaveHighlight(databaseID, data)
	if err != nil {
		return resterror.InternalServerError("Failed to save highlight to Notion")
	}
	return nil
}

func (ns *notionService) GetHighlightByID(databaseID string, highlightID string) (*notion.Highlights, *resterror.RestError) {
	res, err := notion_client.NotionClient.GetHighlightByID(databaseID, highlightID)
	if err != nil {
		return nil, resterror.InternalServerError("Failed to get highlight from Notion")
	}
	response := ns.parseHighlightResponse(res)

	if len(response) == 0 {
		return nil, resterror.NotFound("Highlight not found")
	}
	return &response[0], nil
}

func (ns *notionService) parseHighlightResponse(data *notion_client.HighlightResponse) []notion.Highlights {
	response := make([]notion.Highlights, 0)

	for _, item := range data.Results {
		highlight := notion.Highlights{
			URL:          item.Properties.URL.URL,
			Timestamp:    item.Properties.Date.Date.Start.Unix(),
			NotionPageID: item.ID,
		}
		if len(item.Properties.HighlightID.RichText) > 0 {
			highlight.ID = item.Properties.HighlightID.RichText[0].Text.Content
		}
		if len(item.Properties.Title.RichText) > 0 {
			highlight.Text = item.Properties.Title.RichText[0].Text.Content
		}
		if len(item.Properties.PageTitle.RichText) > 0 {
			highlight.Title = item.Properties.PageTitle.RichText[0].Text.Content
		}
		if len(item.Properties.Domain.RichText) > 0 {
			highlight.Domain = item.Properties.Domain.RichText[0].Text.Content
		}
		if len(item.Properties.Selector.RichText) > 0 {
			highlight.Selector = item.Properties.Selector.RichText[0].Text.Content
		}

		response = append(response, highlight)
	}
	return response
}

func (ns *notionService) DeleteHighlight(notionPageID string) *resterror.RestError {
	err := notion_client.NotionClient.DeleteHighlight(notionPageID)
	if err != nil {
		return resterror.InternalServerError("Failed to delete highlight from Notion")
	}
	return nil
}

func (ns *notionService) ListDatabases() ([]interface{}, *resterror.RestError) {
	res, err := notion_client.NotionClient.ListDatabases()
	if err != nil {
		return nil, resterror.InternalServerError("Failed to list databases")
	}
	return res, nil
}

func (ns *notionService) EnsureDatabaseProperties(databaseID string) *resterror.RestError {
	err := notion_client.NotionClient.EnsureDatabaseProperties(databaseID)
	if err != nil {
		return resterror.InternalServerError("Failed to ensure database properties")
	}
	return nil
}

func (ns *notionService) CreateDatabaseAuto() (map[string]interface{}, *resterror.RestError) {
	databases, err := notion_client.NotionClient.ListDatabases()
	if err != nil {
		return nil, resterror.InternalServerError("Failed to list databases")
	}

	for _, dbRaw := range databases {
		db, ok := dbRaw.(map[string]interface{})
		if !ok {
			continue
		}
		titleArr, _ := db["title"].([]interface{})
		if len(titleArr) > 0 {
			titleObj, _ := titleArr[0].(map[string]interface{})
			plainText, _ := titleObj["plain_text"].(string)
			if strings.Contains(strings.ToLower(plainText), "webnotes") || strings.Contains(strings.ToLower(plainText), "highlights") {
				return map[string]interface{}{
					"success":    true,
					"databaseId": db["id"],
					"created":    false,
				}, nil
			}
		}
	}

	page, err := notion_client.NotionClient.CreatePage("WebNotes - My Highlights")
	var parentPageID string
	if err == nil {
		parentPageID, _ = page["id"].(string)
	} else {
		pages, searchErr := notion_client.NotionClient.SearchPages()
		if searchErr != nil || len(pages) == 0 {
			return nil, resterror.InternalServerError("Failed to create or find parent page")
		}
		p, _ := pages[0].(map[string]interface{})
		parentPageID, _ = p["id"].(string)
	}

	database, err := notion_client.NotionClient.CreateDatabase(parentPageID)
	if err != nil {
		return nil, resterror.InternalServerError("Database creation failed")
	}

	return map[string]interface{}{
		"success":    true,
		"databaseId": database["id"],
		"created":    true,
	}, nil
}

func (ns *notionService) CreateDatabase(parentPageID string) (map[string]interface{}, *resterror.RestError) {
	database, err := notion_client.NotionClient.CreateDatabase(parentPageID)
	if err != nil {
		return nil, resterror.InternalServerError("Database creation failed")
	}
	return database, nil
}

func (ns *notionService) LoadAllHighlights(databaseID string) ([]notion.Highlights, *resterror.RestError) {
	var allHighlights []notion.Highlights
	var startCursor *string

	for {
		res, err := notion_client.NotionClient.QueryDatabase(databaseID, startCursor)
		if err != nil {
			return nil, resterror.InternalServerError("Failed to query database")
		}

		parsed := ns.parseHighlightResponse(res)
		allHighlights = append(allHighlights, parsed...)

		if !res.HasMore {
			break
		}
		cursorStr, ok := res.NextCursor.(string)
		if !ok {
			break
		}
		startCursor = &cursorStr
	}

	if allHighlights == nil {
		allHighlights = make([]notion.Highlights, 0)
	}

	return allHighlights, nil
}
