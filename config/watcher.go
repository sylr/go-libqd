package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	toml "github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	flags "github.com/jessevdk/go-flags"
	json "github.com/tailscale/hujson"
	yaml "gopkg.in/yaml.v3"
)

type watcher struct {
	manager *Manager
	logger  Logger
	name    interface{}
	config  Config
}

// reload configuration
func (w *watcher) reload() {
	newConfig := w.config.DeepCopyConfig()

	// Load config from cli args and then from config file if exists
	err := w.loadConfig(newConfig)

	if err != nil {
		w.logger.Errorf("Error while loading conf: %v", err)
		return
	}

	// Execute validators
	errs := w.manager.runValidators(w.name, w.config, newConfig)

	if len(errs) > 0 {
		for _, err := range errs {
			w.logger.Errorf("Error while validating new conf: %v", err)
		}

		err = errors.New("New configuration not applied because error(s) have been found")
		w.logger.Errorf("%v", err)
		return
	}

	// Execute appliers
	err = w.manager.runAppliers(w.name, w.config, newConfig)

	if err != nil {
		w.logger.Errorf("Error while applying new conf: %v", err)
	}

	// update current configuration
	w.config = newConfig
	w.manager.broadcastNewConfig(w.name)
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
	configFile := w.config.ConfigFile()

	w.logger.Debugf("Watching config files %s", configFile)

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		w.logger.Fatalf("%x", err)
	}

	err = watcher.Add(configFile)

	if err != nil {
		w.logger.Fatalf("%x", err)
	}

	if len(os.Getenv("KUBERNETES_PORT")) > 0 {
		dir := filepath.Dir(configFile)
		w.logger.Infof("In kubernetes context, adding %s to watch list", dir)
		err := watcher.Add(dir)

		if err != nil {
			w.logger.Fatalf("%x", err)
		}
	}

	defer watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				w.logger.Errorf("fsnotify: error %x", event)
				break
			}

			w.logger.Debugf("fsnotify: %s -> %s", event.Name, event.Op.String())

			if event.Op&fsnotify.Write == fsnotify.Write {
				if event.Name == configFile {
					w.logger.Debugf("Config file changed")
				}
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				if event.Name == configFile {
					w.logger.Debugf("Config file created")
				} else if filepath.Base(event.Name) == "..data" {
					w.logger.Debugf("Configmap volume updated")
				} else {
					break
				}
			} else {
				break
			}

			w.logger.Infof("Reloading config")

			// Reload configuration
			w.reload()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			w.logger.Errorf("fsnotify: %s", err)
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
			w.logger.Fatalf("%v", err)
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

	ext := path.Ext(filename)

	switch ext {
	case ".yaml", ".yml":
		err = w.parseYAML(conf, content)
	case ".json":
		err = w.parseJSON(conf, content)
	case ".toml":
		err = w.parseTOML(conf, content)
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

// parseTOML parses the TOML input into a Config.
func (w *watcher) parseTOML(conf Config, bytes []byte) error {
	err := toml.Unmarshal([]byte(bytes), conf)

	if err != nil {
		return err
	}

	return nil
}
