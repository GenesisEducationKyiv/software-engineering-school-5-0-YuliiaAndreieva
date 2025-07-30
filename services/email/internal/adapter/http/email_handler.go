package http

import (
	"net/http"
	"strings"

	"email/internal/adapter/dto"
	"email/internal/adapter/mappings"
	"email/internal/core/domain"
	"email/internal/core/ports/in"
	"email/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	sendEmailUseCase in.SendEmailUseCase
	logger           out.Logger
	validator        in.EmailValidator
	mapper           in.EmailMapper
}

func NewEmailHandler(sendEmailUseCase in.SendEmailUseCase, logger out.Logger) *EmailHandler {
	return &EmailHandler{
		sendEmailUseCase: sendEmailUseCase,
		logger:           logger,
		validator:        NewEmailValidator(),
		mapper:           mappings.NewEmailMapper(),
	}
}

func (h *EmailHandler) SendConfirmationEmail(c *gin.Context) {
	req, err := h.bindConfirmationEmailRequest(c)
	if err != nil {
		h.handleBindingError(c, err, "confirmation email")
		return
	}

	if err := h.validateConfirmationEmailRequest(c, req); err != nil {
		return
	}

	domainReq := h.mapper.MapConfirmationEmailRequest(req)
	result, err := h.sendEmailUseCase.SendConfirmationEmail(c.Request.Context(), domainReq)
	if err != nil {
		h.handleUseCaseError(c, err, "confirmation email")
		return
	}

	h.handleSuccess(c, req.To, "confirmation email", result)
}

func (h *EmailHandler) SendWeatherUpdateEmail(c *gin.Context) {
	req, err := h.bindWeatherUpdateEmailRequest(c)
	if err != nil {
		h.handleBindingError(c, err, "weather update email")
		return
	}

	if err := h.validateWeatherUpdateEmailRequest(c, req); err != nil {
		return
	}

	domainReq := h.mapper.MapWeatherUpdateEmailRequest(req)
	result, err := h.sendEmailUseCase.SendWeatherUpdateEmail(c.Request.Context(), domainReq)
	if err != nil {
		h.handleUseCaseError(c, err, "weather update email")
		return
	}

	h.handleSuccess(c, req.To, "weather update email", result)
}

func (h *EmailHandler) bindConfirmationEmailRequest(c *gin.Context) (dto.ConfirmationEmailRequest, error) {
	var req dto.ConfirmationEmailRequest
	err := c.ShouldBindJSON(&req)
	return req, err
}

func (h *EmailHandler) bindWeatherUpdateEmailRequest(c *gin.Context) (dto.WeatherUpdateEmailRequest, error) {
	var req dto.WeatherUpdateEmailRequest
	err := c.ShouldBindJSON(&req)
	return req, err
}

func (h *EmailHandler) validateConfirmationEmailRequest(c *gin.Context, req dto.ConfirmationEmailRequest) error {
	validationErrors := h.validator.ValidateConfirmationEmailRequest(req)
	if len(validationErrors) > 0 {
		h.handleValidationError(c, validationErrors, "confirmation email")
		return errValidationFailed
	}
	return nil
}

func (h *EmailHandler) validateWeatherUpdateEmailRequest(c *gin.Context, req dto.WeatherUpdateEmailRequest) error {
	validationErrors := h.validator.ValidateWeatherUpdateEmailRequest(req)
	if len(validationErrors) > 0 {
		h.handleValidationError(c, validationErrors, "weather update email")
		return errValidationFailed
	}
	return nil
}

func (h *EmailHandler) handleBindingError(c *gin.Context, err error, emailType string) {
	h.logger.Errorf("Failed to bind JSON for %s: %v", emailType, err)
	c.JSON(http.StatusBadRequest, domain.EmailResponse{
		Success: false,
		Message: "Invalid request: " + err.Error(),
	})
}

func (h *EmailHandler) handleValidationError(c *gin.Context, errors []string, emailType string) {
	h.logger.Errorf("Validation errors for %s: %v", emailType, errors)
	c.JSON(http.StatusBadRequest, domain.EmailResponse{
		Success: false,
		Message: strings.Join(errors, ", "),
	})
}

func (h *EmailHandler) handleUseCaseError(c *gin.Context, err error, emailType string) {
	h.logger.Errorf("Failed to send %s: %v", emailType, err)
	c.JSON(http.StatusInternalServerError, domain.EmailResponse{
		Success: false,
		Message: "Failed to send " + emailType + ": " + err.Error(),
	})
}

func (h *EmailHandler) handleSuccess(c *gin.Context, email string, emailType string, result *domain.EmailDeliveryResult) {
	h.logger.Infof("%s sent successfully to %s", emailType, email)
	if result != nil && result.Error != "" {
		h.logger.Warnf("%s sent with warning: %s", emailType, result.Error)
	}

	c.JSON(http.StatusOK, domain.EmailResponse{
		Success: true,
		Message: emailType + " sent successfully",
	})
}

type validationError struct{}

var errValidationFailed = &validationError{}

func (e *validationError) Error() string {
	return "validation failed"
}
