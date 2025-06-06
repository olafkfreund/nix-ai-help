package functionbase

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"nix-ai-help/pkg/logger"
)

// FunctionInterface defines the interface that all AI functions must implement
type FunctionInterface interface {
	Name() string
	Description() string
	Schema() FunctionSchema
	Execute(ctx context.Context, params map[string]interface{}, options *FunctionOptions) (*FunctionResult, error)
	ValidateParameters(params map[string]interface{}) error
}

// BaseFunction provides common functionality for all AI functions
type BaseFunction struct {
	name        string
	description string
	schema      FunctionSchema
	logger      *logger.Logger
}

// NewBaseFunction creates a new base function
func NewBaseFunction(name, description string, parameters []FunctionParameter) *BaseFunction {
	return &BaseFunction{
		name:        name,
		description: description,
		schema: FunctionSchema{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
		logger: logger.NewLogger(),
	}
}

// Name returns the function name
func (bf *BaseFunction) Name() string {
	return bf.name
}

// Description returns the function description
func (bf *BaseFunction) Description() string {
	return bf.description
}

// Schema returns the function schema
func (bf *BaseFunction) Schema() FunctionSchema {
	return bf.schema
}

// SetSchema sets the function schema (for functions that need to modify their schema)
func (bf *BaseFunction) SetSchema(schema FunctionSchema) {
	bf.schema = schema
}

// ValidateParameters validates the input parameters against the schema
func (bf *BaseFunction) ValidateParameters(params map[string]interface{}) error {
	for _, param := range bf.schema.Parameters {
		value, exists := params[param.Name]

		// Check if required parameter is missing
		if param.Required && !exists {
			return ValidationError{
				Parameter: param.Name,
				Message:   fmt.Sprintf("required parameter '%s' is missing", param.Name),
			}
		}

		// Skip validation if parameter is not provided and not required
		if !exists {
			continue
		}

		// Validate parameter type and constraints
		if err := bf.validateParameterValue(param, value); err != nil {
			return err
		}
	}

	return nil
}

// validateParameterValue validates a single parameter value
func (bf *BaseFunction) validateParameterValue(param FunctionParameter, value interface{}) error {
	// Type validation
	if err := bf.validateType(param, value); err != nil {
		return err
	}

	// Enum validation
	if len(param.Enum) > 0 {
		if err := bf.validateEnum(param, value); err != nil {
			return err
		}
	}

	// Pattern validation for strings
	if param.Pattern != "" && param.Type == "string" {
		if err := bf.validatePattern(param, value); err != nil {
			return err
		}
	}

	// Length validation for strings
	if param.Type == "string" {
		if err := bf.validateStringLength(param, value); err != nil {
			return err
		}
	}

	// Numeric range validation
	if param.Type == "number" || param.Type == "integer" {
		if err := bf.validateNumericRange(param, value); err != nil {
			return err
		}
	}

	return nil
}

// validateType validates the parameter type
func (bf *BaseFunction) validateType(param FunctionParameter, value interface{}) error {
	switch param.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return ValidationError{
				Parameter: param.Name,
				Message:   fmt.Sprintf("parameter '%s' must be a string", param.Name),
				Value:     value,
			}
		}
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// Valid numeric types
		default:
			return ValidationError{
				Parameter: param.Name,
				Message:   fmt.Sprintf("parameter '%s' must be a number", param.Name),
				Value:     value,
			}
		}
	case "integer":
		switch value.(type) {
		case int, int32, int64:
			// Valid integer types
		default:
			return ValidationError{
				Parameter: param.Name,
				Message:   fmt.Sprintf("parameter '%s' must be an integer", param.Name),
				Value:     value,
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return ValidationError{
				Parameter: param.Name,
				Message:   fmt.Sprintf("parameter '%s' must be a boolean", param.Name),
				Value:     value,
			}
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return ValidationError{
				Parameter: param.Name,
				Message:   fmt.Sprintf("parameter '%s' must be an object", param.Name),
				Value:     value,
			}
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return ValidationError{
				Parameter: param.Name,
				Message:   fmt.Sprintf("parameter '%s' must be an array", param.Name),
				Value:     value,
			}
		}
	}
	return nil
}

// validateEnum validates enum constraints
func (bf *BaseFunction) validateEnum(param FunctionParameter, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	for _, enumValue := range param.Enum {
		if valueStr == enumValue {
			return nil
		}
	}
	return ValidationError{
		Parameter: param.Name,
		Message:   fmt.Sprintf("parameter '%s' must be one of: %v", param.Name, param.Enum),
		Value:     value,
	}
}

