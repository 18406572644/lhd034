package controllers

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var schedulerStop chan struct{}
var schedulerRunning bool

func StartBackupScheduler() {
	if schedulerRunning {
		return
	}
	schedulerRunning = true
	schedulerStop = make(chan struct{})

	go runScheduler()
	log.Println("Backup scheduler started")
}

func StopBackupScheduler() {
	if !schedulerRunning {
		return
	}
	close(schedulerStop)
	schedulerRunning = false
	log.Println("Backup scheduler stopped")
}

func RestartBackupScheduler() {
	StopBackupScheduler()
	StartBackupScheduler()
}

func runScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-schedulerStop:
			return
		case <-ticker.C:
			loadConfig()
			if currentConfig.Enabled {
				checkAndRunBackup()
			}
		}
	}
}

func checkAndRunBackup() {
	now := time.Now()

	switch currentConfig.Frequency {
	case "daily":
		lastBackup := getLastBackupTime()
		if now.Sub(lastBackup) >= 24*time.Hour {
			runScheduledBackup()
		}
	case "weekly":
		lastBackup := getLastBackupTime()
		if now.Sub(lastBackup) >= 7*24*time.Hour {
			runScheduledBackup()
		}
	case "monthly":
		lastBackup := getLastBackupTime()
		if now.Sub(lastBackup) >= 30*24*time.Hour {
			runScheduledBackup()
		}
	}
}

func getLastBackupTime() time.Time {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return time.Time{}
	}

	var latest time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "backup_config") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "pre_restore_snapshot") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest
}

func runScheduledBackup() {
	log.Println("Running scheduled backup...")

	if err := ensureBackupDir(); err != nil {
		log.Printf("Failed to create backup directory: %v", err)
		return
	}

	dbPath := getDBPath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Println("Database file not found, skipping scheduled backup")
		return
	}

	src, err := os.Open(dbPath)
	if err != nil {
		log.Printf("Failed to open database for backup: %v", err)
		return
	}
	defer src.Close()

	timestamp := time.Now().Format("20060102_150405")
	backupFilename := fmt.Sprintf("cartridge_archive_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	dst, err := os.Create(backupPath)
	if err != nil {
		log.Printf("Failed to create backup file: %v", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(backupPath)
		log.Printf("Failed to write backup file: %v", err)
		return
	}

	log.Printf("Scheduled backup created: %s", backupFilename)
	cleanupOldBackups()
}

func cleanupOldBackups() {
	loadConfig()
	retention := currentConfig.Retention
	if retention <= 0 {
		retention = 7
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	var backupFiles []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "backup_config") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "pre_restore_snapshot") {
			continue
		}
		backupFiles = append(backupFiles, entry)
	}

	sort.Slice(backupFiles, func(i, j int) bool {
		infoI, _ := backupFiles[i].Info()
		infoJ, _ := backupFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	if len(backupFiles) > retention {
		for _, entry := range backupFiles[retention:] {
			path := filepath.Join(backupDir, entry.Name())
			if err := os.Remove(path); err != nil {
				log.Printf("Failed to delete old backup %s: %v", entry.Name(), err)
			} else {
				log.Printf("Cleaned up old backup: %s", entry.Name())
			}
		}
	}
}
