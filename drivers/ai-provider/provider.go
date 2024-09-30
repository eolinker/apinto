package ai_provider

import (
	"embed"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type ModelType string

const (
	ModelTypeLLM           ModelType = "llm"
	ModelTypeTextEmbedding ModelType = "text-embedding"
	ModelTypeSpeech2Text   ModelType = "speech2text"
	ModelTypeModeration    ModelType = "moderation"
	ModelTypeTTS           ModelType = "tts"
)

const (
	ModeChat     Mode = "chat"
	ModeComplete Mode = "complete"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

type Provider struct {
	Provider            string   `json:"provider" yaml:"provider"`
	SupportedModelTypes []string `json:"supported_model_types" yaml:"supported_model_types"`
}

type Model struct {
	Model           string     `json:"model" yaml:"model"`
	ModelType       ModelType  `json:"model_type" yaml:"model_type"`
	ModelProperties *ModelMode `json:"model_properties" yaml:"model_properties"`
}

type ModelMode struct {
	Mode        string `json:"mode" yaml:"mode"`
	ContextSize int    `json:"context_size" yaml:"context_size"`
}

func LoadModels(providerContent []byte, dirFs embed.FS) (map[string]*Model, error) {
	var provider Provider
	err := yaml.Unmarshal(providerContent, &provider)
	if err != nil {
		return nil, err
	}
	models := make(map[string]*Model)
	for _, modelType := range provider.SupportedModelTypes {
		dirFiles, err := dirFs.ReadDir(modelType)
		if err != nil {
			// 未找到模型目录
			continue
		}
		for _, dirFile := range dirFiles {
			if dirFile.IsDir() || !strings.HasSuffix(dirFile.Name(), ".yaml") {
				continue
			}
			modelContent, err := dirFs.ReadFile(modelType + "/" + dirFile.Name())
			if err != nil {
				return nil, err
			}
			var m Model
			err = yaml.Unmarshal(modelContent, &m)
			if err != nil {
				return nil, err
			}
			models[m.Model] = &m
		}

	}
	return models, nil
}
