package handlers

import (
	"gptimg/internal/repository"
	"gptimg/internal/services"
	"gptimg/pkg/response"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *services.ImageService
	imageRepo    *repository.ImageRepository
}

func NewImageHandler(imageService *services.ImageService, imageRepo *repository.ImageRepository) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
		imageRepo:    imageRepo,
	}
}

func (h *ImageHandler) Generate(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req services.GenerateImageRequest
	if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		if err := c.ShouldBind(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		form, err := c.MultipartForm()
		if err == nil && form != nil {
			fileHeaders := append([]*multipart.FileHeader{}, form.File["reference_image"]...)
			fileHeaders = append(fileHeaders, form.File["reference_images"]...)

			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					response.BadRequest(c, "Failed to read reference image")
					return
				}

				data, err := io.ReadAll(file)
				file.Close()
				if err != nil {
					response.BadRequest(c, "Failed to read reference image")
					return
				}

				referenceImage := services.ReferenceImage{
					Data:        data,
					Name:        fileHeader.Filename,
					ContentType: fileHeader.Header.Get("Content-Type"),
				}
				req.ReferenceImages = append(req.ReferenceImages, referenceImage)
			}

			if len(req.ReferenceImages) > 0 {
				req.ReferenceImageData = req.ReferenceImages[0].Data
				req.ReferenceImageName = req.ReferenceImages[0].Name
				req.ReferenceImageType = req.ReferenceImages[0].ContentType
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	record, err := h.imageService.GenerateImage(userID.(int), &req)
	if err != nil {
		if err.Error() == "insufficient credits" {
			response.BadRequest(c, "Insufficient credits")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, record)
}

func (h *ImageHandler) GetByID(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	record, err := h.imageRepo.FindByID(id)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if record == nil {
		response.NotFound(c, "Image not found")
		return
	}

	if record.UserID != userID.(int) {
		response.Forbidden(c, "Access denied")
		return
	}

	response.Success(c, record)
}

func (h *ImageHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := h.imageRepo.Delete(id, userID.(int)); err != nil {
		response.InternalError(c, "Failed to delete image")
		return
	}

	response.Success(c, gin.H{"message": "Image deleted successfully"})
}

func (h *ImageHandler) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, err := h.imageRepo.FindByUserID(userID.(int), limit, offset)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	response.Success(c, records)
}
