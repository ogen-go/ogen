package validate

import (
	"fmt"
	"sync"
)

// OgenValidator is a function that performs custom validation.
type OgenValidator func(value any, params any) error

// OgenValidatorRegistry holds custom validators that can be registered and used for validation.
type OgenValidatorRegistry struct {
	mu         sync.RWMutex
	validators map[string]OgenValidator
}

// NewOgenValidatorRegistry creates a new OgenValidatorRegistry.
func NewOgenValidatorRegistry() *OgenValidatorRegistry {
	return &OgenValidatorRegistry{
		validators: make(map[string]OgenValidator),
	}
}

func (r *OgenValidatorRegistry) Register(name string, validator OgenValidator) error {
	if name == "" {
		return fmt.Errorf("validator name cannot be empty")
	}
	if validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	r.mu.Lock()
	r.validators[name] = validator
	r.mu.Unlock()

	return nil
}

func (r *OgenValidatorRegistry) Get(name string) (OgenValidator, bool) {
	r.mu.RLock()
	validator, ok := r.validators[name]
	r.mu.RUnlock()
	return validator, ok
}

// Validate validates a value using the specified validator and parameters.
func (r *OgenValidatorRegistry) Validate(validatorName string, value, params any) error {
	validator, exists := r.Get(validatorName)
	if !exists {
		return &ValidationError{
			ValidatorName: validatorName,
			Value:         value,
			Params:        params,
			Message:       fmt.Sprintf("validator '%s' not found", validatorName),
		}
	}

	if err := validator(value, params); err != nil {
		// Wrap in ValidationError if it's not already one
		if _, ok := err.(*ValidationError); !ok {
			return &ValidationError{
				ValidatorName: validatorName,
				Value:         value,
				Params:        params,
				Message:       err.Error(),
			}
		}
		return err
	}

	return nil
}

// ValidationError represents a validation error from a custom validator.
type ValidationError struct {
	ValidatorName string
	Value         any
	Params        any
	Message       string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed (%s): %s", e.ValidatorName, e.Message)
}

// Default global registry
var defaultRegistry = NewOgenValidatorRegistry()

// RegisterValidator registers a validator in the default global registry.
func RegisterValidator(name string, validator OgenValidator) error {
	return defaultRegistry.Register(name, validator)
}

// GetValidator returns a validator from the default global registry.
func GetValidator(name string) (OgenValidator, bool) {
	return defaultRegistry.Get(name)
}

// Ogen validates using the default global registry.
func Ogen(name string, value, params any) error {
	return defaultRegistry.Validate(name, value, params)
}
