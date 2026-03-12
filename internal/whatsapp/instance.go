package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"go.mau.fi/whatsmeow"
	waStore "go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"

	"github.com/jholhewres/whatsgo-api/internal/config"
	"github.com/jholhewres/whatsgo-api/internal/store"
)

// Instance wraps a single whatsmeow client.
type Instance struct {
	client    *whatsmeow.Client
	container *sqlstore.Container
	dbStore   store.Store
	dbInst    *store.Instance
	cfg       *config.Config
	logger    *slog.Logger
	eventCh   chan<- Event

	connected atomic.Bool
	mu        sync.Mutex // protects dbInst fields

	// QR code subscription
	qrMu        sync.Mutex
	qrObservers []chan QREvent

	// Settings
	settingsMu sync.RWMutex
	settings   *store.InstanceSettings

	cancel context.CancelFunc
}

func newInstance(
	dbInst *store.Instance,
	settings *store.InstanceSettings,
	container *sqlstore.Container,
	dbStore store.Store,
	cfg *config.Config,
	eventCh chan<- Event,
	logger *slog.Logger,
) *Instance {
	return &Instance{
		container: container,
		dbStore:   dbStore,
		dbInst:    dbInst,
		cfg:       cfg,
		logger:    logger.With("instance", dbInst.Name),
		eventCh:   eventCh,
		settings:  settings,
	}
}

// Connect establishes the WhatsApp connection.
func (i *Instance) Connect(ctx context.Context) error {
	// Get or create device
	var device *waStore.Device
	var err error

	if i.dbInst.WhatsmeowJID != "" {
		jid, parseErr := waTypes.ParseJID(i.dbInst.WhatsmeowJID)
		if parseErr == nil {
			device, err = i.container.GetDevice(ctx, jid)
			if err != nil {
				i.logger.Warn("failed to get device, creating new", "error", err)
			}
		}
	}

	if device == nil {
		device = i.container.NewDevice()
	}

	i.client = whatsmeow.NewClient(device, nil)
	i.client.AddEventHandler(i.handleEvent)
	i.client.EnableAutoReconnect = true
	i.client.AutoTrustIdentity = true

	_, cancel := context.WithCancel(ctx)
	i.cancel = cancel

	// Update status to connecting
	_ = i.dbStore.UpdateInstanceStatus(ctx, i.dbInst.ID, string(StateConnecting))
	i.dbInst.Status = string(StateConnecting)

	if device.ID == nil {
		// New device - need QR login
		return i.loginWithQR(ctx)
	}

	// Existing session - just connect
	return i.client.Connect()
}

func (i *Instance) loginWithQR(ctx context.Context) error {
	qrChan, err := i.client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("getting QR channel: %w", err)
	}

	if err := i.client.Connect(); err != nil {
		return fmt.Errorf("connecting: %w", err)
	}

	// Process QR events in background
	go func() {
		for evt := range qrChan {
			switch evt.Event {
			case "code":
				i.notifyQR(QREvent{Type: "code", Code: evt.Code})
				i.emit(EventQRCodeUpdated, map[string]string{"qr_code": evt.Code})
			case "success":
				i.notifyQR(QREvent{Type: "success"})
			case "timeout":
				i.notifyQR(QREvent{Type: "timeout"})
			default:
				i.notifyQR(QREvent{Type: evt.Event})
			}
		}
	}()

	return nil
}

// ConnectWithPairingCode uses phone number pairing instead of QR.
func (i *Instance) ConnectWithPairingCode(ctx context.Context, phone string) (string, error) {
	if i.client == nil {
		return "", fmt.Errorf("client not initialized, call Connect first")
	}

	code, err := i.client.PairPhone(ctx, phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return "", fmt.Errorf("pairing phone: %w", err)
	}

	return code, nil
}

// Disconnect disconnects without logging out.
func (i *Instance) Disconnect() {
	if i.cancel != nil {
		i.cancel()
	}
	if i.client != nil && i.client.IsConnected() {
		i.client.Disconnect()
	}
	i.connected.Store(false)
}

// Logout logs out from WhatsApp and clears session data.
func (i *Instance) Logout(ctx context.Context) {
	if i.client != nil {
		if i.client.IsLoggedIn() {
			_ = i.client.Logout(ctx)
		} else if i.client.IsConnected() {
			i.client.Disconnect()
		}
	}
	i.connected.Store(false)
}

// IsConnected returns whether the client is connected.
func (i *Instance) IsConnected() bool {
	return i.connected.Load()
}

// SubscribeQR registers a channel to receive QR events.
func (i *Instance) SubscribeQR() (chan QREvent, func()) {
	ch := make(chan QREvent, 10)
	i.qrMu.Lock()
	i.qrObservers = append(i.qrObservers, ch)
	i.qrMu.Unlock()

	unsub := func() {
		i.qrMu.Lock()
		defer i.qrMu.Unlock()
		for idx, obs := range i.qrObservers {
			if obs == ch {
				i.qrObservers = append(i.qrObservers[:idx], i.qrObservers[idx+1:]...)
				break
			}
		}
		close(ch)
	}

	return ch, unsub
}

func (i *Instance) notifyQR(evt QREvent) {
	i.qrMu.Lock()
	defer i.qrMu.Unlock()
	for _, ch := range i.qrObservers {
		select {
		case ch <- evt:
		default:
		}
	}
}

// UpdateSettings updates the runtime settings.
func (i *Instance) UpdateSettings(settings *store.InstanceSettings) {
	i.settingsMu.Lock()
	i.settings = settings
	i.settingsMu.Unlock()

	// Apply always_online if changed
	if settings.AlwaysOnline && i.client != nil && i.client.IsConnected() {
		_ = i.client.SendPresence(context.Background(), waTypes.PresenceAvailable)
	}
}

func (i *Instance) getSettings() *store.InstanceSettings {
	i.settingsMu.RLock()
	defer i.settingsMu.RUnlock()
	return i.settings
}

// emit sends an event to the manager's event channel.
func (i *Instance) emit(eventType EventType, data any) {
	select {
	case i.eventCh <- Event{
		Event:     eventType,
		Instance:  i.dbInst.Name,
		Data:      data,
		Timestamp: time.Now(),
		ServerURL: i.cfg.Server.BaseURL,
	}:
	default:
		i.logger.Warn("event channel full, dropping event", "event", eventType)
	}
}