// validatePattern validates regex pattern constraints
func (bf *BaseFunction) validatePattern(param FunctionParameter, value interface{}) error {
	valueStr, ok := value.(string)
	if !ok {
		return nil // Type validation should catch this
	}

	matched, err := regexp.MatchString(param.Pattern, valueStr)
	if err != nil {
		return ValidationError{
			Parameter: param.Name,
			Message:   fmt.Sprintf("parameter '%s' has invalid pattern: %v", param.Name, err),
			Value:     value,
		}
	}

	if !matched {
		return ValidationError{
			Parameter: param.Name,
			Message:   fmt.Sprintf("parameter '%s' does not match required pattern: %s", param.Name, param.Pattern),
			Value:     value,
		}
	}

	return nil
}

// validateStringLength validates string length constraints
func (bf *BaseFunction) validateStringLength(param FunctionParameter, value interface{}) error {
	valueStr, ok := value.(string)
	if !ok {
		return nil // Type validation should catch this
	}

	length := len(valueStr)

	if param.MinLength != nil && length < *param.MinLength {
		return ValidationError{
			Parameter: param.Name,
			Message:   fmt.Sprintf("parameter '%s' must be at least %d characters long", param.Name, *param.MinLength),
			Value:     value,
		}
	}

	if param.MaxLength != nil && length > *param.MaxLength {
		return ValidationError{
			Parameter: param.Name,
			Message:   fmt.Sprintf("parameter '%s' must be at most %d characters long", param.Name, *param.MaxLength),
			Value:     value,
		}
	}

	return nil
}

// validateNumericRange validates numeric range constraints
func (bf *BaseFunction) validateNumericRange(param FunctionParameter, value interface{}) error {
	var numValue float64

	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int32:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	default:
		return nil // Type validation should catch this
	}

	if param.Minimum != nil && numValue < *param.Minimum {
		return ValidationError{
			Parameter: param.Name,
			Message:   fmt.Sprintf("parameter '%s' must be at least %f", param.Name, *param.Minimum),
			Value:     value,
		}
	}

	if param.Maximum != nil && numValue > *param.Maximum {
		return ValidationError{
			Parameter: param.Name,
			Message:   fmt.Sprintf("parameter '%s' must be at most %f", param.Name, *param.Maximum),
			Value:     value,
		}
	}

	return nil
}

// Helper functions for creating common parameter types

// StringParam creates a string parameter
func StringParam(name, description string, required bool) FunctionParameter {
	return FunctionParameter{
		Name:        name,
		Type:        "string",
		Description: description,
		Required:    required,
	}
}

// StringParamWithEnum creates a string parameter with enum values
func StringParamWithEnum(name, description string, required bool, enum []string) FunctionParameter {
	return FunctionParameter{
		Name:        name,
		Type:        "string",
		Description: description,
		Required:    required,
		Enum:        enum,
	}
}

// StringParamWithOptions creates a string parameter with additional options (alias for StringParamWithEnum)
func StringParamWithOptions(name, description string, required bool, enum []string, minLen, maxLen *int) FunctionParameter {
	return FunctionParameter{
		Name:        name,
		Type:        "string",
		Description: description,
		Required:    required,
		Enum:        enum,
		MinLength:   minLen,
		MaxLength:   maxLen,
	}
}

// BoolParam creates a boolean parameter
func BoolParam(name, description string, required bool, defaultValue ...bool) FunctionParameter {
	param := FunctionParameter{
		Name:        name,
		Type:        "boolean",
		Description: description,
		Required:    required,
	}
	if len(defaultValue) > 0 {
		param.Default = defaultValue[0]
	}
	return param
}

// ObjectParam creates an object parameter
func ObjectParam(name, description string, required bool) FunctionParameter {
	return FunctionParameter{
		Name:        name,
		Type:        "object",
		Description: description,
		Required:    required,
	}
}

// Helper functions for creating function results

// CreateSuccessResult creates a successful function result with optional message
func CreateSuccessResult(data interface{}, message string) *FunctionResult {
	result := &FunctionResult{
		Success:   true,
		Data:      data,
		Duration:  0, // Will be set by the caller if needed
		Timestamp: time.Now(),
	}
	if message != "" {
		// Store message in metadata since FunctionResult doesn't have a Message field
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["message"] = message
	}
	return result
}

// CreateErrorResult creates an error function result with optional message
func CreateErrorResult(err error, message string) *FunctionResult {
	result := &FunctionResult{
		Success:   false,
		Error:     err.Error(),
		Duration:  0, // Will be set by the caller if needed
		Timestamp: time.Now(),
	}
	if message != "" {
		// Store message in metadata since FunctionResult doesn't have a Message field
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["message"] = message
	}
	return result
}

// SuccessResult creates a successful function result
func SuccessResult(data interface{}, duration time.Duration) *FunctionResult {
	return &FunctionResult{
		Success:   true,
		Data:      data,
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// ErrorResult creates an error function result
func ErrorResult(err error, duration time.Duration) *FunctionResult {
	return &FunctionResult{
		Success:   false,
		Error:     err.Error(),
		Duration:  duration,
		Timestamp: time.Now(),
	}
}
