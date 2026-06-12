package controllers

import (
	"cartridge-archive/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type BackupController struct{}

func NewBackupController() *BackupController {
	return &BackupController{}
}

type BackupInfo struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"createdAt"`
}

type BackupConfig struct {
	Enabled   bool   `json:"enabled"`
	Frequency string `json:"frequency"`
	Retention int    `json:"retention"`
}

var defaultConfig = BackupConfig{
	Enabled:   false,
	Frequency: "daily",
	Retention: 7,
}

var currentConfig = defaultConfig

const backupDir = "./backups"
const configFile = "./backups/backup_config.json"

func getDBPath() string {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/cartridge_archive.db"
	}
	return dbPath
}

func ensureBackupDir() error {
	return os.MkdirAll(backupDir, 0755)
}

func (ctrl *BackupController) CreateBackup(c *gin.Context) {
	if err := ensureBackupDir(); err != nil {
		utils.InternalError(c, "Failed to create backup directory")
		return
	}

	dbPath := getDBPath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		utils.NotFound(c, "Database file not found")
		return
	}

	src, err := os.Open(dbPath)
	if err != nil {
		utils.InternalError(c, "Failed to open database file")
		return
	}
	defer src.Close()

	timestamp := time.Now().Format("20060102_150405")
	backupFilename := fmt.Sprintf("cartridge_archive_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	dst, err := os.Create(backupPath)
	if err != nil {
		utils.InternalError(c, "Failed to create backup file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(backupPath)
		utils.InternalError(c, "Failed to write backup file")
		return
	}

	info, _ := os.Stat(backupPath)
	utils.Success(c, BackupInfo{
		Filename:  backupFilename,
		Size:      info.Size(),
		CreatedAt: info.ModTime().Format(time.RFC3339),
	})
}

func (ctrl *BackupController) ListBackups(c *gin.Context) {
	if err := ensureBackupDir(); err != nil {
		utils.InternalError(c, "Failed to access backup directory")
		return
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		utils.InternalError(c, "Failed to read backup directory")
		return
	}

	backups := make([]BackupInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "backup_config") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Filename:  entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime().Format(time.RFC3339),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	utils.Success(c, backups)
}

func (ctrl *BackupController) RestoreBackup(c *gin.Context) {
	filename := c.Param("filename")
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		utils.BadRequest(c, "Invalid backup filename")
		return
	}

	backupPath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		utils.NotFound(c, "Backup file not found")
		return
	}

	dbPath := getDBPath()

	snapshotFilename := fmt.Sprintf("pre_restore_snapshot_%s.db", time.Now().Format("20060102_150405"))
	snapshotPath := filepath.Join(backupDir, snapshotFilename)

	if _, err := os.Stat(dbPath); err == nil {
		src, err := os.Open(dbPath)
		if err != nil {
			utils.InternalError(c, "Failed to open current database for snapshot")
			return
		}
		dst, err := os.Create(snapshotPath)
		if err != nil {
			src.Close()
			utils.InternalError(c, "Failed to create pre-restore snapshot")
			return
		}
		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}

	src, err := os.Open(backupPath)
	if err != nil {
		utils.InternalError(c, "Failed to open backup file")
		return
	}
	defer src.Close()

	dst, err := os.Create(dbPath)
	if err != nil {
		utils.InternalError(c, "Failed to write to database path")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		utils.InternalError(c, "Failed to restore database")
		return
	}

	utils.Success(c, gin.H{
		"message":          "Database restored successfully",
		"snapshotFilename": snapshotFilename,
	})
}

func (ctrl *BackupController) DeleteBackup(c *gin.Context) {
	filename := c.Param("filename")
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		utils.BadRequest(c, "Invalid backup filename")
		return
	}

	backupPath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		utils.NotFound(c, "Backup file not found")
		return
	}

	if err := os.Remove(backupPath); err != nil {
		utils.InternalError(c, "Failed to delete backup file")
		return
	}

	utils.Success(c, gin.H{"message": "Backup deleted successfully"})
}

func (ctrl *BackupController) GetConfig(c *gin.Context) {
	loadConfig()
	utils.Success(c, currentConfig)
}

func (ctrl *BackupController) UpdateConfig(c *gin.Context) {
	var config BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "Invalid config data")
		return
	}

	if config.Frequency != "daily" && config.Frequency != "weekly" && config.Frequency != "monthly" {
		utils.BadRequest(c, "Frequency must be daily, weekly, or monthly")
		return
	}

	if config.Retention < 1 || config.Retention > 100 {
		utils.BadRequest(c, "Retention must be between 1 and 100")
		return
	}

	currentConfig = config
	saveConfig()

	utils.Success(c, currentConfig)
}

func loadConfig() {
	data, err := os.ReadFile(configFile)
	if err != nil {
		currentConfig = defaultConfig
		return
	}
	if err := json.Unmarshal(data, &currentConfig); err != nil {
		currentConfig = defaultConfig
	}
}

func saveConfig() {
	data, err := json.MarshalIndent(currentConfig, "", "  ")
	if err != nil {
		return
	}
	ensureBackupDir()
	os.WriteFile(configFile, data, 0644)
}
