package validate

// WithValidatorRegistry creates a server option that allows setting a custom validator registry.
func WithValidatorRegistry(registry *Registry) any {
	return &validatorRegistryOption{registry}
}

type validatorRegistryOption struct {
	registry *Registry
}

// Apply the validator registry to the global registry.
func (v *validatorRegistryOption) Apply() {
	if v.registry != nil {
		defaultRegistry = v.registry
	}
}

// SetGlobalRegistry sets the global validator registry.
func SetGlobalRegistry(registry *Registry) {
	if registry != nil {
		defaultRegistry = registry
	}
}
