package llm

import (
	"fmt"
	"net/http"

	"github.com/shezw/ai-terminal/internal/config"
)

func NewLocalClient(cfg *config.ServerConfig) Client {
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/v1", cfg.Port)
	return &remoteClient{
		endpoint: endpoint,
		apiKey:   "",
		model:    "local",
		http:     &http.Client{},
	}
}
