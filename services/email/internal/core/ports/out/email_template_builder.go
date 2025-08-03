//go:generate mockery --dir . --output ../../../../tests/mocks --outpkg mocks --filename email_template_builder_mock.go --structname EmailTemplateBuilder
package out

import "context"

type EmailTemplateBuilder interface {
	BuildConfirmationEmail(ctx context.Context, email, city, confirmationLink string) (string, error)
	BuildWeatherUpdateEmail(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int) (string, error)
}
