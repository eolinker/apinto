package ai_prompt

type Config struct {
	Prompt    string     `json:"prompt"`
	Variables []Variable `json:"variables"`
}

type Variable struct {
	Key     string `json:"key"`
	Require bool   `json:"require"`
}
