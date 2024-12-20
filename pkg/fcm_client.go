// Package fcmclient provides a Firebase Cloud Messaging client for sending notifications
// to both Android and iOS devices. It handles authentication, retries, and proper
// error handling according to FCM best practices.
package fcmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Common errors that might occur during FCM operations
var (
	ErrInvalidToken        = errors.New("invalid registration token")
	ErrNotRegistered       = errors.New("token is no longer registered")
	ErrMessageTooBig       = errors.New("message payload is too big")
	ErrInvalidMessage      = errors.New("message format is invalid")
	ErrServerError         = errors.New("FCM server error")
	ErrAuthenticationError = errors.New("authentication error with FCM")
	ErrContextTimeout      = errors.New("context deadline must be set")
	ErrInvalidPriority     = errors.New("invalid notification priority")
)

// FCMClient handles communication with Firebase Cloud Messaging
type FCMClient struct {
	serverKey    string
	httpClient   *http.Client
	maxRetries   int
	retryBackoff time.Duration
	projectID    string
	rateLimiter  <-chan time.Time
	mu           sync.Mutex // mu protects concurrent access to shared resources in FCMClient.
}

// ClientConfig holds the configuration options for FCMClient
type ClientConfig struct {
	ServerKey      string
	ProjectID      string
	HTTPClient     *http.Client
	MaxRetries     int
	RetryBackoff   time.Duration
	RequestsPerSec int
}

// DefaultConfig returns a ClientConfig with sensible defaults
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		MaxRetries:     3,
		RetryBackoff:   1 * time.Second,
		RequestsPerSec: 100,
	}
}

// NewFCMClient creates a new FCM client with the provided configuration
func NewFCMClient(config *ClientConfig) (*FCMClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate required fields
	if config.ServerKey == "" {
		return nil, errors.New("server key is required")
	}
	if config.ProjectID == "" {
		return nil, errors.New("project ID is required")
	}

	client := &FCMClient{
		serverKey:    config.ServerKey,
		projectID:    config.ProjectID,
		httpClient:   config.HTTPClient,
		maxRetries:   config.MaxRetries,
		retryBackoff: config.RetryBackoff,
		rateLimiter:  time.Tick(time.Second / time.Duration(config.RequestsPerSec)),
	}

	return client, nil
}

