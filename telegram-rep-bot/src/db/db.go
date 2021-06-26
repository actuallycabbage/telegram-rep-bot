package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//type account struct {
//	ID                uint `gorm:"primaryKey"`
//	UUID              string
//	CreatedAt         time.Time
//	UpdatedAt         time.Time
//	UpdateInformation string
//	Status            string `gorm:"index"`
//}
//
//type account_config struct {
//	ID                uint `gorm:"primaryKey"`
//	UUID              string
//	CreatedAt         time.Time
//	UpdateInformation string
//	Status            string `gorm:"index"`
//	Data              string
//}
//
//type telegram_chat_link struct {
//	ID             uint `gorm:"primaryKey"`
//	CreatedAt      time.Time
//	Status         string `gorm:"index"` // active or historical
//	TelegramChatID int64  `gorm:"index"`
//	AccountUUID    string `gorm:"index"`
//}

type rep_event struct {
	ID             uint      `gorm:"primaryKey"`
	CreatedAt      time.Time `gorm:"index"`
	DeletedAt      sql.NullTime
	TelegramChatID int64 `gorm:"index"`
	RepChange      int
	TargetUserID   int64 `gorm:"index"`
	RequestUserID  int64
	Metadata       datatypes.JSON
}

// Unused
//type chat_event struct {
//	ID             uint `gorm:"primaryKey"`
//	CreatedAt      time.Time
//	TelegramChatID int64 `gorm:"index"`
//	EventType      string
//	EventData      string
//	EventOutcome   string
//}

// For possible optimisations down the line
//type rep_event_rollup struct {
//	ID             uint `gorm:"primaryKey"`
//	CreatedAt      time.Time
//	TelegramChatID int64  `gorm:"index"`
//	UserID         string `gorm:"index"`
//	Rep            int
//}

// DB Library Settings
type Config struct {
	Type             string
	ConnectionString string
}

type DB struct {
	Config *Config
	db     *gorm.DB
}

func Connect(config *Config) (*DB, error) {
	db := &DB{
		Config: config,
	}

	// Global gorm config
	gc := &gorm.Config{}

	// TODO: Postgres support
	// Open database
	if db.Config.Type == "sqlite" {
		d, err := gorm.Open(sqlite.Open(db.Config.ConnectionString), gc)
		if err != nil {
			log.Fatal(err)
		}
		db.db = d
	} else {
		return nil, ErrUnsupportedDatabase
	}

	//NOTE: When adding new tables, ensure they also get added to this auto migrate list.
	//db.db.AutoMigrate(&account{}, &account_config{}, &telegram_chat_link{}, &chat_event{}, &rep_event{})
	db.db.AutoMigrate(&rep_event{})

	return db, nil

}

//func (*DB) GetAssociatedAccount(chatID int64) (*account, error) {
//	return nil, ErrNotImplemented
//}
//
//func (*DB) createChatLink(chatID int64, accountUUID string) (*telegram_chat_link, error) {
//	return nil, ErrNotImplemented
//}

func (self *DB) CreateRepEvent(chatID int64, targetID int, requesterID int, repChange int, metadata map[string]interface{}) (*rep_event, error) {

	m, err := json.Marshal(metadata)

	if err != nil {
		return nil, err // TODO: Need to figure out how to chain & warp errors.
	}

	p := &rep_event{
		CreatedAt:      time.Now(),
		TelegramChatID: chatID,
		RepChange:      repChange,
		TargetUserID:   int64(targetID),
		RequestUserID:  int64(requesterID),
		Metadata:       m,
	}

	result := self.db.Create(p)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, ErrNoRowsAffected
	}

	return p, nil
}
