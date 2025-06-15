package repo

import (
	"github.com/ormasia/swiftstream/internal/oss/model"
	"gorm.io/gorm"
)

// ============================================================================
// 数据库初始化
// ============================================================================

// CreateTable 自动迁移所有模型到数据库
func CreateTable(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.OssObject{},
		&model.UploadTask{},
		&model.ChunkRecord{},
	); err != nil {
		return err
	}
	return nil
}

// ============================================================================
// OssObject 操作
// ============================================================================

// CreateObject 创建 OSS 对象记录
func CreateObject(db *gorm.DB, object *model.OssObject) error {
	if err := db.Create(object).Error; err != nil {
		return err
	}
	return nil
}

// GetObject 根据对象 ID 获取 OSS 对象记录
func GetObject(db *gorm.DB, objectID string) (*model.OssObject, error) {
	var object model.OssObject
	if err := db.Where("object_id = ?", objectID).First(&object).Error; err != nil {
		return nil, err
	}
	return &object, nil
}

func GetObjectByEtag(db *gorm.DB, etag string) (*model.OssObject, error) {
	var object model.OssObject
	if err := db.Where("e_tag = ?", etag).First(&object).Error; err != nil {
		return nil, err
	}
	return &object, nil
}

// ============================================================================
// UploadTask 操作
// ============================================================================

// CreateUploadTask 创建上传任务
func CreateUploadTask(db *gorm.DB, uploadTask *model.UploadTask) error {
	if err := db.Create(uploadTask).Error; err != nil {
		return err
	}
	return nil
}

// GetUploadTask 根据 uploadID 获取上传任务
func GetUploadTask(db *gorm.DB, uploadID string) (*model.UploadTask, error) {
	var uploadTask model.UploadTask
	if err := db.Where("upload_id = ?", uploadID).First(&uploadTask).Error; err != nil {
		return nil, err
	}
	return &uploadTask, nil
}

// SaveUploadTask 保存上传任务
func SaveUploadTask(db *gorm.DB, uploadTask *model.UploadTask) error {
	if err := db.Save(uploadTask).Error; err != nil {
		return err
	}
	return nil
}

// ============================================================================
// ChunkRecord 操作
// ============================================================================

// CreateChunkRecords 批量创建分片记录
func CreateChunkRecords(db *gorm.DB, chunkRecords []model.ChunkRecord) error {
	if err := db.CreateInBatches(&chunkRecords, 100).Error; err != nil {
		return err
	}
	return nil
}

// GetChunkRecord 获取单个分片记录
func GetChunkRecord(db *gorm.DB, uploadID string, chunkIndex int) (*model.ChunkRecord, error) {
	var chunkRecord model.ChunkRecord
	if err := db.Where("upload_id = ? AND chunk_index = ?", uploadID, chunkIndex).First(&chunkRecord).Error; err != nil {
		return nil, err
	}
	return &chunkRecord, nil
}

// GetChunkRecords 获取上传任务的所有分片记录，按索引排序
func GetChunkRecords(db *gorm.DB, uploadID string) ([]model.ChunkRecord, error) {
	var chunkRecords []model.ChunkRecord
	if err := db.Where("upload_id = ?", uploadID).
		Order("chunk_index ASC").Find(&chunkRecords).Error; err != nil {
		return nil, err
	}
	return chunkRecords, nil
}

// SaveChunkRecord 保存分片记录
func SaveChunkRecord(db *gorm.DB, chunkRecord *model.ChunkRecord) error {
	if err := db.Save(chunkRecord).Error; err != nil {
		return err
	}
	return nil
}

// DeleteChunkRecords 删除上传任务的所有分片记录
func DeleteChunkRecords(db *gorm.DB, uploadID string) error {
	if err := db.Where("upload_id = ?", uploadID).Delete(&model.ChunkRecord{}).Error; err != nil {
		return err
	}
	return nil
}

// CountUploadedChunks 统计已上传的分片数量
func CountUploadedChunks(db *gorm.DB, uploadID string) (int64, error) {
	var count int64
	if err := db.Model(&model.ChunkRecord{}).
		Where("upload_id = ? AND status = ?", uploadID, "uploaded").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
