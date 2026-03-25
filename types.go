package jsonforms

// AST represents the complete parsed structure of a JSON Forms definition
type AST struct {
	UISchema UISchemaElement `json:"uischema"`
	Schema   any             `json:"schema"` // Raw JSON Schema
}

// UISchemaElement is the base interface for all UI schema elements
type UISchemaElement interface {
	GetType() string
	GetRule() *Rule
	GetOptions() map[string]any
	GetI18n() *string
}

// BaseUISchemaElement contains common fields shared by all UI schema elements
type BaseUISchemaElement struct {
	Type    string         `json:"type"`
	Rule    *Rule          `json:"rule,omitempty"`
	Options map[string]any `json:"options,omitempty"`
	I18n    *string        `json:"i18n,omitempty"`
}

// GetType returns the type of the UI schema element
func (b *BaseUISchemaElement) GetType() string {
	return b.Type
}

// GetRule returns the rule associated with this element
func (b *BaseUISchemaElement) GetRule() *Rule {
	return b.Rule
}

// GetOptions returns the options map for this element
func (b *BaseUISchemaElement) GetOptions() map[string]any {
	return b.Options
}

// GetI18n returns the internationalization key for this element
func (b *BaseUISchemaElement) GetI18n() *string {
	return b.I18n
}

// SchemaProperty represents a resolved JSON Schema property definition.
// Fields are nil/zero when not present in the schema.
type SchemaProperty struct {
	Type       string                    `json:"type,omitempty"`
	Format     string                    `json:"format,omitempty"`
	Enum       []any                     `json:"enum,omitempty"`
	Const      any                       `json:"const,omitempty"`
	Default    any                       `json:"default,omitempty"`
	Pattern    string                    `json:"pattern,omitempty"`
	MinLength  *int                      `json:"minLength,omitempty"`
	MaxLength  *int                      `json:"maxLength,omitempty"`
	Minimum    *float64                  `json:"minimum,omitempty"`
	Maximum    *float64                  `json:"maximum,omitempty"`
	Required   bool                      `json:"-"`                    // true if this property appears in parent's "required" array
	Properties map[string]*SchemaProperty `json:"properties,omitempty"` // for type: "object"
}

// Control binds a UI input to a specific data property
type Control struct {
	BaseUISchemaElement
	Scope          string          `json:"scope"`
	Label          any             `json:"label,omitempty"`  // Can be string, bool, or LabelDescription
	SchemaProperty *SchemaProperty `json:"-"`                // Resolved from Scope against the data schema
}

// LabelDescription provides detailed label configuration
type LabelDescription struct {
	Text string `json:"text"`
	Show *bool  `json:"show,omitempty"`
}

// VerticalLayout stacks UI elements vertically
type VerticalLayout struct {
	BaseUISchemaElement
	Elements []UISchemaElement `json:"elements"`
}

// HorizontalLayout arranges UI elements side-by-side
type HorizontalLayout struct {
	BaseUISchemaElement
	Elements []UISchemaElement `json:"elements"`
}

// Group is a vertical layout with a descriptive label
type Group struct {
	BaseUISchemaElement
	Label    string            `json:"label"`
	Elements []UISchemaElement `json:"elements"`
}

// Categorization provides tab-like organization of related sections
type Categorization struct {
	BaseUISchemaElement
	Label    *string           `json:"label,omitempty"`
	Elements []CategoryElement `json:"elements"` // Can contain Category or nested Categorization
}

// CategoryElement is a marker interface for elements that can be in a Categorization
type CategoryElement interface {
	UISchemaElement
	IsCategoryElement()
}

// Category represents an individual tab/category within a Categorization
type Category struct {
	BaseUISchemaElement
	Label    string            `json:"label"`
	Elements []UISchemaElement `json:"elements"`
}

// IsCategoryElement marks Category as a valid Categorization child
func (c *Category) IsCategoryElement() {}

// IsCategoryElement marks Categorization as a valid Categorization child (recursive)
func (c *Categorization) IsCategoryElement() {}

// Label displays static text in the form
type Label struct {
	BaseUISchemaElement
	Text string `json:"text"`
}

// CustomElement represents an unknown/custom element type that is not a standard JSON Forms element
type CustomElement struct {
	BaseUISchemaElement
	RawData  map[string]any    `json:"-"`                  // Complete raw element data
	Elements []UISchemaElement `json:"elements,omitempty"` // Child elements (recursively parsed)
}

// Rule defines conditional behavior for UI elements
type Rule struct {
	Effect    RuleEffect `json:"effect"`
	Condition Condition  `json:"condition"`
}

// RuleEffect specifies what happens when a rule's condition is met
type RuleEffect string

const (
	RuleEffectHIDE    RuleEffect = "HIDE"
	RuleEffectSHOW    RuleEffect = "SHOW"
	RuleEffectENABLE  RuleEffect = "ENABLE"
	RuleEffectDISABLE RuleEffect = "DISABLE"
)

// Condition is the base interface for all condition types
type Condition interface {
	GetType() string
}

// SchemaBasedCondition validates a scope against a JSON Schema
type SchemaBasedCondition struct {
	Type              string `json:"type,omitempty"` // Optional, defaults to SCHEMA_BASED
	Scope             string `json:"scope"`
	Schema            any    `json:"schema"` // JSON Schema object
	FailWhenUndefined *bool  `json:"failWhenUndefined,omitempty"`
}

// GetType returns the condition type
func (s *SchemaBasedCondition) GetType() string {
	if s.Type != "" {
		return s.Type
	}

	return "SCHEMA_BASED"
}

// LeafCondition performs simple value comparison
type LeafCondition struct {
	Type          string `json:"type"` // "LEAF"
	Scope         string `json:"scope"`
	ExpectedValue any    `json:"expectedValue"`
}

// GetType returns the condition type
func (l *LeafCondition) GetType() string {
	return l.Type
}

// AndCondition combines multiple conditions with AND logic
type AndCondition struct {
	Type       string      `json:"type"` // "AND"
	Conditions []Condition `json:"conditions"`
}

// GetType returns the condition type
func (a *AndCondition) GetType() string {
	return a.Type
}

// OrCondition combines multiple conditions with OR logic
type OrCondition struct {
	Type       string      `json:"type"` // "OR"
	Conditions []Condition `json:"conditions"`
}

// GetType returns the condition type
func (o *OrCondition) GetType() string {
	return o.Type
}
