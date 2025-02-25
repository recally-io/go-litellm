package llms

// RequestOption is a function that configures a ChatCompletionRequest.
type RequestOption func(*ChatCompletionRequest)

// WithModel specifies which model name to use.
func WithModel(model string) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.Model = model
	}
}

// WithMaxTokens specifies the max number of tokens to generate.
func WithMaxTokens(maxTokens int) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.MaxTokens = maxTokens
	}
}

// WithMaxCompletionTokens specifies the max number of completion tokens to generate.
func WithMaxCompletionTokens(maxCompletionTokens int) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.MaxCompletionTokens = maxCompletionTokens
	}
}

// WithTemperature specifies the model temperature, a hyperparameter that
// regulates the randomness, or creativity, of the AI's responses.
func WithTemperature(temperature float32) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.Temperature = temperature
	}
}

// WithTopP	will add an option to use top-p sampling.
func WithTopP(topP float32) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.TopP = topP
	}
}

// WithSeed will add an option to use deterministic sampling.
func WithSeed(seed int) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.Seed = &seed
	}
}

// WithN will add an option to set how many chat completion choices to generate for each input message.
func WithN(n int) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.N = n
	}
}

func WithStream(stream bool) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.Stream = stream
	}
}

func WithExtraHeaders(headers map[string]string) RequestOption {
	return func(o *ChatCompletionRequest) {
		o.ExtraHeaders = headers
	}
}
