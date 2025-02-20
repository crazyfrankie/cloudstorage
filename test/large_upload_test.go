/*
单机测试
[GIN] 2025/02/20 - 13:21:10 | 200 |     47.2648ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:10 | 200 |     36.5033ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:10 | 200 |     40.7846ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:10 | 200 |       31.46ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:10 | 200 |     27.6855ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:10 | 200 |     43.7245ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:10 | 200 |     28.5434ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     27.8008ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     32.5122ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     27.2115ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     30.6971ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     48.5015ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     28.7722ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     27.4866ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
[GIN] 2025/02/20 - 13:21:11 | 200 |     54.0089ms |       127.0.0.1 | POST     "/api/files/upload/chunk"
性能统计：
最快响应时间: 27.2115ms
最慢响应时间: 54.0089ms
平均响应时间: 约 35.5ms
总片数: 15片
单片大小: 约 5MB
性能表现：
单片吞吐量: 5MB/35.5ms ≈ 140MB/s
整体速度: 74.2MB/15片/35.5ms ≈ 140MB/s 或 1.12Gbps
响应时间波动范围在 27ms-54ms 之间，相对稳定
*/

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const (
	ChunkSize = 5 * 1024 * 1024 // 5MB
	UploadURL = "http://localhost:9091/api/files/upload/chunk"
	Token     = "token" // 替换成实际的token
)

func TestLargeFile(t *testing.T) {
	// 打开测试文件
	filename := "testfile"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()

	// 计算分片数
	chunks := (fileSize + ChunkSize - 1) / ChunkSize

	// 用于存储上传ID
	var uploadId string

	// 分片上传
	for i := int64(0); i < chunks; i++ {
		// 读取分片
		chunk := make([]byte, ChunkSize)
		n, err := file.ReadAt(chunk, i*ChunkSize)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// 创建multipart请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加分片文件
		part, err := writer.CreateFormFile("chunk", filepath.Base(filename))
		if err != nil {
			panic(err)
		}
		part.Write(chunk[:n])

		// 添加其他字段
		writer.WriteField("uploadId", uploadId)
		writer.WriteField("partNumber", strconv.FormatInt(i+1, 10))
		writer.WriteField("fileSize", strconv.FormatInt(fileSize, 10))
		writer.WriteField("folder", "0") // 根目录
		writer.WriteField("isLast", strconv.FormatBool(i == chunks-1))

		writer.Close()

		// 创建请求
		req, err := http.NewRequest("POST", UploadURL, body)
		if err != nil {
			panic(err)
		}

		// 设置请求头
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+Token)

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		// 解析响应
		var result struct {
			Code int `json:"code"`
			Data struct {
				UploadId string `json:"upload_id"`
				Etag     string `json:"etag"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			panic(err)
		}

		fmt.Println(result)

		// 第一个分片返回的uploadId保存下来
		if i == 0 {
			uploadId = result.Data.UploadId
		}

		fmt.Printf("Chunk %d/%d uploaded, ETag: %s\n", i+1, chunks, result.Data.Etag)
	}

	fmt.Println("Upload completed!")
}
