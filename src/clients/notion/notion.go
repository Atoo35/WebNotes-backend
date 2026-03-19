package notion

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Atoo35/WebNotes-backend/src/utils/database"
	"github.com/Atoo35/WebNotes-backend/src/utils/helper"
	zlog "github.com/Atoo35/WebNotes-backend/src/utils/logger"
	"go.uber.org/zap"
)

type NotionClientI interface {
	NewNotionClient()
	StartOAuthFlow() string
	GetToken(code string) (*GetTokenResponse, error)
	LoadHighlightsByDomain(domain string, databaseId string) (*HighlightResponse, error)
	SaveHighlight(databaseID string, highlight *CreateHighlight) error
	GetHighlightByID(databaseID string, highlightID string) (*HighlightResponse, error)
	DeleteHighlight(notionPageID string) error
	ListDatabases() ([]interface{}, error)
	EnsureDatabaseProperties(databaseID string) error
	CreateDatabase(parentPageID string) (map[string]interface{}, error)
	CreatePage(title string) (map[string]interface{}, error)
	SearchPages() ([]interface{}, error)
	QueryDatabase(databaseID string, startCursor *string) (*HighlightResponse, error)
}

type notionClient struct {
	RootURL      string
	redirectURL  string
	clientID     string
	clientSecret string
	accessToken  string
}

var NotionClient NotionClientI = &notionClient{}

func (nc *notionClient) NewNotionClient() {
	nc.RootURL = os.Getenv("NOTION_ROOT_URL")
	nc.clientID = os.Getenv("NOTION_CLIENT_ID")
	nc.clientSecret = os.Getenv("NOTION_CLIENT_SECRET")
	nc.redirectURL = os.Getenv("REDIRECT_URL")
}

func (nc *notionClient) StartOAuthFlow() string {
	zlog.Logger.Info("here are we")
	var authUrl = fmt.Sprintf("%s/oauth/authorize?client_id=%s&response_type=code&owner=user&redirect_uri=%s", nc.RootURL, nc.clientID, nc.redirectURL)
	fmt.Println(authUrl)

	return authUrl
}

func (nc *notionClient) GetToken(code string) (*GetTokenResponse, error) {
	// "client_id:client_secret"
	authValue := fmt.Sprintf("%s:%s", nc.clientID, nc.clientSecret)

	// base64 encode
	basicAuth := base64.StdEncoding.EncodeToString([]byte(authValue))

	// Print for debugging (optional)
	fmt.Println("Basic Auth:", basicAuth)

	// Body you want to POST — change to whatever Notion expects
	requestBody := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": nc.redirectURL,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// Create POST request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth/token", nc.RootURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		zlog.Logger.Error("Error creating request to Notion:", zap.Error(err))
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth)

	zlog.Logger.Info("Request to Notion prepared", zap.String("url", req.URL.String()), zap.String("method", req.Method))
	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Logger.Error("Error making request to Notion:", zap.Error(err))
		return nil, err
	}
	zlog.Logger.Info("Response received from Notion", zap.String("status", resp.Status))
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		zlog.Logger.Error("Error reading response body from Notion:", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		zlog.Logger.Error("Bad request response from Notion", zap.ByteString("body", respBody))
		return nil, errors.New(string(respBody))
	}

	var tokenResponse *GetTokenResponse
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		zlog.Logger.Error("Error unmarshaling response from Notion:", zap.Error(err))
		return nil, err
	}

	zlog.Logger.Info("Successfully obtained token from Notion", zap.String("access_token", tokenResponse.AccessToken))
	nc.accessToken = tokenResponse.AccessToken
	return tokenResponse, nil
}

