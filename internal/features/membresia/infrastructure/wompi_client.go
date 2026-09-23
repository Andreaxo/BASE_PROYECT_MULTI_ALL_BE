package infrastructure

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// WompiClient handles communication with the Wompi payment gateway API.
type WompiClient struct {
	publicKey       string
	privateKey      string
	eventsSecret    string
	integritySecret string
	sandbox         bool
	httpClient      *http.Client
}

// NewWompiClient creates a new WompiClient.
func NewWompiClient(publicKey, privateKey, eventsSecret, integritySecret string, sandbox bool) *WompiClient {
	return &WompiClient{
		publicKey:       publicKey,
		privateKey:      privateKey,
		eventsSecret:    eventsSecret,
		integritySecret: integritySecret,
		sandbox:         sandbox,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *WompiClient) baseURL() string {
	if c.sandbox {
		return "https://sandbox.wompi.co/v1"
	}
	return "https://production.wompi.co/v1"
}

// --- Transaction creation ---

// CreateTransactionRequest is the payload to create a Wompi transaction.
type CreateTransactionRequest struct {
	AmountInCents   int64  `json:"amount_in_cents"`
	Currency        string `json:"currency"`
	CustomerEmail   string `json:"customer_email"`
	Reference       string `json:"reference"`
	RedirectURL     string `json:"redirect_url"`
	PaymentSourceID *int64 `json:"payment_source_id,omitempty"`
}

// CreateTransactionResponse is Wompi's response for transaction creation.
type CreateTransactionResponse struct {
	Data struct {
		ID                string `json:"id"`
		Status            string `json:"status"`
		Reference         string `json:"reference"`
		PublicCheckoutURL string `json:"public_checkout_url"`
	} `json:"data"`
}

// CreateTransaction generates the official Wompi Web Checkout hosted URL.
// Official formula for signature:integrity:
// SHA256(reference + amountInCents + currency + integritySecret)
// Documentation: https://docs.wompi.co/docs/colombia/widget-web-checkout
func (c *WompiClient) CreateTransaction(req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
	checkoutBase := "https://checkout.wompi.co/p/"
	params := url.Values{}
	params.Set("public-key", c.publicKey)
	params.Set("currency", req.Currency)
	params.Set("amount-in-cents", strconv.FormatInt(req.AmountInCents, 10))
	params.Set("reference", req.Reference)
	// Only set redirect-url if it is a valid HTTPS URL.
	// AWS CloudFront WAF on checkout.wompi.co blocks requests with non-HTTPS or localhost redirect URLs with a 403 error.
	if req.RedirectURL != "" && strings.HasPrefix(req.RedirectURL, "https://") {
		params.Set("redirect-url", req.RedirectURL)
	}

	// Calculate integrity signature if integritySecret is configured
	if c.integritySecret != "" {
		concat := fmt.Sprintf("%s%d%s%s", req.Reference, req.AmountInCents, req.Currency, c.integritySecret)
		hash := sha256.Sum256([]byte(concat))
		params.Set("signature:integrity", hex.EncodeToString(hash[:]))
	}

	checkoutURL := fmt.Sprintf("%s?%s", checkoutBase, params.Encode())

	res := &CreateTransactionResponse{}
	res.Data.ID = req.Reference
	res.Data.Status = "PENDING"
	res.Data.Reference = req.Reference
	res.Data.PublicCheckoutURL = checkoutURL

	return res, nil
}

// --- Token-based charges (for renewals) ---

// TokenChargeRequest is the payload to charge using a saved payment token.
type TokenChargeRequest struct {
	AmountInCents   int64  `json:"amount_in_cents"`
	Currency        string `json:"currency"`
	CustomerEmail   string `json:"customer_email"`
	Reference       string `json:"reference"`
	PaymentSourceID int64  `json:"payment_source_id"`
}

// ChargeWithToken creates a transaction using a saved payment source token.
func (c *WompiClient) ChargeWithToken(req *TokenChargeRequest) (*CreateTransactionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token charge request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL()+"/transactions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.privateKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Wompi API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Wompi response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Wompi API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result CreateTransactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode Wompi response: %w", err)
	}

	return &result, nil
}

// --- Webhook signature validation ---

// WompiWebhookEvent represents the structure of a Wompi webhook event.
type WompiWebhookEvent struct {
	Event     string `json:"event"`
	Data      struct {
		Transaction struct {
			ID              string `json:"id"`
			Status          string `json:"status"`
			Reference       string `json:"reference"`
			AmountInCents   int64  `json:"amount_in_cents"`
			Currency        string `json:"currency"`
			PaymentMethodType string `json:"payment_method_type"`
			PaymentMethod   struct {
				Token string `json:"token"`
			} `json:"payment_method"`
		} `json:"transaction"`
	} `json:"data"`
	Signature struct {
		Checksum   string   `json:"checksum"`
		Properties []string `json:"properties"`
	} `json:"signature"`
	Timestamp int64 `json:"timestamp"`
}

// ValidateWebhookSignature validates the HMAC signature of a Wompi webhook event.
// Wompi signs: concatenation of property values in order + timestamp + events_secret, then SHA256.
func (c *WompiClient) ValidateWebhookSignature(event *WompiWebhookEvent) bool {
	if event.Signature.Checksum == "" {
		return false
	}

	// Build the string to hash: property values concatenated in order + timestamp + secret
	tx := event.Data.Transaction
	valueMap := map[string]string{
		"transaction.id":              tx.ID,
		"transaction.status":          tx.Status,
		"transaction.amount_in_cents": fmt.Sprintf("%d", tx.AmountInCents),
		"transaction.amountInCents":   fmt.Sprintf("%d", tx.AmountInCents),
		"transaction.reference":       tx.Reference,
	}

	var concat string
	for _, prop := range event.Signature.Properties {
		if val, ok := valueMap[prop]; ok {
			concat += val
		}
	}
	concat += fmt.Sprintf("%d", event.Timestamp)
	concat += c.eventsSecret

	hash := sha256.Sum256([]byte(concat))
	computed := hex.EncodeToString(hash[:])

	return computed == event.Signature.Checksum
}

// --- Query Transactions by Reference or ID ---

// WompiTransactionItem represents a single transaction item returned by Wompi's API.
type WompiTransactionItem struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Reference     string `json:"reference"`
	AmountInCents int64  `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	PaymentMethod struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	} `json:"payment_method"`
}

// WompiSearchResponse represents the list response from GET /transactions?reference=...
type WompiSearchResponse struct {
	Data []WompiTransactionItem `json:"data"`
}

// GetTransactionsByReference queries Wompi API for transactions with the specified reference.
func (c *WompiClient) GetTransactionsByReference(reference string) ([]WompiTransactionItem, error) {
	if c.privateKey == "" || reference == "" {
		return nil, fmt.Errorf("private key or reference is empty")
	}

	reqURL := fmt.Sprintf("%s/transactions?reference=%s", c.baseURL(), url.QueryEscape(reference))
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.privateKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wompi API error status %d: %s", resp.StatusCode, string(body))
	}

	var res WompiSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
