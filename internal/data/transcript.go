package data

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

func ParseTranscript(filePath string) []Turn {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var turns []Turn
	toolNames := map[string]string{}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil || msg.Role == "" {
			continue
		}

		// Try string content
		var strContent string
		if err := json.Unmarshal(msg.Content, &strContent); err == nil {
			if strings.TrimSpace(strContent) != "" {
				turns = append(turns, Turn{Kind: TurnText, Role: msg.Role, Text: strContent})
			}
			continue
		}

		// Try array of content blocks
		var blocks []json.RawMessage
		if err := json.Unmarshal(msg.Content, &blocks); err != nil {
			continue
		}

		var pendingText string
		for _, rawBlock := range blocks {
			var blockType struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(rawBlock, &blockType); err != nil {
				continue
			}

			switch blockType.Type {
			case "text":
				var b struct {
					Text string `json:"text"`
				}
				if json.Unmarshal(rawBlock, &b) == nil && strings.TrimSpace(b.Text) != "" {
					if pendingText != "" {
						pendingText += "\n"
					}
					pendingText += b.Text
				}

			case "tool_use":
				if pendingText != "" {
					turns = append(turns, Turn{Kind: TurnText, Role: msg.Role, Text: pendingText})
					pendingText = ""
				}
				var b struct {
					ID    string          `json:"id"`
					Name  string          `json:"name"`
					Input json.RawMessage `json:"input"`
				}
				if json.Unmarshal(rawBlock, &b) == nil {
					toolNames[b.ID] = b.Name
					inputStr, _ := json.MarshalIndent(mustUnmarshalAny(b.Input), "", "  ")
					turns = append(turns, Turn{
						Kind:     TurnToolCall,
						ID:       b.ID,
						ToolName: b.Name,
						Input:    string(inputStr),
					})
				}

			case "tool_result":
				if pendingText != "" {
					turns = append(turns, Turn{Kind: TurnText, Role: msg.Role, Text: pendingText})
					pendingText = ""
				}
				var b struct {
					ToolUseID string          `json:"tool_use_id"`
					Content   json.RawMessage `json:"content"`
				}
				if json.Unmarshal(rawBlock, &b) == nil {
					content := extractResultContent(b.Content)
					turns = append(turns, Turn{
						Kind:      TurnToolResult,
						ToolUseID: b.ToolUseID,
						ToolName:  toolNames[b.ToolUseID],
						Content:   content,
					})
				}
			}
		}
		if pendingText != "" {
			turns = append(turns, Turn{Kind: TurnText, Role: msg.Role, Text: pendingText})
		}
	}
	return turns
}

func mustUnmarshalAny(raw json.RawMessage) any {
	var v any
	json.Unmarshal(raw, &v)
	return v
}

func extractResultContent(raw json.RawMessage) string {
	// Try string
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	// Try array of text blocks
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) == nil {
		var parts []string
		for _, b := range blocks {
			if b.Type == "text" {
				parts = append(parts, b.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return string(raw)
}
