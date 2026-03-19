package controllers

import (
	"net/http"

	"github.com/Atoo35/WebNotes-backend/src/domains/notion"
	"github.com/Atoo35/WebNotes-backend/src/services"
	zlog "github.com/Atoo35/WebNotes-backend/src/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NotionControllerI interface {
	StartOAuthFlow(ctx *gin.Context)
	GetToken(ctx *gin.Context)
	LoadHighlightsByDomain(ctx *gin.Context)
	SaveHighlight(ctx *gin.Context)
	GetHighlightByID(ctx *gin.Context)
	DeleteHighlight(ctx *gin.Context)
	ListDatabases(ctx *gin.Context)
	EnsureDatabaseProperties(ctx *gin.Context)
	CreateDatabaseAuto(ctx *gin.Context)
	CreateDatabase(ctx *gin.Context)
	LoadAllHighlights(ctx *gin.Context)
}

type notionController struct{}

var NotionController NotionControllerI = &notionController{}

func (nc *notionController) StartOAuthFlow(ctx *gin.Context) {

	res, err := services.NotionService.StartOAuthFlow()
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}

	ctx.JSON(200, res)
}

func (nc *notionController) GetToken(ctx *gin.Context) {
	code := ctx.Query("code")

	token, err := services.NotionService.GetToken(code)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"access_token": token})
}

func (nc *notionController) LoadHighlightsByDomain(ctx *gin.Context) {
	domain := ctx.Query("domain")
	databaseId := ctx.Query("database_id")
	highlights, err := services.NotionService.LoadHighlightsByDomain(domain, databaseId)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"highlights": highlights})
}

func (nc *notionController) SaveHighlight(ctx *gin.Context) {
	var highlight notion.CreateHighlightRequest
	if err := ctx.ShouldBindJSON(&highlight); err != nil {
		zlog.Logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := services.NotionService.SaveHighlight(highlight.DatabaseID, &highlight)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Highlight saved successfully"})
}

func (nc *notionController) GetHighlightByID(ctx *gin.Context) {
	highlightID := ctx.Query("highlight_id")
	databaseID := ctx.Query("database_id")
	highlight, err := services.NotionService.GetHighlightByID(databaseID, highlightID)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"highlight": highlight})
}

func (nc *notionController) DeleteHighlight(ctx *gin.Context) {
	notionPageID := ctx.Query("notion_page_id")
	err := services.NotionService.DeleteHighlight(notionPageID)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Highlight deleted successfully"})
}

func (nc *notionController) ListDatabases(ctx *gin.Context) {
	databases, err := services.NotionService.ListDatabases()
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"databases": databases})
}

func (nc *notionController) EnsureDatabaseProperties(ctx *gin.Context) {
	var body notion.EnsurePropertiesRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := services.NotionService.EnsureDatabaseProperties(body.DatabaseID)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Properties ensured successfully"})
}

func (nc *notionController) CreateDatabaseAuto(ctx *gin.Context) {
	res, err := services.NotionService.CreateDatabaseAuto()
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (nc *notionController) CreateDatabase(ctx *gin.Context) {
	var body notion.CreateDatabaseRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	res, err := services.NotionService.CreateDatabase(body.ParentPageID)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (nc *notionController) LoadAllHighlights(ctx *gin.Context) {
	databaseID := ctx.Query("database_id")
	highlights, err := services.NotionService.LoadAllHighlights(databaseID)
	if err != nil {
		ctx.JSON(err.Code(), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"highlights": highlights})
}
