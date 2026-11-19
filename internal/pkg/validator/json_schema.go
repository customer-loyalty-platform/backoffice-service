package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

const EarningRuleMetaSchema = `{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"type": "object",
	"required": ["and"],
	"properties": {
		"and": {
			"type": "array",
			"items": {
				"type": "object",
				"required": ["field", "operator", "value"],
				"properties": {
					"field": { "type": "string" },
					"operator": { "type": "string", "enum": ["EQUALS", "NOT_EQUALS", "GREATER_THAN", "LESS_THAN", "GREATER_THAN_OR_EQUAL", "LESS_THAN_OR_EQUAL", "BETWEEN", "IN", "CONTAINS"] },
					"value": {}
				}
			}
		}
	}
}`

type JSONSchemaValidator struct {
	compiledSchema *jsonschema.Schema
}

func NewJSONSchemaValidator() (*JSONSchemaValidator, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("rule_schema.json", strings.NewReader(EarningRuleMetaSchema)); err != nil {
		return nil, err
	}
	schema, err := compiler.Compile("rule_schema.json")
	if err != nil {
		return nil, err
	}
	return &JSONSchemaValidator{compiledSchema: schema}, nil
}

func (v *JSONSchemaValidator) ValidateJSONB(jsonbData []byte) error {
	var vInterface interface{}
	if err := json.Unmarshal(jsonbData, &vInterface); err != nil {
		return fmt.Errorf("invalid json format: %w", err)
	}
	if err := v.compiledSchema.Validate(vInterface); err != nil {
		return fmt.Errorf("json rule schema validation failed: %w", err)
	}
	return nil
}

