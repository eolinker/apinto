package google

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	context_label2 "github.com/eolinker/apinto/common/context-label"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sashabaranov/go-openai"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewOpenAIChat(provider, c.APIKey, c.BaseUrl, mt, 10*time.Minute)
	})
}

// thoughtSignatureCache stores Gemini thought signatures keyed by function
// name + canonical arguments. Some OpenAI-compatible clients drop the signature
// on round-trip (neither extra_content nor the _ts_ ID survive), which triggers
// a 400 from Gemini 3 on multi-turn function calling. The gateway caches the
// signature from the response and re-injects it when the same function call is
// sent back, so we don't depend on the client echoing it.
var thoughtSignatureCache *lru.Cache

func init() {
	thoughtSignatureCache, _ = lru.New(8192)
}

// thoughtSignatureKey builds a stable cache key from a function name and its
// arguments. Arguments are serialized canonically (sorted keys) so the request
// side (parsed from the client's raw arguments string) and the response side
// (parsed from Gemini's args map) produce the same key.
func thoughtSignatureKey(name string, args map[string]interface{}) string {
	canonical := canonicalJSON(args)
	sum := sha256.Sum256([]byte(name + "|" + canonical))
	return hex.EncodeToString(sum[:])
}

// canonicalJSON serializes a value with map keys sorted recursively.
func canonicalJSON(v interface{}) string {
	var b strings.Builder
	writeCanonical(&b, v)
	return b.String()
}

func writeCanonical(b *strings.Builder, v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			kb, _ := json.Marshal(k)
			b.Write(kb)
			b.WriteByte(':')
			writeCanonical(b, val[k])
		}
		b.WriteByte('}')
	case []interface{}:
		b.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				b.WriteByte(',')
			}
			writeCanonical(b, item)
		}
		b.WriteByte(']')
	default:
		vb, _ := json.Marshal(val)
		b.Write(vb)
	}
}

// cacheThoughtSignature stores a signature for a function call.
func cacheThoughtSignature(name string, args map[string]interface{}, signature string) {
	if signature == "" || thoughtSignatureCache == nil {
		return
	}
	thoughtSignatureCache.Add(thoughtSignatureKey(name, args), signature)
}

// lookupCachedThoughtSignature retrieves a cached signature for a function call.
func lookupCachedThoughtSignature(name string, args map[string]interface{}) string {
	if thoughtSignatureCache == nil {
		return ""
	}
	if v, ok := thoughtSignatureCache.Get(thoughtSignatureKey(name, args)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// labelStreamRemain holds the incomplete tail of the previous stream chunk. The
// framework feeds raw TCP-sized byte slices (~4KB) to the handler with no line
// or SSE-event boundary guarantee, so a single "data: {json}\n\n" event may be
// split across calls. We buffer the incomplete tail here and prepend it to the
// next chunk to avoid dropping (and thus losing) partial function-call events.
const labelStreamRemain = "google_stream_remain"

func NewOpenAIChat(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &OpenAIChat{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
	}
	orgBase := generativeAIBase
	orgPath := generativeAIPath
	if baseUrl == "" {
		baseUrl = orgBase
	}

	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, baseUrl, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	if strings.TrimSuffix(u.Path, "/") == "" {
		c.path = "/v1beta" + orgPath
	} else {
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), orgPath)
	}
	return c, nil
}

type OpenAIChat struct {
	apikey         string
	provider       string
	path           string
	checkErr       ai_convert.CheckError
	errorCallback  func(ctx http_service.IHttpContext, body []byte)
	balanceHandler eocontext.BalanceHandler
	modelType      ai_convert.ModelType
}

func (o *OpenAIChat) Provider() string {
	return o.provider
}

func (o *OpenAIChat) ModelType() ai_convert.ModelType {
	return o.modelType
}

