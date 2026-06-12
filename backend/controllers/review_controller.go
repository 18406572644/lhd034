package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReviewController struct{}

func NewReviewController() *ReviewController {
	return &ReviewController{}
}

func (ctrl *ReviewController) GetList(c *gin.Context) {
	var reviews []models.Review
	database.DB.Preload("Cartridge").Order("created_at DESC").Find(&reviews)
	utils.Success(c, reviews)
}

func (ctrl *ReviewController) GetByCartridge(c *gin.Context) {
	cartridgeId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的卡带ID")
		return
	}

	var review models.Review
	result := database.DB.Where("cartridge_id = ?", cartridgeId).First(&review)
	if result.Error != nil {
		utils.Success(c, nil)
		return
	}

	utils.Success(c, review)
}

func (ctrl *ReviewController) Create(c *gin.Context) {
	var review models.Review
	if err := c.ShouldBindJSON(&review); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	review.CreatedAt = time.Now()
	review.OverallRating = calculateOverallRating(
		review.ContentRating,
		review.GameplayRating,
		review.GraphicsRating,
		review.SoundRating,
	)

	result := database.DB.Create(&review)
	if result.Error != nil {
		utils.InternalError(c, "创建失败: "+result.Error.Error())
		return
	}

	utils.Success(c, review)
}

func (ctrl *ReviewController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var existing models.Review
	if err := database.DB.First(&existing, id).Error; err != nil {
		utils.NotFound(c, "评价不存在")
		return
	}

	var review models.Review
	if err := c.ShouldBindJSON(&review); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	review.ID = uint(id)
	review.CreatedAt = existing.CreatedAt
	review.CartridgeID = existing.CartridgeID
	review.OverallRating = calculateOverallRating(
		review.ContentRating,
		review.GameplayRating,
		review.GraphicsRating,
		review.SoundRating,
	)

	result := database.DB.Save(&review)
	if result.Error != nil {
		utils.InternalError(c, "更新失败: "+result.Error.Error())
		return
	}

	utils.Success(c, review)
}

func calculateOverallRating(content, gameplay, graphics, sound int) float64 {
	sum := content + gameplay + graphics + sound
	return float64(sum) / 4.0
}
