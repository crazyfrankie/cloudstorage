package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DownloadTaskPrefix  = "download:task:"  // 任务信息
	DownloadQueueKey    = "download:queue"  // 下载队列
	DownloadFilesPrefix = "download:files:" // 任务文件列表
)

type FileCache struct {
	cmd redis.Cmdable
}

type DownloadTask struct {
	UserId     int32             `json:"user_id"`
	Status     string            `json:"status"` // pending/processing/completed/failed
	FolderName string            `json:"folder_name"`
	TotalSize  int64             `json:"total_size"`
	Progress   int64             `json:"progress"`
	CreatedAt  time.Time         `json:"created_at"`
	Files      []*DownloadedFile `json:"files"` // 添加文件列表
}

type DownloadedFile struct {
	FileId int64  `json:"file_id"`
	Name   string `json:"name"` // 添加文件名
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Status string `json:"status"` // pending/processing/completed/failed
}

func NewFileCache(cmd redis.Cmdable) *FileCache {
	return &FileCache{cmd: cmd}
}

// CreateDownloadTask 创建下载任务
func (c *FileCache) CreateDownloadTask(ctx context.Context, taskId string, info *DownloadTask) error {
	taskKey := DownloadTaskPrefix + taskId
	filesKey := DownloadFilesPrefix + taskId

	pipe := c.cmd.Pipeline()

	// 存储任务信息
	taskData, _ := json.Marshal(info)
	pipe.HSet(ctx, taskKey, "info", taskData)

	// 加入下载队列
	pipe.LPush(ctx, DownloadQueueKey, taskId)

	// 设置过期时间
	pipe.Expire(ctx, taskKey, time.Hour*24)
	pipe.Expire(ctx, filesKey, time.Hour*24)

	_, err := pipe.Exec(ctx)
	return err
}

// GetNextDownloadTask 获取下载队列中的下一个任务
func (c *FileCache) GetNextDownloadTask(ctx context.Context) (string, error) {
	result, err := c.cmd.BRPop(ctx, 0, DownloadQueueKey).Result()
	if err != nil {
		return "", err
	}
	return result[1], nil
}

// GetDownloadTaskInfo 获取任务详细信息
func (c *FileCache) GetDownloadTaskInfo(ctx context.Context, taskId string) (*DownloadTask, error) {
	taskKey := DownloadTaskPrefix + taskId

	// 获取任务信息
	data, err := c.cmd.HGet(ctx, taskKey, "info").Bytes()
	if err != nil {
		return nil, err
	}

	var task DownloadTask
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateTaskStatus 更新任务状态
func (c *FileCache) UpdateTaskStatus(ctx context.Context, taskId string, status string, progress int64) error {
	taskKey := DownloadTaskPrefix + taskId

	info, err := c.GetDownloadTaskInfo(ctx, taskId)
	if err != nil {
		return err
	}

	info.Status = status
	info.Progress = progress

	taskData, _ := json.Marshal(info)
	return c.cmd.HSet(ctx, taskKey, "info", taskData).Err()
}