// Gemini API Request and Response structures
type GeminiRequest struct {
	Contents          []GeminiContent         `json:"contents"`
	SystemInstruction *GeminiContent          `json:"systemInstruction,omitempty"`
	Tools             []GeminiTool            `json:"tools,omitempty"`
	ToolConfig        *GeminiToolConfig       `json:"toolConfig,omitempty"`
	GenerationConfig  *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"` // "user" or "model"
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
	InlineData       *GeminiInlineData       `json:"inlineData,omitempty"`
	FileData         *GeminiFileData         `json:"fileData,omitempty"`
	ThoughtSignature string                  `json:"thoughtSignature,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiFileData struct {
	MimeType string `json:"mimeType,omitempty"`
	FileURI  string `json:"fileUri"`
}

type GeminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

type GeminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type GeminiToolConfig struct {
	FunctionCallingConfig *GeminiFunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

type GeminiFunctionCallingConfig struct {
	Mode                 string   `json:"mode,omitempty"` // "AUTO", "ANY", "NONE"
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

type GeminiGenerationConfig struct {
	Temperature      *float32               `json:"temperature,omitempty"`
	TopP             *float32               `json:"topP,omitempty"`
	TopK             *int                   `json:"topK,omitempty"`
	MaxOutputTokens  *int                   `json:"maxOutputTokens,omitempty"`
	CandidateCount   *int                   `json:"candidateCount,omitempty"`
	StopSequences    []string               `json:"stopSequences,omitempty"`
	ResponseMimeType string                 `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]interface{} `json:"responseSchema,omitempty"`
	FrequencyPenalty *float32               `json:"frequencyPenalty,omitempty"`
	PresencePenalty  *float32               `json:"presencePenalty,omitempty"`
	Seed             *int                   `json:"seed,omitempty"`
}

type GeminiResponse struct {
	Candidates    []GeminiCandidate    `json:"candidates"`
	UsageMetadata *GeminiUsageMetadata `json:"usageMetadata,omitempty"`
}

type GeminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
	Index        int           `json:"index"`
}

type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// convertSchemaToGemini recursively changes lowercase parameter types (like "string", "object") to Gemini uppercase (like "STRING", "OBJECT"),
// normalizes multi-types / nullables (like ["number", "null"]), and removes JSON Schema fields not supported by Gemini.
func convertSchemaToGemini(schema map[string]interface{}) {
	if schema == nil {
		return
	}

	// 1. Remove unsupported JSON Schema keywords for Gemini
	unsupportedKeys := []string{
		"$schema", "$id", "$comment", "$defs", "definitions",
		"title",
		"additionalProperties", "patternProperties", "propertyNames",
		"dependencies", "dependentRequired", "dependentSchemas",
		"minProperties", "maxProperties", "unevaluatedProperties",
		"exclusiveMinimum", "exclusiveMaximum", "minimum", "maximum", "multipleOf",
		"minLength", "maxLength", "pattern",
		"uniqueItems", "minContains", "maxContains", "contains", "unevaluatedItems",
		"const", "not",
	}
	for _, key := range unsupportedKeys {
		delete(schema, key)
	}

	// Convert oneOf to anyOf since Gemini only supports anyOf
	if oneOf, ok := schema["oneOf"]; ok {
		if _, hasAnyOf := schema["anyOf"]; !hasAnyOf {
			schema["anyOf"] = oneOf
		}
		delete(schema, "oneOf")
	}

	// 2. Handle type field (can be string or array of strings, e.g. ["number", "null"])
	switch t := schema["type"].(type) {
	case string:
		if strings.EqualFold(t, "null") {
			schema["nullable"] = true
			delete(schema, "type")
		} else {
			schema["type"] = strings.ToUpper(t)
		}
	case []interface{}:
		hasNull := false
		var nonNullTypes []string
		for _, item := range t {
			if str, ok := item.(string); ok {
				if strings.EqualFold(str, "null") {
					hasNull = true
				} else {
					nonNullTypes = append(nonNullTypes, str)
				}
			}
		}
		if hasNull {
			schema["nullable"] = true
		}
		if len(nonNullTypes) == 1 {
			schema["type"] = strings.ToUpper(nonNullTypes[0])
		} else if len(nonNullTypes) > 1 {
			delete(schema, "type")
			var anyOfList []interface{}
			for _, tp := range nonNullTypes {
				anyOfList = append(anyOfList, map[string]interface{}{
					"type": strings.ToUpper(tp),
				})
			}
			if existingAnyOf, ok := schema["anyOf"].([]interface{}); ok {
				schema["anyOf"] = append(existingAnyOf, anyOfList...)
			} else {
				schema["anyOf"] = anyOfList
			}
		} else {
			delete(schema, "type")
		}
	}

	// 3. Recurse into properties
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for _, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				convertSchemaToGemini(propMap)
			}
		}
	}

	// 4. Recurse into items
	if items, ok := schema["items"].(map[string]interface{}); ok {
		convertSchemaToGemini(items)
	} else if itemsArr, ok := schema["items"].([]interface{}); ok {
		for _, item := range itemsArr {
			if itemMap, ok := item.(map[string]interface{}); ok {
				convertSchemaToGemini(itemMap)
			}
		}
	}

	// 5. Recurse into anyOf / allOf and clean null branches
	for _, key := range []string{"anyOf", "allOf"} {
		if arr, ok := schema[key].([]interface{}); ok {
			var newArr []interface{}
			for _, elem := range arr {
				if elemMap, ok := elem.(map[string]interface{}); ok {
					convertSchemaToGemini(elemMap)
					// If the branch is just a null type, hoist nullable=true and filter it out
					t, hasType := elemMap["type"].(string)
					isPureNull := (hasType && strings.EqualFold(t, "null")) || (!hasType && len(elemMap) == 0) || (elemMap["nullable"] == true && len(elemMap) == 1)
					if isPureNull {
						schema["nullable"] = true
						continue
					}
					newArr = append(newArr, elemMap)
				} else {
					newArr = append(newArr, elem)
				}
			}
			if len(newArr) == 0 {
				delete(schema, key)
			} else {
				schema[key] = newArr
			}
		}
	}
}

