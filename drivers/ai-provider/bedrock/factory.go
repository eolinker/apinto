package bedrock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc"
	"io"

	"github.com/mitchellh/mapstructure"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/private/protocol/eventstream"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc/common/bean"
)

const (
	provider = "bedrock"
)

var (
	driverCreate        = eosc.BuildUntyped[ai_convert.ModelType, ai_convert.IConvertDriverCreateFunc[Config]]()
	accessConfigManager ai_convert.IModelAccessConfigManager
)

func init() {
	bean.Autowired(&accessConfigManager)
	ai_convert.RegisterConverterCreateFunc("bedrock", Create)
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}
	return ai_convert.NewConverter(provider, &conf, driverCreate.List())
}

// EventStreamToJSON 将 Amazon EventStream 格式的数据转换为 JSON 格式
func EventStreamToJSON(eventStreamData []byte) ([]StreamResponse, error) {
	// 创建一个结果数组
	var result []StreamResponse

	// 创建一个 EventStream 解码器
	decoder := eventstream.NewDecoder(bytes.NewReader(eventStreamData))

	// 循环读取所有事件
	for {

		// 读取下一个消息
		msg, err := decoder.Decode(nil)
		if err != nil {
			if err == io.EOF {
				break // 正常结束
			}

			// 处理 AWS 错误
			if awsErr, ok := err.(awserr.Error); ok {
				return nil, fmt.Errorf("AWS Error: %s - %s", awsErr.Code(), awsErr.Message())
			}

			return nil, fmt.Errorf("解析 EventStream 时出错: %v", err)
		}

		// 将消息转换为 map
		eventMap := make(map[string]interface{})

		// 处理消息头
		headers := make(map[string]interface{})
		for _, header := range msg.Headers {
			headers[header.Name] = header.Value
		}

		eventMap["headers"] = headers

		// 处理消息体
		if len(msg.Payload) > 0 {
			// 尝试将负载解析为 JSON
			var payload interface{}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				eventMap["payload"] = payload
			} else {
				// 如果不是有效的 JSON，则作为字符串处理
				eventMap["payload"] = string(msg.Payload)
			}
		}
		var streamResponse StreamResponse
		err = mapstructure.Decode(eventMap, &streamResponse)
		if err != nil {
			return nil, err
		}

		// 将事件添加到结果数组
		result = append(result, streamResponse)
	}
	return result, nil
}

// ExtractTextFromEventStream 从 EventStream 中提取文本内容
// 这个函数专门用于从 Bedrock 模型响应中提取生成的文本
func ExtractTextFromEventStream(eventStreamData []byte) (string, error) {
	var fullText string

	decoder := eventstream.NewDecoder(bytes.NewReader(eventStreamData))

	for {
		msg, err := decoder.Decode(nil)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		// 解析消息负载
		var response map[string]interface{}
		if err := json.Unmarshal(msg.Payload, &response); err != nil {
			continue // 跳过无法解析的消息
		}

		// 根据 Bedrock 的响应格式提取文本
		// 注意：具体的字段名可能需要根据使用的模型进行调整
		if completion, ok := response["completion"].(string); ok {
			fullText += completion
		} else if output, ok := response["output"].(map[string]interface{}); ok {
			if text, ok := output["text"].(string); ok {
				fullText += text
			}
		}
	}

	return fullText, nil
}
