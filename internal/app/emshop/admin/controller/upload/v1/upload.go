package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/pkg/common/core"
	"emshop/pkg/log"
)

type uploadController struct {
	trans restserver.I18nTranslator
}

func NewUploadController(trans restserver.I18nTranslator) *uploadController {
	return &uploadController{
		trans: trans,
	}
}

// 允许的图片格式
var allowedImageTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// UploadImage 上传商品图片（管理员专用）
func (uc *uploadController) UploadImage(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to get file from request"})
		return
	}
	defer file.Close()

	// 验证文件大小（最大5MB）
	if header.Size > 5*1024*1024 {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "file size too large, max 5MB"})
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageTypes[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid file type, only jpg, jpeg, png, gif, webp allowed"})
		return
	}

	// 创建上传目录
	uploadDir := "./uploads/goods/images/" + time.Now().Format("2006/01/02")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Errorf("Failed to create upload directory: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := uc.saveUploadedFile(file, filePath); err != nil {
		log.Errorf("Failed to save uploaded file: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 返回文件URL
	fileURL := "/uploads/goods/images/" + time.Now().Format("2006/01/02") + "/" + filename
	
	response := map[string]interface{}{
		"url":      fileURL,
		"filename": filename,
		"size":     header.Size,
		"type":     ext,
	}

	core.WriteResponse(ctx, nil, response)
}

// BatchUploadImages 批量上传商品图片（管理员专用）
func (uc *uploadController) BatchUploadImages(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "no files provided"})
		return
	}

	// 限制批量上传数量（最多10个文件）
	if len(files) > 10 {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "too many files, max 10 files allowed"})
		return
	}

	var results []map[string]interface{}
	var errors []string

	// 创建上传目录
	uploadDir := "./uploads/goods/images/" + time.Now().Format("2006/01/02")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Errorf("Failed to create upload directory: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	for _, header := range files {
		// 验证单个文件
		if header.Size > 5*1024*1024 {
			errors = append(errors, fmt.Sprintf("%s: file size too large", header.Filename))
			continue
		}

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if !allowedImageTypes[ext] {
			errors = append(errors, fmt.Sprintf("%s: invalid file type", header.Filename))
			continue
		}

		// 打开文件
		file, err := header.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to open file", header.Filename))
			continue
		}

		// 生成文件名并保存
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
		filePath := filepath.Join(uploadDir, filename)

		if err := uc.saveUploadedFile(file, filePath); err != nil {
			file.Close()
			errors = append(errors, fmt.Sprintf("%s: failed to save file", header.Filename))
			continue
		}

		file.Close()

		// 添加成功结果
		fileURL := "/uploads/goods/images/" + time.Now().Format("2006/01/02") + "/" + filename
		results = append(results, map[string]interface{}{
			"url":      fileURL,
			"filename": filename,
			"size":     header.Size,
			"type":     ext,
		})
	}

	response := map[string]interface{}{
		"success": results,
		"errors":  errors,
		"total":   len(files),
		"successCount": len(results),
		"errorCount": len(errors),
	}

	core.WriteResponse(ctx, nil, response)
}

// saveUploadedFile 保存上传的文件
func (uc *uploadController) saveUploadedFile(src multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}