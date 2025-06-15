package handlers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ormasia/swiftstream/internal/oss/repo"

	"github.com/gofiber/fiber/v2"
)

type ChunkUploadResp struct {
	ChunkIndex int    `json:"chunkIndex"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
}

// Chunk handles the upload of a chunk by its ID and index.
func (h *Handlers) Chunk(c *fiber.Ctx) error {
	uploadID := c.Params("uploadid")
	chunkIndexStr := c.Params("chunkIndex")

	// TODO: 判断uploadID在不在数据库中
	if uploadID == "" || chunkIndexStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uploadId and chunkIndex are required",
		})
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chunkIndex",
		})
	}

	// 验证 uploadID 是否存在
	uploadTask, err := repo.GetUploadTask(h.db, uploadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Upload task not found",
		})
	}
	// 检查分片索引是否有效
	if chunkIndex < 0 || chunkIndex >= uploadTask.ChunkCount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chunkIndex",
		})
	}
	// 检查上传任务状态 TODO: 这里是不是秒传的逻辑？？
	if uploadTask.Status != "uploading" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Upload task is not in uploading status",
		})
	}
	// 查找对应的分片记录 所有的分片记录都提前写入了ChunkRecord
	chunkRecord, err := repo.GetChunkRecord(h.db, uploadID, chunkIndex)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Chunk record not found",
		})
	}
	// 检查分片是否已经上传
	if chunkRecord.Status == "uploaded" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Chunk already uploaded",
		})
	}
	// 获取上传的文件 从 fiber context 中获取分片文件
	file, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No chunk file provided",
		})
	}
	// 验证分片大小，TODO: 初步验证，没有进行完整的验证
	if file.Size != chunkRecord.ChunkSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Chunk size mismatch. Expected: %d, Got: %d", chunkRecord.ChunkSize, file.Size),
		})
	}
	// 创建存储目录
	uploadDir := filepath.Join("data", "uploads", uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload directory",
		})
	}
	// 保存分片文件
	chunkPath := filepath.Join(uploadDir, fmt.Sprintf("chunk_%d", chunkIndex))
	if err := c.SaveFile(file, chunkPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save chunk file",
		})
	}
	// 更新分片记录 状态和分片路径
	chunkRecord.Status = "uploaded"
	chunkRecord.FilePath = chunkPath
	if err := repo.SaveChunkRecord(h.db, chunkRecord); err != nil {
		// 如果更新失败，清理已保存的文件
		if err := os.Remove(chunkPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete chunk file",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update chunk record",
		})
	}

	// 更新上传任务进度 TODO: 后续可以优化为异步更新
	uploadedCount, err := repo.CountUploadedChunks(h.db, uploadID)
	if err != nil {
		log.Printf("Failed to count uploaded chunks: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update upload task progress",
		})
	}
	uploadTask.UploadedChunks = int(uploadedCount)
	uploadTask.UpdateProgress()

	if err := repo.SaveUploadTask(h.db, uploadTask); err != nil {
		log.Printf("Failed to update upload task progress: %v\n", err)
	}

	return c.Status(fiber.StatusOK).JSON(ChunkUploadResp{
		ChunkIndex: chunkIndex,
		Status:     "Chunk uploaded",
		Progress:   uploadTask.Progress,
	})
}
