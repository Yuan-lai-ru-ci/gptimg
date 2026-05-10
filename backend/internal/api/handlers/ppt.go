package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/internal/services"
	"gptimg/pkg/response"
	"html"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PPTHandler struct {
	llmService   *services.LLMService
	imageService *services.ImageService
	sessionRepo  *repository.SessionRepository
}

func NewPPTHandler(llmService *services.LLMService, imageService *services.ImageService, sessionRepo *repository.SessionRepository) *PPTHandler {
	return &PPTHandler{llmService: llmService, imageService: imageService, sessionRepo: sessionRepo}
}

func (h *PPTHandler) Plan(c *gin.Context) {
	var req services.PlanPPTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	plan, err := h.llmService.PlanPPT(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, plan)
}

func (h *PPTHandler) PlanDocument(c *gin.Context) {
	fileHeader, err := c.FormFile("document")
	if err != nil {
		response.BadRequest(c, "missing document")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.BadRequest(c, "failed to open document")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, 12*1024*1024))
	file.Close()
	if err != nil {
		response.BadRequest(c, "failed to read document")
		return
	}

	documentText, err := extractDocumentText(fileHeader.Filename, data)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	requirement := strings.TrimSpace(c.PostForm("requirement"))
	if requirement == "" {
		requirement = "请根据上传文档生成一套中文科创类PPT。"
	}

	combinedRequirement := fmt.Sprintf(
		"%s\n\n请先理解下面文档，提炼主题、章节和关键论点，再拆成适合逐页生成图片的PPT结构。不要照抄长段落，要把每页压缩为清晰标题和短要点。\n\n文档名称：%s\n\n文档正文：\n%s",
		requirement,
		fileHeader.Filename,
		documentText,
	)

	slideCount, _ := strconv.Atoi(strings.TrimSpace(c.PostForm("slide_count")))

	plan, err := h.llmService.PlanPPT(&services.PlanPPTRequest{
		Requirement: combinedRequirement,
		SlideCount:  slideCount,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, plan)
}

func (h *PPTHandler) Generate(c *gin.Context) {
	userID, _ := c.Get("user_id")
	req, err := bindGeneratePPTRequest(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	session, err := h.sessionRepo.FindByID(req.SessionID)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if session == nil || session.UserID != userID.(int) {
		response.Forbidden(c, "Invalid session")
		return
	}

	if strings.TrimSpace(req.DeckTitle) != "" {
		session.Title = req.DeckTitle
		_ = h.sessionRepo.Update(session)
	}

	if strings.TrimSpace(req.Size) == "" {
		req.Size = "1792x1024"
	}
	if strings.TrimSpace(req.Quality) == "" {
		req.Quality = "high"
	}

	go h.generateSlides(userID.(int), req)

	response.Success(c, gin.H{
		"deck_title": req.DeckTitle,
		"slides":     []interface{}{},
		"status":     "generating",
	})
}

func bindGeneratePPTRequest(c *gin.Context) (services.GeneratePPTImagesRequest, error) {
	var req services.GeneratePPTImagesRequest
	if !strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		if err := c.ShouldBindJSON(&req); err != nil {
			return req, err
		}
		return req, nil
	}

	req.SessionID = c.PostForm("session_id")
	req.GenerationMode = c.PostForm("generation_mode")
	req.DeckTitle = c.PostForm("deck_title")
	req.DeckGoal = c.PostForm("deck_goal")
	req.VisualDirection = c.PostForm("visual_direction")
	req.MasterStyleDescription = c.PostForm("master_style_description")
	req.Size = c.PostForm("size")
	req.Quality = c.PostForm("quality")
	req.Style = c.PostForm("style")

	if rawRules := c.PostForm("consistency_rules"); strings.TrimSpace(rawRules) != "" {
		if err := json.Unmarshal([]byte(rawRules), &req.ConsistencyRules); err != nil {
			return req, fmt.Errorf("invalid consistency_rules")
		}
	}

	if rawSlides := c.PostForm("slides"); strings.TrimSpace(rawSlides) != "" {
		if err := json.Unmarshal([]byte(rawSlides), &req.Slides); err != nil {
			return req, fmt.Errorf("invalid slides")
		}
	}

	form, err := c.MultipartForm()
	if err == nil && form != nil {
		fileHeaders := append([]*multipart.FileHeader{}, form.File["reference_image"]...)
		fileHeaders = append(fileHeaders, form.File["reference_images"]...)
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				return req, fmt.Errorf("failed to read reference image")
			}
			data, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				return req, fmt.Errorf("failed to read reference image")
			}
			req.ReferenceImages = append(req.ReferenceImages, services.ReferenceImage{
				Data:        data,
				Name:        fileHeader.Filename,
				ContentType: fileHeader.Header.Get("Content-Type"),
			})
		}
	}

	if strings.TrimSpace(req.SessionID) == "" || len(req.Slides) == 0 {
		return req, fmt.Errorf("missing session_id or slides")
	}

	return req, nil
}

