package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DataQualityReport struct {
	TotalCartridges    int64              `json:"totalCartridges"`
	CompletenessScore  float64            `json:"completenessScore"`
	DuplicateCount     int                `json:"duplicateCount"`
	DuplicateGroups    []DuplicateGroup   `json:"duplicateGroups"`
	MissingFields      []MissingFieldStat `json:"missingFields"`
	Anomalies          []AnomalyItem      `json:"anomalies"`
	AnomalyCount       int                `json:"anomalyCount"`
	LastScanTime       string             `json:"lastScanTime"`
}

type DuplicateGroup struct {
	Key        string             `json:"key"`
	Title      string             `json:"title"`
	Platform   string             `json:"platform"`
	ReleaseYear int               `json:"releaseYear"`
	Count      int                `json:"count"`
	Cartridges []models.Cartridge `json:"cartridges"`
}

type MissingFieldStat struct {
	Field    string `json:"field"`
	Label    string `json:"label"`
	MissingCount int64 `json:"missingCount"`
	MissingRate float64 `json:"missingRate"`
}

type AnomalyItem struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Field    string `json:"field"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Reason   string `json:"reason"`
	Severity string `json:"severity"`
}

type DataQualityController struct{}

func NewDataQualityController() *DataQualityController {
	return &DataQualityController{}
}

func (ctrl *DataQualityController) GetQualityReport(c *gin.Context) {
	var total int64
	database.DB.Model(&models.Cartridge{}).Count(&total)

	if total == 0 {
		utils.Success(c, DataQualityReport{
			TotalCartridges:   0,
			CompletenessScore: 100,
			LastScanTime:      time.Now().Format(time.RFC3339),
		})
		return
	}

	duplicateGroups := findDuplicates()
	missingFields := checkMissingFields(total)
	anomalies := checkAnomalies()

	completenessScore := calculateCompletenessScore(total, missingFields, anomalies)

	duplicateCount := 0
	for _, group := range duplicateGroups {
		duplicateCount += group.Count - 1
	}

	report := DataQualityReport{
		TotalCartridges:   total,
		CompletenessScore: completenessScore,
		DuplicateCount:    duplicateCount,
		DuplicateGroups:   duplicateGroups,
		MissingFields:     missingFields,
		Anomalies:         anomalies,
		AnomalyCount:      len(anomalies),
		LastScanTime:      time.Now().Format(time.RFC3339),
	}

	utils.Success(c, report)
}

func (ctrl *DataQualityController) ScanDuplicates(c *gin.Context) {
	groups := findDuplicates()
	utils.Success(c, groups)
}

func (ctrl *DataQualityController) ScanMissingFields(c *gin.Context) {
	var total int64
	database.DB.Model(&models.Cartridge{}).Count(&total)
	fields := checkMissingFields(total)
	utils.Success(c, fields)
}

func (ctrl *DataQualityController) ScanAnomalies(c *gin.Context) {
	anomalies := checkAnomalies()
	utils.Success(c, anomalies)
}

func (ctrl *DataQualityController) FixDuplicates(c *gin.Context) {
	var req struct {
		KeepID uint   `json:"keepId" binding:"required"`
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var keepCartridge models.Cartridge
	if err := database.DB.First(&keepCartridge, req.KeepID).Error; err != nil {
		utils.NotFound(c, "保留的卡带不存在")
		return
	}

	groups := findDuplicates()
	var targetGroup *DuplicateGroup
	for i := range groups {
		for _, cartridge := range groups[i].Cartridges {
			if cartridge.ID == req.KeepID {
				targetGroup = &groups[i]
				break
			}
		}
		if targetGroup != nil {
			break
		}
	}

	if targetGroup == nil {
		utils.BadRequest(c, "未找到重复组")
		return
	}

	deletedCount := 0
	for _, cartridge := range targetGroup.Cartridges {
		if cartridge.ID != req.KeepID {
			database.DB.Delete(&models.Cartridge{}, cartridge.ID)
			deletedCount++
		}
	}

	utils.Success(c, gin.H{
		"deletedCount": deletedCount,
		"keptId":       req.KeepID,
	})
}

func (ctrl *DataQualityController) FixMissingFields(c *gin.Context) {
	var req struct {
		Field   string      `json:"field" binding:"required"`
		Value   interface{} `json:"value"`
		IDs     []uint      `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	query := database.DB.Model(&models.Cartridge{})
	if len(req.IDs) > 0 {
		query = query.Where("id IN ?", req.IDs)
	} else {
		switch req.Field {
		case "platform":
			query = query.Where("platform = '' OR platform IS NULL")
		case "publisher":
			query = query.Where("publisher = '' OR publisher IS NULL")
		case "region":
			query = query.Where("region = '' OR region IS NULL")
		case "condition":
			query = query.Where("condition = '' OR condition IS NULL")
		}
	}

	result := query.Update(req.Field, req.Value)

	utils.Success(c, gin.H{
		"updatedCount": result.RowsAffected,
		"field":        req.Field,
		"value":        req.Value,
	})
}

