package parser

import (
	"encoding/json"
	"fmt"

	"concurrent-downloader/models"
)

func Parse(data []byte) (models.Todo, error) {
	var todo models.Todo

	err := json.Unmarshal(data, &todo)
	if err != nil {
		return models.Todo{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return todo, nil
}
