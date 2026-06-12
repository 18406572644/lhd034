package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type PlaythroughController struct{}

func NewPlaythroughController() *PlaythroughController {
	return &PlaythroughController{}
}

func (ctrl *PlaythroughController) GetList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	cartridgeId := c.Query("cartridgeId")
	year := c.Query("year")
	difficulty := c.Query("difficulty")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.Playthrough{})

	if cartridgeId != "" {
		query = query.Where("cartridge_id = ?", cartridgeId)
	}
	if year != "" {
		query = query.Where("strftime('%Y', completion_date) = ?", year)
	}
	if difficulty != "" {
		query = query.Where("difficulty_rating = ?", difficulty)
	}

	var total int64
	query.Count(&total)

	var playthroughs []models.Playthrough
	query.Preload("Cartridge").
		Order("completion_date DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&playthroughs)

	utils.SuccessPaged(c, playthroughs, total, page, pageSize)
}

func (ctrl *PlaythroughController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var playthrough models.Playthrough
	result := database.DB.Preload("Cartridge").First(&playthrough, id)
	if result.Error != nil {
		utils.NotFound(c, "通关记录不存在")
		return
	}

	utils.Success(c, playthrough)
}

func (ctrl *PlaythroughController) GetByCartridge(c *gin.Context) {
	cartridgeId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的卡带ID")
		return
	}

	var playthroughs []models.Playthrough
	database.DB.Where("cartridge_id = ?", cartridgeId).
		Order("completion_date DESC").
		Find(&playthroughs)

	utils.Success(c, playthroughs)
}

func (ctrl *PlaythroughController) Create(c *gin.Context) {
	var playthrough models.Playthrough
	if err := c.ShouldBindJSON(&playthrough); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	playthrough.CreatedAt = time.Now()

	result := database.DB.Create(&playthrough)
	if result.Error != nil {
		utils.InternalError(c, "创建失败: "+result.Error.Error())
		return
	}

	utils.Success(c, playthrough)
}

func (ctrl *PlaythroughController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var existing models.Playthrough
	if err := database.DB.First(&existing, id).Error; err != nil {
		utils.NotFound(c, "通关记录不存在")
		return
	}

	var playthrough models.Playthrough
	if err := c.ShouldBindJSON(&playthrough); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	playthrough.ID = uint(id)
	playthrough.CreatedAt = existing.CreatedAt

	result := database.DB.Save(&playthrough)
	if result.Error != nil {
		utils.InternalError(c, "更新失败: "+result.Error.Error())
		return
	}

	utils.Success(c, playthrough)
}

func (ctrl *PlaythroughController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	result := database.DB.Delete(&models.Playthrough{}, id)
	if result.Error != nil {
		utils.InternalError(c, "删除失败: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFound(c, "通关记录不存在")
		return
	}

	utils.Success(c, nil)
}
