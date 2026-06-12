package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type BatchUpdateRequest struct {
	IDs       []uint                 `json:"ids" binding:"required"`
	Fields    map[string]interface{} `json:"fields" binding:"required"`
	Mode      string                 `json:"mode"`
	Increment float64                `json:"increment"`
}

type BatchUpdatePreview struct {
	TotalRecords int                    `json:"totalRecords"`
	FieldChanges []FieldChangePreview   `json:"fieldChanges"`
	SampleBefore []models.Cartridge     `json:"sampleBefore"`
	SampleAfter  []map[string]interface{} `json:"sampleAfter"`
}

type FieldChangePreview struct {
	Field       string      `json:"field"`
	Label       string      `json:"label"`
	BeforeValue interface{} `json:"beforeValue"`
	AfterValue  interface{} `json:"afterValue"`
	ChangeCount int         `json:"changeCount"`
}

type BatchActionRequest struct {
	IDs          []uint   `json:"ids" binding:"required"`
	CollectionID uint     `json:"collectionId"`
	Tags         []string `json:"tags"`
	Action       string   `json:"action"`
}

type BatchController struct{}

func NewBatchController() *BatchController {
	return &BatchController{}
}

func (ctrl *BatchController) BatchUpdatePreview(c *gin.Context) {
	var req BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequest(c, "请选择要修改的卡带")
		return
	}

	var cartridges []models.Cartridge
	database.DB.Where("id IN ?", req.IDs).Find(&cartridges)

	if len(cartridges) == 0 {
		utils.NotFound(c, "未找到匹配的卡带")
		return
	}

	fieldLabels := map[string]string{
		"platform":      "平台",
		"condition":     "品相",
		"region":        "区域",
		"purchasePrice": "购买价格",
		"publisher":     "发行商",
		"releaseYear":   "发行年份",
		"status":        "游玩状态",
		"notes":         "备注",
	}

	var fieldChanges []FieldChangePreview
	for field, value := range req.Fields {
		changeCount := 0
		var beforeSample interface{}
		var afterSample interface{}

		for i, cartridge := range cartridges {
			beforeVal := getFieldValue(&cartridge, field)
			afterVal := calculateNewValue(beforeVal, value, req.Mode, req.Increment)

			if i == 0 {
				beforeSample = beforeVal
				afterSample = afterVal
			}

			if fmt.Sprintf("%v", beforeVal) != fmt.Sprintf("%v", afterVal) {
				changeCount++
			}
		}

		label := fieldLabels[field]
		if label == "" {
			label = field
		}

		fieldChanges = append(fieldChanges, FieldChangePreview{
			Field:       field,
			Label:       label,
			BeforeValue: beforeSample,
			AfterValue:  afterSample,
			ChangeCount: changeCount,
		})
	}

	sampleCount := 3
	if len(cartridges) < 3 {
		sampleCount = len(cartridges)
	}
	sampleBefore := cartridges[:sampleCount]

	var sampleAfter []map[string]interface{}
	for i := 0; i < sampleCount; i++ {
		after := make(map[string]interface{})
		for field, value := range req.Fields {
			beforeVal := getFieldValue(&cartridges[i], field)
			afterVal := calculateNewValue(beforeVal, value, req.Mode, req.Increment)
			after[field] = afterVal
		}
		sampleAfter = append(sampleAfter, after)
	}

	preview := BatchUpdatePreview{
		TotalRecords: len(cartridges),
		FieldChanges: fieldChanges,
		SampleBefore: sampleBefore,
		SampleAfter:  sampleAfter,
	}

	utils.Success(c, preview)
}

func (ctrl *BatchController) BatchUpdate(c *gin.Context) {
	var req BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequest(c, "请选择要修改的卡带")
		return
	}

	var cartridges []models.Cartridge
	database.DB.Where("id IN ?", req.IDs).Find(&cartridges)

	if len(cartridges) == 0 {
		utils.NotFound(c, "未找到匹配的卡带")
		return
	}

	now := time.Now()
	updatedCount := 0

	for i := range cartridges {
		changed := false
		updateFields := make(map[string]interface{})

		for field, value := range req.Fields {
			beforeVal := getFieldValue(&cartridges[i], field)
			afterVal := calculateNewValue(beforeVal, value, req.Mode, req.Increment)

			if fmt.Sprintf("%v", beforeVal) != fmt.Sprintf("%v", afterVal) {
				setFieldValue(&cartridges[i], field, afterVal)
				dbColumn := getDBColumnName(field)
				updateFields[dbColumn] = getFieldValue(&cartridges[i], field)
				changed = true
			}
		}

		if changed {
			updateFields["updated_at"] = now
			result := database.DB.Model(&models.Cartridge{}).Where("id = ?", cartridges[i].ID).Updates(updateFields)
			if result.Error == nil {
				updatedCount++
			}
		}
	}

	utils.Success(c, gin.H{
		"updatedCount": updatedCount,
		"totalCount":   len(cartridges),
	})
}

