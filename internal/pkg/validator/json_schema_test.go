package validator_test

import (
	"testing"

	"backoffice-service/internal/pkg/validator"
)

func TestJSONSchemaValidator_ValidConditions(t *testing.T) {
	v, err := validator.NewJSONSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to initialize validator: %v", err)
	}

	validJSON := []byte(`{
		"and": [
			{"field": "transaction.category", "operator": "EQUALS", "value": "F&B"},
			{"field": "transaction.amount", "operator": "GREATER_THAN", "value": 10000}
		]
	}`)

	if err := v.ValidateJSONB(validJSON); err != nil {
		t.Errorf("Expected valid JSON to pass validation, got: %v", err)
	}
}

func TestJSONSchemaValidator_InvalidOperator(t *testing.T) {
	v, err := validator.NewJSONSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to initialize validator: %v", err)
	}

	invalidJSON := []byte(`{
		"and": [
			{"field": "transaction.category", "operator": "INVALID_OP", "value": "F&B"}
		]
	}`)

	if err := v.ValidateJSONB(invalidJSON); err == nil {
		t.Errorf("Expected invalid operator to fail validation, but it passed")
	}
}

func TestJSONSchemaValidator_MissingRequiredField(t *testing.T) {
	v, err := validator.NewJSONSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to initialize validator: %v", err)
	}

	missingFieldJSON := []byte(`{
		"and": [
			{"operator": "EQUALS", "value": "F&B"}
		]
	}`)

	if err := v.ValidateJSONB(missingFieldJSON); err == nil {
		t.Errorf("Expected missing 'field' to fail validation, but it passed")
	}
}