// sanitizeFunctionResponse recursively walks a function-response object and removes
// any key whose name starts with "$" (e.g. "$ref", "$schema"). Gemini treats these
// as JSON Schema references inside functionResponse.response and tries to resolve
// them against parts carrying a matching display_name. When no such part exists
// (the common case for tool results that happen to contain "$ref"), the API
// rejects the request with a 400. Stripping these keys (or converting the whole
// sub-object to its JSON string when it only contains $-keys) keeps the response
// shape intact for the model while avoiding the reference-resolution path.
func sanitizeFunctionResponse(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// Check if every key in this map starts with "$". If so, this entire
		// object is a schema fragment (like {"$ref": "#/foo"}) and the safest
		// representation is its JSON string, which Gemini will treat as plain
		// text rather than trying to resolve references.
		allDollar := len(val) > 0
		hasDollar := false
		for k := range val {
			if !strings.HasPrefix(k, "$") {
				allDollar = false
			} else {
				hasDollar = true
			}
		}
		if allDollar {
			b, _ := json.Marshal(val)
			return string(b)
		}
		result := make(map[string]interface{}, len(val))
		for k, sub := range val {
			if strings.HasPrefix(k, "$") {
				// Drop $-prefixed keys entirely; they are schema metadata that
				// Gemini cannot resolve inside a function response.
				continue
			}
			result[k] = sanitizeFunctionResponse(sub)
		}
		// If sanitization left us with nothing (e.g. input was {"$ref": "..."}),
		// fall back to the original JSON string so the model still sees the data.
		if len(result) == 0 && hasDollar {
			b, _ := json.Marshal(val)
			return string(b)
		}
		return result
	case []interface{}:
		for i, item := range val {
			val[i] = sanitizeFunctionResponse(item)
		}
		return val
	default:
		return v
	}
}

