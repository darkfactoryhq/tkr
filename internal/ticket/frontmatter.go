package ticket

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseFile(path string) (*Ticket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading ticket file: %w", err)
	}

	content := string(data)

	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing opening frontmatter delimiter")
	}

	rest := content[4:]
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return nil, fmt.Errorf("missing closing frontmatter delimiter")
	}

	frontmatter := rest[:end]
	body := strings.TrimLeft(rest[end+4:], "\n")

	var t Ticket
	if err := yaml.Unmarshal([]byte(frontmatter), &t); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	t.Body = body
	t.FilePath = path

	return &t, nil
}

func Marshal(t *Ticket) ([]byte, error) {
	var buf bytes.Buffer

	fm, err := yaml.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("marshaling frontmatter: %w", err)
	}

	buf.WriteString("---\n")
	buf.Write(fm)
	buf.WriteString("---\n")

	if t.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(t.Body)
		if !strings.HasSuffix(t.Body, "\n") {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}
