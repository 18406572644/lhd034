package database

import (
	"cartridge-archive/models"
	"fmt"
	"log"
	"os"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbDir := "./data"
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
		dbPath = fmt.Sprintf("%s/cartridge_archive.db", dbDir)
	} else {
		dbDir := dbPath[:len(dbPath)-len("/cartridge_archive.db")]
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
	}

	dsn := dbPath
	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(
		&models.Cartridge{},
		&models.Playthrough{},
		&models.Review{},
		&models.WishlistItem{},
		&models.BorrowRecord{},
		&models.Collection{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	seedData()
}

func seedData() {
	var count int64
	DB.Model(&models.Cartridge{}).Count(&count)
	if count > 0 {
		return
	}

	cartridges := []models.Cartridge{
		{
			Title:         "超级马里奥兄弟",
			Platform:      "NES",
			Publisher:     "任天堂",
			ReleaseYear:   1985,
			Condition:     "excellent",
			PurchasePrice: 128.50,
			PurchaseDate:  "2023-03-15",
			Region:        "日版",
			Notes:         "经典初代，保存完好",
			CoverImage:    "",
			Screenshots:   models.StringArray{},
		},
		{
			Title:         "塞尔达传说 时之笛",
			Platform:      "N64",
			Publisher:     "任天堂",
			ReleaseYear:   1998,
			Condition:     "mint",
			PurchasePrice: 350.00,
			PurchaseDate:  "2022-08-20",
			Region:        "美版",
			Notes:         "全新未拆封收藏",
			CoverImage:    "",
			Screenshots:   models.StringArray{},
		},
		{
			Title:         "最终幻想7",
			Platform:      "PS1",
			Publisher:     "史克威尔",
			ReleaseYear:   1997,
			Condition:     "good",
			PurchasePrice: 88.00,
			PurchaseDate:  "2023-01-10",
			Region:        "日版",
			Notes:         "双碟装，说明书齐全",
			CoverImage:    "",
			Screenshots:   models.StringArray{},
		},
		{
			Title:         "索尼克大冒险",
			Platform:      "DC",
			Publisher:     "世嘉",
			ReleaseYear:   1998,
			Condition:     "fair",
			PurchasePrice: 45.00,
			PurchaseDate:  "2023-05-01",
			Region:        "日版",
			Notes:         "外壳有轻微划痕",
			CoverImage:    "",
			Screenshots:   models.StringArray{},
		},
	}

	for i := range cartridges {
		DB.Create(&cartridges[i])
	}

	playthroughs := []models.Playthrough{
		{
			CartridgeID:     1,
			StartDate:       "2024-01-05",
			CompletionDate:  "2024-01-12",
			PlayTimeHours:   8.5,
			DifficultyRating: 2,
			EndingType:      "标准结局",
			MultipleEndings: false,
			AchievedEndings: models.StringArray{},
			Notes:           "重温童年经典",
		},
		{
			CartridgeID:     3,
			StartDate:       "2024-02-01",
			CompletionDate:  "2024-03-15",
			PlayTimeHours:   60.0,
			DifficultyRating: 4,
			EndingType:      "A结局",
			MultipleEndings: true,
			AchievedEndings: models.StringArray{"A结局"},
			Notes:           "终于达成A结局，下次尝试其他路线",
		},
	}

	for i := range playthroughs {
		DB.Create(&playthroughs[i])
	}

	reviews := []models.Review{
		{
			CartridgeID:    1,
			ContentRating:  5,
			GameplayRating: 5,
			GraphicsRating: 4,
			SoundRating:    5,
			OverallRating:  4.8,
			ReviewText:     "永远的经典，每次玩都有新感受",
			EasterEggs:     models.StringArray{"负一关", "跳关秘籍"},
		},
		{
			CartridgeID:    3,
			ContentRating:  5,
			GameplayRating: 5,
			GraphicsRating: 5,
			SoundRating:    5,
			OverallRating:  5.0,
			ReviewText:     "RPG史上的里程碑，爱丽丝之死至今难忘",
			EasterEggs:     models.StringArray{"萨菲罗斯彩蛋", "隐藏召唤兽"},
		},
	}

	for i := range reviews {
		DB.Create(&reviews[i])
	}

	wishlist := []models.WishlistItem{
		{
			CartridgeID:     2,
			Priority:        "high",
			PlannedStartDate: "2024-06-01",
			Tags:            models.StringArray{"神作", "必玩"},
			Notes:           "听说这是最好的塞尔达",
		},
		{
			CartridgeID:     4,
			Priority:        "medium",
			PlannedStartDate: "",
			Tags:            models.StringArray{"世嘉", "怀旧"},
			Notes:           "世嘉粉丝必玩",
		},
	}

	for i := range wishlist {
		DB.Create(&wishlist[i])
	}

	log.Println("Database seeded with sample data")
}
