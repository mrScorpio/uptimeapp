package store

import (
	"context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Login        string    `gorm:"uniqueIndex;not null" json:"login"`
	Name         string    `gorm:"not null;default:''" json:"name"`
	AvatarURL    string    `gorm:"not null;default:''" json:"avatar_url"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"-"`
}

type Session struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	User      User      `gorm:"foreignKey:UserID"`
	TokenHash string    `gorm:"uniqueIndex;size:64;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Monitor struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	URL       string    `gorm:"not null" json:"url"`
	Interval  int       `gorm:"not null" json:"interval_seconds"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`
}

func Open(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).AutoMigrate(&User{}, &Session{}, &Monitor{}); err != nil {
		return nil, err
	}
	return db, nil
}
