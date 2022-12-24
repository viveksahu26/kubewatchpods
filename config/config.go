package config

// Resource contains resource configuration
type Resource struct {
	Pod bool `json:"po"`
}

// Config struct contains kubewatch configuration
type Config struct {
	// Handlers know how to send notifications to specific services.
	// Handler Handler `json:"handler"`

	// Reason   []string `json:"reason"`

	// Resources to watch.
	Resource Resource `json:"resource"`

	// For watching specific namespace, leave it empty for watching all.
	// this config is ignored when watching namespaces
	Namespace string `json:"namespace,omitempty"`
}
