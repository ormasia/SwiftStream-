package handlers

import (
	"github.com/ormasia/swiftstream/internal/oss/model"
	"github.com/ormasia/swiftstream/internal/oss/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type InitReq struct {
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	FileType  string `json:"file_type"`
	ChunkSize int64  `json:"chunk_size"`
	FileMD5   string `json:"file_md5"` // 文件的MD5值，用于秒传 对应 model.OssObject.ETag
}

type InitResp struct {
	UploadID   string `json:"uploadId"`
	ChunkCount int    `json:"chunkCount"` // 分片总数
}

func (h *Handlers) Init(c *fiber.Ctx) error {
	// 解析请求体
	var req InitReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ //返回json格式
			"error": err.Error(),
		})
	}

	// 验证请求参数
	if req.FileName == "" || req.FileSize <= 0 || req.ChunkSize <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid parameters",
		})
	}

	// 生成唯一上传UploadID
	uploadID := uuid.New().String()

	// 检查文件是否已存在（秒传）
	if req.FileMD5 != "" {
		existingObject, err := repo.GetObjectByEtag(h.db, req.FileMD5)
		if err == nil && existingObject != nil {
			// 文件已存在，创建上传任务并直接标记完成
			uploadTask := model.UploadTask{
				UploadID:   uploadID,
				FileName:   req.FileName,
				FileSize:   req.FileSize,
				FileType:   req.FileType,
				ChunkSize:  req.ChunkSize,
				ChunkCount: 0, // 秒传时分片数为0
				Status:     "completed",
				Progress:   100,
				URL:        existingObject.URL,
				ObjectKey:  existingObject.ObjectKey,
				ETag:       existingObject.ETag, //
			}

			if err := repo.SaveUploadTask(h.db, &uploadTask); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create upload task",
				})
			}

			// 返回秒传结果
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"uploadId":      uploadID,
				"chunkCount":    0,
				"instantUpload": true,
				"fileUrl":       existingObject.URL,
				"objectKey":     existingObject.ObjectKey,
			})
		}
	}

	// 计算分片总数
	chunkCount := int(req.FileSize / req.ChunkSize)
	if req.FileSize%req.ChunkSize != 0 {
		chunkCount++
	}

	// 创建上传任务记录
	uploadTask := model.UploadTask{
		UploadID:   uploadID,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		FileType:   req.FileType,
		ChunkSize:  req.ChunkSize,
		ChunkCount: chunkCount,
		Status:     "uploading",
		// TODO: 从context中获取用户ID
		// UserID: getUserID(c),
	}

	// 保存上传任务到数据库
	if err := repo.SaveUploadTask(h.db, &uploadTask); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload task",
		})
	}

	// 创建分片记录
	var chunkRecords []model.ChunkRecord
	for i := range chunkCount {
		chunkSize := req.ChunkSize
		// 最后一个分片可能小于标准分片大小
		if i == chunkCount-1 {
			chunkSize = req.FileSize - int64(i)*req.ChunkSize // 不允许不同数值类型之间的隐式转换
		}

		chunkRecords = append(chunkRecords, model.ChunkRecord{
			UploadTaskID: uploadTask.ID,
			UploadID:     uploadID,
			ChunkIndex:   i,
			ChunkSize:    chunkSize,
			Status:       "pending",
		})
	}
	// 批量创建分片记录
	if err := repo.CreateChunkRecords(h.db, chunkRecords); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create chunk records",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(InitResp{
		UploadID:   uploadID,
		ChunkCount: chunkCount,
	})
}
