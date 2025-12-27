package bootstrap

import (
	"context"

	"evening-gown/internal/cache"
	"evening-gown/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// InitAdminCounters initializes Redis-backed admin counters from the database.
//
// This is used to keep counters strongly consistent across restarts.
// Best-effort: if Redis is nil/disabled, it does nothing.
func InitAdminCounters(ctx context.Context, db *gorm.DB, rdb *redis.Client) error {
	if db == nil || rdb == nil {
		return nil
	}

	var newLeads int64
	if err := db.WithContext(ctx).Model(&model.ContactLead{}).Where("status = ?", "new").Count(&newLeads).Error; err != nil {
		return err
	}

	return rdb.Set(ctx, cache.AdminContactsNewCountKey, newLeads, 0).Err()
}
