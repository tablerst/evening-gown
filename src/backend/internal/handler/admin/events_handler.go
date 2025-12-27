package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"evening-gown/internal/cache"
	"evening-gown/internal/logging"
	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type EventsHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewEventsHandler(db *gorm.DB) *EventsHandler {
	return NewEventsHandlerWithRedis(db, nil)
}

func NewEventsHandlerWithRedis(db *gorm.DB, rdb *redis.Client) *EventsHandler {
	return &EventsHandler{db: db, rdb: rdb}
}

type eventsMetricsDay struct {
	Date   string           `json:"date"`
	Total  int64            `json:"total"`
	ByType map[string]int64 `json:"byType"`
}

type eventsMetricsResponse struct {
	Range struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"range"`
	TZ     string            `json:"tz"`
	Series []eventsMetricsDay `json:"series"`
	Totals struct {
		Total  int64            `json:"total"`
		ByType map[string]int64 `json:"byType"`
	} `json:"totals"`
	Cache struct {
		Hit bool   `json:"hit"`
		Key string `json:"key"`
	} `json:"cache"`
	AsOf string `json:"asOf"`
}

// Metrics returns time-series analytics for events.
//
// Query params:
// - range: 7d|30d|90d (default 7d)
// - from/to: RFC3339 (optional, overrides range)
// - tz: IANA tz (default UTC)
// - event_type, product_id: optional filters
// - force=true: bypass cache and recompute
func (h *EventsHandler) Metrics(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	ctx := c.Request.Context()
	force := strings.EqualFold(strings.TrimSpace(c.Query("force")), "true")

	tzName := strings.TrimSpace(c.Query("tz"))
	if tzName == "" {
		tzName = "UTC"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc = time.UTC
		tzName = "UTC"
	}

	// Time range
	var (
		from time.Time
		to   time.Time
	)
	fromStr := strings.TrimSpace(c.Query("from"))
	toStr := strings.TrimSpace(c.Query("to"))
	if fromStr != "" && toStr != "" {
		ft, ferr := time.Parse(time.RFC3339, fromStr)
		tt, terr := time.Parse(time.RFC3339, toStr)
		if ferr == nil && terr == nil {
			from = ft.UTC()
			to = tt.UTC()
		}
	}
	if from.IsZero() || to.IsZero() || !from.Before(to) {
		// Default to last N days in the chosen TZ (including today).
		rangeRaw := strings.TrimSpace(c.Query("range"))
		days := 7
		switch strings.ToLower(rangeRaw) {
		case "30d":
			days = 30
		case "90d":
			days = 90
		}

		now := time.Now().In(loc)
		end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		start := end.AddDate(0, 0, -days)
		from = start.UTC()
		to = end.UTC()
	}

	// Optional filters
	filterEventType := strings.TrimSpace(c.Query("event_type"))
	filterProductID := uint(0)
	if pid := strings.TrimSpace(c.Query("product_id")); pid != "" {
		if v, err := strconv.ParseUint(pid, 10, 64); err == nil && v > 0 {
			filterProductID = uint(v)
		}
	}

	key := buildEventsMetricsCacheKey(from, to, tzName, filterEventType, filterProductID)
	if h.rdb != nil && !force {
		b, err := h.rdb.Get(ctx, key).Bytes()
		if err == nil && len(b) > 0 {
			var resp eventsMetricsResponse
			if json.Unmarshal(b, &resp) == nil {
				resp.Cache.Hit = true
				resp.Cache.Key = key
				c.JSON(http.StatusOK, resp)
				return
			}
		}
	}

	resp, err := h.computeMetrics(ctx, from, to, loc, tzName, filterEventType, filterProductID)
	if err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin events metrics compute failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	resp.Cache.Hit = false
	resp.Cache.Key = key

	if h.rdb != nil {
		if b, err := json.Marshal(resp); err == nil {
			ttl := cache.TTLWithKeyJitter(time.Hour, key, 0.2)
			_ = h.rdb.Set(ctx, key, b, ttl).Err()
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *EventsHandler) computeMetrics(ctx context.Context, fromUTC, toUTC time.Time, loc *time.Location, tzName, eventType string, productID uint) (eventsMetricsResponse, error) {
	q := h.db.WithContext(ctx).Model(&model.Event{}).
		Where("occurred_at >= ?", fromUTC).
		Where("occurred_at < ?", toUTC)
	if strings.TrimSpace(eventType) != "" {
		q = q.Where("event_type = ?", strings.TrimSpace(eventType))
	}
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}

	// For cross-dialect compatibility (tests use SQLite), aggregate in Go.
	type row struct {
		OccurredAt time.Time `gorm:"column:occurred_at"`
		EventType  string    `gorm:"column:event_type"`
	}
	var rows []row
	if err := q.Select("occurred_at, event_type").Find(&rows).Error; err != nil {
		return eventsMetricsResponse{}, err
	}

	byDay := map[string]map[string]int64{}
	totalsByType := map[string]int64{}
	var total int64
	for _, r := range rows {
		day := r.OccurredAt.In(loc).Format("2006-01-02")
		m := byDay[day]
		if m == nil {
			m = map[string]int64{}
			byDay[day] = m
		}
		et := strings.TrimSpace(r.EventType)
		m[et]++
		totalsByType[et]++
		total++
	}

	// Build continuous series day-by-day.
	start := fromUTC.In(loc)
	end := toUTC.In(loc)
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)

	series := make([]eventsMetricsDay, 0)
	for d := startDay; d.Before(endDay); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		m := byDay[dateStr]
		if m == nil {
			m = map[string]int64{}
		}
		var dayTotal int64
		for _, v := range m {
			dayTotal += v
		}
		series = append(series, eventsMetricsDay{Date: dateStr, Total: dayTotal, ByType: m})
	}

	var resp eventsMetricsResponse
	resp.Range.From = fromUTC.Format(time.RFC3339)
	resp.Range.To = toUTC.Format(time.RFC3339)
	resp.TZ = tzName
	resp.Series = series
	resp.Totals.Total = total
	resp.Totals.ByType = totalsByType
	resp.AsOf = time.Now().UTC().Format(time.RFC3339)
	return resp, nil
}

func buildEventsMetricsCacheKey(fromUTC, toUTC time.Time, tzName, eventType string, productID uint) string {
	// Keep key readable and stable.
	fromPart := fromUTC.Format("20060102")
	toPart := toUTC.Format("20060102")
	if strings.TrimSpace(tzName) == "" {
		tzName = "UTC"
	}
	et := strings.TrimSpace(eventType)
	if et == "" {
		et = "-"
	}
	pid := "-"
	if productID > 0 {
		pid = strconv.FormatUint(uint64(productID), 10)
	}

	// Avoid ':' in key parts to keep separators stable.
	sanitize := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return "-"
		}
		s = strings.ReplaceAll(s, ":", "_")
		s = strings.ReplaceAll(s, " ", "_")
		return s
	}

	return fmt.Sprintf("eg:admin:events:metrics:v1:from=%s:to=%s:tz=%s:event_type=%s:product_id=%s", fromPart, toPart, sanitize(tzName), sanitize(et), sanitize(pid))
}

func (h *EventsHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	q := h.db.WithContext(c.Request.Context()).Model(&model.Event{})

	if et := strings.TrimSpace(c.Query("event_type")); et != "" {
		q = q.Where("event_type = ?", et)
	}
	if pid := strings.TrimSpace(c.Query("product_id")); pid != "" {
		if v, err := strconv.ParseUint(pid, 10, 64); err == nil && v > 0 {
			q = q.Where("product_id = ?", uint(v))
		}
	}

	if from := strings.TrimSpace(c.Query("from")); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where("occurred_at >= ?", t)
		}
	}
	if to := strings.TrimSpace(c.Query("to")); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where("occurred_at <= ?", t)
		}
	}

	limit := parseIntQuery(c, "limit", 100)
	offset := parseIntQuery(c, "offset", 0)
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin events query count failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var items []model.Event
	if err := q.Order("occurred_at desc, id desc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin events query list failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *EventsHandler) Get(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var e model.Event
	if err := h.db.WithContext(c.Request.Context()).First(&e, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, e)
}

func (h *EventsHandler) Delete(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res := h.db.WithContext(c.Request.Context()).Delete(&model.Event{}, uint(id))
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
