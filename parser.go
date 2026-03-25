package jsonforms

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Static errors for err113 compliance
var (
	ErrMissingTypeField              = errors.New("missing or invalid 'type' field")
	ErrControlMissingScope           = errors.New("Control missing required 'scope' field")
	ErrGroupMissingLabel             = errors.New("Group missing required 'label' field")
	ErrCategorizationMissingElements = errors.New("Categorization missing required 'elements' field")
	ErrElementNotObject              = errors.New("element is not an object")
	ErrCategoryMissingLabel          = errors.New("Category missing required 'label' field")
	ErrLabelMissingText              = errors.New("Label missing required 'text' field")
	ErrMissingElements               = errors.New("missing or invalid 'elements' field")
	ErrRuleMissingEffect             = errors.New("Rule missing required 'effect' field")
	ErrRuleMissingCondition          = errors.New("Rule missing required 'condition' field")
	ErrUnknownConditionType          = errors.New("unknown condition type")
	ErrSchemaConditionMissingScope   = errors.New("SchemaBasedCondition missing required 'scope' field")
	ErrSchemaConditionMissingSchema  = errors.New("SchemaBasedCondition missing required 'schema' field")
	ErrLeafConditionMissingScope     = errors.New("LeafCondition missing required 'scope' field")
	ErrLeafConditionMissingValue     = errors.New("LeafCondition missing required 'expectedValue' field")
	ErrAndConditionMissingConditions = errors.New("AndCondition missing required 'conditions' field")
	ErrOrConditionMissingConditions  = errors.New("OrCondition missing required 'conditions' field")
)

// Parse parses JSON Forms UI schema and data schema into an AST
func Parse(uiSchemaJSON, schemaJSON []byte) (*AST, error) {
	// Parse UI Schema
	uiSchema, err := parseUISchema(uiSchemaJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UI schema: %w", err)
	}

	// Parse Data Schema (stored as raw any)
	var schema any
	if len(schemaJSON) > 0 {
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			return nil, fmt.Errorf("failed to parse data schema: %w", err)
		}
	}

	ast := &AST{
		UISchema: uiSchema,
		Schema:   schema,
	}

	// Post-parse: resolve schema properties onto controls
	if schema != nil {
		resolveSchemaProperties(ast)
	}

	return ast, nil
}

// resolveSchemaProperties walks the AST and resolves schema property definitions onto Control nodes
func resolveSchemaProperties(ast *AST) {
	_ = Walk(ast.UISchema, &schemaResolver{schema: ast.Schema})
}

type schemaResolver struct {
	BaseVisitor
	schema any
}

func (r *schemaResolver) VisitControl(c *Control) error {
	c.SchemaProperty = resolveScope(c.Scope, r.schema)
	return nil
}

// resolveScope walks the JSON Schema following a JSON Pointer scope path
// and returns the property definition at that path.
func resolveScope(scope string, schema any) *SchemaProperty {
	if schema == nil {
		return nil
	}

	// Strip leading "#/" or "#"
	path := strings.TrimPrefix(scope, "#/")

	path = strings.TrimPrefix(path, "#")

	if path == "" {
		return nil
	}

	segments := strings.Split(path, "/")

	current, ok := schema.(map[string]any)
	if !ok {
		return nil
	}

	// Track the parent node at each "properties" level for required checking
	var parent map[string]any

	var propertyName string

	for i, segment := range segments {
		val, exists := current[segment]
		if !exists {
			return nil
		}

		next, ok := val.(map[string]any)
		if !ok {
			return nil
		}

		// Track parent: when we see "properties", the next segment is a property name
		if segment == "properties" && i+1 < len(segments) {
			parent = current
		} else if i > 0 && segments[i-1] == "properties" {
			propertyName = segment
		}

		current = next
	}

	return buildSchemaProperty(current, parent, propertyName)
}

func buildSchemaProperty(node, parent map[string]any, propertyName string) *SchemaProperty {
	sp := &SchemaProperty{}

	if t, ok := node["type"].(string); ok {
		sp.Type = t
	}

	if f, ok := node["format"].(string); ok {
		sp.Format = f
	}

	if p, ok := node["pattern"].(string); ok {
		sp.Pattern = p
	}

	if e, ok := node["enum"].([]any); ok {
		sp.Enum = e
	}

	if c, exists := node["const"]; exists {
		sp.Const = c
	}

	if d, exists := node["default"]; exists {
		sp.Default = d
	}

	if v, ok := node["minLength"].(float64); ok {
		i := int(v)
		sp.MinLength = &i
	}

	if v, ok := node["maxLength"].(float64); ok {
		i := int(v)
		sp.MaxLength = &i
	}

	if v, ok := node["minimum"].(float64); ok {
		sp.Minimum = &v
	}

	if v, ok := node["maximum"].(float64); ok {
		sp.Maximum = &v
	}

	// Check if this property is in parent's "required" array
	if parent != nil && propertyName != "" {
		if req, ok := parent["required"].([]any); ok {
			for _, r := range req {
				if s, ok := r.(string); ok && s == propertyName {
					sp.Required = true
					break
				}
			}
		}
	}

	return sp
}

