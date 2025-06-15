package model

import (
	"time"

	"gorm.io/gorm"
)

// OssObject represents an object stored in OSS
type OssObject struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 文件基本信息
	FileName string `json:"file_name" gorm:"not null"`
	FileSize int64  `json:"file_size" gorm:"not null"`
	FileType string `json:"file_type" gorm:"not null"` // image/video/audio
	MimeType string `json:"mime_type" gorm:"not null"`

	// OSS 存储信息
	Bucket    string `json:"bucket" gorm:"not null"`
	ObjectKey string `json:"object_key" gorm:"not null;uniqueIndex"`
	ETag      string `json:"etag"`

	// 访问信息
	URL    string `json:"url"`     // 文件访问URL
	CDNUrl string `json:"cdn_url"` // CDN访问URL

	// 元数据
	Width    int `json:"width"`    // 图片/视频宽度
	Height   int `json:"height"`   // 图片/视频高度
	Duration int `json:"duration"` // 视频/音频时长(秒)

	// 业务信息
	UserID     uint   `json:"user_id" gorm:"index"`
	BusinessID string `json:"business_id" gorm:"index"`       // 业务关联ID
	Status     string `json:"status" gorm:"default:'active'"` // active/deleted
}

// TableName 指定表名
func (OssObject) TableName() string {
	return "oss_objects"
}

// UploadTask 分片上传任务
type UploadTask struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 上传任务信息
	UploadID   string `json:"upload_id" gorm:"uniqueIndex;not null" comment:"唯一上传ID"`
	FileName   string `json:"file_name" gorm:"not null" comment:"上传的文件名"`
	FileSize   int64  `json:"file_size" gorm:"not null" comment:"文件大小"`
	FileType   string `json:"file_type" gorm:"not null" comment:"文件类型"`
	ChunkSize  int64  `json:"chunk_size" gorm:"not null" comment:"分片大小"`
	ChunkCount int    `json:"chunk_count" gorm:"not null" comment:"分片总数"`

	// 上传状态
	Status         string `json:"status" gorm:"default:'uploading'" comment:"上传状态"`   // uploading/completed/failed/cancelled
	UploadedChunks int    `json:"uploaded_chunks" gorm:"default:0" comment:"已上传分片数量"` // 已上传分片数量
	Progress       int    `json:"progress" gorm:"default:0" comment:"上传进度百分比"`        // 上传进度百分比

	// 完成后的文件信息
	ObjectKey string `json:"object_key"`
	URL       string `json:"url"` // 文件访问URL
	ETag      string `json:"etag"`

	// 业务信息
	UserID     uint   `json:"user_id" gorm:"index"`
	BusinessID string `json:"business_id" gorm:"index"`
}

// ChunkRecord 分片记录
type ChunkRecord struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联上传任务
	UploadTaskID uint       `json:"upload_task_id" gorm:"index;not null"`
	UploadTask   UploadTask `json:"upload_task" gorm:"foreignKey:UploadTaskID"` // 将 UploadTask 设置为外键关联
	UploadID     string     `json:"upload_id" gorm:"index;not null"`

	// 分片信息
	ChunkIndex int    `json:"chunk_index" gorm:"not null"`
	ChunkSize  int64  `json:"chunk_size" gorm:"not null"`
	Status     string `json:"status" gorm:"default:'pending'"` // pending/uploaded/failed 	待上传/上传成功/失败
	FilePath   string `json:"file_path"`                       // 分片文件存储路径
	ETag       string `json:"etag"`                            // 分片文件的ETag
}

// TableName 指定表名
func (UploadTask) TableName() string {
	return "upload_tasks"
}

// TableName 指定表名
func (ChunkRecord) TableName() string {
	return "chunk_records"
}

// IsCompleted 检查上传任务是否完成
func (ut *UploadTask) IsCompleted() bool {
	return ut.UploadedChunks >= ut.ChunkCount
}

// UpdateProgress 更新上传进度
func (ut *UploadTask) UpdateProgress() {
	if ut.ChunkCount > 0 {
		ut.Progress = (ut.UploadedChunks * 100) / ut.ChunkCount
	}
}