// extractExtraContentSignatures parses the raw request body and extracts thought
// signatures carried by the client via tool_calls[].extra_content.google.thought_signature.
// It returns a map: message index -> (tool_call index -> signature).
func extractExtraContentSignatures(body []byte) map[int]map[int]string {
	var raw struct {
		Messages []struct {
			ToolCalls []struct {
				ExtraContent *struct {
					Google *struct {
						ThoughtSignature string `json:"thought_signature"`
					} `json:"google"`
				} `json:"extra_content"`
			} `json:"tool_calls"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}
	result := make(map[int]map[int]string)
	for msgIdx, msg := range raw.Messages {
		for tcIdx, tc := range msg.ToolCalls {
			if tc.ExtraContent != nil && tc.ExtraContent.Google != nil && tc.ExtraContent.Google.ThoughtSignature != "" {
				if result[msgIdx] == nil {
					result[msgIdx] = make(map[int]string)
				}
				result[msgIdx][tcIdx] = tc.ExtraContent.Google.ThoughtSignature
			}
		}
	}
	return result
}

// lookupExtraSignature returns the signature for the given message/tool_call position, if any.
func lookupExtraSignature(sigs map[int]map[int]string, msgIdx, tcIdx int) string {
	if sigs == nil {
		return ""
	}
	if m, ok := sigs[msgIdx]; ok {
		return m[tcIdx]
	}
	return ""
}

// mimeTypeFromURL infers the image MIME type from the URL extension.
// Returns empty string when unknown, letting Gemini infer it.
func mimeTypeFromURL(u string) string {
	lower := strings.ToLower(u)
	if idx := strings.IndexAny(lower, "?#"); idx != -1 {
		lower = lower[:idx]
	}
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	default:
		return ""
	}
}

func (o *OpenAIChat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	context_label2.SetBillingMode(ctx, context_label2.BillingModeImmediate)
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	if o.apikey != "" {
		httpContext.Proxy().Header().SetHeader("x-goog-api-key", o.apikey)
		httpContext.Proxy().Header().DelHeader("authorization")
	}
	model := ai_convert.GetAIModel(ctx)
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	chatRequest := eosc.NewBase[openai.ChatCompletionRequest](extender)
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}

	// Extract thought signatures carried by the client via
	// tool_calls[].extra_content.google.thought_signature. This is the standard
	// round-trip format for Gemini's OpenAI-compatible layer and survives clients
	// that rewrite tool_call IDs. Indexed by message position -> tool_call position.
	extraSignatures := extractExtraContentSignatures(body)

	// Transform request to Gemini format
	geminiReq := &GeminiRequest{}

	// 1. Convert messages
	var systemParts []GeminiPart
	for msgIdx, msg := range chatRequest.Config.Messages {
		if msg.Role == "system" || msg.Role == "developer" {
			systemParts = append(systemParts, GeminiPart{Text: msg.Content})
			continue
		}

		content := GeminiContent{}
		switch msg.Role {
		case "user":
			content.Role = "user"
			if len(msg.MultiContent) > 0 {
				for _, part := range msg.MultiContent {
					if part.Type == openai.ChatMessagePartTypeText {
						content.Parts = append(content.Parts, GeminiPart{Text: part.Text})
					} else if part.Type == openai.ChatMessagePartTypeImageURL && part.ImageURL != nil {
						if strings.HasPrefix(part.ImageURL.URL, "data:") {
							// Inline base64 image
							commaIdx := strings.Index(part.ImageURL.URL, ",")
							if commaIdx != -1 {
								meta := part.ImageURL.URL[:commaIdx]
								data := part.ImageURL.URL[commaIdx+1:]
								// meta format: data:<mimeType>;base64
								mimeType := "image/jpeg"
								if semiIdx := strings.Index(meta, ";"); semiIdx > len("data:") {
									mimeType = meta[len("data:"):semiIdx]
								} else if len(meta) > len("data:") {
									mimeType = meta[len("data:"):]
								}
								content.Parts = append(content.Parts, GeminiPart{
									InlineData: &GeminiInlineData{
										MimeType: mimeType,
										Data:     data,
									},
								})
							}
						} else if part.ImageURL.URL != "" {
							// Remote URL: pass through as fileData reference
							content.Parts = append(content.Parts, GeminiPart{
								FileData: &GeminiFileData{
									MimeType: mimeTypeFromURL(part.ImageURL.URL),
									FileURI:  part.ImageURL.URL,
								},
							})
						}
					}
				}
			} else {
				content.Parts = append(content.Parts, GeminiPart{Text: msg.Content})
			}

		case "assistant":
			content.Role = "model"
			if msg.Content != "" {
				content.Parts = append(content.Parts, GeminiPart{Text: msg.Content})
			}
			for tcIdx, toolCall := range msg.ToolCalls {
				var args map[string]interface{}
				if toolCall.Function.Arguments != "" {
					_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				}
				// Prefer the signature carried via extra_content (survives ID rewrites);
				// fall back to the signature encoded in the tool_call ID; finally fall
				// back to the gateway-side cache keyed by function name + args, for
				// clients that drop the signature entirely.
				ts := lookupExtraSignature(extraSignatures, msgIdx, tcIdx)
				if ts == "" && strings.Contains(toolCall.ID, "_ts_") {
					parts := strings.SplitN(toolCall.ID, "_ts_", 2)
					if len(parts) == 2 {
						ts = parts[1]
					}
				}
				if ts == "" {
					ts = lookupCachedThoughtSignature(toolCall.Function.Name, args)
				}
				content.Parts = append(content.Parts, GeminiPart{
					FunctionCall: &GeminiFunctionCall{
						Name: toolCall.Function.Name,
						Args: args,
					},
					ThoughtSignature: ts,
				})
			}

		case "tool":
			content.Role = "user"
			var responseObj map[string]interface{}
			err := json.Unmarshal([]byte(msg.Content), &responseObj)
			if err != nil {
				responseObj = map[string]interface{}{"result": msg.Content}
			}

			// Sanitize $-prefixed keys (e.g. "$ref") from the response. Gemini
			// interprets these as JSON Schema references inside functionResponse
			// and rejects the request with a 400 when the reference target is
			// not found among the parts. See sanitizeFunctionResponse for details.
			sanitized := make(map[string]interface{}, len(responseObj))
			for k, v := range responseObj {
				if strings.HasPrefix(k, "$") {
					continue
				}
				sanitized[k] = sanitizeFunctionResponse(v)
			}
			if len(sanitized) == 0 && len(responseObj) > 0 {
				// All keys were $-prefixed; preserve the raw content as a string
				// so the model still sees the data.
				sanitized["result"] = msg.Content
			}
			responseObj = sanitized

			funcName := msg.Name
			if funcName == "" && msg.ToolCallID != "" {
				// Search backward to find the tool call name by ToolCallID
				for i := len(chatRequest.Config.Messages) - 1; i >= 0; i-- {
					prevMsg := chatRequest.Config.Messages[i]
					if prevMsg.Role == "assistant" {
						for _, tc := range prevMsg.ToolCalls {
							if tc.ID == msg.ToolCallID {
								funcName = tc.Function.Name
								break
							}
						}
					}
					if funcName != "" {
						break
					}
				}
			}

			// Fallback to avoid empty name rejection from Gemini API
			if funcName == "" {
				if msg.ToolCallID != "" {
					funcName = msg.ToolCallID
				} else {
					funcName = "unknown_function"
				}
			}

			content.Parts = append(content.Parts, GeminiPart{
				FunctionResponse: &GeminiFunctionResponse{
					Name:     funcName,
					Response: responseObj,
				},
			})
		}

		if len(content.Parts) > 0 {
			geminiReq.Contents = append(geminiReq.Contents, content)
		}
	}

	if len(systemParts) > 0 {
		geminiReq.SystemInstruction = &GeminiContent{
			Role:  "user",
			Parts: systemParts,
		}
	}

	// 2. Convert tools
	if len(chatRequest.Config.Tools) > 0 {
		var declarations []GeminiFunctionDeclaration
		for _, tool := range chatRequest.Config.Tools {
			if tool.Type == openai.ToolTypeFunction && tool.Function != nil {
				decl := GeminiFunctionDeclaration{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
				}
				if tool.Function.Parameters != nil {
					if paramBytes, err := json.Marshal(tool.Function.Parameters); err == nil {
						var paramMap map[string]interface{}
						if err := json.Unmarshal(paramBytes, &paramMap); err == nil {
							convertSchemaToGemini(paramMap)
							decl.Parameters = paramMap
						}
					}
				}
				declarations = append(declarations, decl)
			}
		}
		if len(declarations) > 0 {
			geminiReq.Tools = []GeminiTool{{FunctionDeclarations: declarations}}
		}
	}

	// 3. Convert tool choice
	if chatRequest.Config.ToolChoice != nil {
		config := &GeminiToolConfig{
			FunctionCallingConfig: &GeminiFunctionCallingConfig{},
		}
		// Try string first
		var toolChoiceStr string
		if bytes, err := json.Marshal(chatRequest.Config.ToolChoice); err == nil {
			if err := json.Unmarshal(bytes, &toolChoiceStr); err == nil {
				switch toolChoiceStr {
				case "none":
					config.FunctionCallingConfig.Mode = "NONE"
				case "auto":
					config.FunctionCallingConfig.Mode = "AUTO"
				case "required":
					config.FunctionCallingConfig.Mode = "ANY"
				}
			} else {
				// Object type tool choice
				var toolChoiceObj struct {
					Type     string `json:"type"`
					Function struct {
						Name string `json:"name"`
					} `json:"function"`
				}
				if err := json.Unmarshal(bytes, &toolChoiceObj); err == nil {
					if toolChoiceObj.Function.Name != "" {
						config.FunctionCallingConfig.Mode = "ANY"
						config.FunctionCallingConfig.AllowedFunctionNames = []string{toolChoiceObj.Function.Name}
					}
				}
			}
		}
		if config.FunctionCallingConfig.Mode != "" {
			geminiReq.ToolConfig = config
		}
	}

	// 4. Convert generation parameters
	genConfig := &GeminiGenerationConfig{}
	if chatRequest.Config.Temperature != 0 {
		temp := chatRequest.Config.Temperature
		genConfig.Temperature = &temp
	}
	if chatRequest.Config.TopP != 0 {
		topp := chatRequest.Config.TopP
		genConfig.TopP = &topp
	}
	if chatRequest.Config.MaxTokens > 0 {
		genConfig.MaxOutputTokens = &chatRequest.Config.MaxTokens
	}
	if len(chatRequest.Config.Stop) > 0 {
		genConfig.StopSequences = chatRequest.Config.Stop
	}
	if chatRequest.Config.N > 1 {
		genConfig.CandidateCount = &chatRequest.Config.N
	}
	if chatRequest.Config.FrequencyPenalty != 0 {
		fp := chatRequest.Config.FrequencyPenalty
		genConfig.FrequencyPenalty = &fp
	}
	if chatRequest.Config.PresencePenalty != 0 {
		pp := chatRequest.Config.PresencePenalty
		genConfig.PresencePenalty = &pp
	}
	if chatRequest.Config.Seed != nil {
		genConfig.Seed = chatRequest.Config.Seed
	}
	if chatRequest.Config.ResponseFormat != nil {
		switch strings.ToLower(string(chatRequest.Config.ResponseFormat.Type)) {
		case "json_object":
			genConfig.ResponseMimeType = "application/json"
		case "json_schema":
			genConfig.ResponseMimeType = "application/json"
			if js := chatRequest.Config.ResponseFormat.JSONSchema; js != nil && js.Schema != nil {
				if schemaBytes, err := js.Schema.MarshalJSON(); err == nil {
					var schemaMap map[string]interface{}
					if err := json.Unmarshal(schemaBytes, &schemaMap); err == nil {
						convertSchemaToGemini(schemaMap)
						genConfig.ResponseSchema = schemaMap
					}
				}
			}
		}
	}
	geminiReq.GenerationConfig = genConfig

	newBody, err := json.Marshal(geminiReq)
	if err != nil {
		return fmt.Errorf("marshal gemini request error: %v", err)
	}
	httpContext.Proxy().Body().SetRaw("application/json", newBody)

	// Set URL and Stream handling
	path := fmt.Sprintf("%s:generateContent", model)
	if chatRequest.Config.Stream {
		path = fmt.Sprintf("%s/%s:%s", o.path, model, "streamGenerateContent")
		httpContext.Proxy().URI().SetQuery("alt", "sse")
		httpContext.Proxy().AppendStreamBodyHandle(o.streamHandler)
		context_label2.SetModelCompletionStreamTag(ctx)
	} else {
		path = fmt.Sprintf("%s/%s:%s", o.path, model, "generateContent")
		context_label2.SetDisableStream(ctx, true)
		context_label2.SetModelCompletionTag(ctx)
	}

	httpContext.Proxy().URI().SetPath(path)

	if o.balanceHandler != nil {
		ctx.SetBalance(o.balanceHandler)
	}

	return nil
}

func convertOpenAIFormat(ctx http_service.IHttpContext, body []byte) ([]byte, error) {
	var geminiResp GeminiResponse
	err := json.Unmarshal(body, &geminiResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal gemini response error: %v", err)
	}

	// Map to OpenAI Response
	openaiResp := openai.ChatCompletionResponse{
		ID:      "chatcmpl-" + ctx.RequestId(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   ai_convert.GetAIModel(ctx),
	}

	if geminiResp.UsageMetadata != nil {
		outputToken := geminiResp.UsageMetadata.CandidatesTokenCount + geminiResp.UsageMetadata.ThoughtsTokenCount
		openaiResp.Usage = openai.Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: outputToken,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		}
		ai_convert.SetAIModelInputToken(ctx, geminiResp.UsageMetadata.PromptTokenCount)
		ai_convert.SetAIModelOutputToken(ctx, outputToken)
		ai_convert.SetAIModelTotalToken(ctx, geminiResp.UsageMetadata.TotalTokenCount)
	}

	toolSignatures := make(map[string]string)
	for _, candidate := range geminiResp.Candidates {
		choice := openai.ChatCompletionChoice{
			Index: candidate.Index,
			Message: openai.ChatCompletionMessage{
				Role: "assistant",
			},
		}

		switch candidate.FinishReason {
		case "STOP":
			choice.FinishReason = openai.FinishReasonStop
		case "MAX_TOKENS":
			choice.FinishReason = openai.FinishReasonLength
		case "SAFETY", "RECITATION":
			choice.FinishReason = openai.FinishReasonContentFilter
		default:
			choice.FinishReason = openai.FinishReasonStop
		}

		var toolCalls []openai.ToolCall
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				choice.Message.Content += part.Text
			}
			if part.FunctionCall != nil {
				argsBytes, _ := json.Marshal(part.FunctionCall.Args)
				toolCallID := "call_" + strconv.FormatInt(time.Now().UnixNano(), 10)
				if part.ThoughtSignature != "" {
					// Keep the signature encoded in the ID as a fallback, and
					// also surface it via extra_content below.
					toolCallID = fmt.Sprintf("call_%d_ts_%s", time.Now().UnixNano(), part.ThoughtSignature)
					toolSignatures[toolCallID] = part.ThoughtSignature
					// Cache so we can re-inject even if the client drops it.
					cacheThoughtSignature(part.FunctionCall.Name, part.FunctionCall.Args, part.ThoughtSignature)
				}
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   toolCallID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: string(argsBytes),
					},
				})
			}
		}

		if len(toolCalls) > 0 {
			choice.Message.ToolCalls = toolCalls
			choice.FinishReason = openai.FinishReasonToolCalls
		}

		openaiResp.Choices = append(openaiResp.Choices, choice)
	}

	newBody, err := json.Marshal(openaiResp)
	if err != nil {
		return nil, err
	}
	newBody = injectToolCallSignatures(newBody, toolSignatures)
	return newBody, nil
}

func (o *OpenAIChat) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	if httpContext.Response().StatusCode() != 200 {
		o.convertErrorResponse(httpContext)
		return nil
	}
	body := httpContext.Response().GetBody()
	newBody, err := convertOpenAIFormat(httpContext, body)
	if err != nil {
		log.Errorf("[dynamic-billing] failed to convert Gemini response to OpenAI format: %v", err)
		o.convertErrorResponse(httpContext)
		return err

	}
	httpContext.Response().SetBody(newBody)
	return nil
}

// injectToolCallSignatures rewrites an already-marshalled OpenAI chat response
// (completion or chunk) to attach thought signatures onto tool_calls via the
// extra_content.google.thought_signature field, which is the format Gemini's
// OpenAI-compatible layer expects to be echoed back in the next turn.
// signatures maps tool_call ID -> thought signature.
func injectToolCallSignatures(body []byte, signatures map[string]string) []byte {
	if len(signatures) == 0 {
		return body
	}
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return body
	}
	choices, ok := root["choices"].([]interface{})
	if !ok {
		return body
	}
	for _, ch := range choices {
		choiceMap, ok := ch.(map[string]interface{})
		if !ok {
			continue
		}
		// non-stream uses "message", stream uses "delta"
		for _, key := range []string{"message", "delta"} {
			msg, ok := choiceMap[key].(map[string]interface{})
			if !ok {
				continue
			}
			toolCalls, ok := msg["tool_calls"].([]interface{})
			if !ok {
				continue
			}
			for _, tc := range toolCalls {
				tcMap, ok := tc.(map[string]interface{})
				if !ok {
					continue
				}
				id, _ := tcMap["id"].(string)
				sig, exist := signatures[id]
				if !exist || sig == "" {
					continue
				}
				tcMap["extra_content"] = map[string]interface{}{
					"google": map[string]interface{}{
						"thought_signature": sig,
					},
				}
			}
		}
	}
	newBody, err := json.Marshal(root)
	if err != nil {
		return body
	}
	return newBody
}

// convertErrorResponse converts a Gemini error body into the OpenAI error format
// so that OpenAI-compatible clients can parse the error consistently.
func (o *OpenAIChat) convertErrorResponse(httpContext http_service.IHttpContext) {
	body := httpContext.Response().GetBody()
	var geminiErr struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &geminiErr); err != nil || geminiErr.Error.Message == "" {
		// Not a recognizable Gemini error, leave the body untouched.
		return
	}

	openaiErr := map[string]interface{}{
		"error": map[string]interface{}{
			"message": geminiErr.Error.Message,
			"type":    geminiErr.Error.Status,
			"code":    geminiErr.Error.Code,
		},
	}
	newBody, err := json.Marshal(openaiErr)
	if err != nil {
		return
	}
	httpContext.Response().SetBody(newBody)
}

func (o *OpenAIChat) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	var sseBuffer bytes.Buffer
	requestID := "chatcmpl-" + ctx.RequestId()
	model := ai_convert.GetAIModel(ctx)

	// Prepend any incomplete tail buffered from the previous chunk, then split
	// off a new tail so we only parse complete lines. The framework hands us raw
	// TCP-sized slices with no event boundary, so a data line may be cut in half.
	data := ctx.GetLabel(labelStreamRemain) + string(p)
	lastNL := strings.LastIndexByte(data, '\n')
	if lastNL < 0 {
		// No complete line yet; buffer everything and emit nothing.
		ctx.SetLabel(labelStreamRemain, data)
		return []byte{}, nil
	}
	// Keep the bytes after the last newline as the new remainder.
	ctx.SetLabel(labelStreamRemain, data[lastNL+1:])
	complete := data[:lastNL+1]

	scanner := bufio.NewScanner(strings.NewReader(complete))
	// Raise the line limit well above the default 64KB for large tool-call args.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimPrefix(line, "data:")
		dataStr = strings.TrimSpace(dataStr)
		if dataStr == "" {
			continue
		}

		var geminiResp GeminiResponse
		err := json.Unmarshal([]byte(dataStr), &geminiResp)
		if err != nil {
			log.Errorf("unmarshal gemini stream chunk error: %v, data: %s", err, dataStr)
			continue
		}

		if geminiResp.UsageMetadata != nil {
			ai_convert.SetAIModelInputToken(ctx, geminiResp.UsageMetadata.PromptTokenCount)
			ai_convert.SetAIModelOutputToken(ctx, geminiResp.UsageMetadata.CandidatesTokenCount+geminiResp.UsageMetadata.ThoughtsTokenCount)
			ai_convert.SetAIModelTotalToken(ctx, geminiResp.UsageMetadata.TotalTokenCount)
		}

		for _, candidate := range geminiResp.Candidates {
			streamResp := openai.ChatCompletionStreamResponse{
				ID:      requestID,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
			}

			// isFinal marks the terminating chunk from Gemini (finishReason set).
			isFinal := candidate.FinishReason != ""

			var finishReason openai.FinishReason
			switch candidate.FinishReason {
			case "":
				// Not a terminating chunk.
			case "STOP":
				finishReason = openai.FinishReasonStop
			case "MAX_TOKENS":
				finishReason = openai.FinishReasonLength
			case "SAFETY", "RECITATION":
				finishReason = openai.FinishReasonContentFilter
			default:
				// Unmapped terminating reason (e.g. OTHER, MALFORMED_FUNCTION_CALL):
				// fall back to stop so the stream still ends properly.
				finishReason = openai.FinishReasonStop
			}

			var delta openai.ChatCompletionStreamChoiceDelta
			delta.Role = "assistant"

			var toolCalls []openai.ToolCall
			toolSignatures := make(map[string]string)
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					delta.Content = part.Text
				}
				if part.FunctionCall != nil {
					argsBytes, _ := json.Marshal(part.FunctionCall.Args)
					toolCallIdx := len(toolCalls)
					toolCallID := "call_" + strconv.FormatInt(time.Now().UnixNano(), 10)
					if part.ThoughtSignature != "" {
						toolCallID = fmt.Sprintf("call_%d_ts_%s", time.Now().UnixNano(), part.ThoughtSignature)
						toolSignatures[toolCallID] = part.ThoughtSignature
						cacheThoughtSignature(part.FunctionCall.Name, part.FunctionCall.Args, part.ThoughtSignature)
					}
					toolCalls = append(toolCalls, openai.ToolCall{
						Index: &toolCallIdx,
						ID:    toolCallID,
						Type:  openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      part.FunctionCall.Name,
							Arguments: string(argsBytes),
						},
					})
				}
			}

			if len(toolCalls) > 0 {
				delta.ToolCalls = toolCalls
				if isFinal {
					finishReason = openai.FinishReasonToolCalls
				}
			}

			// Attach usage on the final chunk so token stats are always
			// forwarded, regardless of stream_options.include_usage. Only the
			// terminating chunk carries it to avoid repeating usage per chunk.
			if isFinal && geminiResp.UsageMetadata != nil {
				streamResp.Usage = &openai.Usage{
					PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
					CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount + geminiResp.UsageMetadata.ThoughtsTokenCount,
					TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
				}
			}

			choice := openai.ChatCompletionStreamChoice{
				Index:        candidate.Index,
				Delta:        delta,
				FinishReason: finishReason,
			}
			streamResp.Choices = []openai.ChatCompletionStreamChoice{choice}

			content, _ := json.Marshal(streamResp)
			content = injectToolCallSignatures(content, toolSignatures)
			sseBuffer.WriteString(fmt.Sprintf("data: %s\n\n", string(content)))

			if isFinal {
				sseBuffer.WriteString("data: [DONE]\n\n")
			}
		}
	}

	return sseBuffer.Bytes(), nil
}
