package ticket

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TicketConfig represents the configuration stored in ticket.yaml
type TicketConfig struct {
	SubDirs     []string           `yaml:"sub_dirs" json:"sub_dirs"`
	ExtraFields []ExtraFieldConfig `yaml:"extra_fields" json:"extra_fields"`
}

// ExtraFieldConfig defines validation rules for a custom metadata field
type ExtraFieldConfig struct {
	Name     string `yaml:"name" json:"name"`
	Required bool   `yaml:"required" json:"required"`
}

// LoadConfig parses a ticket.yaml config in the given directory.
// If the file doesn't exist, it returns an empty configuration.
func LoadConfig(dir string) (*TicketConfig, error) {
	configPath := filepath.Join(dir, "ticket.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TicketConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read ticket config %s: %w", configPath, err)
	}

	var config TicketConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config %s: %w", configPath, err)
	}

	return &config, nil
}
