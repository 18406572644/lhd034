package routes

import (
	"cartridge-archive/controllers"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		cartridgeCtrl := controllers.NewCartridgeController()
		playthroughCtrl := controllers.NewPlaythroughController()
		reviewCtrl := controllers.NewReviewController()
		wishlistCtrl := controllers.NewWishlistController()
		borrowCtrl := controllers.NewBorrowController()
		statsCtrl := controllers.NewStatisticsController()
		sessionCtrl := controllers.NewPlayingSessionController()

		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		cartridges := api.Group("/cartridges")
		{
			cartridges.GET("", cartridgeCtrl.GetList)
			cartridges.GET("/platforms", cartridgeCtrl.GetPlatforms)
			cartridges.GET("/publishers", cartridgeCtrl.GetPublishers)
			cartridges.GET("/:id", cartridgeCtrl.GetByID)
			cartridges.POST("", cartridgeCtrl.Create)
			cartridges.PUT("/:id", cartridgeCtrl.Update)
			cartridges.DELETE("/:id", cartridgeCtrl.Delete)
			cartridges.POST("/upload", cartridgeCtrl.Upload)
			cartridges.GET("/:id/playthroughs", playthroughCtrl.GetByCartridge)
			cartridges.GET("/:id/review", reviewCtrl.GetByCartridge)
		}

		playthroughs := api.Group("/playthroughs")
		{
			playthroughs.GET("", playthroughCtrl.GetList)
			playthroughs.GET("/:id", playthroughCtrl.GetByID)
			playthroughs.POST("", playthroughCtrl.Create)
			playthroughs.PUT("/:id", playthroughCtrl.Update)
			playthroughs.DELETE("/:id", playthroughCtrl.Delete)
		}

		reviews := api.Group("/reviews")
		{
			reviews.GET("", reviewCtrl.GetList)
			reviews.POST("", reviewCtrl.Create)
			reviews.PUT("/:id", reviewCtrl.Update)
		}

		wishlist := api.Group("/wishlist")
		{
			wishlist.GET("", wishlistCtrl.GetList)
			wishlist.POST("", wishlistCtrl.Create)
			wishlist.PUT("/:id", wishlistCtrl.Update)
			wishlist.DELETE("/:id", wishlistCtrl.Delete)
		}

		borrows := api.Group("/borrows")
		{
			borrows.GET("", borrowCtrl.GetList)
			borrows.GET("/:id", borrowCtrl.GetByID)
			borrows.POST("", borrowCtrl.Create)
			borrows.PUT("/:id", borrowCtrl.Update)
			borrows.PUT("/:id/return", borrowCtrl.Return)
			borrows.DELETE("/:id", borrowCtrl.Delete)
		}

		statistics := api.Group("/statistics")
		{
			statistics.GET("/overview", statsCtrl.GetOverview)
			statistics.GET("/annual", statsCtrl.GetAnnual)
			statistics.GET("/platforms", statsCtrl.GetPlatforms)
			statistics.GET("/publishers", statsCtrl.GetPublishers)
			statistics.GET("/conditions", statsCtrl.GetConditions)
			statistics.GET("/ratings", statsCtrl.GetRatingDistribution)
			statistics.GET("/playtime-top10", statsCtrl.GetPlayTimeTop10)
			statistics.GET("/difficulty", statsCtrl.GetDifficultyAnalysis)
			statistics.GET("/value-trend", statsCtrl.GetValueTrend)
			statistics.GET("/regions", statsCtrl.GetRegionDistribution)
			statistics.GET("/completion-rate", statsCtrl.GetCompletionRate)
		}

		sessions := api.Group("/sessions")
		{
			sessions.GET("", sessionCtrl.GetList)
			sessions.GET("/playing", sessionCtrl.GetPlayingCartridges)
			sessions.GET("/:id", sessionCtrl.GetProgress)
			sessions.POST("", sessionCtrl.Create)
			sessions.PUT("/:id", sessionCtrl.Update)
			sessions.DELETE("/:id", sessionCtrl.Delete)
		}

		cartridges.GET("/:id/sessions", sessionCtrl.GetByCartridge)
		cartridges.GET("/:id/progress", sessionCtrl.GetProgress)

		backups := api.Group("/backups")
		{
			backupCtrl := controllers.NewBackupController()
			backups.GET("", backupCtrl.ListBackups)
			backups.POST("", backupCtrl.CreateBackup)
			backups.DELETE("/:filename", backupCtrl.DeleteBackup)
			backups.POST("/:filename/restore", backupCtrl.RestoreBackup)
			backups.GET("/config", backupCtrl.GetConfig)
			backups.PUT("/config", backupCtrl.UpdateConfig)
		}
	}

	return r
}
