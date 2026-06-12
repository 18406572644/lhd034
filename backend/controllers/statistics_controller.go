package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StatisticsController struct{}

func NewStatisticsController() *StatisticsController {
	return &StatisticsController{}
}

type OverviewStats struct {
	TotalCartridges    int64   `json:"totalCartridges"`
	TotalPlaythroughs  int64   `json:"totalPlaythroughs"`
	TotalPlayTime      float64 `json:"totalPlayTime"`
	WishlistCount      int64   `json:"wishlistCount"`
	BorrowedCount      int64   `json:"borrowedCount"`
	TotalValue         float64 `json:"totalValue"`
	NewThisMonth       int64   `json:"newThisMonth"`
	CompletedThisMonth int64   `json:"completedThisMonth"`
}

type MonthlyData struct {
	Month        string  `json:"month"`
	Count        int     `json:"count"`
	PlayTimeHours float64 `json:"playTimeHours"`
}

type PlatformStat struct {
	Platform string `json:"platform"`
	Count    int64  `json:"count"`
}

type PublisherStat struct {
	Publisher string `json:"publisher"`
	Count     int64  `json:"count"`
}

type ConditionStat struct {
	Condition string `json:"condition"`
	Count     int64  `json:"count"`
}

func (ctrl *StatisticsController) GetOverview(c *gin.Context) {
	var stats OverviewStats

	database.DB.Model(&models.Cartridge{}).Count(&stats.TotalCartridges)
	database.DB.Model(&models.Playthrough{}).Count(&stats.TotalPlaythroughs)
	database.DB.Model(&models.WishlistItem{}).Count(&stats.WishlistCount)
	database.DB.Model(&models.BorrowRecord{}).Where("status = ?", "borrowed").Or("status = ?", "overdue").Count(&stats.BorrowedCount)

	database.DB.Model(&models.Playthrough{}).Select("COALESCE(SUM(play_time_hours), 0)").Scan(&stats.TotalPlayTime)
	database.DB.Model(&models.Cartridge{}).Select("COALESCE(SUM(purchase_price), 0)").Scan(&stats.TotalValue)

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	database.DB.Model(&models.Cartridge{}).Where("created_at >= ?", monthStart).Count(&stats.NewThisMonth)
	database.DB.Model(&models.Playthrough{}).Where("completion_date >= ?", monthStart).Count(&stats.CompletedThisMonth)

	utils.Success(c, stats)
}

func (ctrl *StatisticsController) GetAnnual(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		year = time.Now().Year()
	}

	monthlyData := make([]MonthlyData, 12)
	for i := 0; i < 12; i++ {
		month := i + 1
		monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		nextMonth := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")

		var count int64
		database.DB.Model(&models.Playthrough{}).
			Where("completion_date >= ? AND completion_date < ?", monthStart, nextMonth).
			Count(&count)

		var playTime float64
		database.DB.Model(&models.Playthrough{}).
			Where("completion_date >= ? AND completion_date < ?", monthStart, nextMonth).
			Select("COALESCE(SUM(play_time_hours), 0)").
			Scan(&playTime)

		monthlyData[i] = MonthlyData{
			Month:         time.Month(month).String(),
			Count:         int(count),
			PlayTimeHours: playTime,
		}
	}

	utils.Success(c, monthlyData)
}

func (ctrl *StatisticsController) GetPlatforms(c *gin.Context) {
	var stats []PlatformStat
	database.DB.Model(&models.Cartridge{}).
		Select("platform, COUNT(*) as count").
		Group("platform").
		Order("count DESC").
		Scan(&stats)
	utils.Success(c, stats)
}

func (ctrl *StatisticsController) GetPublishers(c *gin.Context) {
	var stats []PublisherStat
	database.DB.Model(&models.Cartridge{}).
		Select("publisher, COUNT(*) as count").
		Group("publisher").
		Order("count DESC").
		Limit(10).
		Scan(&stats)
	utils.Success(c, stats)
}

func (ctrl *StatisticsController) GetConditions(c *gin.Context) {
	var stats []ConditionStat
	database.DB.Model(&models.Cartridge{}).
		Select("condition, COUNT(*) as count").
		Group("condition").
		Scan(&stats)
	utils.Success(c, stats)
}
