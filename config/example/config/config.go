package config

//go:generate deepcopy-gen --input-dirs . --output-package . --output-file-base config_deepcopy --go-header-file /dev/null
//+k8s:deepcopy-gen=true
//+k8s:deepcopy-gen:interfaces=github.com/sylr/go-libqd/config.Config

// MyAppConfiguration implements github.com/sylr/go-libqd/config.Config
type MyAppConfiguration struct {
	Reloads  int32  `yaml:"-"`
	File     string `                           short:"f" long:"config"`
	Verbose  []bool `yaml:"verbose"             short:"v" long:"verbose"`
	Version  bool   `                                     long:"version"`
	HTTPPort int    `yaml:"port"    json:"port" short:"p" long:"port"`
}

func (c *MyAppConfiguration) ConfigFile() string {
	return c.File
}
