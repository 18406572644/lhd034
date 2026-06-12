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

type RatingStat struct {
	Rating float64 `json:"rating"`
	Count  int64   `json:"count"`
	Label  string  `json:"label"`
}

type PlayTimeRankItem struct {
	ID            uint    `json:"id"`
	CartridgeID   uint    `json:"cartridgeId"`
	Title         string  `json:"title"`
	Platform      string  `json:"platform"`
	PlayTimeHours float64 `json:"playTimeHours"`
	CompletionDate string `json:"completionDate"`
}

type DifficultyStat struct {
	Difficulty       int     `json:"difficulty"`
	Label            string  `json:"label"`
	Count            int64   `json:"count"`
	AvgPlayTimeHours float64 `json:"avgPlayTimeHours"`
}

type ValueTrendItem struct {
	Date         string  `json:"date"`
	Value        float64 `json:"value"`
	Cumulative   float64 `json:"cumulative"`
}

type RegionStat struct {
	Region string `json:"region"`
	Count  int64  `json:"count"`
	Label  string `json:"label"`
}

type CompletionRate struct {
	TotalCartridges     int64   `json:"totalCartridges"`
	CompletedCartridges int64   `json:"completedCartridges"`
	Rate                float64 `json:"rate"`
	PlayingCount        int64   `json:"playingCount"`
	UnstartedCount      int64   `json:"unstartedCount"`
	ShelvedCount        int64   `json:"shelvedCount"`
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

func (ctrl *StatisticsController) GetRatingDistribution(c *gin.Context) {
	ratingRanges := []struct {
		Min   float64
		Max   float64
		Label string
	}{
		{0.5, 1.5, "1星"},
		{1.5, 2.5, "2星"},
		{2.5, 3.5, "3星"},
		{3.5, 4.5, "4星"},
		{4.5, 5.5, "5星"},
	}

	result := make([]RatingStat, 0, len(ratingRanges))
	for _, r := range ratingRanges {
		var count int64
		database.DB.Model(&models.Review{}).
			Where("overall_rating >= ? AND overall_rating < ?", r.Min, r.Max).
			Count(&count)
		result = append(result, RatingStat{
			Rating: (r.Min + r.Max) / 2,
			Count:  count,
			Label:  r.Label,
		})
	}
	utils.Success(c, result)
}

func (ctrl *StatisticsController) GetPlayTimeTop10(c *gin.Context) {
	var rankings []PlayTimeRankItem
	database.DB.Model(&models.Playthrough{}).
		Select("playthroughs.id, playthroughs.cartridge_id, cartridges.title, cartridges.platform, playthroughs.play_time_hours, playthroughs.completion_date").
		Joins("LEFT JOIN cartridges ON cartridges.id = playthroughs.cartridge_id").
		Order("playthroughs.play_time_hours DESC").
		Limit(10).
		Scan(&rankings)
	utils.Success(c, rankings)
}

func (ctrl *StatisticsController) GetDifficultyAnalysis(c *gin.Context) {
	difficultyLabels := map[int]string{
		1: "非常简单",
		2: "简单",
		3: "中等",
		4: "困难",
		5: "非常困难",
	}

	result := make([]DifficultyStat, 0, 5)
	for diff := 1; diff <= 5; diff++ {
		var count int64
		var avgTime float64

		database.DB.Model(&models.Playthrough{}).
			Where("difficulty_rating = ?", diff).
			Count(&count)

		if count > 0 {
			database.DB.Model(&models.Playthrough{}).
				Where("difficulty_rating = ?", diff).
				Select("COALESCE(AVG(play_time_hours), 0)").
				Scan(&avgTime)
		}

		result = append(result, DifficultyStat{
			Difficulty:       diff,
			Label:            difficultyLabels[diff],
			Count:            count,
			AvgPlayTimeHours: avgTime,
		})
	}
	utils.Success(c, result)
}

func (ctrl *StatisticsController) GetValueTrend(c *gin.Context) {
	type RawPurchase struct {
		PurchaseDate string  `gorm:"column:purchase_date"`
		PurchasePrice float64 `gorm:"column:purchase_price"`
	}

	var rawData []RawPurchase
	database.DB.Model(&models.Cartridge{}).
		Select("purchase_date, purchase_price").
		Where("purchase_date IS NOT NULL AND purchase_date != '' AND purchase_price > 0").
		Order("purchase_date ASC").
		Scan(&rawData)

	monthlyMap := make(map[string]float64)
	for _, item := range rawData {
		if len(item.PurchaseDate) >= 7 {
			monthKey := item.PurchaseDate[:7]
			monthlyMap[monthKey] += item.PurchasePrice
		}
	}

	monthKeys := make([]string, 0, len(monthlyMap))
	for k := range monthlyMap {
		monthKeys = append(monthKeys, k)
	}

	for i := 0; i < len(monthKeys)-1; i++ {
		for j := i + 1; j < len(monthKeys); j++ {
			if monthKeys[i] > monthKeys[j] {
				monthKeys[i], monthKeys[j] = monthKeys[j], monthKeys[i]
			}
		}
	}

	result := make([]ValueTrendItem, 0, len(monthKeys))
	var cumulative float64
	for _, month := range monthKeys {
		value := monthlyMap[month]
		cumulative += value
		result = append(result, ValueTrendItem{
			Date:       month,
			Value:      value,
			Cumulative: cumulative,
		})
	}

	utils.Success(c, result)
}

func (ctrl *StatisticsController) GetRegionDistribution(c *gin.Context) {
	regionLabels := map[string]string{
		"jp":    "日版",
		"japan": "日版",
		"us":    "美版",
		"usa":   "美版",
		"eu":    "欧版",
		"europe": "欧版",
		"cn":    "国行",
		"china": "国行",
		"hk":    "港版",
		"asia":  "亚洲版",
		"global": "全球版",
		"worldwide": "全球版",
	}

	var rawStats []struct {
		Region string `json:"region"`
		Count  int64  `json:"count"`
	}

	database.DB.Model(&models.Cartridge{}).
		Select("COALESCE(NULLIF(region, ''), 'unknown') as region, COUNT(*) as count").
		Group("region").
		Scan(&rawStats)

	groupedMap := make(map[string]int64)
	groupedMap["其他"] = 0

	for _, item := range rawStats {
		regionKey := item.Region
		label, exists := regionLabels[regionKey]
		if exists {
			groupedMap[label] += item.Count
		} else if regionKey == "unknown" {
			groupedMap["未指定"] = item.Count
		} else {
			groupedMap["其他"] += item.Count
		}
	}

	result := make([]RegionStat, 0, len(groupedMap))
	for region, count := range groupedMap {
		if count > 0 {
			result = append(result, RegionStat{
				Region: region,
				Count:  count,
				Label:  region,
			})
		}
	}

	utils.Success(c, result)
}

func (ctrl *StatisticsController) GetCompletionRate(c *gin.Context) {
	var result CompletionRate

	database.DB.Model(&models.Cartridge{}).Count(&result.TotalCartridges)
	database.DB.Model(&models.Cartridge{}).Where("status = ?", "completed").Count(&result.CompletedCartridges)
	database.DB.Model(&models.Cartridge{}).Where("status = ?", "playing").Count(&result.PlayingCount)
	database.DB.Model(&models.Cartridge{}).Where("status = ?", "unstarted").Count(&result.UnstartedCount)
	database.DB.Model(&models.Cartridge{}).Where("status = ?", "shelved").Count(&result.ShelvedCount)

	if result.TotalCartridges > 0 {
		result.Rate = float64(result.CompletedCartridges) / float64(result.TotalCartridges) * 100
	}

	utils.Success(c, result)
}
