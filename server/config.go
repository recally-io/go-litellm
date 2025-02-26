package server

import (
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/recally-io/polyllm/llms"
)

type ProviderModel struct {
	Name  string `env:"NAME"`
	ID    string `env:"ID"`
	Alias string `env:"ALIAS"`
}

type Provider struct {
	BaseURL string `env:"BASE_URL"`
	APIKey  string `env:"API_KEY"`
	Prefix  string `env:"PREFIX"`
	// Model ids, in env it should be set as a comma separated string: "model1,model2"
	Models []string `env:"MODELS"`
	// Model aliases, in env it should be set as a key value value: "alias1=model1,alias2=model2"
	ModelAlias map[string]string `env:"MODEL_ALIAS"`

	llm llms.LLM
}

func (p *Provider) GetRealModel(model string) string {
	if realModel, ok := p.ModelAlias[model]; ok {
		return realModel
	}
	if p.Prefix != "" {
		parts := strings.Split(model, "/")
		return strings.Join(parts[1:], "/")
	}
	return model
}

type config struct {
	PORT   int    `env:"PORT" envdefault:"8088"`
	APIKey string `env:"API_KEY"`

	OpenAI     Provider `envPrefix:"OPENAI_"`
	DeepSeek   Provider `envPrefix:"DEEPSEEK_"`
	Qwen       Provider `envPrefix:"QWEN_"`
	Gemini     Provider `envPrefix:"GEMINI_"`
	OpenRouter Provider `envPrefix:"OPENROUTER_"`
	Volcengine Provider `envPrefix:"VOLCENGINE_"`
	Groq       Provider `envPrefix:"GROQ_"`
	Xai        Provider `envPrefix:"XAI_"`
}

var Settings = &config{}

func init() {
	// Load env to settings
	if err := env.ParseWithOptions(Settings, env.Options{}); err != nil {
		slog.Error("failed to load settings", "err", err)
		return
	}
	slog.Info("settings loaded", "settings", Settings)
}
