package ai_provider

import ai_convert "github.com/eolinker/apinto/ai-convert"

var (
	values = []*ai_convert.ValueRule{
		{
			Value: "frequency_penalty",
			Type:  "float64",
		},
		{
			Value: "n",
			Type:  "int",
		},
		{
			Value: "logprobs",
			Type:  "bool",
		},
		{
			Value: "max_tokens",
			Type:  "int",
		},
		{
			Value: "top_p",
			Type:  "float64",
		},
		{
			Value: "presence_penalty",
			Type:  "float64",
		},
		{
			Value: "seed",
			Type:  "int",
		},
		{
			Value: "temperature",
			Type:  "float64",
		},
		{
			Value: "top_k",
			Type:  "int",
		},
		//{
		//	Value: "response_format",
		//	Type:  "string",
		//},
		{
			Value: "stream",
			Type:  "bool",
		},
	}
	providerMappings = map[string]string{
		"context_length_exceeded_behavior": "",
		"count_penalty":                    "",
		"disable_search":                   "",
		"do_sample":                        "",
		"enable_enhance":                   "",
		"enable_search":                    "",
		"frequency_penalt":                 "frequency_penalty",
		"frequency_penalty":                "frequency_penalty",
		"json_schema":                      "json_schema",
		"k":                                "top_k",
		"logprobs":                         "logprobs",
		"mask_sensitive_info":              "",
		"maxTokenCount":                    "max_tokens",
		"maxTokens":                        "max_tokens",
		"max_gen_len":                      "max_tokens",
		"max_new_tokens":                   "max_tokens",
		"max_output_tokens":                "max_tokens",
		"max_tokens":                       "max_tokens",
		"max_tokens_to_sample":             "max_tokens",
		"min_output_tokens":                "max_tokens",
		"p":                                "top_p",
		"plugin_web_search":                "",
		"preamble_override":                "",
		"presence_penalty":                 "presence_penalty",
		"prompt_truncation":                "",
		"random_seed":                      "seed",
		"reasoning_effort":                 "",
		"repetition_penalty":               "",
		"res_format":                       "response_format",
		"response_format":                  "response_format",
		"return_type":                      "response_format",
		"safe_prompt":                      "",
		"seed":                             "seed",
		"show_ref_label":                   "",
		"stream":                           "stream",
		"temperature":                      "temperature",
		"topP":                             "top_p",
		"top_k":                            "top_k",
		"top_logprobs":                     "top_logprobs",
		"top_p":                            "top_p",
		"web_search":                       "",
		"with_search_enhance":              "",
	}
	providerMapValue = make(map[string]*ai_convert.ValueRule)
)

func init() {
	tmpMap := map[string]*ai_convert.ValueRule{}
	for _, v := range values {
		tmpMap[v.Value] = v
	}
	for k, v := range providerMappings {
		if v == "" {
			continue
		}
		t, ok := tmpMap[v]
		if !ok {
			continue
		}
		providerMapValue[k] = t
	}
}