// parseUISchema parses the UI schema JSON into a UISchemaElement
func parseUISchema(data []byte) (UISchemaElement, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return parseUISchemaElement(raw)
}

// parseUISchemaElement recursively parses a UI schema element
func parseUISchemaElement(data map[string]any) (UISchemaElement, error) {
	elementType, ok := data["type"].(string)
	if !ok {
		return nil, ErrMissingTypeField
	}

	// Parse common base fields
	base, err := parseBaseElement(data)
	if err != nil {
		return nil, err
	}

	// Parse specific element types
	switch elementType {
	case "Control":
		return parseControl(data, base)
	case "VerticalLayout":
		return parseVerticalLayout(data, base)
	case "HorizontalLayout":
		return parseHorizontalLayout(data, base)
	case "Group":
		return parseGroup(data, base)
	case "Categorization":
		return parseCategorization(data, base)
	case "Category":
		return parseCategory(data, base)
	case "Label":
		return parseLabel(data, base)
	default:
		// Create a CustomElement for unknown element types
		return parseCustomElement(data, base), nil
	}
}

// parseBaseElement parses common fields shared by all UI schema elements
func parseBaseElement(data map[string]any) (BaseUISchemaElement, error) {
	base := BaseUISchemaElement{
		Type: data["type"].(string),
	}

	// Parse optional rule
	if ruleData, ok := data["rule"].(map[string]any); ok {
		rule, err := parseRule(ruleData)
		if err != nil {
			return base, fmt.Errorf("failed to parse rule: %w", err)
		}

		base.Rule = rule
	}

	// Parse optional options
	if options, ok := data["options"].(map[string]any); ok {
		base.Options = options
	}

	// Parse optional i18n
	if i18n, ok := data["i18n"].(string); ok {
		base.I18n = &i18n
	}

	return base, nil
}

// parseControl parses a Control element
func parseControl(data map[string]any, base BaseUISchemaElement) (*Control, error) {
	scope, ok := data["scope"].(string)
	if !ok {
		return nil, ErrControlMissingScope
	}

	control := &Control{
		BaseUISchemaElement: base,
		Scope:               scope,
	}

	if label, ok := data["label"]; ok {
		control.Label = label
	}

	return control, nil
}

// parseVerticalLayout parses a VerticalLayout element
func parseVerticalLayout(data map[string]any, base BaseUISchemaElement) (*VerticalLayout, error) {
	elements, err := parseElementsArray(data)
	if err != nil {
		return nil, err
	}

	return &VerticalLayout{
		BaseUISchemaElement: base,
		Elements:            elements,
	}, nil
}

// parseHorizontalLayout parses a HorizontalLayout element
func parseHorizontalLayout(data map[string]any, base BaseUISchemaElement) (*HorizontalLayout, error) {
	elements, err := parseElementsArray(data)
	if err != nil {
		return nil, err
	}

	return &HorizontalLayout{
		BaseUISchemaElement: base,
		Elements:            elements,
	}, nil
}

// parseGroup parses a Group element
func parseGroup(data map[string]any, base BaseUISchemaElement) (*Group, error) {
	label, ok := data["label"].(string)
	if !ok {
		return nil, ErrGroupMissingLabel
	}

	elements, err := parseElementsArray(data)
	if err != nil {
		return nil, err
	}

	return &Group{
		BaseUISchemaElement: base,
		Label:               label,
		Elements:            elements,
	}, nil
}

// parseCategorization parses a Categorization element
func parseCategorization(data map[string]any, base BaseUISchemaElement) (*Categorization, error) {
	elementsData, ok := data["elements"].([]any)
	if !ok {
		return nil, ErrCategorizationMissingElements
	}

	var elements []CategoryElement

	for i, elemData := range elementsData {
		elemMap, ok := elemData.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("element %d: %w", i, ErrElementNotObject)
		}

		elem, err := parseUISchemaElement(elemMap)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}

		// Ensure element is a Category or Categorization (skip custom elements in categorizations)
		categoryElem, ok := elem.(CategoryElement)
		if !ok {
			// Skip non-category elements (like CustomElement)
			continue
		}

		elements = append(elements, categoryElem)
	}

	categorization := &Categorization{
		BaseUISchemaElement: base,
		Elements:            elements,
	}

	if label, ok := data["label"].(string); ok {
		categorization.Label = &label
	}

	return categorization, nil
}

// parseCategory parses a Category element
func parseCategory(data map[string]any, base BaseUISchemaElement) (*Category, error) {
	label, ok := data["label"].(string)
	if !ok {
		return nil, ErrCategoryMissingLabel
	}

	elements, err := parseElementsArray(data)
	if err != nil {
		return nil, err
	}

	return &Category{
		BaseUISchemaElement: base,
		Label:               label,
		Elements:            elements,
	}, nil
}