func (ctrl *BatchController) BatchAddToWishlist(c *gin.Context) {
	var req BatchActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequest(c, "请选择要添加的卡带")
		return
	}

	addedCount := 0
	skippedCount := 0
	now := time.Now()

	for _, id := range req.IDs {
		var existing models.WishlistItem
		result := database.DB.Where("cartridge_id = ?", id).First(&existing)
		if result.Error == nil {
			skippedCount++
			continue
		}

		item := models.WishlistItem{
			CartridgeID:      id,
			Priority:         "medium",
			Tags:             models.StringArray(req.Tags),
			AddedAt:          now,
		}

		if database.DB.Create(&item).Error == nil {
			addedCount++
		}
	}

	utils.Success(c, gin.H{
		"addedCount":   addedCount,
		"skippedCount": skippedCount,
		"totalCount":   len(req.IDs),
	})
}

func (ctrl *BatchController) BatchSetTags(c *gin.Context) {
	var req BatchActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequest(c, "请选择要设置标签的卡带")
		return
	}

	updatedCount := 0

	for _, id := range req.IDs {
		var wishlist models.WishlistItem
		result := database.DB.Where("cartridge_id = ?", id).First(&wishlist)
		if result.Error != nil {
			continue
		}

		if req.Action == "append" {
			existingTags := make(map[string]bool)
			for _, tag := range wishlist.Tags {
				existingTags[tag] = true
			}
			for _, tag := range req.Tags {
				if !existingTags[tag] {
					wishlist.Tags = append(wishlist.Tags, tag)
				}
			}
		} else {
			wishlist.Tags = models.StringArray(req.Tags)
		}

		if database.DB.Save(&wishlist).Error == nil {
			updatedCount++
		}
	}

	utils.Success(c, gin.H{
		"updatedCount": updatedCount,
		"totalCount":   len(req.IDs),
	})
}

func (ctrl *BatchController) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequest(c, "请选择要删除的卡带")
		return
	}

	result := database.DB.Delete(&models.Cartridge{}, req.IDs)

	utils.Success(c, gin.H{
		"deletedCount": result.RowsAffected,
		"totalCount":   len(req.IDs),
	})
}

func getDBColumnName(field string) string {
	switch field {
	case "platform":
		return "platform"
	case "condition":
		return "condition"
	case "region":
		return "region"
	case "purchasePrice":
		return "purchase_price"
	case "publisher":
		return "publisher"
	case "releaseYear":
		return "release_year"
	case "status":
		return "status"
	case "notes":
		return "notes"
	default:
		return field
	}
}

func getFieldValue(cartridge *models.Cartridge, field string) interface{} {
	switch field {
	case "platform":
		return cartridge.Platform
	case "condition":
		return cartridge.Condition
	case "region":
		return cartridge.Region
	case "purchasePrice":
		return cartridge.PurchasePrice
	case "publisher":
		return cartridge.Publisher
	case "releaseYear":
		return cartridge.ReleaseYear
	case "status":
		return cartridge.Status
	case "notes":
		return cartridge.Notes
	default:
		return nil
	}
}

func setFieldValue(cartridge *models.Cartridge, field string, value interface{}) {
	switch field {
	case "platform":
		cartridge.Platform = toString(value)
	case "condition":
		cartridge.Condition = toString(value)
	case "region":
		cartridge.Region = toString(value)
	case "purchasePrice":
		cartridge.PurchasePrice = toFloat64(value)
	case "publisher":
		cartridge.Publisher = toString(value)
	case "releaseYear":
		cartridge.ReleaseYear = toInt(value)
	case "status":
		cartridge.Status = toString(value)
	case "notes":
		cartridge.Notes = toString(value)
	}
}

func calculateNewValue(beforeVal interface{}, newValue interface{}, mode string, increment float64) interface{} {
	switch mode {
	case "increment":
		beforeFloat := toFloat64(beforeVal)
		return beforeFloat + increment
	case "percentage":
		beforeFloat := toFloat64(beforeVal)
		return beforeFloat * (1 + increment/100)
	case "append":
		beforeStr := toString(beforeVal)
		newStr := toString(newValue)
		if beforeStr == "" {
			return newStr
		}
		return beforeStr + ", " + newStr
	case "overwrite":
		fallthrough
	default:
		return newValue
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case string:
		var result float64
		fmt.Sscanf(val, "%f", &result)
		return result
	default:
		return 0
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	case uint:
		return int(val)
	case string:
		var result int
		fmt.Sscanf(val, "%d", &result)
		return result
	default:
		return 0
	}
}