func findDuplicates() []DuplicateGroup {
	var cartridges []models.Cartridge
	database.DB.Find(&cartridges)

	groupMap := make(map[string]*DuplicateGroup)

	for _, cartridge := range cartridges {
		normalizedTitle := normalizeTitle(cartridge.Title)
		key := fmt.Sprintf("%s|%s|%d", normalizedTitle, cartridge.Platform, cartridge.ReleaseYear)

		if group, exists := groupMap[key]; exists {
			group.Count++
			group.Cartridges = append(group.Cartridges, cartridge)
		} else {
			groupMap[key] = &DuplicateGroup{
				Key:         key,
				Title:       cartridge.Title,
				Platform:    cartridge.Platform,
				ReleaseYear: cartridge.ReleaseYear,
				Count:       1,
				Cartridges:  []models.Cartridge{cartridge},
			}
		}
	}

	var groups []DuplicateGroup
	for _, group := range groupMap {
		if group.Count > 1 {
			groups = append(groups, *group)
		}
	}

	return groups
}

func normalizeTitle(title string) string {
	title = strings.TrimSpace(strings.ToLower(title))
	title = strings.ReplaceAll(title, " ", "")
	title = strings.ReplaceAll(title, "-", "")
	title = strings.ReplaceAll(title, "_", "")
	title = strings.ReplaceAll(title, "：", "")
	title = strings.ReplaceAll(title, ":", "")
	return title
}

func checkMissingFields(total int64) []MissingFieldStat {
	fields := []struct {
		Field string
		Label string
		Query string
	}{
		{"title", "标题", "title = '' OR title IS NULL"},
		{"platform", "平台", "platform = '' OR platform IS NULL"},
		{"publisher", "发行商", "publisher = '' OR publisher IS NULL"},
		{"releaseYear", "发行年份", "release_year = 0 OR release_year IS NULL"},
		{"condition", "品相", "condition = '' OR condition IS NULL"},
		{"region", "区域", "region = '' OR region IS NULL"},
		{"purchasePrice", "购买价格", "purchase_price = 0 OR purchase_price IS NULL"},
	}

	var result []MissingFieldStat
	for _, f := range fields {
		var count int64
		database.DB.Model(&models.Cartridge{}).Where(f.Query).Count(&count)

		rate := 0.0
		if total > 0 {
			rate = float64(count) / float64(total) * 100
		}

		result = append(result, MissingFieldStat{
			Field:        f.Field,
			Label:        f.Label,
			MissingCount: count,
			MissingRate:  rate,
		})
	}

	return result
}

func checkAnomalies() []AnomalyItem {
	var cartridges []models.Cartridge
	database.DB.Find(&cartridges)

	var anomalies []AnomalyItem
	currentYear := time.Now().Year()

	for _, c := range cartridges {
		if c.ReleaseYear < 1950 || c.ReleaseYear > currentYear {
			severity := "high"
			if c.ReleaseYear == 0 {
				severity = "medium"
			}
			anomalies = append(anomalies, AnomalyItem{
				ID:       c.ID,
				Title:    c.Title,
				Field:    "releaseYear",
				Label:    "发行年份",
				Value:    fmt.Sprintf("%d", c.ReleaseYear),
				Reason:   fmt.Sprintf("年份不在合理范围内 (1950-%d)", currentYear),
				Severity: severity,
			})
		}

		if c.PurchasePrice > 100000 {
			anomalies = append(anomalies, AnomalyItem{
				ID:       c.ID,
				Title:    c.Title,
				Field:    "purchasePrice",
				Label:    "购买价格",
				Value:    fmt.Sprintf("%.2f", c.PurchasePrice),
				Reason:   "价格超过10万，可能是输入错误",
				Severity: "high",
			})
		}

		if c.PurchasePrice < 0 {
			anomalies = append(anomalies, AnomalyItem{
				ID:       c.ID,
				Title:    c.Title,
				Field:    "purchasePrice",
				Label:    "购买价格",
				Value:    fmt.Sprintf("%.2f", c.PurchasePrice),
				Reason:   "价格为负数",
				Severity: "high",
			})
		}

		validConditions := map[string]bool{
			"mint": true, "excellent": true, "good": true, "fair": true, "poor": true,
		}
		if c.Condition != "" && !validConditions[c.Condition] {
			anomalies = append(anomalies, AnomalyItem{
				ID:       c.ID,
				Title:    c.Title,
				Field:    "condition",
				Label:    "品相",
				Value:    c.Condition,
				Reason:   "品相值不在有效范围内",
				Severity: "medium",
			})
		}

		validStatuses := map[string]bool{
			"unstarted": true, "playing": true, "completed": true, "shelved": true,
		}
		if c.Status != "" && !validStatuses[c.Status] {
			anomalies = append(anomalies, AnomalyItem{
				ID:       c.ID,
				Title:    c.Title,
				Field:    "status",
				Label:    "游玩状态",
				Value:    c.Status,
				Reason:   "状态值不在有效范围内",
				Severity: "medium",
			})
		}
	}

	return anomalies
}

func calculateCompletenessScore(total int64, missingFields []MissingFieldStat, anomalies []AnomalyItem) float64 {
	if total == 0 {
		return 100
	}

	totalFields := float64(total) * float64(len(missingFields))
	totalMissing := float64(0)
	for _, f := range missingFields {
		totalMissing += float64(f.MissingCount)
	}

	completeness := 100.0
	if totalFields > 0 {
		completeness = (1 - totalMissing/totalFields) * 70
	}

	anomalyPenalty := float64(len(anomalies)) / float64(total) * 30
	score := completeness - anomalyPenalty

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return mathRound(score, 1)
}

func mathRound(val float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(val*multiplier+0.5)) / multiplier
}
