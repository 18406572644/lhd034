package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

type Cartridge struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	Title         string       `gorm:"not null" json:"title"`
	Platform      string       `gorm:"not null;index" json:"platform"`
	Publisher     string       `gorm:"not null;index" json:"publisher"`
	ReleaseYear   int          `gorm:"index" json:"releaseYear"`
	Condition     string       `gorm:"not null;default:'good';index" json:"condition"`
	PurchasePrice float64      `gorm:"default:0" json:"purchasePrice"`
	PurchaseDate  string       `json:"purchaseDate"`
	CoverImage    string       `json:"coverImage"`
	Screenshots   StringArray  `gorm:"type:text" json:"screenshots"`
	Region        string       `json:"region"`
	Notes         string       `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
	Playthroughs  []Playthrough `gorm:"foreignKey:CartridgeID" json:"playthroughs,omitempty"`
	Review        *Review      `gorm:"foreignKey:CartridgeID" json:"review,omitempty"`
	Wishlist      *WishlistItem `gorm:"foreignKey:CartridgeID" json:"wishlist,omitempty"`
	BorrowRecords []BorrowRecord `gorm:"foreignKey:CartridgeID" json:"borrowRecords,omitempty"`
}

type Playthrough struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	CartridgeID     uint        `gorm:"not null;index" json:"cartridgeId"`
	StartDate       string      `json:"startDate"`
	CompletionDate  string      `gorm:"not null;index" json:"completionDate"`
	PlayTimeHours   float64     `gorm:"default:0" json:"playTimeHours"`
	DifficultyRating int        `gorm:"not null;default:3" json:"difficultyRating"`
	EndingType      string      `json:"endingType"`
	MultipleEndings bool        `gorm:"default:false" json:"multipleEndings"`
	AchievedEndings StringArray `gorm:"type:text" json:"achievedEndings"`
	Notes           string      `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time   `json:"createdAt"`
	Cartridge       *Cartridge  `gorm:"foreignKey:CartridgeID" json:"cartridge,omitempty"`
}

type Review struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	CartridgeID    uint        `gorm:"not null;uniqueIndex" json:"cartridgeId"`
	ContentRating  int         `gorm:"not null;default:3" json:"contentRating"`
	GameplayRating int         `gorm:"not null;default:3" json:"gameplayRating"`
	GraphicsRating int         `gorm:"not null;default:3" json:"graphicsRating"`
	SoundRating    int         `gorm:"not null;default:3" json:"soundRating"`
	OverallRating  float64     `gorm:"not null;default:3" json:"overallRating"`
	ReviewText     string      `gorm:"type:text" json:"reviewText"`
	StoryNotes     string      `gorm:"type:text" json:"storyNotes"`
	EasterEggs     StringArray `gorm:"type:text" json:"easterEggs"`
	CreatedAt      time.Time   `json:"createdAt"`
	Cartridge      *Cartridge  `gorm:"foreignKey:CartridgeID" json:"cartridge,omitempty"`
}

type WishlistItem struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	CartridgeID     uint        `gorm:"not null" json:"cartridgeId"`
	Priority        string      `gorm:"not null;default:'medium'" json:"priority"`
	PlannedStartDate string     `json:"plannedStartDate"`
	Tags            StringArray `gorm:"type:text" json:"tags"`
	Notes           string      `gorm:"type:text" json:"notes"`
	AddedAt         time.Time   `json:"addedAt"`
	Cartridge       *Cartridge  `gorm:"foreignKey:CartridgeID" json:"cartridge,omitempty"`
}

type BorrowRecord struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	CartridgeID        uint       `gorm:"not null;index" json:"cartridgeId"`
	BorrowerName       string     `gorm:"not null" json:"borrowerName"`
	BorrowerContact    string     `json:"borrowerContact"`
	BorrowDate         string     `gorm:"not null" json:"borrowDate"`
	ExpectedReturnDate string     `json:"expectedReturnDate"`
	ActualReturnDate   *string    `json:"actualReturnDate"`
	Status             string     `gorm:"not null;default:'borrowed';index" json:"status"`
	Notes              string     `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time  `json:"createdAt"`
	Cartridge          *Cartridge `gorm:"foreignKey:CartridgeID" json:"cartridge,omitempty"`
}

type Collection struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `gorm:"not null" json:"name"`
	Description string      `gorm:"type:text" json:"description"`
	CoverImage  string      `json:"coverImage"`
	CreatedAt   time.Time   `json:"createdAt"`
	Cartridges  []Cartridge `gorm:"many2many:cartridge_collections;" json:"cartridges,omitempty"`
}
