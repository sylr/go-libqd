package main

import (
	"errors"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/sylr/go-libqd/config"
)

// MyAppConfiguration implements github.com/sylr/go-libqd/config.Config
type MyAppConfiguration struct {
	Logger   *logrus.Logger `yaml:"-"       json:"-"`
	Reloads  int32          `yaml:"-"`
	File     string         `                           short:"f" long:"config"`
	Verbose  []bool         `yaml:"verbose"             short:"v" long:"verbose"`
	Version  bool           `                                     long:"version"`
	HTTPPort int            `yaml:"port"    json:"port" short:"p" long:"port"`
}

func (c *MyAppConfiguration) Copy() config.Config {
	return &MyAppConfiguration{
		Logger:   c.Logger,
		Reloads:  c.Reloads,
		File:     c.File,
		Verbose:  c.Verbose,
		Version:  c.Version,
		HTTPPort: c.HTTPPort,
	}
}

func (c *MyAppConfiguration) ConfigFile() string {
	return c.File
}

func (c *MyAppConfiguration) Validate(currentConfig config.Config) []error {
	var errs []error
	var current *MyAppConfiguration

	c.Logger.Trace("Validating config")

	if currentConfig != nil {
		current = currentConfig.(*MyAppConfiguration)
	}

	if current != nil && current.HTTPPort != c.HTTPPort {
		errs = append(errs, errors.New("port cannot be changed"))
	}

	if c.HTTPPort < 0 || c.HTTPPort > 65535 {
		errs = append(errs, errors.New("port should be between 0 and 65535"))
	}

	return errs
}

func (c *MyAppConfiguration) Apply(currentConfig config.Config) error {
	c.Logger.Trace("Applying config")

	if currentConfig != nil {
		atomic.AddInt32(&c.Reloads, 1)
	}

	switch len(c.Verbose) {
	case 0:
		c.Logger.SetLevel(logrus.ErrorLevel)
	case 1:
		c.Logger.SetLevel(logrus.WarnLevel)
	case 2:
		c.Logger.SetLevel(logrus.InfoLevel)
	case 3:
		c.Logger.SetLevel(logrus.DebugLevel)
	case 4:
		fallthrough
	default:
		c.Logger.SetLevel(logrus.TraceLevel)
	}

	return nil
}
