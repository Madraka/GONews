package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"news/internal/models"

	"gorm.io/datatypes"
)

// EmbedDetector handles automatic URL detection and embed creation
type EmbedDetector struct {
	patterns map[string]*EmbedPattern
}

// EmbedPattern defines how to detect and handle different embed types
type EmbedPattern struct {
	Regex       *regexp.Regexp
	EmbedType   string
	ExtractData func(url string) (EmbedData, error)
}

// EmbedData contains extracted information from URLs
type EmbedData struct {
	Type        string                 `json:"type"`
	URL         string                 `json:"url"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
}

// EmbedSuggestion represents a suggested embed from detected URL
type EmbedSuggestion struct {
	URL         string                 `json:"url"`
	EmbedType   string                 `json:"embed_type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
	Preview     string                 `json:"preview"`
}

// NewEmbedDetector creates a new embed detector with predefined patterns
func NewEmbedDetector() *EmbedDetector {
	detector := &EmbedDetector{
		patterns: make(map[string]*EmbedPattern),
	}

	// YouTube patterns
	youtubeRegex := regexp.MustCompile(`(?i)(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/embed/|youtube\.com/shorts/)([a-zA-Z0-9_-]+)`)
	detector.patterns["youtube"] = &EmbedPattern{
		Regex:     youtubeRegex,
		EmbedType: "youtube",
		ExtractData: func(urlStr string) (EmbedData, error) {
			matches := youtubeRegex.FindStringSubmatch(urlStr)
			if len(matches) < 2 {
				return EmbedData{}, fmt.Errorf("invalid YouTube URL")
			}

			videoID := matches[1]
			return EmbedData{
				Type:        "youtube",
				URL:         urlStr,
				Title:       "YouTube Video",
				Description: "Embedded YouTube video",
				Settings: map[string]interface{}{
					"embed_url":     urlStr,
					"embed_type":    "youtube",
					"video_id":      videoID,
					"embed_width":   "100%",
					"embed_height":  "auto",
					"autoplay":      false,
					"show_controls": true,
					"muted":         false,
				},
			}, nil
		},
	}

	// Twitter patterns
	twitterRegex := regexp.MustCompile(`(?i)(?:twitter\.com|x\.com)/[a-zA-Z0-9_]+/status/([0-9]+)`)
	detector.patterns["twitter"] = &EmbedPattern{
		Regex:     twitterRegex,
		EmbedType: "twitter",
		ExtractData: func(urlStr string) (EmbedData, error) {
			matches := twitterRegex.FindStringSubmatch(urlStr)
			if len(matches) < 2 {
				return EmbedData{}, fmt.Errorf("invalid Twitter URL")
			}

			tweetID := matches[1]
			return EmbedData{
				Type:        "twitter",
				URL:         urlStr,
				Title:       "Twitter Post",
				Description: "Embedded Twitter/X post",
				Settings: map[string]interface{}{
					"embed_url":    urlStr,
					"embed_type":   "twitter",
					"tweet_id":     tweetID,
					"embed_width":  "100%",
					"embed_height": "auto",
					"show_thread":  false,
				},
			}, nil
		},
	}

	// Instagram patterns
	instagramRegex := regexp.MustCompile(`(?i)instagram\.com/p/([a-zA-Z0-9_-]+)`)
	detector.patterns["instagram"] = &EmbedPattern{
		Regex:     instagramRegex,
		EmbedType: "instagram",
		ExtractData: func(urlStr string) (EmbedData, error) {
			matches := instagramRegex.FindStringSubmatch(urlStr)
			if len(matches) < 2 {
				return EmbedData{}, fmt.Errorf("invalid Instagram URL")
			}

			postID := matches[1]
			return EmbedData{
				Type:        "instagram",
				URL:         urlStr,
				Title:       "Instagram Post",
				Description: "Embedded Instagram post",
				Settings: map[string]interface{}{
					"embed_url":    urlStr,
					"embed_type":   "instagram",
					"post_id":      postID,
					"embed_width":  "100%",
					"embed_height": "auto",
				},
			}, nil
		},
	}

	// TikTok patterns
	tiktokRegex := regexp.MustCompile(`(?i)tiktok\.com/@[^/]+/video/([0-9]+)`)
	detector.patterns["tiktok"] = &EmbedPattern{
		Regex:     tiktokRegex,
		EmbedType: "tiktok",
		ExtractData: func(urlStr string) (EmbedData, error) {
			matches := tiktokRegex.FindStringSubmatch(urlStr)
			if len(matches) < 2 {
				return EmbedData{}, fmt.Errorf("invalid TikTok URL")
			}

			videoID := matches[1]
			return EmbedData{
				Type:        "tiktok",
				URL:         urlStr,
				Title:       "TikTok Video",
				Description: "Embedded TikTok video",
				Settings: map[string]interface{}{
					"embed_url":    urlStr,
					"embed_type":   "tiktok",
					"video_id":     videoID,
					"embed_width":  "100%",
					"embed_height": "auto",
				},
			}, nil
		},
	}

	// LinkedIn patterns
	linkedinRegex := regexp.MustCompile(`(?i)linkedin\.com/posts/[a-zA-Z0-9_-]+-([0-9]+)-`)
	detector.patterns["linkedin"] = &EmbedPattern{
		Regex:     linkedinRegex,
		EmbedType: "linkedin",
		ExtractData: func(urlStr string) (EmbedData, error) {
			return EmbedData{
				Type:        "linkedin",
				URL:         urlStr,
				Title:       "LinkedIn Post",
				Description: "Embedded LinkedIn post",
				Settings: map[string]interface{}{
					"embed_url":    urlStr,
					"embed_type":   "linkedin",
					"embed_width":  "100%",
					"embed_height": "auto",
				},
			}, nil
		},
	}

	return detector
}

