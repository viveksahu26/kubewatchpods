package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

// ConfigFileName stores file of config
var ConfigFileName = ".kubewatch.yaml"

// Handler contains handler configuration
type Handler struct{}

// Resource contains resource configuration
type Resource struct {
	Pod bool `json:"po"`
}

// Config struct contains kubewatch configuration
type Config struct {
	// Handlers know how to send notifications to specific services.
	Handler Handler `json:"handler"`

	// Reason   []string `json:"reason"`

	// Resources to watch.
	Resource Resource `json:"resource"`

	// For watching specific namespace, leave it empty for watching all.
	// this config is ignored when watching namespaces
	Namespace string `json:"namespace,omitempty"`
}

// New creates new config object
func New() (*Config, error) {
	c := &Config{}
	if err := c.Load(); err != nil {
		return c, err
	}

	return c, nil
}

func createIfNotExist() error {
	// create file if not exist
	configFile := filepath.Join(configDir(), ConfigFileName)
	_, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(configFile)
			if err != nil {
				return err
			}
			file.Close()
		} else {
			return err
		}
	}
	return nil
}

// Load loads configuration from config file
func (c *Config) Load() error {
	err := createIfNotExist()
	if err != nil {
		return err
	}

	file, err := os.Open(getConfigFile())
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if len(b) != 0 {
		return yaml.Unmarshal(b, c)
	}

	return nil
}

// CheckMissingResourceEnvvars will read the environment for equivalent config variables to set
func (c *Config) CheckMissingResourceEnvvars() {
	if !c.Resource.Pod && os.Getenv("KW_POD") == "true" {
		c.Resource.Pod = true
	}
}

func (c *Config) Write() error {
	f, err := os.OpenFile(getConfigFile(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	// compat with old versions of kubewatch
	return enc.Encode(c)
}

func getConfigFile() string {
	configFile := filepath.Join(configDir(), ConfigFileName)
	if _, err := os.Stat(configFile); err == nil {
		return configFile
	}

	return ""
}

func configDir() string {
	if configDir := os.Getenv("KW_CONFIG"); configDir != "" {
		return configDir
	}

	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		return home
	}
	return os.Getenv("HOME")
}
