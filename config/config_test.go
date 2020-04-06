package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type MyConfig struct {
	Logger     *logrus.Logger `yaml:"-"`
	File       string         `                   short:"f" long:"config"   description:"Yaml config"`
	Verbose    []bool         `yaml:"verbose"     short:"v" long:"verbose"  description:"Show verbose debug information"`
	JSONOutput bool           `yaml:"json_output" short:"j" long:"json"     description:"Use json format for output"`
	Version    bool           `                             long:"version"  description:"Show version"`
}

func (c MyConfig) Copy() Config {
	return &MyConfig{
		Logger:     c.Logger,
		File:       c.File,
		Verbose:    c.Verbose,
		JSONOutput: c.JSONOutput,
		Version:    c.Version,
	}
}

func (c MyConfig) ConfigFile() string {
	return c.File
}

func (c MyConfig) Validate(currentConfig Config) []error {
	return nil
}

func (c MyConfig) Apply(currentConfig Config) error {
	if c.JSONOutput {
		c.Logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		c.Logger.SetFormatter(&logrus.TextFormatter{})
	}
	return nil
}

type LogWriter struct {
	t *testing.T
}

func (l *LogWriter) Write(p []byte) (n int, err error) {
	str := strings.TrimSpace(string(p))
	l.t.Log(str)

	return len(str), nil
}

func TestMyConfig(t *testing.T) {
	// Create temporary file for config
	tmpFile, err := ioutil.TempFile(os.TempDir(), "libqd-config-")

	if err != nil {
		t.Error(err)
		return
	}

	defer os.Remove(tmpFile.Name())

	// Logger and test log wrapper
	logger := logrus.New()
	logger.SetOutput(&LogWriter{t})
	logger.SetLevel(logrus.TraceLevel)

	myConfig := &MyConfig{
		// We need to define it otherwise yaml.Marshal will set it to empty
		File:       tmpFile.Name(),
		Logger:     logger,
		Verbose:    []bool{true, true, true, true, true, true},
		JSONOutput: false,
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

	// Init configuration
	confManager := GetManager()
	confManager.logger = logger
	err = confManager.NewConfig(nil, myConfig)

	if err != nil {
		t.Error(err)
		return
	}

	if _, ok := logger.Formatter.(*logrus.TextFormatter); !ok {
		t.Error("Log formatter should be a *logrus.TextFormatter")
		return
	}

	m := confManager.GetConfig(nil).(*MyConfig)

	t.Logf("%#v", m)

	c := confManager.NewConfigChan(nil)

	// Write yaml in the config file with JSON output switch on
	myConfig.JSONOutput = true
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

	// Check that the config has been applied
	if _, ok := logger.Formatter.(*logrus.JSONFormatter); !ok {
		t.Error("Log formatter should be a *logrus.JSONFormatter")
		return
	}
}
