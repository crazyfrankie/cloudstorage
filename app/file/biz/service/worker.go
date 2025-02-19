package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/file/mws"
)

type DownloadWorker interface {
	Run()
}

type RedisWorker struct {
	repo      *repository.UploadRepo
	minio     *mws.MinioServer
	stopCh    chan struct{}
	workerNum int
}

func NewRedisWorker(repo *repository.UploadRepo, minio *mws.MinioServer) DownloadWorker {
	return &RedisWorker{
		repo:      repo,
		minio:     minio,
		stopCh:    make(chan struct{}),
		workerNum: 3,
	}
}

func (w *RedisWorker) Start() {
	for i := 0; i < w.workerNum; i++ {
		go w.Run()
	}
}

func (w *RedisWorker) Stop() {
	close(w.stopCh)
}

func (w *RedisWorker) Run() {
	for {
		select {
		case <-w.stopCh:
			return
		default:
			// 获取下一个任务
			taskId, err := w.repo.GetNextDownloadTask(context.Background())
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			w.processTask(taskId)
		}
	}
}

func (w *RedisWorker) processTask(taskId string) {
	ctx := context.Background()
	task, err := w.repo.GetDownloadTaskInfo(ctx, taskId)
	if err != nil {
		return
	}

	// 创建临时下载目录
	tmpDir := filepath.Join(os.TempDir(), "downloads", taskId)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	var wg sync.WaitGroup
	for _, file := range task.Files {
		wg.Add(1)
		go func(file *cache.DownloadedFile) {
			defer wg.Done()

			// 检查是否已下载
			if file.Status == "completed" {
				return
			}

			// 获取文件对象
			obj, err := w.minio.GetObject(ctx, w.minio.BucketName, file.Name)
			if err != nil {
				file.Status = "failed"
				return
			}
			defer obj.Close()

			// 创建本地文件
			filePath := filepath.Join(tmpDir, file.Path)
			os.MkdirAll(filepath.Dir(filePath), 0755)

			f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				file.Status = "failed"
				return
			}
			defer f.Close()

			// 设置偏移量，支持断点续传
			if file.Downloaded > 0 {
				obj.Seek(file.Downloaded, io.SeekStart)
				f.Seek(file.Downloaded, io.SeekStart)
			}

			// 拷贝数据并更新进度
			written, err := io.Copy(f, obj)
			if err != nil {
				file.Status = "failed"
				return
			}

			file.Downloaded += written
			file.Status = "completed"

			// 更新任务进度
			w.repo.UpdateTaskStatus(ctx, taskId, "processing", task.Progress+written)
		}(file)
	}

	wg.Wait()
}
