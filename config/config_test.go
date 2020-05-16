package config

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type MyConfig struct {
	Logger  Logger `yaml:"-"`
	File    string `               short:"f" long:"config"  description:"Yaml config"`
	Verbose []bool `yaml:"verbose" short:"v" long:"verbose" description:"Show verbose debug information"`
	Version bool   `                         long:"version" description:"Show version"`
}

func (c MyConfig) Copy() Config {
	return &MyConfig{
		Logger:  c.Logger,
		File:    c.File,
		Verbose: c.Verbose,
		Version: c.Version,
	}
}

func (c MyConfig) ConfigFile() string {
	return c.File
}

type testLogger struct {
	*testing.T
}

func (t *testLogger) Tracef(format string, vals ...interface{}) {
	t.Logf("go-libqd/config: "+format, vals...)
}

func (t *testLogger) Debugf(format string, vals ...interface{}) {
	t.Logf("go-libqd/config: "+format, vals...)
}

func (t *testLogger) Infof(format string, vals ...interface{}) {
	t.Logf("go-libqd/config: "+format, vals...)
}

func (t *testLogger) Warnf(format string, vals ...interface{}) {
	t.Logf("go-libqd/config: "+format, vals...)
}

func TestMyConfig(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create temporary file for config
	tmpFile, err := ioutil.TempFile(os.TempDir(), "libqd-config-")

	if err != nil {
		t.Error(err)
		return
	}

	defer os.Remove(tmpFile.Name())

	// Logger and test log wrapper
	logger := &testLogger{t}

	myConfig := &MyConfig{
		// We need to define it otherwise yaml.Marshal will set it to empty
		File:    tmpFile.Name(),
		Logger:  logger,
		Verbose: []bool{true, true, true, true, true, true},
	}

	yml, err := yaml.Marshal(myConfig)

	if err != nil {
		t.Error(err)
		return
	}

	err = ioutil.WriteFile(tmpFile.Name(), yml, 0)

	if err != nil {
		t.Error(err)
		return
	}

	// Override go test os.Args
	os.Args = []string{
		"test", "-vvvvvv", "-f", tmpFile.Name(),
	}

	// Some variable
	a := 0

	validator := func(currentConfig Config, newConfig Config) []error {
		var errs []error
		var ok bool
		var currentConf *MyConfig

		if currentConfig != nil {
			currentConf, ok = currentConfig.(*MyConfig)

			if !ok {
				errs = append(errs, errors.New("Can not cast currentConfig to MyConfig"))
				return errs
			}
		}

		newConf, ok := newConfig.(*MyConfig)

		if !ok {
			errs = append(errs, errors.New("Can not cast newConfig to MyConfig"))
			return errs
		}

		if len(newConf.Verbose) > 6 {
			errs = append(errs, errors.New("Verbose can not be greater than 6"))
		}

		if currentConf != nil {
			if currentConf.File != newConf.File {
				errs = append(errs, errors.New("File is immutable"))
			}
		}

		return errs
	}

	applier := func(currentConfig Config, newConfig Config) error {
		var err error
		var ok bool
		var currentConf *MyConfig

		if currentConfig != nil {
			currentConf, ok = currentConfig.(*MyConfig)

			if !ok {
				return errors.New("Can not cast currentConfig to MyConfig")
			}
		}

		newConf, ok := newConfig.(*MyConfig)

		if !ok {
			return errors.New("Can not cast newConfig to MyConfig")
		}

		// Increment `a` only after first reload
		if currentConf != nil && newConf != nil {
			a++
		}

		return err
	}

	confManager := GetManager(logger)
	confManager.AddValidators(nil, validator)
	confManager.AddAppliers(nil, applier)

	// Launch config
	err = confManager.MakeConfig(ctx, nil, myConfig)

	if err != nil {
		t.Error(err)
		return
	}

	if a != 0 {
		t.Errorf("a=%d but should be 0", a)
	}

	m := confManager.GetConfig(nil).(*MyConfig)

	t.Logf("%#v", m)

	c := confManager.NewConfigChan(nil)

	myConfig.Version = true
	yml, err = yaml.Marshal(myConfig)

	if err != nil {
		t.Error(err)
		return
	}

	err = ioutil.WriteFile(tmpFile.Name(), yml, 0)

	if err != nil {
		t.Error(err)
		return
	}

	// Check that a new config is sent via the channel
	select {
	case newConf := <-c:
		t.Logf("%#v", newConf)
	case <-time.After(5 * time.Second):
		t.Error("No new configuration received")
		return
	}

	if a != 1 {
		t.Errorf("a=%d but should be 1", a)
	}
}
