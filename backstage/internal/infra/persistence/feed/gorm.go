package infrafeed

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

// EnsureTimelineIndex 创建 timeline 回源查询所需索引。
func EnsureTimelineIndex(db *gorm.DB) error {
	var count int64
	err := db.Raw(
		"SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		"video",
		"idx_video_timeline",
	).Scan(&count).Error
	if err != nil {
		return err
	}
	// 索引已经存在
	if count > 0 {
		return nil
	}
	// 索引不存在，则创建
	return db.Exec("CREATE INDEX idx_video_timeline ON video(status, published_at DESC, id DESC)").Error
}
