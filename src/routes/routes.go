package routes

import (
	"net/http"

	"github.com/Atoo35/WebNotes-backend/src/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1Routes := r.Group("/v1")

	v1Routes.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/notion/save-highlight", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	notionRoutes := v1Routes.Group("/notion")
	{
		notionRoutes.GET("/start-oauth", controllers.NotionController.StartOAuthFlow)
		notionRoutes.GET("/get-token", controllers.NotionController.GetToken)
		notionRoutes.GET("/load-highlights-by-domain", controllers.NotionController.LoadHighlightsByDomain)
		notionRoutes.GET("/get-highlight-by-id", controllers.NotionController.GetHighlightByID)

		notionRoutes.POST("/save-highlight", controllers.NotionController.SaveHighlight)
		notionRoutes.DELETE("/delete-highlight", controllers.NotionController.DeleteHighlight)

		notionRoutes.GET("/databases/list", controllers.NotionController.ListDatabases)
		notionRoutes.POST("/databases/ensure-properties", controllers.NotionController.EnsureDatabaseProperties)
		notionRoutes.POST("/databases/create-auto", controllers.NotionController.CreateDatabaseAuto)
		notionRoutes.POST("/databases/create", controllers.NotionController.CreateDatabase)
		notionRoutes.GET("/highlights/all", controllers.NotionController.LoadAllHighlights)
	}
}
