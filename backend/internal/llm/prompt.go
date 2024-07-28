package llm

import (
	"bytes"
	"text/template"

	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultPromptTemplate is the default template for generating a prompt
	DefaultPromptTemplate = `
Please provide the answers of the following question based on the context provided. The contexts contains documents and the recent conversations.
Always provide answer in <answser></answer> tag.

Question:
------------
{{.Query}}
-----------

Contexts:
-----------
{{range .Documents}}
Title: {{.Title}}
Content: {{.Content}}
UUID: {{.Metadata.UUID}}
Source: {{.Metadata.Source}}
{{end}}
-----------
Recent conversations
-----------
{{range .ConversationHistories}}
Role: {{.Role}}
Message: {{.Message}}
{{end}}
-----------
`
)

// Documents is a struct that represents a document
type Documents struct {
	Title    string
	Content  string
	Metadata map[string]string
	URL      string
}

// PromptTemplateData is the data used to generate a prompt
type PromptTemplateData struct {
	Query                 string
	Documents             []Documents
	ConversationHistories []types.Conversation
}

// PromptTemplate is a struct that represents a prompt template
type PromptTemplate struct {
	Template string
}

// PromptGenerator is a struct that generates prompts
type PromptGenerator struct {
	logger   *logrus.Logger
	template PromptTemplate
}

// Prompt is a struct that represents a prompt
type Prompt struct {
	Text string
}

// NewPromptGenerator creates a new prompt generator with the given logger and template
func NewPromptGenerator(logger *logrus.Logger, template PromptTemplate) *PromptGenerator {
	if template.Template == "" {
		template.Template = DefaultPromptTemplate
	}
	return &PromptGenerator{
		logger:   logger,
		template: template,
	}
}

// GeneratePrompt generates a prompt with the given data
func (p *PromptGenerator) GeneratePrompt(data PromptTemplateData) (Prompt, error) {
	if len(data.Documents) == 0 {
		p.logger.Warn("No documents found in the context")
		return Prompt{Text: data.Query}, nil
	}

	tmpl, err := template.New("prompt").Parse(p.template.Template)
	if err != nil {
		p.logger.WithError(err).Error("Failed to parse template")
		return Prompt{}, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		p.logger.WithError(err).Error("Failed to execute template")
		return Prompt{}, err
	}

	return Prompt{
		Text: buf.String(),
	}, nil
}
