module github.com/sylr/go-libqd/config

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/jessevdk/go-flags v1.4.0
	github.com/tailscale/hujson v0.0.0-20190930033718-5098e564d9b3
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200506231410-2ff61e1afc86
)

replace (
	github.com/niemeyer/pretty => github.com/sylr/go-pretty v0.0.0-20200517092739-d3252c08b3ba
	gopkg.in/yaml.v3 => github.com/sylr/go-yaml v0.0.0-20200517101938-4dbbae02f875
)
