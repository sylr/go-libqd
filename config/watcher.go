package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	json "github.com/tailscale/hujson"
	yaml "gopkg.in/yaml.v3"
)

type watcher struct {
	logger *log.Logger
	name   interface{}
	config Config
}

// reload configuration
func (w *watcher) reload() {
	newConfig := w.config.Copy()

	// local logger
	llogger := w.logger.WithFields(log.Fields{
		"_func": "reload",
	})

	// Load config from cli args and then from config file if
	err := w.loadConfig(newConfig)

	if err != nil {
		llogger.Errorf("Error while loading conf: %v", err)
		return
	}

	// Validate new config
	errs := newConfig.Validate(w.config)

	if len(errs) > 0 {
		for _, err := range errs {
			llogger.Errorf("Error while validating new conf: %v", err)
		}

		err = errors.New("New configuration not applied because error(s) have been found")
		llogger.Error(err)
		return
	}

	// Apply config
	err = newConfig.Apply(w.config)

	if err != nil {
		llogger.Errorf("Error while applying conf: %v", err)
		return
	}

	// update current configuration
	w.config = newConfig
	manager := GetManager()
	manager.broadcastNewConfig(w.name)
}

func (w *watcher) loadConfig(conf Config) error {
	// Read cli arguments and loads in into config, it will exit if errors occurs
	w.readConfigCLIOptions(conf)

	// Read config file content and loads in into config
	err := w.readConfigFile(conf)

	if err != nil {
		w.logger.Errorf("Configuration not applied because parsing of config file failed: %s", err)
		return err
	}

	return nil
}

func (w *watcher) watchConfigFile(ctx context.Context) {
	// local logger
	llogger := w.logger.WithFields(log.Fields{
		"_func": "watchConfigFile",
	})

	configFile := w.config.ConfigFile()

	llogger.Debugf("Watching config files %s", configFile)

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		llogger.Fatal(err)
	}

	err = watcher.Add(configFile)

	if err != nil {
		llogger.Fatal(err)
	}

	if len(os.Getenv("KUBERNETES_PORT")) > 0 {
		dir := filepath.Dir(configFile)
		llogger.Infof("In kubernetes context, adding %s to watch list", dir)
		err := watcher.Add(dir)

		if err != nil {
			llogger.Fatal(err)
		}
	}

	defer watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				llogger.Error("fsnotify: error")
				break
			}

			llogger.Debugf("fsnotify: %s -> %s", event.Name, event.Op.String())

			if event.Op&fsnotify.Write == fsnotify.Write {
				if event.Name == configFile {
					llogger.Debugf("config: file changed")
				}
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				if event.Name == configFile {
					llogger.Debugf("config: file created")
				} else if filepath.Base(event.Name) == "..data" {
					llogger.Debugf("config: configmap volume updated")
				} else {
					break
				}
			} else {
				break
			}

			llogger.Info("config: reloading config")

			// Reload configuration
			w.reload()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			llogger.Errorf("fsnotify: %s", err)
		}
	}
}

// readConfigCLIOptions loads config from cli arguments
func (w *watcher) readConfigCLIOptions(conf Config) {
	parser := flags.NewParser(conf, flags.Default)

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Fatal(err)
			os.Exit(1)
		}
	}
}

// readConfigFile parses the config file defined by -f/--config
func (w *watcher) readConfigFile(conf Config) error {
	configFile := conf.ConfigFile()
	if conf == nil || len(configFile) == 0 {
		return nil
	}

	err := w.loadFile(conf, configFile)

	if err != nil {
		return err
	}

	return nil
}

// LoadFile parses the given YAML file into a Config.
func (w *watcher) loadFile(conf Config, filename string) error {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	switch path.Ext(filename) {
	case "yaml", "yml":
		err = w.parseYAML(conf, content)
	case "json":
		err = w.parseJSON(conf, content)
	}

	if err != nil {
		return fmt.Errorf("parsing YAML file %s: %v", filename, err)
	}

	return nil
}

// parseYAML parses the YAML input into a Config.
func (w *watcher) parseYAML(conf Config, bytes []byte) error {
	err := yaml.Unmarshal([]byte(bytes), conf)

	if err != nil {
		return err
	}

	return nil
}

// parseJSON parses the JSON input into a Config.
func (w *watcher) parseJSON(conf Config, bytes []byte) error {
	err := json.Unmarshal([]byte(bytes), conf)

	if err != nil {
		return err
	}

	return nil
}
