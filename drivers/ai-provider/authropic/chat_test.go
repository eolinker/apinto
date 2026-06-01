package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func getAPIKey() string {
	return os.Getenv("ANTHROPIC_API_KEY")
}

func getBaseUrl() string {
	baseUrl := os.Getenv("ANTHROPIC_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://api.anthropic.com/v1"
	}
	return baseUrl
}

func TestCalculateAnthropicStreamUsageWithMockData(t *testing.T) {
	mockSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_01XFDUDYJgAACzvn","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":25,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" there"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":12}}

event: message_stop
data: {"type":"message_stop"}`
	input, output := calculateAnthropicStreamUsage([]byte(mockSSE))

	fmt.Printf("===== Mock Anthropic Stream Usage =====\n")
	fmt.Printf("Input tokens:  %d\n", input)
	fmt.Printf("Output tokens: %d\n", output)
	fmt.Printf("Total tokens:  %d\n", input+output)

	if input != 25 {
		t.Errorf("expected input tokens = 25, got %d", input)
	}
	if output != 12 {
		t.Errorf("expected output tokens = 12, got %d", output)
	}

	contentBuilder := extractContentFromSSE([]byte(mockSSE))
	fmt.Printf("\n===== Mock Model Response Content =====\n")
	fmt.Printf("%s\n", contentBuilder)
	if contentBuilder != "Hello there!" {
		t.Errorf("expected content 'Hello there!', got '%s'", contentBuilder)
	}
}

func TestCalculateAnthropicStreamUsage(t *testing.T) {
	apiKey := getAPIKey()
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY environment variable is not set, skipping real API test")
	}

	reqBody := map[string]any{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 1024,
		"stream":     true,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": "Say hello in one sentence.",
			},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body error: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/messages", getBaseUrl()), bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("create request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body error: %v", err)
	}

	input, output := calculateAnthropicStreamUsage(respBody)
	fmt.Printf("===== Anthropic Stream Usage =====\n")
	fmt.Printf("Input tokens:  %d\n", input)
	fmt.Printf("Output tokens: %d\n", output)
	fmt.Printf("Total tokens:  %d\n", input+output)

	if input <= 0 {
		t.Errorf("expected input tokens > 0, got %d", input)
	}
	if output <= 0 {
		t.Errorf("expected output tokens > 0, got %d", output)
	}

	contentBuilder := extractContentFromSSE(respBody)
	fmt.Printf("\n===== Model Response Content =====\n")
	fmt.Printf("%s\n", contentBuilder)
}

func extractContentFromSSE(body []byte) string {
	contentBuilder := &strings.Builder{}
	scanner := bufio.NewScanner(bytes.NewReader(body))
	var eventType string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		switch eventType {
		case "content_block_delta":
			var delta struct {
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &delta); err == nil && delta.Delta.Text != "" {
				contentBuilder.WriteString(delta.Delta.Text)
			}
		case "message_start":
			var msg struct {
				Message struct {
					Usage struct {
						InputTokens int `json:"input_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(data), &msg); err == nil {
				fmt.Printf("Message start input_tokens: %d\n", msg.Message.Usage.InputTokens)
			}
		case "message_delta":
			var delta struct {
				Delta struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
				Usage struct {
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}
			if err := json.Unmarshal([]byte(data), &delta); err == nil {
				fmt.Printf("Message delta stop_reason: %s, output_tokens: %d\n", delta.Delta.StopReason, delta.Usage.OutputTokens)
			}
		}
		eventType = ""
	}
	return contentBuilder.String()
}
