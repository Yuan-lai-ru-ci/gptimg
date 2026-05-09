package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gptimg/internal/config"
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/internal/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLMService struct {
	repo *repository.LLMConfigRepository
	cfg  *config.Config
}

type PlanPPTRequest struct {
	Requirement    string `json:"requirement" binding:"required"`
	SlideCount     int    `json:"slide_count"`
	GenerationMode string `json:"generation_mode"`
}

type GeneratePPTImagesRequest struct {
	SessionID              string                `json:"session_id" binding:"required"`
	GenerationMode         string                `json:"generation_mode"`
	DeckTitle              string                `json:"deck_title"`
	DeckGoal               string                `json:"deck_goal"`
	VisualDirection        string                `json:"visual_direction"`
	MasterStyleDescription string                `json:"master_style_description"`
	ConsistencyRules       []string              `json:"consistency_rules"`
	Slides                 []models.PPTSlidePlan `json:"slides" binding:"required"`
	Size                   string                `json:"size"`
	Quality                string                `json:"quality"`
	Style                  string                `json:"style"`
	ReferenceImages        []ReferenceImage      `json:"-" form:"-"`
}

func NewLLMService(repo *repository.LLMConfigRepository, cfg *config.Config) *LLMService {
	return &LLMService{repo: repo, cfg: cfg}
}

func (s *LLMService) PlanPPT(req *PlanPPTRequest) (*models.PPTPlan, error) {
	prompt := buildPPTPlannerPrompt(req)
	raw, err := s.completeJSON(prompt)
	if err != nil {
		return nil, err
	}

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var plan models.PPTPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse PPT plan: %w", err)
	}
	if len(plan.Slides) == 0 {
		return nil, errors.New("planner returned empty slides")
	}
	normalizePPTPlan(&plan)
	if plan.GenerationMode == "" {
		plan.GenerationMode = req.GenerationMode
	}
	return &plan, nil
}

func normalizePPTPlan(plan *models.PPTPlan) {
	plan.DeckTitle = strings.TrimSpace(plan.DeckTitle)
	plan.DeckGoal = strings.TrimSpace(plan.DeckGoal)
	plan.VisualDirection = strings.TrimSpace(plan.VisualDirection)
	plan.MasterStyleDescription = strings.TrimSpace(plan.MasterStyleDescription)
	plan.GenerationMode = strings.TrimSpace(plan.GenerationMode)

	if plan.MasterStyleDescription == "" {
		plan.MasterStyleDescription = fmt.Sprintf(
			"%s Maintain these deck-wide rules: %s",
			plan.VisualDirection,
			strings.Join(plan.ConsistencyRules, "; "),
		)
	}

	for i := range plan.Slides {
		slide := &plan.Slides[i]
		slide.Title = strings.TrimSpace(slide.Title)
		slide.Objective = strings.TrimSpace(slide.Objective)
		slide.LayoutNotes = strings.TrimSpace(slide.LayoutNotes)
		slide.PageDescription = strings.TrimSpace(slide.PageDescription)
		slide.ImagePrompt = strings.TrimSpace(slide.ImagePrompt)
		slide.SpeakerNotes = strings.TrimSpace(slide.SpeakerNotes)

		if slide.PageDescription == "" {
			slide.PageDescription = fmt.Sprintf(
				"这一页是：内容页\n标题是：%s\n核心内容是：%s",
				slide.Title,
				slide.Objective,
			)
		}

		if slide.ImagePrompt == "" {
			slide.ImagePrompt = slide.PageDescription
		}
	}
}

func (s *LLMService) completeJSON(userPrompt string) (string, error) {
	configs, err := s.repo.FindActiveConfigs()
	if err != nil {
		return "", err
	}
	if len(configs) == 0 {
		return "", errors.New("no active LLM configuration found")
	}

	var lastErr error
	for _, item := range configs {
		apiKey, err := utils.DecryptAPIKey(item.APIKeyEncrypted, s.cfg.EncryptionKey)
		if err != nil {
			lastErr = err
			continue
		}
		content, err := callOpenAICompatibleChat(item.APIBaseURL, apiKey, item.Model, userPrompt)
		if err == nil {
			return content, nil
		}
		lastErr = fmt.Errorf("%s: %w", item.ConfigName, err)
	}
	if lastErr == nil {
		lastErr = errors.New("all LLM configs failed")
	}
	return "", lastErr
}

func buildPPTPlannerPrompt(req *PlanPPTRequest) string {
	slideHint := "let the AI choose an appropriate slide count between 4 and 8"
	if req.SlideCount > 0 {
		slideHint = fmt.Sprintf("create exactly %d slides", req.SlideCount)
	}
	modeInstruction := buildPPTModePlannerInstruction(req.GenerationMode)

	return fmt.Sprintf(`你是一位中文科创答辩PPT策划助手。

你的目标是给图片生成模型提供“尽量少但足够用”的内容结构，不要过度规定设计细节，不要替模型写死构图。

要求：
- 只输出严格 JSON，不要输出 markdown，不要解释。
- %s
- JSON schema 必须是：
{
  "deck_title": "string",
  "deck_goal": "string",
  "visual_direction": "string",
  "master_style_description": "string",
  "consistency_rules": ["string"],
  "slides": [
    {
      "slide_number": 1,
      "title": "string",
      "objective": "string",
      "layout_notes": "string",
      "page_description": "string",
      "image_prompt": "string",
      "speaker_notes": "string"
    }
  ]
}
- 当用户需求属于科创、项目答辩、平台方案时，优先按“封面 -> 背景/痛点 -> 方案/架构 -> 优势/场景 -> 结尾”组织。
- visual_direction、master_style_description、consistency_rules 都要尽量简短，不要堆砌设计黑话。
- 每一页最重要的是把“这一页是什么页、标题是什么、核心内容是什么”说清楚。
- page_description 尽量用这种弱约束格式：
  这一页是：[封面页/背景页/方案页/架构页/优势页/场景页/结尾页]
  标题是：[标题]
  核心内容是：[2到5个短点]
- image_prompt 也保持极简，只保留页类型、标题、核心内容和整体风格倾向，不要写死版式，不要写太多修饰词。
- 不要强行规定留白、卡片数量、左右布局、字体层级、装饰元素数量，让图片模型自由发挥。
- 文字内容控制在适合一页 PPT 的范围，不要长段落。
%s

用户需求：
%s`, slideHint, modeInstruction, req.Requirement)
}

func buildPPTModePlannerInstruction(mode string) string {
	switch mode {
	case "style_text":
		return `- 当前模式是“根据图片风格 + 文字内容生成PPT”：用户上传图片只作为整体视觉风格参考，真正页面内容必须来自用户文字。若用户按“第1页/第2页/Page 1/Page 2”分段，必须严格按这些分段拆页；不要把参考图里的文字或内容写入PPT计划。`
	case "image_content_style":
		return `- 当前模式是“根据图片内容 + 文字描述风格生成PPT”：每页核心内容来自对应上传图片，用户文字只描述整体设计风格。计划应保持每页独立，不要把其他页图片内容混入当前页。`
	case "free_mix":
		return `- 当前模式是“根据图片 + 文本直接补全或美化PPT”：允许根据图片和文字自由整理、扩写、补全和美化，但每页仍要内容清楚，不要串页。`
	default:
		return ""
	}
}

func callOpenAICompatibleChat(baseURL, apiKey, model, prompt string) (string, error) {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-chat"
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a precise JSON-only assistant."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.5,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}
