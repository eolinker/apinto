package openAI

import (
	_ "embed"
	"testing"

	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"
)

func TestLoad(t *testing.T) {
	models, err := ai_provider.LoadModels(providerContent, providerDir)
	if err != nil {
		t.Fatal(err)
	}
	for key, model := range models {
		t.Logf("key:%s,type:%+v", key, model.ModelType)
		if model.ModelProperties != nil {
			t.Logf("mode:%s,context_size:%d", model.ModelProperties.Mode, model.ModelProperties.ContextSize)
		}
	}
}
