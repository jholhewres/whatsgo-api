package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/store/sqlstore"

	"github.com/jholhewres/whatsgo-api/internal/config"
	"github.com/jholhewres/whatsgo-api/internal/store"
)

// Manager manages the lifecycle of all WhatsApp instances.
type Manager struct {
	mu        sync.RWMutex
	instances map[string]*Instance

	store     store.Store
	cfg       *config.Config
	logger    *slog.Logger
	container *sqlstore.Container
	eventCh   chan Event
}

// NewManager creates a new WhatsApp instance manager.
func NewManager(s store.Store, cfg *config.Config, logger *slog.Logger) (*Manager, error) {
	container, err := initSessionContainer(cfg)
	if err != nil {
		return nil, fmt.Errorf("initializing session container: %w", err)
	}

	return &Manager{
		instances: make(map[string]*Instance),
		store:     s,
		cfg:       cfg,
		logger:    logger,
		container: container,
		eventCh:   make(chan Event, 4096),
	}, nil
}

func initSessionContainer(cfg *config.Config) (*sqlstore.Container, error) {
	switch cfg.Database.Backend {
	case "sqlite":
		return sqlstore.New(context.Background(), "sqlite3",
			fmt.Sprintf("file:%s?_foreign_keys=on", cfg.WhatsApp.SessionDBPath), nil)
	default:
		// PostgreSQL: use pgx via database/sql stdlib adapter
		// Import pgx/v5/stdlib to register "pgx" driver
		dsn := cfg.Database.PostgreSQL.DSN()
		return sqlstore.New(context.Background(), "pgx", dsn, nil)
	}
}

// Events returns the event channel for the webhook dispatcher.
func (m *Manager) Events() <-chan Event {
	return m.eventCh
}

// LoadAll loads all instances from the database and reconnects those with status "open".
func (m *Manager) LoadAll(ctx context.Context) {
	instances, err := m.store.ListInstances(ctx)
	if err != nil {
		m.logger.Error("failed to load instances", "error", err)
		return
	}

	for _, inst := range instances {
		if inst.Status == string(StateOpen) || inst.Status == string(StateConnecting) {
			m.logger.Info("reconnecting instance", "name", inst.Name)
			go func(inst *store.Instance) {
				if _, err := m.reconnect(ctx, inst); err != nil {
					m.logger.Error("failed to reconnect instance", "name", inst.Name, "error", err)
				}
			}(inst)
		}
	}
}

// Create creates a new WhatsApp instance.
func (m *Manager) Create(ctx context.Context, name string) (*store.Instance, error) {
	existing, _ := m.store.GetInstanceByName(ctx, name)
	if existing != nil {
		return nil, fmt.Errorf("instance %q already exists", name)
	}

	now := time.Now()
	inst := &store.Instance{
		ID:        uuid.New().String(),
		Name:      name,
		Token:     uuid.New().String(),
		Status:    string(StateClose),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := m.store.CreateInstance(ctx, inst); err != nil {
		return nil, fmt.Errorf("creating instance: %w", err)
	}

	// Create default settings
	settings := &store.InstanceSettings{InstanceID: inst.ID}
	if err := m.store.UpsertInstanceSettings(ctx, settings); err != nil {
		m.logger.Error("failed to create default settings", "name", name, "error", err)
	}

	return inst, nil
}

// Connect initiates the WhatsApp connection for an instance.
func (m *Manager) Connect(ctx context.Context, name string) (*Instance, error) {
	dbInst, err := m.store.GetInstanceByName(ctx, name)
	if err != nil || dbInst == nil {
		return nil, fmt.Errorf("instance %q not found", name)
	}

	m.mu.Lock()
	// If already running, disconnect first
	if existing, ok := m.instances[name]; ok {
		existing.Disconnect()
		delete(m.instances, name)
	}
	m.mu.Unlock()

	return m.createAndConnect(ctx, dbInst)
}

// reconnect reconnects an instance from its stored session.
func (m *Manager) reconnect(ctx context.Context, dbInst *store.Instance) (*Instance, error) {
	return m.createAndConnect(ctx, dbInst)
}

func (m *Manager) createAndConnect(ctx context.Context, dbInst *store.Instance) (*Instance, error) {
	settings, err := m.store.GetInstanceSettings(ctx, dbInst.ID)
	if err != nil {
		settings = &store.InstanceSettings{InstanceID: dbInst.ID}
	}

	inst := newInstance(dbInst, settings, m.container, m.store, m.cfg, m.eventCh, m.logger)

	if err := inst.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connecting instance %q: %w", dbInst.Name, err)
	}

	m.mu.Lock()
	m.instances[dbInst.Name] = inst
	m.mu.Unlock()

	return inst, nil
}

// Get returns a running instance by name.
func (m *Manager) Get(name string) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inst, ok := m.instances[name]
	return inst, ok
}

// Disconnect disconnects an instance without logging out.
func (m *Manager) Disconnect(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	if ok {
		delete(m.instances, name)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("instance %q not running", name)
	}

	inst.Disconnect()
	_ = m.store.UpdateInstanceStatus(ctx, inst.dbInst.ID, string(StateClose))
	return nil
}

// Restart disconnects and reconnects an instance.
func (m *Manager) Restart(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	if ok {
		inst.Disconnect()
		delete(m.instances, name)
	}
	m.mu.Unlock()

	dbInst, err := m.store.GetInstanceByName(ctx, name)
	if err != nil || dbInst == nil {
		return fmt.Errorf("instance %q not found", name)
	}

	_, err = m.reconnect(ctx, dbInst)
	return err
}

// Logout logs out from WhatsApp and clears the session.
func (m *Manager) Logout(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	if ok {
		delete(m.instances, name)
	}
	m.mu.Unlock()

	if ok {
		inst.Logout(ctx)
	}

	dbInst, err := m.store.GetInstanceByName(ctx, name)
	if err != nil || dbInst == nil {
		return nil
	}
	_ = m.store.UpdateInstanceStatus(ctx, dbInst.ID, string(StateClose))
	_ = m.store.UpdateInstancePhone(ctx, dbInst.ID, "", "", "")
	return nil
}

// Delete deletes an instance completely.
func (m *Manager) Delete(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	if ok {
		delete(m.instances, name)
	}
	m.mu.Unlock()

	if ok {
		inst.Logout(ctx)
	}

	dbInst, err := m.store.GetInstanceByName(ctx, name)
	if err != nil || dbInst == nil {
		return nil
	}

	return m.store.DeleteInstance(ctx, dbInst.ID)
}

// DisconnectAll disconnects all running instances (used during shutdown).
func (m *Manager) DisconnectAll() {
	m.mu.Lock()
	instances := make(map[string]*Instance, len(m.instances))
	for k, v := range m.instances {
		instances[k] = v
	}
	m.instances = make(map[string]*Instance)
	m.mu.Unlock()

	for name, inst := range instances {
		m.logger.Info("disconnecting instance", "name", name)
		inst.Disconnect()
	}
}
