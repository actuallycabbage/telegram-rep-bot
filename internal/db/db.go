package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type account struct {
	ID                uint `gorm:"primaryKey"`
	CreatedAt         time.Time
	UpdatedAt         sql.NullTime
	UpdateInformation string
	Status            string `gorm:"index;default:active"`
}

type account_config struct {
	ID        uint `gorm:"primaryKey"`
	AccountID uint
	CreatedAt time.Time
	Status    string `gorm:"index;default:active"`
	Settings  datatypes.JSON
	account   account `gorm:"foreignKey:AccountID"`
}

type AccountSettings struct {
	Rep *AccountRepSettings `json:"rep"`
}

type AccountRepSettings struct {
	Enabled          bool                        `json:"enabled"`
	Silent           bool                        `json:"silent"`
	Cooldown         *AccountRepCooldownSettings `json:"cooldown"`
	PositiveTriggers []string                    `json:"positiveTriggers"`
	PositiveStickers []string                    `json:"positiveStickers"`
	NegativeTriggers []string                    `json:"negativeTriggers"`
	NegativeStickers []string                    `json:"negativeStickers"`
}

type AccountRepCooldownSettings struct {
	Enabled  bool   `json:"enabled"`
	Duration string `json:"duration"`
}

type telegram_chat_link struct {
	ID             uint `gorm:"primaryKey"`
	CreatedAt      time.Time
	Status         string `gorm:"index;default:active"` // active or historical
	TelegramChatID int64  `gorm:"index"`
	AccountID      uint   `gorm:"index"`
}

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
	db.db.AutoMigrate(&rep_event{}, &account{}, &telegram_chat_link{}, &account_config{})

	return db, nil

}

func (self *DB) CreateChatLink(chatID int64, accountID uint) (*telegram_chat_link, error) {
	log.Printf("Creating chat link for chat %d and account %d\n", chatID, accountID)

	result := self.db.Model(&telegram_chat_link{}).Where("telegram_chat_id = ? AND status = ?", chatID, "active").Update("status", "historical")

	if result.Error != nil {
		return nil, result.Error
	}

	row := &telegram_chat_link{
		CreatedAt:      time.Now(),
		Status:         "active",
		TelegramChatID: chatID,
		AccountID:      accountID,
	}
	result = self.db.Create(row)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("Only 0 rows were inserted")
	}

	return row, nil

}

func (self *DB) GetChatLink(chatID int64) (*telegram_chat_link, error) {
	record := &telegram_chat_link{}
	result := self.db.Where("telegram_chat_id = ? and status = ?", chatID, "active").Take(&record, &telegram_chat_link{})

	if result.Error != nil {
		return nil, result.Error
	}

	return record, nil
}

// Tries to get a chat link for a group, if it fails it will create one and link it to an account
func (self *DB) MustGetChatLink(chatID int64) *telegram_chat_link {
	link, err := self.GetChatLink(chatID)

	if err == nil {
		return link
	}

	if err != gorm.ErrRecordNotFound {
		// if another error happened something went very wrong
		panic(err.Error())
	}

	// Ok looks like there's no chatlink. Make some
	// TODO: This whole acc creation thing should be moved into its own function
	var acc = &account{}
	// Step 1: Aquire the closest account
	// TODO: For now, there is only the master account. Will move them into their own individualised things soon.
	// gonna assume that if acc exists, settings does too
	result := self.db.First(acc, &account{})

	// In the case of the db being deleted and there being no account, just make one with settings quickly.
	if result.Error == gorm.ErrRecordNotFound {
		acc.CreatedAt = time.Now()
		self.db.Create(acc)

		// Create accconf struct
		accconf := &account_config{
			CreatedAt: time.Now(),
			AccountID: acc.ID,
			Settings:  []byte("{}"), // TODO: Fill this JSON out?
		}
		self.db.Create(accconf)

	}

	// Ok lets try that again.
	link, err = self.CreateChatLink(chatID, acc.ID)
	if err != nil {
		panic(err)
	}

	return link

}

func (self *DB) GetChatSettings(chatID int64) (*AccountSettings, error) {
	cl := self.MustGetChatLink(chatID)

	ac := account_config{}

	result := self.db.Where("account_id = ? and status = ?", cl.AccountID, "active").First(&ac, &account_config{})

	if result.Error != nil {
		panic(result.Error)
	}

	accsetting := &AccountSettings{}

	err := json.Unmarshal(ac.Settings, accsetting)

	if err != nil {
		return nil, err
	}

	return accsetting, nil

}

func (self *DB) CreateRepEvent(chatID int64, targetID int64, requesterID int64, repChange int, metadata map[string]interface{}) (*rep_event, error) {

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

type LeaderboardEntry struct {
	UserID int64
	Rep    int
}

func (self *DB) GetChatRep(chatID int64, order string, limit int) []LeaderboardEntry {
	// Make sure that limit is > 0

	var query string
	switch order {
	case "asc":
		query = "select target_user_id, sum(rep_change) from rep_events where telegram_chat_id = ? group by target_user_id order by sum(rep_change) asc limit ?"
		break
	case "desc":
		query = "select target_user_id, sum(rep_change) from rep_events where telegram_chat_id = ? group by target_user_id order by sum(rep_change) desc limit ?"
		break
	}

	rows, err := self.db.Raw(query, chatID, limit).Rows()

	if err != nil {
		log.Println(err)
		return nil
	}

	var l []LeaderboardEntry

	defer rows.Close()
	for rows.Next() {
		buffer := LeaderboardEntry{}
		rows.Scan(&buffer.UserID, &buffer.Rep)
		l = append(l, buffer)
	}

	return l

}
