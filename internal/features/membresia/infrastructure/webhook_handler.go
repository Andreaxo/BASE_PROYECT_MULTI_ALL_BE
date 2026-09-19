package infrastructure

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/membresia/domain"
)

// WebhookHandler handles Wompi webhook notifications.
type WebhookHandler struct {
	service domain.MembresiaService
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(service domain.MembresiaService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// HandleWompiWebhook handles POST /api/webhooks/wompi
// This endpoint is PUBLIC (no JWT) — Wompi calls it from their servers.
func (h *WebhookHandler) HandleWompiWebhook(c *gin.Context) {
	// Read raw body for signature validation
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	signature := c.GetHeader("X-Event-Checksum")

	if err := h.service.ProcessWebhook(body, signature); err != nil {
		// Log but return 200 to prevent Wompi retries on business logic errors
		// Only return non-200 on actual processing failures
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
