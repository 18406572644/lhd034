package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type BorrowController struct{}

func NewBorrowController() *BorrowController {
	return &BorrowController{}
}

func (ctrl *BorrowController) GetList(c *gin.Context) {
	status := c.Query("status")
	cartridgeId := c.Query("cartridgeId")

	query := database.DB.Model(&models.BorrowRecord{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if cartridgeId != "" {
		query = query.Where("cartridge_id = ?", cartridgeId)
	}

	var records []models.BorrowRecord
	query.Preload("Cartridge").
		Order("created_at DESC").
		Find(&records)

	utils.Success(c, records)
}

func (ctrl *BorrowController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var record models.BorrowRecord
	result := database.DB.Preload("Cartridge").First(&record, id)
	if result.Error != nil {
		utils.NotFound(c, "借还记录不存在")
		return
	}

	utils.Success(c, record)
}

func (ctrl *BorrowController) Create(c *gin.Context) {
	var record models.BorrowRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	record.Status = "borrowed"
	record.CreatedAt = time.Now()

	result := database.DB.Create(&record)
	if result.Error != nil {
		utils.InternalError(c, "创建失败: "+result.Error.Error())
		return
	}

	utils.Success(c, record)
}

func (ctrl *BorrowController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var existing models.BorrowRecord
	if err := database.DB.First(&existing, id).Error; err != nil {
		utils.NotFound(c, "借还记录不存在")
		return
	}

	var record models.BorrowRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	record.ID = uint(id)
	record.CreatedAt = existing.CreatedAt

	result := database.DB.Save(&record)
	if result.Error != nil {
		utils.InternalError(c, "更新失败: "+result.Error.Error())
		return
	}

	utils.Success(c, record)
}

func (ctrl *BorrowController) Return(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var record models.BorrowRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		utils.NotFound(c, "借还记录不存在")
		return
	}

	now := time.Now().Format("2006-01-02")
	record.ActualReturnDate = &now
	record.Status = "returned"

	result := database.DB.Save(&record)
	if result.Error != nil {
		utils.InternalError(c, "更新失败: "+result.Error.Error())
		return
	}

	utils.Success(c, record)
}

func (ctrl *BorrowController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	result := database.DB.Delete(&models.BorrowRecord{}, id)
	if result.Error != nil {
		utils.InternalError(c, "删除失败: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFound(c, "借还记录不存在")
		return
	}

	utils.Success(c, nil)
}

func (ctrl *BorrowController) UpdateOverdueStatus() {
	now := time.Now().Format("2006-01-02")
	database.DB.Model(&models.BorrowRecord{}).
		Where("status = ? AND expected_return_date < ?", "borrowed", now).
		Update("status", "overdue")
}
