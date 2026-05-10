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
	slideHint := "choose 5-8 slides based on content complexity"
	if req.SlideCount > 0 {
		slideHint = fmt.Sprintf("create exactly %d slides", req.SlideCount)
	}
	modeInstruction := buildPPTModePlannerInstruction(req.GenerationMode)

	return fmt.Sprintf(`You are a competition-grade PPT content planner. Your job is to produce STRUCTURED DATA that will be fed into an image generation prompt template. You do NOT write image prompts directly.

CRITICAL RULES:
1. Output ONLY valid JSON. No markdown, no explanation, no commentary.
2. %s
3. The JSON must follow this exact schema:
{
  "deck_title": "string - the main title of the presentation",
  "deck_goal": "string - one sentence describing the presentation's purpose",
  "visual_direction": "string - 15-25 words describing the overall visual mood (e.g. 'dark tech, cyan accents, geometric, minimal, professional competition defense')",
  "master_style_description": "string - 80-120 words, VERY SPECIFIC visual system description including: exact background colors/gradients, accent color hex codes, title font style and position, body text color and size, decorative elements, spacing rhythm. This will be repeated verbatim on every slide to ensure consistency.",
  "consistency_rules": ["array of 3-5 short rules that must hold across all slides"],
  "slides": [
    {
      "slide_number": 1,
      "title": "string - the exact title text to render on this slide (Chinese)",
      "objective": "string - 1-2 sentences of what this slide communicates",
      "layout_notes": "string - brief layout hint: cover/toc/content-bullets/content-diagram/content-comparison/closing",
      "page_description": "string - the EXACT body text content for this slide. Use bullet points with • prefix. Keep to 3-5 bullets max, each under 20 Chinese characters. This text will be rendered verbatim on the slide image.",
      "image_prompt": "string - leave empty, will be auto-generated by template engine",
      "speaker_notes": "string - what the presenter would say (not rendered on slide)"
    }
  ]
}

CONTENT GUIDELINES FOR COMPETITION PPT:
- Slide 1: Cover page (title + subtitle/team + competition name)
- Slide 2: Background/Problem (why this matters, pain points)
- Middle slides: Solution/Architecture/Implementation/Advantages (core content)
- Second-to-last: Results/Demo/Comparison
- Last slide: Summary + Thank you
- Each slide's page_description must contain the ACTUAL TEXT to display, not descriptions of what to show
- Write in Chinese. Keep text concise - this is a PPT, not a document
- Bullet points should be information-dense but short (under 20 chars each)
- Title should be 4-10 Chinese characters

MASTER_STYLE_DESCRIPTION EXAMPLE (write something similar but tailored to the topic):
"Dark navy (#1a2332) to deep teal (#0d3b4f) 135-degree diagonal gradient background. Subtle hexagonal grid pattern at 8%% opacity. Title: 36pt bold white Source Han Sans at top-left, 80px from edges, with 2px cyan (#00d4ff) underline extending 40%% of slide width. Body: 22pt regular light gray (#e0e8f0) with 1.5 line height. Accent: electric cyan (#00d4ff) for bullet markers, highlight borders, and key metrics. Cards: rounded-corner containers with 1px cyan border at 25%% opacity, dark fill (#0f2a3a). Page number: 14pt gray bottom-right. Decorative: thin diagonal lines in top-right corner at 5%% opacity."
%s

USER REQUIREMENT:
%s`, slideHint, modeInstruction, req.Requirement)
}

func buildPPTModePlannerInstruction(mode string) string {
	switch mode {
	case "style_text":
		return `- 当前模式是"根据图片风格 + 文字内容生成PPT"：用户上传图片只作为整体视觉风格参考，真正页面内容必须来自用户文字。若用户按"第1页/第2页/Page 1/Page 2"分段，必须严格按这些分段拆页；不要把参考图里的文字或内容写入PPT计划。`
	case "image_content_style":
		return `- 当前模式是"根据图片内容 + 文字描述风格生成PPT"：每页核心内容来自对应上传图片，用户文字只描述整体设计风格。计划应保持每页独立，不要把其他页图片内容混入当前页。`
	case "free_mix":
		return `- 当前模式是"根据图片 + 文本直接补全或美化PPT"：允许根据图片和文字自由整理、扩写、补全和美化，但每页仍要内容清楚，不要串页。`
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
