package strategies

import (
	"bytes"
	"context"
	"email/internal/adapter/email/templates"
	"email/internal/core/domain"
	"email/internal/core/ports/out"
	"encoding/json"
	"fmt"
	"html/template"
)

type ConfirmationEmailStrategy struct {
	logger out.Logger
}

func NewConfirmationEmailStrategy(logger out.Logger) out.EmailTemplateStrategy {
	return &ConfirmationEmailStrategy{
		logger: logger,
	}
}

func (s *ConfirmationEmailStrategy) BuildTemplate(ctx context.Context, data interface{}) (string, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected map[string]interface{}, got %T", data)
	}

	jsonData, err := json.Marshal(dataMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	var confirmationData domain.ConfirmationEmailData
	if err := json.Unmarshal(jsonData, &confirmationData); err != nil {
		return "", fmt.Errorf("failed to unmarshal confirmation data: %w", err)
	}

	tmpl, err := template.New("confirmation").Parse(templates.ConfirmationEmailTemplate)
	if err != nil {
		return "", err
	}

	templateData := templates.ConfirmationEmailData{
		City:             confirmationData.City,
		ConfirmationLink: confirmationData.ConfirmationLink,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *ConfirmationEmailStrategy) GetType() string {
	return string(domain.EmailTypeConfirmation)
}
