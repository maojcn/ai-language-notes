package ai

import (
	"ai-language-notes/internal/api/dto"
	"encoding/json"
	"fmt"
	"strings"
)

// parseJSONContent extracts structured content from LLM JSON response
func ParseJSONContent(content string) (*dto.ProcessedContent, error) {
	// Strip markdown code blocks if present
	if strings.HasPrefix(content, "```") {
		// Find the position of the first opening brace
		startPos := strings.Index(content, "{")
		if startPos != -1 {
			// Find the position of the last closing brace
			endPos := strings.LastIndex(content, "}")
			if endPos != -1 && endPos > startPos {
				content = content[startPos : endPos+1]
			}
		}
	}

	// Try parsing as a direct JSON object first
	var processedContent dto.ProcessedContent
	if err := json.Unmarshal([]byte(content), &processedContent); err == nil {
		// Successful direct parsing
		if processedContent.Content != "" && len(processedContent.Tags) > 0 {
			return &processedContent, nil
		}
	}

	// If direct parsing failed, try the flexible approach
	var rawContent map[string]interface{}
	if err := json.Unmarshal([]byte(content), &rawContent); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Process the content field
	var finalContent string
	if contentVal, exists := rawContent["content"]; exists {
		switch c := contentVal.(type) {
		case string:
			finalContent = c
		case map[string]interface{}:
			contentBytes, err := json.MarshalIndent(c, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to process content object: %w", err)
			}
			finalContent = string(contentBytes)
		default:
			return nil, fmt.Errorf("unexpected content format in LLM response")
		}
	} else {
		return nil, fmt.Errorf("content field missing in LLM response")
	}

	// Process tags
	var tags []string
	if tagsVal, exists := rawContent["tags"]; exists {
		if tagsSlice, ok := tagsVal.([]interface{}); ok {
			for _, tag := range tagsSlice {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}
	}

	return &dto.ProcessedContent{
		Content: finalContent,
		Tags:    tags,
	}, nil
}
