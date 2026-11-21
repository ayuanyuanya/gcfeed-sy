package infravideo

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

// EnsureStats 确保每个视频都有一条统计记录。
func EnsureStats(db *gorm.DB) error {
	// video_stat:视频关联表，统计视频的喜欢，评论，收藏数量
	// video: 视频表
	tx := db.Exec(`
       INSERT INTO video_stat (video_id, like_count, comment_count, favorite_count, created_at, updated_at)
       SELECT v.id,0,0,0,NOW(),NOW()
       FROM video AS v
       LEFT JOIN video_stat AS vs ON vs.video_id = v.id
       WHERE vs.video_id IS NULL
       `)
	return tx.Error
}