// Notification represents a push notification to be sent
type Notification struct {
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Sound    string            `json:"sound,omitempty"`
	Badge    int               `json:"badge,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	Priority string            `json:"priority,omitempty"` // high or normal
}

// Message represents the FCM message format
type Message struct {
	Token        string         `json:"token"`
	Notification *Notification  `json:"notification"`
	Data         interface{}    `json:"data,omitempty"`
	Android      *AndroidConfig `json:"android,omitempty"`
	APNS         *APNSConfig    `json:"apns,omitempty"`
}

// AndroidConfig contains Android-specific message configuration
type AndroidConfig struct {
	Priority    string `json:"priority,omitempty"` // high or normal
	CollapseKey string `json:"collapse_key,omitempty"`
	TTL         string `json:"ttl,omitempty"`
}

// APNSConfig contains iOS-specific message configuration
type APNSConfig struct {
	Headers map[string]string `json:"headers,omitempty"`
	Payload APNSPayload       `json:"payload"`
}

// APNSPayload represents the APNS payload structure
type APNSPayload struct {
	Aps Aps `json:"aps"`
}

// Aps represents the core APNS notification payload
type Aps struct {
	Alert            interface{} `json:"alert,omitempty"`
	Badge            int         `json:"badge,omitempty"`
	Sound            string      `json:"sound,omitempty"`
	ContentAvailable int         `json:"content-available,omitempty"`
	MutableContent   int         `json:"mutable-content,omitempty"`
	Category         string      `json:"category,omitempty"`
}

// Option defines a function type for configuring the FCM client
type Option func(*FCMClient)

// WithRetryConfig sets custom retry parameters
func WithRetryConfig(maxRetries int, backoff time.Duration) Option {
	return func(c *FCMClient) {
		c.maxRetries = maxRetries
		c.retryBackoff = backoff
	}
}

// validateToken performs comprehensive token validation
func (c *FCMClient) validateToken(token string) error {
	if token == "" {
		return ErrInvalidToken
	}

	// FCM tokens are typically 163 characters
	if len(token) > 163 {
		return fmt.Errorf("%w: token exceeds maximum length", ErrInvalidToken)
	}

	// Add basic character validation (could be enhanced based on FCM requirements)
	for _, char := range token {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return fmt.Errorf("%w: invalid characters in token", ErrInvalidToken)
		}
	}

	return nil
}

// validateNotification performs comprehensive notification validation
func (c *FCMClient) validateNotification(notification *Notification) error {
	if notification == nil {
		return ErrInvalidMessage
	}

	if notification.Title == "" && notification.Body == "" {
		return fmt.Errorf("%w: either title or body must be present", ErrInvalidMessage)
	}

	// Validate priority values
	if notification.Priority != "" &&
		notification.Priority != "high" &&
		notification.Priority != "normal" {
		return ErrInvalidPriority
	}

	// Initialize empty data map if nil
	if notification.Data == nil {
		notification.Data = make(map[string]string)
	}

	return nil
}

// SendNotification sends a push notification to a specific device token
func (c *FCMClient) SendNotification(ctx context.Context, token string, notification *Notification) error {
	// Ensure context has a deadline
	if deadline, ok := ctx.Deadline(); !ok || deadline.IsZero() {
		return ErrContextTimeout
	}

	// Validate inputs
	if err := c.validateToken(token); err != nil {
		return fmt.Errorf("token validation error: %w", err)
	}
	if err := c.validateNotification(notification); err != nil {
		return fmt.Errorf("notification validation error: %w", err)
	}

	message := &Message{
		Token:        token,
		Notification: notification,
		Data:         notification.Data,
		Android: &AndroidConfig{
			Priority: notification.Priority,
		},
		APNS: &APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: APNSPayload{
				Aps: Aps{
					Alert: map[string]string{
						"title": notification.Title,
						"body":  notification.Body,
					},
					Sound: notification.Sound,
					Badge: notification.Badge,
				},
			},
		},
	}

	return c.sendWithRetry(ctx, message)
}

// validateInputs performs basic validation of notification parameters
func (c *FCMClient) validateInputs(token string, notification *Notification) error {
	if token == "" {
		return ErrInvalidToken
	}
	if notification == nil {
		return ErrInvalidMessage
	}
	if notification.Title == "" && notification.Body == "" {
		return ErrInvalidMessage
	}
	return nil
}

// sendWithRetry attempts to send the message with retries on certain errors
func (c *FCMClient) sendWithRetry(ctx context.Context, message *Message) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Apply rate limiting
		select {
		case <-c.rateLimiter:
		case <-ctx.Done():
			return ctx.Err()
		}

		if attempt > 0 {
			// Wait before retrying, using exponential backoff
			backoff := c.retryBackoff * time.Duration(1<<uint(attempt-1))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := c.send(ctx, message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Enhanced retry logic based on error type and response codes
		if !c.shouldRetry(err) {
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// send performs the actual HTTP request to FCM
func (c *FCMClient) send(ctx context.Context, message *Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	// Use dynamic project-specific URL
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.projectID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Might want to replace with OAuth 2.0 token when implementing proper authentication
	req.Header.Set("Authorization", "Bearer "+c.serverKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp)
}

// shouldRetry determines if a retry attempt should be made based on the error
func (c *FCMClient) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Retry on server errors and rate limiting
	if err == ErrServerError {
		return true
	}

	// Don't retry on client errors or invalid tokens
	if errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrNotRegistered) ||
		errors.Is(err, ErrMessageTooBig) ||
		errors.Is(err, ErrInvalidMessage) {
		return false
	}

	// Retry on network errors and unexpected errors
	return true
}

// handleResponse processes the FCM API response and returns appropriate errors
func (c *FCMClient) handleResponse(resp *http.Response) error {
	// Read response body for error logging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", ErrInvalidMessage, string(body))
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", ErrAuthenticationError, string(body))
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", ErrNotRegistered, string(body))
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", ErrServerError, string(body))
	default:
		if resp.StatusCode >= 500 {
			return fmt.Errorf("%w: %s", ErrServerError, string(body))
		}
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

// SendBatchNotifications sends notifications to multiple tokens in batches
func (c *FCMClient) SendBatchNotifications(ctx context.Context, tokens []string, notification *Notification) []error {
	if len(tokens) == 0 {
		return nil
	}

	// Create buffered error channel
	errChan := make(chan error, len(tokens))
	var wg sync.WaitGroup

	// Process tokens in batches of 500 (FCM limit)
	batchSize := 500
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}

		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()
			for _, token := range batch {
				if err := c.SendNotification(ctx, token, notification); err != nil {
					errChan <- fmt.Errorf("error sending to token %s: %w", token, err)
				}
			}
		}(tokens[i:end])
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	return errors
}
