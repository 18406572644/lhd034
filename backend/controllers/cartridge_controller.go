package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CartridgeController struct{}

func NewCartridgeController() *CartridgeController {
	return &CartridgeController{}
}

func (ctrl *CartridgeController) GetList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "12"))
	platform := c.Query("platform")
	publisher := c.Query("publisher")
	condition := c.Query("condition")
	year := c.Query("year")
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 12
	}

	query := database.DB.Model(&models.Cartridge{})

	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if publisher != "" {
		query = query.Where("publisher = ?", publisher)
	}
	if condition != "" {
		query = query.Where("condition = ?", condition)
	}
	if year != "" {
		query = query.Where("release_year = ?", year)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("title LIKE ? OR publisher LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var cartridges []models.Cartridge
	query.Preload("Review").Preload("Wishlist").Preload("Sessions").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&cartridges)

	utils.SuccessPaged(c, cartridges, total, page, pageSize)
}

func (ctrl *CartridgeController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var cartridge models.Cartridge
	result := database.DB.Preload("Playthroughs").
		Preload("Review").
		Preload("Wishlist").
		Preload("BorrowRecords").
		Preload("Sessions").
		First(&cartridge, id)

	if result.Error != nil {
		utils.NotFound(c, "卡带不存在")
		return
	}

	utils.Success(c, cartridge)
}

func (ctrl *CartridgeController) Create(c *gin.Context) {
	var cartridge models.Cartridge
	if err := c.ShouldBindJSON(&cartridge); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	cartridge.CreatedAt = time.Now()
	cartridge.UpdatedAt = time.Now()

	result := database.DB.Create(&cartridge)
	if result.Error != nil {
		utils.InternalError(c, "创建失败: "+result.Error.Error())
		return
	}

	utils.Success(c, cartridge)
}

func (ctrl *CartridgeController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var existing models.Cartridge
	if err := database.DB.First(&existing, id).Error; err != nil {
		utils.NotFound(c, "卡带不存在")
		return
	}

	var cartridge models.Cartridge
	if err := c.ShouldBindJSON(&cartridge); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	cartridge.ID = uint(id)
	cartridge.CreatedAt = existing.CreatedAt
	cartridge.UpdatedAt = time.Now()

	result := database.DB.Save(&cartridge)
	if result.Error != nil {
		utils.InternalError(c, "更新失败: "+result.Error.Error())
		return
	}

	utils.Success(c, cartridge)
}

func (ctrl *CartridgeController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	result := database.DB.Delete(&models.Cartridge{}, id)
	if result.Error != nil {
		utils.InternalError(c, "删除失败: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFound(c, "卡带不存在")
		return
	}

	utils.Success(c, nil)
}

func (ctrl *CartridgeController) Upload(c *gin.Context) {
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.InternalError(c, "创建上传目录失败")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "获取文件失败: "+err.Error())
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		utils.BadRequest(c, "不支持的文件格式")
		return
	}

	maxSize := int64(10 * 1024 * 1024)
	if header.Size > maxSize {
		utils.BadRequest(c, "文件大小不能超过10MB")
		return
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), generateRandomString(8), ext)
	filepath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filepath)
	if err != nil {
		utils.InternalError(c, "创建文件失败: "+err.Error())
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		utils.InternalError(c, "保存文件失败: "+err.Error())
		return
	}

	url := fmt.Sprintf("/uploads/%s", filename)
	utils.Success(c, gin.H{"url": url, "filename": filename})
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func (ctrl *CartridgeController) GetPlatforms(c *gin.Context) {
	var platforms []string
	database.DB.Model(&models.Cartridge{}).Distinct("platform").Pluck("platform", &platforms)
	utils.Success(c, platforms)
}

func (ctrl *CartridgeController) GetPublishers(c *gin.Context) {
	var publishers []string
	database.DB.Model(&models.Cartridge{}).Distinct("publisher").Pluck("publisher", &publishers)
	utils.Success(c, publishers)
}
