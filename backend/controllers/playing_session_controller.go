package controllers

import (
	"cartridge-archive/database"
	"cartridge-archive/models"
	"cartridge-archive/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type PlayingSessionController struct{}

func NewPlayingSessionController() *PlayingSessionController {
	return &PlayingSessionController{}
}

func (ctrl *PlayingSessionController) GetList(c *gin.Context) {
	cartridgeId := c.Query("cartridgeId")

	query := database.DB.Model(&models.PlayingSession{})

	if cartridgeId != "" {
		query = query.Where("cartridge_id = ?", cartridgeId)
	}

	var sessions []models.PlayingSession
	query.Preload("Cartridge").
		Order("session_date DESC, created_at DESC").
		Find(&sessions)

	utils.Success(c, sessions)
}

func (ctrl *PlayingSessionController) GetByCartridge(c *gin.Context) {
	cartridgeId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的卡带ID")
		return
	}

	var sessions []models.PlayingSession
	database.DB.Where("cartridge_id = ?", cartridgeId).
		Order("session_date DESC, created_at DESC").
		Find(&sessions)

	utils.Success(c, sessions)
}

func (ctrl *PlayingSessionController) Create(c *gin.Context) {
	var session models.PlayingSession
	if err := c.ShouldBindJSON(&session); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if session.ProgressPercent < 0 {
		session.ProgressPercent = 0
	}
	if session.ProgressPercent > 100 {
		session.ProgressPercent = 100
	}

	session.CreatedAt = time.Now()

	result := database.DB.Create(&session)
	if result.Error != nil {
		utils.InternalError(c, "创建失败: "+result.Error.Error())
		return
	}

	var cartridge models.Cartridge
	if err := database.DB.First(&cartridge, session.CartridgeID).Error; err == nil {
		newStatus := "playing"
		if session.ProgressPercent >= 100 {
			newStatus = "completed"
		}
		if cartridge.Status == "unstarted" || cartridge.Status == "playing" {
			database.DB.Model(&cartridge).Update("status", newStatus)
		}
	}

	utils.Success(c, session)
}

func (ctrl *PlayingSessionController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var existing models.PlayingSession
	if err := database.DB.First(&existing, id).Error; err != nil {
		utils.NotFound(c, "游玩会话不存在")
		return
	}

	var session models.PlayingSession
	if err := c.ShouldBindJSON(&session); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if session.ProgressPercent < 0 {
		session.ProgressPercent = 0
	}
	if session.ProgressPercent > 100 {
		session.ProgressPercent = 100
	}

	session.ID = uint(id)
	session.CreatedAt = existing.CreatedAt

	result := database.DB.Save(&session)
	if result.Error != nil {
		utils.InternalError(c, "更新失败: "+result.Error.Error())
		return
	}

	utils.Success(c, session)
}

func (ctrl *PlayingSessionController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	result := database.DB.Delete(&models.PlayingSession{}, id)
	if result.Error != nil {
		utils.InternalError(c, "删除失败: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFound(c, "游玩会话不存在")
		return
	}

	utils.Success(c, nil)
}

func (ctrl *PlayingSessionController) GetProgress(c *gin.Context) {
	cartridgeId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的卡带ID")
		return
	}

	var sessions []models.PlayingSession
	database.DB.Where("cartridge_id = ?", cartridgeId).
		Order("session_date ASC, created_at ASC").
		Find(&sessions)

	if len(sessions) == 0 {
		utils.Success(c, gin.H{
			"cartridgeId":       cartridgeId,
			"currentProgress":   0,
			"totalSessions":     0,
			"totalMinutes":      0,
			"estimatedRemaining": nil,
			"sessions":          []models.PlayingSession{},
		})
		return
	}

	var totalMinutes int
	var maxProgress int
	for _, s := range sessions {
		totalMinutes += s.DurationMinutes
		if s.ProgressPercent > maxProgress {
			maxProgress = s.ProgressPercent
		}
	}

	var estimatedRemaining *float64
	if maxProgress > 0 && maxProgress < 100 {
		minutesPerPercent := float64(totalMinutes) / float64(maxProgress)
		remaining := minutesPerPercent * float64(100-maxProgress)
		estimatedRemaining = &remaining
	}

	utils.Success(c, gin.H{
		"cartridgeId":        cartridgeId,
		"currentProgress":    maxProgress,
		"totalSessions":      len(sessions),
		"totalMinutes":       totalMinutes,
		"estimatedRemaining": estimatedRemaining,
		"sessions":           sessions,
	})
}

func (ctrl *PlayingSessionController) GetPlayingCartridges(c *gin.Context) {
	var cartridges []models.Cartridge
	database.DB.Where("status = ?", "playing").
		Preload("Sessions").
		Find(&cartridges)

	type PlayingCartridgeProgress struct {
		Cartridge         models.Cartridge    `json:"cartridge"`
		CurrentProgress   int                 `json:"currentProgress"`
		TotalSessions     int                 `json:"totalSessions"`
		TotalMinutes      int                 `json:"totalMinutes"`
		EstimatedRemaining *float64           `json:"estimatedRemaining"`
		LatestSession     *models.PlayingSession `json:"latestSession"`
	}

	var result []PlayingCartridgeProgress
	for _, cart := range cartridges {
		sessions := cart.Sessions
		if len(sessions) == 0 {
			continue
		}

		var totalMinutes int
		var maxProgress int
		var latest *models.PlayingSession
		for i, s := range sessions {
			totalMinutes += s.DurationMinutes
			if s.ProgressPercent > maxProgress {
				maxProgress = s.ProgressPercent
			}
			if latest == nil || s.CreatedAt.After(latest.CreatedAt) {
				latest = &sessions[i]
			}
		}

		var estimatedRemaining *float64
		if maxProgress > 0 && maxProgress < 100 {
			minutesPerPercent := float64(totalMinutes) / float64(maxProgress)
			remaining := minutesPerPercent * float64(100 - maxProgress)
			estimatedRemaining = &remaining
		}

		result = append(result, PlayingCartridgeProgress{
			Cartridge:         cart,
			CurrentProgress:   maxProgress,
			TotalSessions:     len(sessions),
			TotalMinutes:      totalMinutes,
			EstimatedRemaining: estimatedRemaining,
			LatestSession:     latest,
		})
	}

	utils.Success(c, result)
}
