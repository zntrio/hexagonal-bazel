package actor

import (
	"github.com/mcuadros/go-defaults"
)

// Configuration represents the server configuration state.
type Configuration struct {
	Server struct {
		Listener struct {
			Network string `json:"network" toml:"network" default:"tcp" comment:"The network type to listen to (tcp, tcp4, tcp6, unix)."`
			Address string `json:"address" toml:"address" default:"localhost:3001" comment:"The listener address according to the network type selected."`
		} `comment:"# Listener"`
	} `comment:"#################### Server ####################"`
	Persistence struct {
		Type   string `json:"type" toml:"type" default:"badger" comment:"The persistence implementation to use (badger)."`
		Badger struct {
			Dir      string `json:"dir" toml:"dir" default:"data" comment:"The directory where the data are stored."`
			ValueDir string `json:"value_dir" toml:"value_dir" default:"data" comment:"The directory to store."`
		}
	} `comment:"#################### Persistence ####################"`
	Broker struct {
		Type string `json:"type" toml:"type" default:"noop" comment:"The broker implementation to use (noop,console,jetstream)."`
	} `comment:"#################### Broker ####################"`
	Generator struct {
		WorkerID        uint8 `json:"worker_id" toml:"worker_id" default:"1" comment:"The worker identifier. Must be different for each instance."`
		SecretWordCount int   `json:"secret_word_count" toml:"secret_word_count" default:"6" comment:"The secret word count for passphrase protected links."`
	} `comment:"#################### Generator ####################"`
}

// -----------------------------------------------------------------------------

func DefaultConfiguration() *Configuration {
	cfg := &Configuration{}
	defaults.SetDefaults(cfg)
	return cfg
}