func (h *PPTHandler) generateSlides(userID int, req services.GeneratePPTImagesRequest) {
	slides := append([]models.PPTSlidePlan(nil), req.Slides...)
	sort.Slice(slides, func(i, j int) bool {
		return slides[i].SlideNumber < slides[j].SlideNumber
	})

	var masterRecord *models.ImageRecord
	for index, slide := range slides {
		finalPrompt := buildSlideImagePrompt(&req, slides, index, &slide, masterRecord)
		generateReq := &services.GenerateImageRequest{
			Prompt:          finalPrompt,
			SessionID:       req.SessionID,
			Size:            req.Size,
			Quality:         req.Quality,
			Style:           req.Style,
			ReferenceImages: resolveUploadedReferenceImages(req.GenerationMode, req.ReferenceImages, len(slides), index),
		}

		referenceRecord := resolveReferenceRecord(req.GenerationMode, slides, index, masterRecord)
		if referenceRecord != nil && strings.TrimSpace(referenceRecord.LocalPath) != "" {
			referenceData, referenceName, referenceType, err := h.imageService.LoadReferenceImage(referenceRecord.LocalPath)
			if err == nil {
				generateReq.ReferenceImages = append(generateReq.ReferenceImages, services.ReferenceImage{
					Data:        referenceData,
					Name:        referenceName,
					ContentType: referenceType,
				})
			}
		}

		record, err := h.imageService.GenerateImage(userID, &services.GenerateImageRequest{
			Prompt:             generateReq.Prompt,
			SessionID:          generateReq.SessionID,
			Size:               generateReq.Size,
			Quality:            generateReq.Quality,
			Style:              generateReq.Style,
			ReferenceImageData: generateReq.ReferenceImageData,
			ReferenceImageName: generateReq.ReferenceImageName,
			ReferenceImageType: generateReq.ReferenceImageType,
			ReferenceImages:    generateReq.ReferenceImages,
		})
		if err != nil {
			log.Printf("PPT slide %d generation failed: %v", slide.SlideNumber, err)
			return
		}

		if shouldBecomeMaster(slides, index) {
			masterRecord = record
		}

		if index < len(slides)-1 {
			time.Sleep(10 * time.Second)
		}
	}
}

func resolveUploadedReferenceImages(mode string, referenceImages []services.ReferenceImage, slideCount int, index int) []services.ReferenceImage {
	if len(referenceImages) == 0 {
		return nil
	}

	switch mode {
	case "style_text":
		return []services.ReferenceImage{referenceImages[0]}
	case "image_content_style":
		if index < len(referenceImages) {
			return []services.ReferenceImage{referenceImages[index]}
		}
		return nil
	}

	if slideCount > 0 && len(referenceImages) >= slideCount && index < len(referenceImages) {
		return []services.ReferenceImage{referenceImages[index]}
	}

	return append([]services.ReferenceImage(nil), referenceImages...)
}

func extractDocumentText(filename string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	var text string
	var err error

	switch ext {
	case ".txt", ".md", ".markdown", ".csv":
		text = string(data)
	case ".docx":
		text, err = extractZipXMLText(data, "word/document.xml")
	case ".pptx":
		text, err = extractZipXMLText(data, "ppt/slides/")
	default:
		return "", fmt.Errorf("unsupported document type: %s. Please upload txt, md, csv, docx, or pptx", ext)
	}
	if err != nil {
		return "", err
	}

	text = normalizeDocumentText(text)
	if text == "" {
		return "", fmt.Errorf("document text is empty")
	}
	if len([]rune(text)) > 18000 {
		text = string([]rune(text)[:18000]) + "\n\n[文档较长，已截取前 18000 字用于规划]"
	}
	return text, nil
}

