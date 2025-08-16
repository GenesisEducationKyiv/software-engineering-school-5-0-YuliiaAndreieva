package out

import (
	"context"
)

type EmailTemplateStrategy interface {
	BuildTemplate(ctx context.Context, data interface{}) (string, error)
	GetType() string
}
