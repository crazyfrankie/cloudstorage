package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/file/mws"
)

type DownloadWorker struct {
	repo      *repository.UploadRepo
	minio     *mws.MinioServer
	stopCh    chan struct{}
	workerNum int
}

func NewDownloadWorker(repo *repository.UploadRepo, minio *mws.MinioServer) *DownloadWorker {
	return &DownloadWorker{
		repo:      repo,
		minio:     minio,
		stopCh:    make(chan struct{}),
		workerNum: 3,
	}
}

func (w *DownloadWorker) Start() {
	for i := 0; i < w.workerNum; i++ {
		go w.run()
	}
}

func (w *DownloadWorker) Stop() {
	close(w.stopCh)
}

func (w *DownloadWorker) run() {
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

func (w *DownloadWorker) processTask(taskId string) {
	ctx := context.Background()

	// 获取任务信息
	task, err := w.repo.GetDownloadTaskInfo(ctx, taskId)
	if err != nil {
		err := w.repo.UpdateTaskStatus(ctx, taskId, "failed", 0)
		if err != nil {
			fmt.Printf("update task status failed, taskId:%s", taskId)
			return
		}
		return
	}

	// 更新任务状态为处理中
	err = w.repo.UpdateTaskStatus(ctx, taskId, "processing", 0)
	if err != nil {
		fmt.Printf("update task status failed, taskId:%s", taskId)
		return
	}

	// 创建临时下载目录
	tmpDir := filepath.Join(os.TempDir(), "downloads", taskId)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	var totalSize int64
	var wg sync.WaitGroup
	for _, file := range task.Files {
		wg.Add(1)
		go func(file *cache.DownloadedFile) {
			defer wg.Done()

			// 下载文件
			obj, err := w.minio.GetObject(ctx, w.minio.BucketName, file.Name)
			if err != nil {
				file.Status = "failed"
				return
			}
			defer obj.Close()

			// 保存到临时目录
			filePath := filepath.Join(tmpDir, file.Path)
			os.MkdirAll(filepath.Dir(filePath), 0755)

			f, err := os.Create(filePath)
			if err != nil {
				file.Status = "failed"
				return
			}
			defer f.Close()

			written, err := io.Copy(f, obj)
			if err != nil {
				file.Status = "failed"
				return
			}

			file.Status = "completed"
			atomic.AddInt64(&totalSize, written)

			// 更新进度
			err = w.repo.UpdateTaskStatus(ctx, taskId, "processing", atomic.LoadInt64(&totalSize))
			if err != nil {
				fmt.Printf("update task status failed, taskId:%s", taskId)
				return
			}
		}(file)
	}

	wg.Wait()

	// 更新任务状态为完成
	err = w.repo.UpdateTaskStatus(ctx, taskId, "completed", totalSize)
	if err != nil {
		fmt.Printf("update task status failed, taskId:%s", taskId)
		return
	}
}
