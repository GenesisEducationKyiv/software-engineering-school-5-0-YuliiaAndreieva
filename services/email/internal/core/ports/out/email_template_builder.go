//go:generate mockery --dir . --output ../../../../tests/mocks --outpkg mocks --filename email_template_builder_mock.go --structname EmailTemplateBuilder --name EmailTemplateBuilder
package out

import (
	"context"
	"email/internal/core/domain"
)

type EmailTemplateBuilder interface {
	BuildEmailTemplate(ctx context.Context, req domain.SendEmailRequest) (string, error)
}
