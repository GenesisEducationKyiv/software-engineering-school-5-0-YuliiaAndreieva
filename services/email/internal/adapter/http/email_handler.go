package http

import (
	"net/http"

	"email-service/internal/core/domain"
	"email-service/internal/core/ports/in"
	"email-service/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	sendEmailUseCase in.SendEmailUseCase
	logger           out.Logger
}

func NewEmailHandler(sendEmailUseCase in.SendEmailUseCase, logger out.Logger) *EmailHandler {
	return &EmailHandler{
		sendEmailUseCase: sendEmailUseCase,
		logger:           logger,
	}
}

func (h *EmailHandler) SendConfirmationEmail(c *gin.Context) {
	var req domain.ConfirmationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for confirmation email: %v", err)
		c.JSON(http.StatusBadRequest, domain.EmailResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.sendEmailUseCase.SendConfirmationEmail(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to send confirmation email: %v", err)
		c.JSON(http.StatusInternalServerError, domain.EmailResponse{
			Success: false,
			Message: "Failed to send confirmation email: " + err.Error(),
		})
		return
	}

	h.logger.Infof("Confirmation email sent successfully to %s", req.To)
	if result != nil && result.Error != "" {
		h.logger.Warnf("Confirmation email sent with warning: %s", result.Error)
	}

	c.JSON(http.StatusOK, domain.EmailResponse{
		Success: true,
		Message: "Confirmation email sent successfully",
	})
}

func (h *EmailHandler) SendWeatherUpdateEmail(c *gin.Context) {
	var req domain.WeatherUpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for weather update email: %v", err)
		c.JSON(http.StatusBadRequest, domain.EmailResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.sendEmailUseCase.SendWeatherUpdateEmail(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to send weather update email: %v", err)
		c.JSON(http.StatusInternalServerError, domain.EmailResponse{
			Success: false,
			Message: "Failed to send weather update email: " + err.Error(),
		})
		return
	}

	h.logger.Infof("Weather update email sent successfully to %s", req.To)
	if result != nil && result.Error != "" {
		h.logger.Warnf("Weather update email sent with warning: %s", result.Error)
	}

	c.JSON(http.StatusOK, domain.EmailResponse{
		Success: true,
		Message: "Weather update email sent successfully",
	})
}
