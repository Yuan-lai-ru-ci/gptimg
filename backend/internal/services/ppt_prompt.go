package services

import (
	"fmt"
	"gptimg/internal/models"
	"strings"
)

func BuildCompetitionSlidePrompt(
	req *GeneratePPTImagesRequest,
	slides []models.PPTSlidePlan,
	index int,
	slide *models.PPTSlidePlan,
) string {
	pageType := classifyPageType(slides, index)
	masterStyle := buildMasterStyleBlock(req)
	layoutTemplate := getLayoutTemplate(pageType)
	contentBlock := buildContentBlock(slide, pageType)
	typographyBlock := buildTypographyBlock(pageType)
	constraints := buildConstraints()

	return fmt.Sprintf(`[Canvas] A single 16:9 presentation slide, 1792×1024 pixels. Chinese competition/defense style PPT page.

[Style] %s

[Layout] %s

[Content] %s

[Typography] %s

[Constraints] %s`,
		masterStyle,
		layoutTemplate,
		contentBlock,
		typographyBlock,
		constraints,
	)
}

func classifyPageType(slides []models.PPTSlidePlan, index int) string {
	if index == 0 {
		return "cover"
	}
	if index == len(slides)-1 {
		return "closing"
	}
	if index == 1 && len(slides) > 4 {
		return "toc"
	}
	return "content"
}

func buildMasterStyleBlock(req *GeneratePPTImagesRequest) string {
	if desc := strings.TrimSpace(req.MasterStyleDescription); desc != "" {
		return desc
	}

	direction := strings.TrimSpace(req.VisualDirection)
	if direction == "" {
		direction = "Modern technology competition style"
	}

	rules := strings.Join(req.ConsistencyRules, "; ")
	if rules == "" {
		rules = "consistent color palette across all slides; same font hierarchy; same decorative elements"
	}

	return fmt.Sprintf(
		"%s. Deck-wide rules: %s. "+
			"Use a dark gradient background (deep navy #1a2332 to teal #0d3b4f). "+
			"Title: bold white sans-serif at top-left with generous margin. "+
			"Body text: light gray (#e0e8f0) with clear hierarchy. "+
			"Accent color: electric cyan (#00d4ff) for highlights, borders, and markers. "+
			"Decorative: thin accent-colored horizontal rule below title, subtle geometric grid at 10%% opacity. "+
			"Clean, minimal, technology-forward. High information density but well-organized whitespace.",
		direction, rules,
	)
}

func getLayoutTemplate(pageType string) string {
	switch pageType {
	case "cover":
		return "Cover page layout: Large bold title centered vertically in upper 60%% of slide. " +
			"Subtitle or team name below title at 60%% vertical position. " +
			"Competition name and date at bottom. " +
			"Background: full-bleed gradient with subtle tech-themed decorative elements (circuit lines, geometric shapes at low opacity). " +
			"No content cards or bullet points on this page."

	case "toc":
		return "Table of contents layout: Page title 'CONTENTS' or '目录' at top-left. " +
			"3-6 numbered items listed vertically with clear spacing. " +
			"Each item: number in accent color + chapter title in white. " +
			"Right side: subtle decorative graphic or abstract shape. " +
			"Clean and scannable."

	case "closing":
		return "Closing page layout: Large '谢谢' or 'THANK YOU' centered. " +
			"Below: team name, contact info, or key takeaway in smaller text. " +
			"Minimal decorative elements. Calm, conclusive feel. " +
			"Same background system as other slides but simpler composition."

	default:
		return "Content page layout: Title at top-left with accent underline. " +
			"Main content area occupies 70%% of slide below title. " +
			"Content can include: bullet points, key metrics in highlight boxes, small diagrams, or comparison columns. " +
			"Page number at bottom-right. " +
			"Maintain consistent margins (80px left, 60px top, 60px right, 40px bottom)."
	}
}

func buildContentBlock(slide *models.PPTSlidePlan, pageType string) string {
	title := strings.TrimSpace(slide.Title)
	objective := strings.TrimSpace(slide.Objective)
	pageDesc := strings.TrimSpace(slide.PageDescription)

	if pageType == "cover" {
		return fmt.Sprintf(
			"Title text: \"%s\"\n"+
				"Subtitle text: \"%s\"",
			title, objective,
		)
	}

	if pageType == "closing" {
		return fmt.Sprintf(
			"Main text: \"谢谢观看\"\n"+
				"Secondary text: \"%s\"",
			objective,
		)
	}

	if pageDesc != "" {
		lines := strings.Split(pageDesc, "\n")
		var quotedLines []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "这一页是：") || strings.HasPrefix(line, "标题是：") || strings.HasPrefix(line, "核心内容是：") {
				continue
			}
			quotedLines = append(quotedLines, fmt.Sprintf("\"%s\"", line))
		}
		if len(quotedLines) > 0 {
			return fmt.Sprintf(
				"Title text: \"%s\"\nBody content (render each line as a bullet or key point):\n%s",
				title,
				strings.Join(quotedLines, "\n"),
			)
		}
	}

	return fmt.Sprintf(
		"Title text: \"%s\"\nBody text: \"%s\"",
		title, objective,
	)
}

func buildTypographyBlock(pageType string) string {
	switch pageType {
	case "cover":
		return "Title: 56pt bold white sans-serif (like Source Han Sans / Noto Sans CJK), centered. " +
			"Subtitle: 28pt regular, light gray (#c8d4e0), centered below title with 24px gap. " +
			"Bottom info: 18pt, muted gray (#8899aa), centered at bottom."

	case "toc":
		return "Page title: 36pt bold white, top-left. " +
			"List numbers: 48pt bold accent color (#00d4ff). " +
			"List items: 24pt medium white, aligned after numbers."

	case "closing":
		return "Main text: 64pt bold white, centered vertically. " +
			"Secondary: 20pt regular, light gray, centered below."

	default:
		return "Title: 36pt bold white, top-left aligned. " +
			"Body bullets: 22pt regular, light gray (#e0e8f0), left-aligned with 1.6 line spacing. " +
			"Highlight keywords: accent color (#00d4ff) or white bold. " +
			"Page number: 14pt muted gray, bottom-right."
	}
}

func buildConstraints() string {
	return "CRITICAL: Only render the exact Chinese and English text specified in quotation marks above. " +
		"No extra text, no additional words, no random lettering, no watermarks, no placeholder text, no lorem ipsum. " +
		"All Chinese characters must be spelled correctly and clearly readable at presentation viewing distance. " +
		"Do not add any text that is not explicitly quoted in the [Content] section. " +
		"Maintain pixel-perfect text rendering with correct stroke order for all CJK characters."
}
