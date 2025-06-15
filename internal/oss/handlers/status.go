package handlers

import (
	"github.com/ormasia/swiftstream/internal/oss/repo"

	"github.com/gofiber/fiber/v2"
)

type StatusResp struct {
	UploadID       string              `json:"uploadId"`
	Status         string              `json:"status"`
	Progress       int                 `json:"progress"`
	UploadedChunks int                 `json:"uploadedChunks"`
	TotalChunks    int                 `json:"totalChunks"`
	ChunkStatus    []ChunkStatusDetail `json:"chunkStatus"`
}

type ChunkStatusDetail struct {
	ChunkIndex int    `json:"chunkIndex"`
	Status     string `json:"status"`
}

// Status 查询上传任务状态
func (h *Handlers) Status(c *fiber.Ctx) error {
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

	// 获取分片状态
	chunkRecords, err := repo.GetChunkRecords(h.db, uploadID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get chunk status",
		})
	}

	// 构造分片状态详情
	chunkStatus := make([]ChunkStatusDetail, len(chunkRecords))
	for i, chunk := range chunkRecords {
		chunkStatus[i] = ChunkStatusDetail{
			ChunkIndex: chunk.ChunkIndex,
			Status:     chunk.Status,
		}
	}

	return c.JSON(StatusResp{
		UploadID:       uploadTask.UploadID,
		Status:         uploadTask.Status,
		Progress:       uploadTask.Progress,
		UploadedChunks: uploadTask.UploadedChunks,
		TotalChunks:    uploadTask.ChunkCount,
		ChunkStatus:    chunkStatus,
	})
}
