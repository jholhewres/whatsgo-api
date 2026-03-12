package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jholhewres/whatsgo-api/internal/config"
	"github.com/jholhewres/whatsgo-api/internal/store"
	"github.com/jholhewres/whatsgo-api/internal/whatsapp"
)

// Dispatcher dispatches webhook events via HTTP POST with retry.
type Dispatcher struct {
	store     store.Store
	globalCfg config.WebhookConfig
	logger    *slog.Logger
	client    *http.Client
	sem       chan struct{} // concurrency limiter
}

// NewDispatcher creates a new webhook dispatcher.
func NewDispatcher(s store.Store, cfg config.WebhookConfig, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		store:     s,
		globalCfg: cfg,
		logger:    logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		sem: make(chan struct{}, 50), // max 50 concurrent dispatches
	}
}

// Start consumes events from the channel and dispatches them.
func (d *Dispatcher) Start(ctx context.Context, events <-chan whatsapp.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-events:
			d.sem <- struct{}{} // acquire
			go func(evt whatsapp.Event) {
				defer func() { <-d.sem }() // release
				d.dispatch(ctx, evt)
			}(evt)
		}
	}
}

func (d *Dispatcher) dispatch(ctx context.Context, evt whatsapp.Event) {
	payload := Payload{
		Event:     string(evt.Event),
		Instance:  evt.Instance,
		Data:      evt.Data,
		Timestamp: evt.Timestamp,
		ServerURL: evt.ServerURL,
	}

	// Global webhook
	if d.globalCfg.GlobalURL != "" && shouldDispatch(d.globalCfg.GlobalEvents, string(evt.Event)) {
		d.send(ctx, d.globalCfg.GlobalURL, d.globalCfg.GlobalHeaders, payload)
	}

	// Per-instance webhook: look up by instance name directly
	inst, err := d.store.GetInstanceByName(ctx, evt.Instance)
	if err != nil || inst == nil {
		return
	}

	wh, err := d.store.GetWebhookByInstance(ctx, inst.ID)
	if err != nil || wh == nil || !wh.Enabled {
		return
	}

	if shouldDispatch(wh.Events, string(evt.Event)) {
		d.send(ctx, wh.URL, wh.Headers, payload)
	}
}

func (d *Dispatcher) send(ctx context.Context, url string, headers map[string]string, payload Payload) {
	body, err := json.Marshal(payload)
	if err != nil {
		d.logger.Error("failed to marshal webhook payload", "error", err)
		return
	}

	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			d.logger.Error("failed to create webhook request", "url", url, "error", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "WhatsGo-API/1.0")
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := d.client.Do(req)
		if err != nil {
			d.logger.Warn("webhook request failed", "url", url, "attempt", attempt+1, "error", err)
			if !d.backoff(ctx, attempt) {
				return
			}
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return
		}

		// Don't retry client errors (4xx) except 408 and 429
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 408 && resp.StatusCode != 429 {
			d.logger.Warn("webhook returned client error, not retrying",
				"url", url, "status", resp.StatusCode, "event", payload.Event)
			return
		}

		d.logger.Warn("webhook returned error",
			"url", url, "status", resp.StatusCode, "attempt", attempt+1, "event", payload.Event)

		if !d.backoff(ctx, attempt) {
			return
		}
	}

	d.logger.Error("webhook delivery failed after retries",
		"url", url, "event", payload.Event, "instance", payload.Instance)
}

// backoff waits with exponential backoff, respecting context cancellation.
func (d *Dispatcher) backoff(ctx context.Context, attempt int) bool {
	delay := time.Duration(1<<uint(attempt)) * time.Second
	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

func shouldDispatch(allowedEvents []string, eventType string) bool {
	if len(allowedEvents) == 0 {
		return true // empty = all events
	}
	for _, e := range allowedEvents {
		if e == eventType || e == "*" {
			return true
		}
	}
	return false
}