// parseLabel parses a Label element
func parseLabel(data map[string]any, base BaseUISchemaElement) (*Label, error) {
	text, ok := data["text"].(string)
	if !ok {
		return nil, ErrLabelMissingText
	}

	return &Label{
		BaseUISchemaElement: base,
		Text:                text,
	}, nil
}

// parseCustomElement parses an unknown/custom element type
func parseCustomElement(data map[string]any, base BaseUISchemaElement) *CustomElement {
	custom := &CustomElement{
		BaseUISchemaElement: base,
		RawData:             data,
	}

	// Try to parse child elements if they exist
	if _, hasElements := data["elements"]; hasElements {
		elements, err := parseElementsArray(data)
		if err == nil {
			custom.Elements = elements
		}
		// If parsing fails, we still preserve the custom element with raw data
	}

	return custom
}

// parseElementsArray parses the 'elements' array common to many layout types
func parseElementsArray(data map[string]any) ([]UISchemaElement, error) {
	elementsData, ok := data["elements"].([]any)
	if !ok {
		return nil, ErrMissingElements
	}

	var elements []UISchemaElement

	for i, elemData := range elementsData {
		elemMap, ok := elemData.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("element %d: %w", i, ErrElementNotObject)
		}

		elem, err := parseUISchemaElement(elemMap)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}

		elements = append(elements, elem)
	}

	return elements, nil
}

// parseRule parses a Rule object
func parseRule(data map[string]any) (*Rule, error) {
	effect, ok := data["effect"].(string)
	if !ok {
		return nil, ErrRuleMissingEffect
	}

	conditionData, ok := data["condition"].(map[string]any)
	if !ok {
		return nil, ErrRuleMissingCondition
	}

	condition, err := parseCondition(conditionData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse condition: %w", err)
	}

	return &Rule{
		Effect:    RuleEffect(effect),
		Condition: condition,
	}, nil
}

// parseCondition parses a Condition object
func parseCondition(data map[string]any) (Condition, error) {
	conditionType, _ := data["type"].(string)

	// Determine condition type
	switch conditionType {
	case "LEAF":
		return parseLeafCondition(data)
	case "AND":
		return parseAndCondition(data)
	case "OR":
		return parseOrCondition(data)
	case "SCHEMA_BASED", "":
		// Default to SCHEMA_BASED if type is not specified
		return parseSchemaBasedCondition(data)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownConditionType, conditionType)
	}
}

// parseSchemaBasedCondition parses a SchemaBasedCondition
func parseSchemaBasedCondition(data map[string]any) (*SchemaBasedCondition, error) {
	scope, ok := data["scope"].(string)
	if !ok {
		return nil, ErrSchemaConditionMissingScope
	}

	schema, ok := data["schema"]
	if !ok {
		return nil, ErrSchemaConditionMissingSchema
	}

	condition := &SchemaBasedCondition{
		Scope:  scope,
		Schema: schema,
	}

	if condType, ok := data["type"].(string); ok {
		condition.Type = condType
	}

	if failWhenUndefined, ok := data["failWhenUndefined"].(bool); ok {
		condition.FailWhenUndefined = &failWhenUndefined
	}

	return condition, nil
}

// parseLeafCondition parses a LeafCondition
func parseLeafCondition(data map[string]any) (*LeafCondition, error) {
	scope, ok := data["scope"].(string)
	if !ok {
		return nil, ErrLeafConditionMissingScope
	}

	expectedValue, ok := data["expectedValue"]
	if !ok {
		return nil, ErrLeafConditionMissingValue
	}

	return &LeafCondition{
		Type:          "LEAF",
		Scope:         scope,
		ExpectedValue: expectedValue,
	}, nil
}

// parseAndCondition parses an AndCondition
func parseAndCondition(data map[string]any) (*AndCondition, error) {
	conditionsData, ok := data["conditions"].([]any)
	if !ok {
		return nil, ErrAndConditionMissingConditions
	}

	var conditions []Condition

	for i, condData := range conditionsData {
		condMap, ok := condData.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("condition %d: %w", i, ErrElementNotObject)
		}

		cond, err := parseCondition(condMap)
		if err != nil {
			return nil, fmt.Errorf("condition %d: %w", i, err)
		}

		conditions = append(conditions, cond)
	}

	return &AndCondition{
		Type:       "AND",
		Conditions: conditions,
	}, nil
}

// parseOrCondition parses an OrCondition
func parseOrCondition(data map[string]any) (*OrCondition, error) {
	conditionsData, ok := data["conditions"].([]any)
	if !ok {
		return nil, ErrOrConditionMissingConditions
	}

	var conditions []Condition

	for i, condData := range conditionsData {
		condMap, ok := condData.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("condition %d: %w", i, ErrElementNotObject)
		}

		cond, err := parseCondition(condMap)
		if err != nil {
			return nil, fmt.Errorf("condition %d: %w", i, err)
		}

		conditions = append(conditions, cond)
	}

	return &OrCondition{
		Type:       "OR",
		Conditions: conditions,
	}, nil
}
