package types

import "time"

// Project represents an indexed project
type Project struct {
	ID            int64                  `json:"id"`
	Path          string                 `json:"path"`
	Name          string                 `json:"name"`
	LanguageStats map[string]int         `json:"language_stats"` // language -> file count
	LastIndexed   time.Time              `json:"last_indexed"`
	CreatedAt     time.Time              `json:"created_at"`
	Config        map[string]interface{} `json:"config,omitempty"`
}

// ProjectConfig represents project configuration
type ProjectConfig struct {
	Exclude      []string          `json:"exclude"`
	Languages    []string          `json:"languages"`
	IndexOnSave  bool              `json:"index_on_save"`
	Plugins      []string          `json:"plugins"`
	CustomConfig map[string]string `json:"custom_config,omitempty"`
}
