package validate

import (
	"fmt"
	"sync"

	"github.com/go-faster/errors"
)

// ValidatorFunc is a function that validates a value and returns an error if validation fails.
type ValidatorFunc func(value any, params string) error

// Registry holds custom validators that can be registered and used for validation.
type Registry struct {
	mu         sync.RWMutex
	validators map[string]ValidatorFunc
}

// NewRegistry creates a new validator registry.
func NewRegistry() *Registry {
	return &Registry{
		validators: make(map[string]ValidatorFunc),
	}
}

// RegisterValidator registers a custom validator function with the given name.
func (r *Registry) RegisterValidator(name string, validator ValidatorFunc) error {
	if name == "" {
		return errors.New("validator name cannot be empty")
	}
	if validator == nil {
		return errors.New("validator function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.validators[name] = validator
	return nil
}

// GetValidator returns the validator function for the given name.
func (r *Registry) GetValidator(name string) (ValidatorFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	validator, exists := r.validators[name]
	return validator, exists
}

// Validate validates a value using the specified validator and parameters.
func (r *Registry) Validate(validatorName string, value any, params string) error {
	validator, exists := r.GetValidator(validatorName)
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
	Params        string
	Message       string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed (%s): %s", e.ValidatorName, e.Message)
}

// Default global registry
var defaultRegistry = NewRegistry()

// RegisterValidator registers a validator in the default global registry.
func RegisterValidator(name string, validator ValidatorFunc) error {
	return defaultRegistry.RegisterValidator(name, validator)
}

// GetValidator returns a validator from the default global registry.
func GetValidator(name string) (ValidatorFunc, bool) {
	return defaultRegistry.GetValidator(name)
}

// ValidateWith validates using the default global registry.
func ValidateWith(validatorName string, value any, params string) error {
	return defaultRegistry.Validate(validatorName, value, params)
}
