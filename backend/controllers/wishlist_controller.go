package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type WishlistController struct{}

func NewWishlistController() *WishlistController {
	return &WishlistController{}
}

func (ctrl *WishlistController) GetList(c *gin.Context) {
	var wishlist []models.WishlistItem
	database.DB.Preload("Cartridge").
		Order("priority DESC, added_at DESC").
		Find(&wishlist)
	utils.Success(c, wishlist)
}

func (ctrl *WishlistController) Create(c *gin.Context) {
	var item models.WishlistItem
	if err := c.ShouldBindJSON(&item); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	item.AddedAt = time.Now()

	result := database.DB.Create(&item)
	if result.Error != nil {
		utils.InternalError(c, "创建失败: "+result.Error.Error())
		return
	}

	utils.Success(c, item)
}

func (ctrl *WishlistController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var existing models.WishlistItem
	if err := database.DB.First(&existing, id).Error; err != nil {
		utils.NotFound(c, "待玩项目不存在")
		return
	}

	var item models.WishlistItem
	if err := c.ShouldBindJSON(&item); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	item.ID = uint(id)
	item.AddedAt = existing.AddedAt
	item.CartridgeID = existing.CartridgeID

	result := database.DB.Save(&item)
	if result.Error != nil {
		utils.InternalError(c, "更新失败: "+result.Error.Error())
		return
	}

	utils.Success(c, item)
}

func (ctrl *WishlistController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	result := database.DB.Delete(&models.WishlistItem{}, id)
	if result.Error != nil {
		utils.InternalError(c, "删除失败: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFound(c, "待玩项目不存在")
		return
	}

	utils.Success(c, nil)
}
