package dao

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"shangqing/internal/config"
	"shangqing/internal/model"
)

type DB struct {
	*gorm.DB
}

func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("open db error: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB error: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.GetConnMaxLifetime())

	// 自动迁移
	if err := db.AutoMigrate(
		&model.User{},
		&model.Conversation{},
		&model.Message{},
		&model.ConsciousnessEvent{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate error: %w", err)
	}

	return &DB{DB: db}, nil
}

// ----- User -----

func (d *DB) CreateUser(ctx context.Context, user *model.User) error {
	return d.WithContext(ctx).Create(user).Error
}

func (d *DB) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := d.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *DB) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := d.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *DB) UpdateUserEntropy(ctx context.Context, userID int64, entropyDelta float64) error {
	return d.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("entropy_value", gorm.Expr("entropy_value + ?", entropyDelta)).
		Error
}

func (d *DB) UpdateUserLevel(ctx context.Context, userID int64, level string) error {
	return d.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("consciousness_level", level).Error
}

// ----- Conversation -----

func (d *DB) CreateConversation(ctx context.Context, conv *model.Conversation) error {
	return d.WithContext(ctx).Create(conv).Error
}

func (d *DB) GetConversation(ctx context.Context, id string) (*model.Conversation, error) {
	var conv model.Conversation
	err := d.WithContext(ctx).Where("id = ?", id).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (d *DB) GetConversationsByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, int64, error) {
	var convs []*model.Conversation
	var total int64

	query := d.WithContext(ctx).Model(&model.Conversation{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("last_message_at DESC").Offset(offset).Limit(limit).Find(&convs).Error; err != nil {
		return nil, 0, err
	}

	return convs, total, nil
}

func (d *DB) UpdateConversationStats(ctx context.Context, convID string) error {
	return d.WithContext(ctx).Model(&model.Conversation{}).
		Where("id = ?", convID).
		Updates(map[string]interface{}{
			"message_count":  gorm.Expr("message_count + 1"),
			"last_message_at": gorm.Expr("NOW()"),
		}).Error
}

// ----- Message -----

func (d *DB) CreateMessage(ctx context.Context, msg *model.Message) error {
	return d.WithContext(ctx).Create(msg).Error
}

func (d *DB) GetMessagesByConversation(ctx context.Context, convID string, limit int) ([]*model.Message, error) {
	var msgs []*model.Message
	err := d.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at ASC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

// ----- ConsciousnessEvent -----

func (d *DB) CreateConsciousnessEvent(ctx context.Context, event *model.ConsciousnessEvent) error {
	return d.WithContext(ctx).Create(event).Error
}