func (nc *notionClient) LoadHighlightsByDomain(domain string, databaseId string) (*HighlightResponse, error) {
	err := nc.checkAccessToken()
	if err != nil {
		return nil, err
	}

	requestBody := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Domain",
			"rich_text": map[string]interface{}{
				"equals": domain,
			},
		},
		"sorts": []map[string]interface{}{
			{
				"property":  "Date",
				"direction": "descending",
			},
		},
		"page_size": 100,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// Create POST request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/databases/%s/query", nc.RootURL, databaseId), bytes.NewBuffer(bodyBytes))
	if err != nil {
		zlog.Logger.Error("Error creating request to Notion:", zap.Error(err))
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Logger.Error("Error making request to Notion:", zap.Error(err))
		return nil, err
	}
	zlog.Logger.Info("Response received from Notion", zap.String("status", resp.Status))
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		zlog.Logger.Error("Error reading response body from Notion:", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		zlog.Logger.Error("Bad request response from Notion", zap.ByteString("body", respBody))
		return nil, errors.New(string(respBody))
	}

	var highlightResponse *HighlightResponse
	err = json.Unmarshal(respBody, &highlightResponse)
	if err != nil {
		zlog.Logger.Error("Error unmarshaling highlight response from Notion:", zap.Error(err))
		return nil, err
	}
	return highlightResponse, nil
}

func (nc *notionClient) SaveHighlight(databaseID string, highlight *CreateHighlight) error {
	err := nc.checkAccessToken()
	if err != nil {
		return err
	}

	selectionJSON, err := json.Marshal(highlight.Selector)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": databaseID,
		},
		"properties": map[string]interface{}{
			"Title": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": helper.Truncate(highlight.Text, 100),
						},
					},
				},
			},
			"URL": map[string]interface{}{
				"url": highlight.URL,
			},
			"Page Title": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": highlight.Title,
						},
					},
				},
			},
			"Domain": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": highlight.Domain,
						},
					},
				},
			},
			"Date": map[string]interface{}{
				"date": map[string]interface{}{
					"start": highlight.Date,
				},
			},
			"Highlight ID": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": highlight.ID,
						},
					},
				},
			},
			"Selector": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": string(selectionJSON),
						},
					},
				},
			},
		},
		"children": []map[string]interface{}{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]interface{}{
								"content": highlight.Text,
							},
						},
					},
				},
			},
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]interface{}{
								"content": "Source: " + highlight.URL,
							},
						},
					},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/pages", nc.RootURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		zlog.Logger.Error("Error creating request to Notion:", zap.Error(err))
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Logger.Error("Error making request to Notion:", zap.Error(err))
		return err
	}

	zlog.Logger.Info("Response received from Notion", zap.String("status", resp.Status))
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		zlog.Logger.Error("Bad request response from Notion", zap.ByteString("body", respBody))
		return errors.New(string(respBody))
	}

	return nil
}

func (nc *notionClient) checkAccessToken() error {
	if nc.accessToken == "" {
		access_token, err := database.SQLClient.Get("adrooney322@gmail.com")
		if err != nil {
			return err
		}
		nc.accessToken = access_token
	}
	return nil
}

func (nc *notionClient) GetHighlightByID(databaseID string, highlightID string) (*HighlightResponse, error) {
	err := nc.checkAccessToken()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Highlight ID",
			"rich_text": map[string]interface{}{
				"equals": highlightID,
			},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/databases/%s/query", nc.RootURL, databaseID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		zlog.Logger.Error("Error creating request to Notion:", zap.Error(err))
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Logger.Error("Error making request to Notion:", zap.Error(err))
		return nil, err
	}

	zlog.Logger.Info("Response received from Notion", zap.String("status", resp.Status))
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		zlog.Logger.Error("Error reading response body from Notion:", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		zlog.Logger.Error("Bad request response from Notion", zap.ByteString("body", respBody))
		return nil, errors.New(string(respBody))
	}

	var highlightResponse *HighlightResponse
	err = json.Unmarshal(respBody, &highlightResponse)
	if err != nil {
		zlog.Logger.Error("Error unmarshaling highlight response from Notion:", zap.Error(err))
		return nil, err
	}

	return highlightResponse, nil
}

func (nc *notionClient) DeleteHighlight(notionPageID string) error {
	err := nc.checkAccessToken()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"archived": true,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		zlog.Logger.Error("Error marshaling payload for deleting highlight:", zap.Error(err))
		return err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/pages/%s", nc.RootURL, notionPageID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		zlog.Logger.Error("Error creating request to Notion:", zap.Error(err))
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zlog.Logger.Error("Error making request to Notion:", zap.Error(err))
		return err
	}
	zlog.Logger.Info("Response received from Notion", zap.String("status", resp.Status))
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		zlog.Logger.Error("Bad request response from Notion", zap.ByteString("body", respBody))
		return errors.New(string(respBody))
	}
	return nil
}

