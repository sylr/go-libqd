package config

import (
	"fmt"
	"sync"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/sirupsen/logrus"
)

var (
	manager *Manager
)

// -----------------------------------------------------------------------------

// Config is an interface describing functions needed by this module.
type Config interface {
	// Copy returns a copy of the current struct.
	Copy() Config
	// ConfigFile returns the path of the config file.
	ConfigFile() string
	// Validate is the function called upon creation of a new config and when the
	// configuration file associated to the configuration is updated. It makes sure
	// the configuration is valid. The currentConfig argument is nil when Validate()
	// is called the first time, then it will contain the current configuration.
	Validate(currentConfig Config) []error
	// Apply is the function called upon validation of a new or updated
	// configuration. The currentConfig argument is nil when Apply()
	// is called the first time, then it will contain the current configuration.
	Apply(currentConfig Config) error
}

// Chan is a channel within which pointers to a new configuration will be sent.
type Chan chan Config

// -----------------------------------------------------------------------------

// GetManager returns the Manager that will give you access to all your configs.
func GetManager() *Manager {
	if manager == nil {
		manager = &Manager{
			logger: logrus.New(),
		}
	}

	return manager
}

// Manager is a struct that stores configuration watchers and the chans used to
// broadcasts new configurations when configurations files are updated.
type Manager struct {
	logger   *logrus.Logger
	watchers map[interface{}]*watcher
	chans    map[interface{}][]Chan
	mu       sync.RWMutex
}

// GetConfig returns an existing configuration, nil otherwise.
func (m *Manager) GetConfig(name interface{}) Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if c, ok := m.watchers[name]; ok {
		return c.config
	}

	return nil
}

// NewConfig returns a new initialized configuration. This will create a new
// goroutine which watches the configuration file for updates.
func (m *Manager) NewConfig(name interface{}, conf Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// local logger
	llogger := m.logger.WithFields(logrus.Fields{"_func": "Init"})
	llogger.Trace()

	if m.watchers == nil {
		m.watchers = make(map[interface{}]*watcher)
	}

	if _, ok := m.watchers[name]; ok {
		return fmt.Errorf("The configuration `%v` already exists", name)
	}

	m.watchers[name] = &watcher{
		logger: m.logger,
		name:   name,
		config: conf,
	}

	// Load config from cli args and then from config file if
	err := m.watchers[name].loadConfig(m.watchers[name].config)

	if err != nil {
		return errors.Errorf("Error while loading conf: %w", err)
	}

	// Validate config
	errs := m.watchers[name].config.Validate(nil)

	if len(errs) > 0 {
		for _, err := range errs {
			llogger.Errorf("Error while validating conf: %v", err)
		}

		return fmt.Errorf("Configuration not applied because %d error(s) have been found", len(errs))
	}

	// Apply config
	err = m.watchers[name].config.Apply(nil)

	if err != nil {
		return fmt.Errorf("Error while applying conf: %w", err)
	}

	go m.watchers[name].watchConfigFile()

	// Sleep a bit to let the watchConfigFile go routine the time to watch
	// the configuration file it it supposed to.
	time.Sleep(time.Millisecond)

	return err
}

// NewConfigChan returns a channel that will be used to send new configurations
// when the configuration file associated to the Config has been updated.
func (m *Manager) NewConfigChan(name interface{}) Chan {
	c := make(chan Config)
	m.registerChan(name, c)

	return c
}

func (m *Manager) registerChan(name interface{}, c Chan) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.chans == nil {
		m.chans = make(map[interface{}][]Chan)
	}

	m.chans[name] = append(m.chans[name], c)
}

// broadcastNewConfig sends a configuration pointer in all registered channels.
func (m *Manager) broadcastNewConfig(name interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.chans[name] {
		m.logger.Tracef("Signaling new conf %p in chan %p", m.watchers[name].config, m.chans[name][i])
		m.chans[name][i] <- m.watchers[name].config
	}
}
