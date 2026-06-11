package ticket

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TicketMeta represents the standard metadata structure in YAML Frontmatter.
type TicketMeta struct {
	ID          string                 `yaml:"id" json:"id"`
	Title       string                 `yaml:"title,omitempty" json:"title,omitempty"`
	Type        string                 `yaml:"type,omitempty" json:"type,omitempty"`
	Status      string                 `yaml:"status" json:"status"`
	Priority    string                 `yaml:"priority,omitempty" json:"priority,omitempty"`
	Owner       string                 `yaml:"owner,omitempty" json:"owner,omitempty"`
	CreatedAt   string                 `yaml:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt   string                 `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
	ResolvedAt  string                 `yaml:"resolved_at,omitempty" json:"resolved_at,omitempty"`
	Conclusion  string                 `yaml:"conclusion,omitempty" json:"conclusion,omitempty"`
	ExtraFields map[string]interface{} `yaml:",inline" json:"extra_fields,omitempty"` // Preserves all other dynamic properties
}

// ParseFile parses a markdown file containing YAML frontmatter and returns the metadata and body.
func ParseFile(filePath string) (*TicketMeta, string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	metaStr, bodyStr, err := splitFrontmatter(string(content))
	if err != nil {
		return nil, "", fmt.Errorf("invalid frontmatter in %s: %w", filePath, err)
	}

	var meta TicketMeta
	if len(metaStr) > 0 {
		if err := yaml.Unmarshal([]byte(metaStr), &meta); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal yaml: %w", err)
		}
	}

	return &meta, bodyStr, nil
}

// WriteFile writes the ticket metadata and markdown body back to the file.
func WriteFile(filePath string, meta *TicketMeta, body string) error {
	var yamlBuf bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&yamlBuf)
	yamlEncoder.SetIndent(2)

	if err := yamlEncoder.Encode(meta); err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}
	yamlEncoder.Close()

	// Construct final content
	var out strings.Builder
	out.WriteString("---\n")
	out.WriteString(yamlBuf.String())
	out.WriteString("---\n")
	
	// Normalize body spacing (ensure it starts clean but preserves structure)
	cleanBody := strings.TrimLeft(body, "\r\n")
	out.WriteString(cleanBody)

	// Write back to file atomically (write to temp file and rename)
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(out.String()), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file to target: %w", err)
	}

	return nil
}

// splitFrontmatter splits the markdown content into YAML frontmatter and markdown body.
func splitFrontmatter(content string) (string, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var frontmatter lines
	var body lines
	inFrontmatter := false
	delimiterCount := 0

	// We check if the file starts with the frontmatter delimiter
	isFirstLine := true

	for scanner.Scan() {
		line := scanner.Text()
		
		if isFirstLine {
			isFirstLine = false
			if line == "---" {
				inFrontmatter = true
				delimiterCount++
				continue
			} else {
				// No frontmatter found, whole file is body
				return "", content, nil
			}
		}

		if inFrontmatter {
			if line == "---" {
				inFrontmatter = false
				delimiterCount++
				continue
			}
			frontmatter.add(line)
		} else {
			body.add(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	if delimiterCount == 1 {
		return "", "", fmt.Errorf("frontmatter delimiter is not closed")
	}

	return frontmatter.join(), body.joinWithTrailing(), nil
}

type lines []string

func (l *lines) add(line string) {
	*l = append(*l, line)
}

func (l *lines) join() string {
	if len(*l) == 0 {
		return ""
	}
	return strings.Join(*l, "\n") + "\n"
}

func (l *lines) joinWithTrailing() string {
	if len(*l) == 0 {
		return ""
	}
	return strings.Join(*l, "\n")
}