func (nc *notionClient) ListDatabases() ([]interface{}, error) {
	err := nc.checkAccessToken()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "object",
			"value":    "database",
		},
		"sort": map[string]interface{}{
			"direction": "descending",
			"timestamp": "last_edited_time",
		},
	}
	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/search", nc.RootURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(respBody))
	}

	var data DatabaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Results, nil
}

func (nc *notionClient) EnsureDatabaseProperties(databaseID string) error {
	err := nc.checkAccessToken()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/databases/%s", nc.RootURL, databaseID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("could not fetch database schema")
	}

	var db map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&db)

	existingProps, ok := db["properties"].(map[string]interface{})
	if !ok {
		existingProps = make(map[string]interface{})
	}

	requiredProps := map[string]map[string]interface{}{
		"Title":        {"title": map[string]interface{}{}},
		"URL":          {"url": map[string]interface{}{}},
		"Page Title":   {"rich_text": map[string]interface{}{}},
		"Domain":       {"rich_text": map[string]interface{}{}},
		"Date":         {"date": map[string]interface{}{}},
		"Highlight ID": {"rich_text": map[string]interface{}{}},
		"Selector":     {"rich_text": map[string]interface{}{}},
	}

	missingProps := make(map[string]interface{})
	for k, v := range requiredProps {
		if _, exists := existingProps[k]; !exists {
			missingProps[k] = v
		}
	}

	if len(missingProps) == 0 {
		return nil
	}

	updatePayload := map[string]interface{}{"properties": missingProps}
	bodyBytes, _ := json.Marshal(updatePayload)
	patchReq, err := http.NewRequest("PATCH", fmt.Sprintf("%s/databases/%s", nc.RootURL, databaseID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+nc.accessToken)
	patchReq.Header.Set("Notion-Version", "2022-06-28")

	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		return err
	}
	defer patchResp.Body.Close()

	if patchResp.StatusCode != http.StatusOK {
		return errors.New("could not update database properties")
	}

	return nil
}

func (nc *notionClient) CreateDatabase(parentPageID string) (map[string]interface{}, error) {
	err := nc.checkAccessToken()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"parent": map[string]interface{}{
			"type":    "page_id",
			"page_id": parentPageID,
		},
		"title": []map[string]interface{}{
			{"text": map[string]interface{}{"content": "WebNotes Highlights"}},
		},
		"properties": map[string]interface{}{
			"Title":        map[string]interface{}{"title": map[string]interface{}{}},
			"URL":          map[string]interface{}{"url": map[string]interface{}{}},
			"Page Title":   map[string]interface{}{"rich_text": map[string]interface{}{}},
			"Domain":       map[string]interface{}{"rich_text": map[string]interface{}{}},
			"Date":         map[string]interface{}{"date": map[string]interface{}{}},
			"Highlight ID": map[string]interface{}{"rich_text": map[string]interface{}{}},
			"Selector":     map[string]interface{}{"rich_text": map[string]interface{}{}},
		},
	}
	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/databases", nc.RootURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(respBody))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}

func (nc *notionClient) CreatePage(title string) (map[string]interface{}, error) {
	err := nc.checkAccessToken()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"parent": map[string]interface{}{
			"type":      "workspace",
			"workspace": true,
		},
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]interface{}{"content": title}},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/pages", nc.RootURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(respBody))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}

func (nc *notionClient) SearchPages() ([]interface{}, error) {
	err := nc.checkAccessToken()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "object",
			"value":    "page",
		},
		"page_size": 1,
	}
	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/search", nc.RootURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(respBody))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	results, _ := data["results"].([]interface{})
	return results, nil
}

func (nc *notionClient) QueryDatabase(databaseID string, startCursor *string) (*HighlightResponse, error) {
	err := nc.checkAccessToken()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"sorts": []map[string]interface{}{
			{
				"property":  "Date",
				"direction": "descending",
			},
		},
		"page_size": 100,
	}
	if startCursor != nil {
		payload["start_cursor"] = *startCursor
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/databases/%s/query", nc.RootURL, databaseID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nc.accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(respBody))
	}

	var hr HighlightResponse
	if err := json.NewDecoder(resp.Body).Decode(&hr); err != nil {
		return nil, err
	}
	return &hr, nil
}