// DetectEmbeddableURLs scans text content for embeddable URLs
func (ed *EmbedDetector) DetectEmbeddableURLs(content string) []EmbedSuggestion {
	var suggestions []EmbedSuggestion

	// Simple URL regex to find potential URLs
	urlRegex := regexp.MustCompile(`https?://[^\s<>"]+`)
	urls := urlRegex.FindAllString(content, -1)

	for _, urlStr := range urls {
		// Clean up URL (remove trailing punctuation)
		urlStr = strings.TrimRight(urlStr, ".,!?;)")

		suggestion := ed.AnalyzeURL(urlStr)
		if suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	return suggestions
}

// AnalyzeURL analyzes a single URL and returns embed suggestion if applicable
func (ed *EmbedDetector) AnalyzeURL(urlStr string) *EmbedSuggestion {
	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil
	}

	// Check against all patterns
	for _, pattern := range ed.patterns {
		if pattern.Regex.MatchString(urlStr) {
			embedData, err := pattern.ExtractData(urlStr)
			if err != nil {
				continue
			}

			return &EmbedSuggestion{
				URL:         urlStr,
				EmbedType:   embedData.Type,
				Title:       embedData.Title,
				Description: embedData.Description,
				Settings:    embedData.Settings,
				Preview:     ed.generatePreview(embedData),
			}
		}
	}

	return nil
}

// CreateEmbedBlock creates a content block from embed suggestion
func (ed *EmbedDetector) CreateEmbedBlock(suggestion EmbedSuggestion) (*models.ArticleContentBlock, error) {
	settingsJSON, err := json.Marshal(suggestion.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize embed settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		BlockType: "embed",
		Content:   suggestion.Title,
		Settings:  datatypes.JSON(settingsJSON),
		IsVisible: true,
	}

	return block, nil
}

// generatePreview creates a text preview of the embed
func (ed *EmbedDetector) generatePreview(data EmbedData) string {
	switch data.Type {
	case "youtube":
		return fmt.Sprintf("🎥 YouTube: %s", data.URL)
	case "twitter":
		return fmt.Sprintf("🐦 Twitter: %s", data.URL)
	case "instagram":
		return fmt.Sprintf("📸 Instagram: %s", data.URL)
	case "tiktok":
		return fmt.Sprintf("🎵 TikTok: %s", data.URL)
	case "linkedin":
		return fmt.Sprintf("💼 LinkedIn: %s", data.URL)
	default:
		return fmt.Sprintf("🔗 Embed: %s", data.URL)
	}
}

// Global embed detector instance
var embedDetector *EmbedDetector

// GetEmbedDetector returns the global embed detector instance
func GetEmbedDetector() *EmbedDetector {
	if embedDetector == nil {
		embedDetector = NewEmbedDetector()
	}
	return embedDetector
}

// AutoDetectEmbeds analyzes content and returns suggested embeds
func AutoDetectEmbeds(content string) []EmbedSuggestion {
	detector := GetEmbedDetector()
	return detector.DetectEmbeddableURLs(content)
}

// CreateEmbedFromURL creates an embed block from a URL
func CreateEmbedFromURL(urlStr string) (*models.ArticleContentBlock, error) {
	detector := GetEmbedDetector()
	suggestion := detector.AnalyzeURL(urlStr)

	if suggestion == nil {
		return nil, fmt.Errorf("URL is not embeddable: %s", urlStr)
	}

	return detector.CreateEmbedBlock(*suggestion)
}