func extractZipXMLText(data []byte, prefix string) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to parse Office document")
	}

	files := append([]*zip.File(nil), reader.File...)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	var chunks []string
	for _, item := range files {
		if item.FileInfo().IsDir() || !strings.HasPrefix(item.Name, prefix) || !strings.HasSuffix(item.Name, ".xml") {
			continue
		}
		file, err := item.Open()
		if err != nil {
			continue
		}
		raw, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}
		cleaned := xmlToPlainText(string(raw))
		if strings.TrimSpace(cleaned) != "" {
			chunks = append(chunks, cleaned)
		}
	}

	if len(chunks) == 0 {
		return "", fmt.Errorf("no readable text found in Office document")
	}
	return strings.Join(chunks, "\n\n"), nil
}

func xmlToPlainText(raw string) string {
	raw = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(raw, " ")
	raw = html.UnescapeString(raw)
	return raw
}

func normalizeDocumentText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func buildSlideImagePrompt(
	req *services.GeneratePPTImagesRequest,
	slides []models.PPTSlidePlan,
	index int,
	slide *models.PPTSlidePlan,
	masterRecord *models.ImageRecord,
) string {
	if req.GenerationMode == "style_text" || req.GenerationMode == "image_content_style" || req.GenerationMode == "free_mix" {
		if directPrompt := strings.TrimSpace(slide.ImagePrompt); directPrompt != "" {
			return directPrompt
		}
		if directPrompt := strings.TrimSpace(slide.PageDescription); directPrompt != "" {
			return directPrompt
		}
		return strings.TrimSpace(slide.Objective)
	}

	return services.BuildCompetitionSlidePrompt(req, slides, index, slide)
}

func buildGenerationModeInstruction(mode string) string {
	switch mode {
	case "style_text":
		return "模式①：上传图片只作为整体视觉风格样板。请学习配色、标题位置、排版节奏、字体感觉、边框和装饰元素；不要复制参考图里的文字、图片内容或页面主题。页面真实内容只来自当前页用户文字，不要扩写新论点，不要把上一页或下一页内容带进来。"
	case "image_content_style":
		return "模式②：当前页上传图片是本页必须保留的核心内容参考，文字只作为全套统一设计风格描述。请重绘/美化当前页，不要混入其他页图片。"
	case "free_mix":
		return "模式③：允许综合上传图片和文字自由整理、补全和美化，但必须围绕当前页，不要串页。"
	default:
		return "默认PPT生成模式。"
	}
}

func resolveReferenceRecord(mode string, slides []models.PPTSlidePlan, index int, masterRecord *models.ImageRecord) *models.ImageRecord {
	if mode == "style_text" || mode == "image_content_style" {
		return nil
	}
	if len(slides) <= 2 {
		return nil
	}
	if index == 0 || index == len(slides)-1 {
		return nil
	}
	return masterRecord
}

func shouldBecomeMaster(slides []models.PPTSlidePlan, index int) bool {
	if len(slides) <= 2 {
		return false
	}
	return index == 1
}

func buildContinuityNote(
	slides []models.PPTSlidePlan,
	index int,
	slide *models.PPTSlidePlan,
	masterRecord *models.ImageRecord,
) string {
	if len(slides) <= 2 {
		return "This deck is very short. Keep a coherent style across all slides, but tailor the scene directly to this slide's message."
	}

	if index == 0 {
		return "This is the cover slide. Establish the overall visual identity for the deck, but keep it more iconic and open than the internal content slides."
	}

	if index == 1 {
		return "This is the first internal content slide and will become the master style reference for the rest of the middle slides. Establish the definitive visual system here: palette, lighting, rendering texture, composition grammar, diagram language, icon style, information density, and how title/content zones are reserved."
	}

	if index == len(slides)-1 {
		return "This is the closing slide. Keep it related to the deck's world, but allow it to feel cleaner, calmer, and more conclusive than the middle content slides. Do not use the internal master slide composition rigidly."
	}

	if masterRecord != nil {
		return fmt.Sprintf(
			"Use the attached master content slide reference image as the visual anchor for slide %d. If an uploaded reference image for this exact slide is attached, use it only as this slide's content/layout reference. Do not merge text, pictures, titles, or prompt content from other slide numbers. Match the palette, rendering texture, camera language, icon style, contrast, spacing rhythm, and composition grammar while creating a new scene for this slide's content only.",
			slide.SlideNumber,
		)
	}

	return fmt.Sprintf(
		"This is an internal content slide. Keep it highly consistent with the deck's established middle-slide visual system and compose a new scene for slide %d.",
		slide.SlideNumber,
	)
}
