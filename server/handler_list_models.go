package server

import (
	"encoding/json"
	"net/http"

	"github.com/recally-io/polyllm/llms"
)

type listModelsResponse struct {
	Data   []llms.Model `json:"data"`
	Object string       `json:"object"`
}

func listModelsHandler(w http.ResponseWriter, r *http.Request) {

	models := make([]llms.Model, 0)

	for model := range modelProviderMappings {
		models = append(models, llms.Model{
			ID:     model,
			Object: "model",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(listModelsResponse{
		Data:   models,
		Object: "list",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

}
