package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ormasia/swiftstream/internal/oss/model"
	"github.com/ormasia/swiftstream/internal/oss/repo"
)

type CompleteResp struct {
	Status    string `json:"status"`
	FileURL   string `json:"fileUrl"`
	FileSize  int64  `json:"fileSize"`
	FileName  string `json:"fileName"`
	ObjectKey string `json:"objectKey"`
	ETag      string `json:"etag"`
	ObjectID  uint   `json:"objectId"`
}

// Complete处理分片文件的合并，生成最终文件
func (h *Handlers) Complete(c *fiber.Ctx) error {
	uploadID := c.Params("uploadid")
	if uploadID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uploadId is required",
		})
	}

	// 获取上传任务
	uploadTask, err := repo.GetUploadTask(h.db, uploadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Upload task not found",
		})
	}
	// 检查上传任务状态
	if uploadTask.Status == "completed" {
		// // 获取ObjectID
		// ObjectID := repo.GetObjectIDByUploadID(h.db, uploadID)

		return c.JSON(CompleteResp{
			Status:    "completed",
			FileURL:   uploadTask.URL,
			FileSize:  uploadTask.FileSize,
			FileName:  uploadTask.FileName,
			ObjectKey: uploadTask.ObjectKey,
			ETag:      uploadTask.ETag,
			ObjectID:  0, // TODO: 需要从 OssObject 中获取
		})
	}

	if uploadTask.Status != "uploading" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Upload task is not in uploading status",
		})
	}

	// 检查所有分片是否都已上传
	chunkRecords, err := repo.GetChunkRecords(h.db, uploadID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get chunk records",
		})
	}
	uploadedChunksCount := len(chunkRecords)

	if int(uploadedChunksCount) != uploadTask.ChunkCount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Not all chunks uploaded. Expected: %d, Uploaded: %d", uploadTask.ChunkCount, uploadedChunksCount),
		})
	}

	// 合并分片文件
	finalFilePath := filepath.Join("data", "files", uploadTask.FileName)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(finalFilePath), 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create file directory",
		})
	}
	// 创建最终文件
	finalFile, err := os.Create(finalFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create final file",
		})
	}
	defer finalFile.Close()
	// 创建 MD5 哈希计算器 MD5 是流式哈希算法，按顺序流式处理等于一次性计算整个文件
	hash := md5.New()
	multiWriter := io.MultiWriter(finalFile, hash)

	// 合并所有分片
	for _, chunk := range chunkRecords {
		chunkFile, err := os.Open(chunk.FilePath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to open chunk file: %s", chunk.FilePath),
			})
		}

		_, err = io.Copy(multiWriter, chunkFile)
		chunkFile.Close()

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to merge chunk",
			})
		}

		// 删除分片文件
		os.Remove(chunk.FilePath)
	}
	// 计算文件的 ETag (MD5)
	etag := fmt.Sprintf("%x", hash.Sum(nil))

	// 生成文件URL和对象键（使用uploadID确保唯一性）
	objectKey := fmt.Sprintf("uploads/%s/%s_%s", time.Now().Format("2006/01/02"), uploadID, uploadTask.FileName)
	fileURL := fmt.Sprintf("/files/%s", uploadTask.FileName)

	// 创建 OssObject 记录
	ossObject := model.OssObject{
		FileName:   uploadTask.FileName,
		FileSize:   uploadTask.FileSize,
		FileType:   uploadTask.FileType,
		MimeType:   uploadTask.FileType, // TODO: 需要根据文件扩展名确定正确的 MIME 类型
		Bucket:     "default",           // TODO: 从配置中获取
		ObjectKey:  objectKey,
		ETag:       etag,
		URL:        fileURL,
		UserID:     uploadTask.UserID,
		BusinessID: uploadTask.BusinessID,
		Status:     "active",
	}

	// 保存 OSS 对象记录
	if err := repo.CreateObject(h.db, &ossObject); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create OSS object record",
		})
	}
	// 更新上传任务状态
	uploadTask.Status = "completed"
	uploadTask.URL = fileURL
	uploadTask.ObjectKey = objectKey
	uploadTask.ETag = etag
	uploadTask.Progress = 100

	if err := repo.SaveUploadTask(h.db, uploadTask); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update upload task",
		})
	}

	// 清理分片记录和临时目录
	if err := repo.DeleteChunkRecords(h.db, uploadID); err != nil {
		// 记录日志但不影响响应
		fmt.Printf("Failed to delete chunk records: %v\n", err)
	}
	uploadDir := filepath.Join("data", "uploads", uploadID)
	os.RemoveAll(uploadDir)

	return c.Status(fiber.StatusOK).JSON(CompleteResp{
		Status:    "completed",
		FileURL:   fileURL,
		FileSize:  uploadTask.FileSize,
		FileName:  uploadTask.FileName,
		ObjectKey: objectKey,
		ETag:      etag,
		ObjectID:  ossObject.ID,
	})
}
