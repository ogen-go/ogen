package integration

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-faster/errors"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

// CEL validator for testing - demonstrates custom validator implementation.
type CEL struct {
	Expression string
	Program    cel.Program
	Set        bool
}

// SetExpression sets CEL expression and compiles it.
func (c *CEL) SetExpression(expr string) error {
	if expr == "" {
		c.Set = false
		return nil
	}

	env, err := cel.NewEnv(
		cel.Variable("value", cel.DynType),
	)
	if err != nil {
		return errors.Wrap(err, "create CEL environment")
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return errors.Wrap(issues.Err(), "compile CEL expression")
	}

	program, err := env.Program(ast)
	if err != nil {
		return errors.Wrap(err, "create CEL program")
	}

	c.Expression = expr
	c.Program = program
	c.Set = true
	return nil
}

// Validate returns error if value does not match CEL expression.
func (c CEL) Validate(value any) error {
	if !c.Set {
		return nil
	}

	// Convert value to CEL-compatible type
	celValue := convertToCELValue(value)

	result, _, err := c.Program.Eval(map[string]any{
		"value": celValue,
	})
	if err != nil {
		return errors.Wrap(err, "evaluate CEL expression")
	}

	if result.Type() != types.BoolType {
		return errors.Errorf("CEL expression must return boolean, got %s", result.Type())
	}

	if result.Value().(bool) {
		return nil
	}

	return &CELValidationError{
		Expression: c.Expression,
		Value:      value,
	}
}

// convertToCELValue converts Go values to CEL-compatible values
func convertToCELValue(value any) any {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return convertToCELValue(v.Elem().Interface())
	case reflect.Interface:
		return convertToCELValue(v.Elem().Interface())
	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = convertToCELValue(v.Index(i).Interface())
		}
		return result
	case reflect.Map:
		result := make(map[string]any)
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			result[keyStr] = convertToCELValue(v.MapIndex(key).Interface())
		}
		return result
	case reflect.Struct:
		// Convert struct to map for CEL
		result := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				fieldValue := v.Field(i)
				if fieldValue.CanInterface() {
					// Use JSON tag name if available, otherwise use field name
					fieldName := field.Name
					if jsonTag := field.Tag.Get("json"); jsonTag != "" {
						// Handle json:"name,omitempty" format by splitting on comma
						parts := strings.Split(jsonTag, ",")
						if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
							fieldName = parts[0]
						} else if parts[0] == "-" {
							continue // Skip fields with json:"-"
						}
					}
					result[fieldName] = convertToCELValue(fieldValue.Interface())
				}
			}
		}
		return result
	default:
		return value
	}
}

// CELValidationError reports that CEL expression validation failed.
type CELValidationError struct {
	Expression string
	Value      any
}

// Error implements error.
func (e *CELValidationError) Error() string {
	return fmt.Sprintf("CEL validation failed: %s", e.Expression)
}
