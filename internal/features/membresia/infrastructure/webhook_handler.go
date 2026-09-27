package infrastructure

import (
	"io"
	"log"
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
		log.Printf("❌ [Wompi Webhook] Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	signature := c.GetHeader("X-Event-Checksum")
	log.Printf("📩 [Wompi Webhook] Received webhook event (%d bytes). Header signature: %s", len(body), signature)

	if err := h.service.ProcessWebhook(body, signature); err != nil {
		log.Printf("⚠️ [Wompi Webhook Error]: %v", err)
		// Return 200 with status error so Wompi doesn't unnecessarily flood retries on logic errors
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	log.Printf("✅ [Wompi Webhook] Processed successfully")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
