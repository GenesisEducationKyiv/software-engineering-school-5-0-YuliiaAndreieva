package http

import (
	"net/http"

	"token-service/internal/core/domain"
	"token-service/internal/core/ports/in"
	"token-service/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	generateTokenUseCase in.GenerateTokenUseCase
	validateTokenUseCase in.ValidateTokenUseCase
	logger               out.Logger
}

func NewTokenHandler(
	generateTokenUseCase in.GenerateTokenUseCase,
	validateTokenUseCase in.ValidateTokenUseCase,
	logger out.Logger,
) *TokenHandler {
	return &TokenHandler{
		generateTokenUseCase: generateTokenUseCase,
		validateTokenUseCase: validateTokenUseCase,
		logger:               logger,
	}
}

func (h *TokenHandler) GenerateToken(c *gin.Context) {
	var req domain.GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for token generation: %v", err)
		c.JSON(http.StatusBadRequest, domain.GenerateTokenResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.generateTokenUseCase.GenerateToken(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, domain.GenerateTokenResponse{
			Success: false,
			Message: "Failed to generate token: " + err.Error(),
		})
		return
	}

	if result.Success {
		h.logger.Infof("Token generated successfully")
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Token generation failed: %s", result.Message)
		c.JSON(http.StatusBadRequest, result)
	}
}

func (h *TokenHandler) ValidateToken(c *gin.Context) {
	var req domain.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for token validation: %v", err)
		c.JSON(http.StatusBadRequest, domain.ValidateTokenResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.validateTokenUseCase.ValidateToken(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to validate token: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ValidateTokenResponse{
			Success: false,
			Message: "Failed to validate token: " + err.Error(),
		})
		return
	}

	if result.Success {
		if result.Valid {
			h.logger.Infof("Token validated successfully")
		} else {
			h.logger.Warnf("Token validation failed: %s", result.Message)
		}
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Errorf("Token validation error: %s", result.Message)
		c.JSON(http.StatusInternalServerError, result)
	}
} 