package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"evening-gown/internal/bootstrap"
	"evening-gown/internal/config"
	"evening-gown/internal/database"
	"evening-gown/internal/model"

	"gorm.io/gorm"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	dsn := cfg.Postgres.DSN
	if dsn == "" {
		log.Fatalf("POSTGRES_DSN is empty (seed requires Postgres)")
	}

	db, err := database.New(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer func() {
		_ = database.Close(db)
	}()

	if err := bootstrap.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	// Optional: bootstrap admin if requested.
	if cfg.Admin.Email != "" && cfg.Admin.Password != "" {
		if err := bootstrap.EnsureSingleAdmin(db, cfg.Admin.Email, cfg.Admin.Password); err != nil {
			log.Printf("ensure admin skipped/failed: %v", err)
		}
	}

	if err := seedAll(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	log.Println("seed completed")
}

func seedAll(db *gorm.DB) error {
	now := time.Now().UTC()

	detail1 := mustJSON(map[string]any{
		"title_i18n": map[string]any{"zh": "白色幻影礼服", "en": "White Phantom Gown"},
		"specs": []any{
			map[string]any{"k": "Fabric", "v": "Silk"},
			map[string]any{"k": "Color", "v": "Ivory"},
		},
		"option_groups": []any{},
	})

	products := []model.Product{
		{
			Slug:          "style-9001",
			StyleNo:       9001,
			Season:        "ss25",
			Category:      "gown",
			Availability:  "in_stock",
			IsNew:         true,
			NewRank:       10,
			CoverImageURL: "https://picsum.photos/seed/evening-gown-9001/900/1200",
			HoverImageURL: "https://picsum.photos/seed/evening-gown-9001h/900/1200",
			PriceMode:     "negotiable",
			DetailJSON:    detail1,
			PublishedAt:   &now,
		},
		{
			Slug:          "style-9002",
			StyleNo:       9002,
			Season:        "ss25",
			Category:      "couture",
			Availability:  "preorder",
			IsNew:         true,
			NewRank:       9,
			CoverImageURL: "https://picsum.photos/seed/evening-gown-9002/900/1200",
			HoverImageURL: "https://picsum.photos/seed/evening-gown-9002h/900/1200",
			PriceMode:     "negotiable",
			DetailJSON:    detail1,
			PublishedAt:   &now,
		},
		{
			Slug:          "style-9003",
			StyleNo:       9003,
			Season:        "fw25",
			Category:      "bridal",
			Availability:  "in_stock",
			IsNew:         false,
			NewRank:       0,
			CoverImageURL: "https://picsum.photos/seed/evening-gown-9003/900/1200",
			HoverImageURL: "https://picsum.photos/seed/evening-gown-9003h/900/1200",
			PriceMode:     "negotiable",
			DetailJSON:    detail1,
			PublishedAt:   nil, // draft
		},
	}

	for _, p := range products {
		if err := upsertProduct(db, p); err != nil {
			return fmt.Errorf("upsert product styleNo=%d: %w", p.StyleNo, err)
		}
	}

	updates := []model.UpdatePost{
		{
			Type:       "company",
			Status:     "published",
			Tag:        "新品",
			Title:      "2025 春夏系列上新",
			Summary:    "春夏系列新品陆续到店，欢迎预约看样。",
			Body:       "春夏系列新品陆续到店，欢迎预约看样。\n\n（演示数据，可随时删除）",
			RefCode:    "SEED-UPDATE-001",
			PinnedRank: 10,
			PublishedAt: &now,
		},
		{
			Type:       "company",
			Status:     "published",
			Tag:        "活动",
			Title:      "线下展厅预约开放",
			Summary:    "我们已开放线下展厅预约服务。",
			Body:       "我们已开放线下展厅预约服务。\n\n（演示数据，可随时删除）",
			RefCode:    "SEED-UPDATE-002",
			PinnedRank: 5,
			PublishedAt: &now,
		},
		{
			Type:       "company",
			Status:     "draft",
			Tag:        "草稿",
			Title:      "（草稿）品牌故事更新",
			Summary:    "这是一条草稿，前台不会展示。",
			Body:       "这是一条草稿，前台不会展示。\n\n（演示数据，可随时删除）",
			RefCode:    "SEED-UPDATE-003",
			PinnedRank: 0,
			PublishedAt: nil,
		},
	}

	for _, u := range updates {
		if err := upsertUpdate(db, u); err != nil {
			return fmt.Errorf("upsert update ref=%s: %w", u.RefCode, err)
		}
	}

	// Contacts (demo)
	contacts := []model.ContactLead{
		{Name: "Alice", Phone: "13800000000", Wechat: "", Message: "想咨询合作与价格", SourcePage: "/", Status: "new"},
		{Name: "Bob", Phone: "", Wechat: "bob_wechat", Message: "预约看样", SourcePage: "/#contact", Status: "new"},
	}
	for _, c := range contacts {
		if err := upsertContact(db, c); err != nil {
			return fmt.Errorf("upsert contact: %w", err)
		}
	}

	// Events (demo)
	_ = db.Create(&model.Event{EventType: "view", OccurredAt: now, PageURL: "/", Payload: mustJSON(map[string]any{"demo": true})}).Error

	return nil
}

func upsertProduct(db *gorm.DB, p model.Product) error {
	var existing model.Product
	err := db.Where("style_no = ?", p.StyleNo).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&p).Error
	}
	if err != nil {
		return err
	}

	updates := map[string]any{
		"slug":            p.Slug,
		"season":          p.Season,
		"category":        p.Category,
		"availability":    p.Availability,
		"is_new":          p.IsNew,
		"new_rank":        p.NewRank,
		"cover_image_url": p.CoverImageURL,
		"hover_image_url": p.HoverImageURL,
		"price_mode":      p.PriceMode,
		"detail_json":     p.DetailJSON,
		"published_at":    p.PublishedAt,
		"deleted_at":      nil,
	}

	return db.Model(&model.Product{}).Where("id = ?", existing.ID).Updates(updates).Error
}

func upsertUpdate(db *gorm.DB, u model.UpdatePost) error {
	var existing model.UpdatePost
	err := db.Where("ref_code = ?", u.RefCode).Where("deleted_at IS NULL").First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&u).Error
	}
	if err != nil {
		return err
	}

	updates := map[string]any{
		"type":         u.Type,
		"status":       u.Status,
		"tag":          u.Tag,
		"title":        u.Title,
		"summary":      u.Summary,
		"body":         u.Body,
		"ref_code":     u.RefCode,
		"pinned_rank":  u.PinnedRank,
		"published_at": u.PublishedAt,
		"deleted_at":   nil,
	}
	return db.Model(&model.UpdatePost{}).Where("id = ?", existing.ID).Updates(updates).Error
}

func upsertContact(db *gorm.DB, c model.ContactLead) error {
	q := db.Model(&model.ContactLead{})
	if c.Phone != "" {
		q = q.Where("phone = ?", c.Phone)
	} else if c.Wechat != "" {
		q = q.Where("wechat = ?", c.Wechat)
	} else {
		return nil
	}

	var existing model.ContactLead
	err := q.First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&c).Error
	}
	if err != nil {
		return err
	}

	return nil
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}
