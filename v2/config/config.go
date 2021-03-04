// Package config provides configuration for client apps based on the Cryptowatch SDK.
package config // import "code.cryptowat.ch/cw-sdk-go/config"

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
)

// Default URLs used if one of the URLs isn't specified.
const (
	DefaultRESTURL   = "https://api.cryptowat.ch"
	DefaultStreamURL = "wss://stream.cryptowat.ch"
	DefaultTradeURL  = "wss://trading.service.cryptowat.ch"

	Filepath = ".cw/credentials.yml"
)

// Various validation errors.
var (
	ErrNilConfig      = Error{Type: "config", Why: "config is nil", How: "create and load config first"}
	ErrEmptyAPIKey    = Error{Type: "config", What: "api_url", Why: "is empty", How: "specify an api_key"}
	ErrEmptySecretKey = Error{Type: "config", What: "secret_key", Why: "is empty", How: "specify a secret_key"}
	ErrInvalidHTTPURL = Error{Type: "config", Why: "wrong url", How: "URL must be a valid http or https url"}
	ErrInvalidWSURL   = Error{Type: "config", Why: "wrong url", How: "URL must be a valid ws or wss url"}
	ErrInvalidScheme  = Error{Type: "config", Why: "invalid scheme", How: "scheme must be http(s) or ws(s)"}

	ErrNilArgs = Error{Type: "args", Why: "args is nil", How: "create an instance of args"}
)

// CWConfig holds the configuration.
type CWConfig struct {
	mu        sync.Mutex `yaml:"-"` // protects the fields below
	APIKey    string     `yaml:"api_key"`
	SecretKey string     `yaml:"secret_key"`
	RESTURL   string     `yaml:"rest_url"`
	StreamURL string     `yaml:"stream_url"`
	TradeURL  string     `yaml:"trade_url"`
}

func Get() *CWConfig {
	cfg := &CWConfig{}

	defaultPath, dfErr := DefaultFilepath()
	if dfErr == nil {
		cfgFile, _ := NewFromPath(defaultPath)
		if cfgFile != nil {
			cfg = cfgFile
		}
	}

	cfg.setDefaults()

	return cfg
}

// NewFromPath creates a new CW from a file by the given path.
func NewFromPath(path string) (*CWConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cfg := &CWConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, errors.Trace(err)
	}

	cfg.setDefaults()

	return cfg, nil
}

func (c *CWConfig) setDefaults() {
	if c.RESTURL == "" {
		c.RESTURL = DefaultRESTURL
	}

	if c.StreamURL == "" {
		c.StreamURL = DefaultStreamURL
	}

	if c.TradeURL == "" {
		c.TradeURL = DefaultTradeURL
	}
}

// String can't be defined on a value receiver here because of the mutex.
func (c *CWConfig) String() string {
	raw, err := yaml.Marshal(c)
	if err != nil {
		return err.Error()
	}

	return string(raw)
}

// DefaultFilepath determines and returns default config path.
// It can return an error if detecting the user's home directory has failed.
func DefaultFilepath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Trace(err)
	}

	return filepath.Join(home, Filepath), nil
}

// Error holds detials about an error occured during validation process.
type Error struct {
	Type string
	What string
	Why  string
	How  string
}

func (e Error) Error() string {
	if e.What == "" {
		return fmt.Sprintf("invalid %s: %s. Possible fix: %s", e.Type, e.Why, e.How)
	}

	return fmt.Sprintf("invalid %s: %s - %s. Possible fix: %s", e.Type, e.What, e.Why, e.How)
}

// CheckURL checks that the url has the correct scheme.
func CheckURL(given string, schemes ...string) error {
	u, err := url.Parse(given)
	if err != nil {
		return errors.Trace(err)
	}

	for _, s := range schemes {
		if u.Scheme == s {
			return nil
		}
	}

	return ErrInvalidScheme
}
