package test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	DownloadURL = "http://localhost:9091/api/files/download/1"
	//Token       = "your-token"
	ChunkSize = 5 * 1024 * 1024 // 5MB
)

func TestDownloadLargeFile(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, DownloadURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+Token)

	client := &http.Client{}

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	// 获取并处理文件名
	contentDisposition := resp.Header.Get("Content-Disposition")
	filename := "downloaded_file"
	if contentDisposition != "" {
		// 从 Content-Disposition 中提取文件名
		if start := strings.Index(contentDisposition, "filename="); start != -1 {
			// 移除 "filename=" 和引号
			filename = contentDisposition[start+len("filename="):]
			filename = strings.Trim(filename, `"`)
			// URL 解码
			if decodedName, err := url.QueryUnescape(filename); err == nil {
				filename = decodedName
			}
			// 处理非法字符
			filename = sanitizeFilename(filename)
		}
	}

	contentLength := resp.Header.Get("Content-Length")
	fileSize, _ := strconv.ParseInt(contentLength, 10, 64)

	// 创建目标文件
	outputPath := filepath.Join("downloads", filename)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatal(err)
	}
	out, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	// 使用缓冲区接收流式数据
	buffer := make([]byte, ChunkSize)
	var totalReceived int64

	for {
		n, err := resp.Body.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		// 写入文件
		if _, err := out.Write(buffer[:n]); err != nil {
			t.Fatal(err)
		}

		// 更新进度
		totalReceived += int64(n)
		progress := float64(totalReceived) / float64(fileSize) * 100
		fmt.Printf("\rDownloading... %.2f%% (%d/%d bytes)", progress, totalReceived, fileSize)
	}

	// 计算并显示统计信息
	duration := time.Since(startTime)
	speed := float64(totalReceived) / duration.Seconds() / 1024 / 1024 // MB/s

	fmt.Printf("\nDownload completed:\n")
	fmt.Printf("Total size: %.2f MB\n", float64(totalReceived)/1024/1024)
	fmt.Printf("Time taken: %v\n", duration)
	fmt.Printf("Average speed: %.2f MB/s\n", speed)

	// 验证文件大小
	info, err := out.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != totalReceived {
		t.Errorf("file size mismatch: got %d, want %d", info.Size(), totalReceived)
	}
}

// sanitizeFilename 清理文件名中的非法字符
func sanitizeFilename(filename string) string {
	// 替换 Windows 文件名中的非法字符
	illegalChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range illegalChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// 如果文件名过长，进行截断
	const maxLength = 200
	if len(result) > maxLength {
		ext := filepath.Ext(result)
		result = result[:maxLength-len(ext)] + ext
	}

	return result
}
