package jsonforms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseControl(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/name"
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok, "Expected Control, got %T", result.UISchema)

	assert.Equal(t, "#/properties/name", control.Scope)
}

func TestParseControlWithLabel(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/email",
		"label": "Email Address"
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok, "Expected Control, got %T", result.UISchema)

	assert.Equal(t, "Email Address", control.Label)
}

func TestParseVerticalLayout(t *testing.T) {
	uiSchema := []byte(`{
		"type": "VerticalLayout",
		"elements": [
			{
				"type": "Control",
				"scope": "#/properties/name"
			},
			{
				"type": "Control",
				"scope": "#/properties/email"
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	layout, ok := result.UISchema.(*VerticalLayout)
	require.True(t, ok, "Expected VerticalLayout, got %T", result.UISchema)

	assert.Len(t, layout.Elements, 2)

	// Check first control
	control1, ok := layout.Elements[0].(*Control)
	require.True(t, ok, "Expected first element to be Control, got %T", layout.Elements[0])

	assert.Equal(t, "#/properties/name", control1.Scope)

	// Check second control
	control2, ok := layout.Elements[1].(*Control)
	require.True(t, ok, "Expected second element to be Control, got %T", layout.Elements[1])

	assert.Equal(t, "#/properties/email", control2.Scope)
}

func TestParseHorizontalLayout(t *testing.T) {
	uiSchema := []byte(`{
		"type": "HorizontalLayout",
		"elements": [
			{
				"type": "Control",
				"scope": "#/properties/firstName"
			},
			{
				"type": "Control",
				"scope": "#/properties/lastName"
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	layout, ok := result.UISchema.(*HorizontalLayout)
	require.True(t, ok, "Expected HorizontalLayout, got %T", result.UISchema)

	assert.Len(t, layout.Elements, 2)
}

func TestParseGroup(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Group",
		"label": "Personal Info",
		"elements": [
			{
				"type": "Control",
				"scope": "#/properties/name"
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	group, ok := result.UISchema.(*Group)
	require.True(t, ok, "Expected Group, got %T", result.UISchema)

	assert.Equal(t, "Personal Info", group.Label)
	assert.Len(t, group.Elements, 1)
}

func TestParseLabel(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Label",
		"text": "Welcome to the form"
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	label, ok := result.UISchema.(*Label)
	require.True(t, ok, "Expected Label, got %T", result.UISchema)

	assert.Equal(t, "Welcome to the form", label.Text)
}

func TestParseCategorization(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Categorization",
		"elements": [
			{
				"type": "Category",
				"label": "Basic",
				"elements": [
					{
						"type": "Control",
						"scope": "#/properties/name"
					}
				]
			},
			{
				"type": "Category",
				"label": "Advanced",
				"elements": [
					{
						"type": "Control",
						"scope": "#/properties/age"
					}
				]
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	categorization, ok := result.UISchema.(*Categorization)
	require.True(t, ok, "Expected Categorization, got %T", result.UISchema)

	assert.Len(t, categorization.Elements, 2)

	// Check first category
	category1, ok := categorization.Elements[0].(*Category)
	require.True(t, ok, "Expected first element to be Category, got %T", categorization.Elements[0])

	assert.Equal(t, "Basic", category1.Label)
}

func TestParseRuleWithSchemaBasedCondition(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/email",
		"rule": {
			"effect": "SHOW",
			"condition": {
				"scope": "#/properties/subscribe",
				"schema": {
					"const": true
				}
			}
		}
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok, "Expected Control, got %T", result.UISchema)

	require.NotNil(t, control.Rule, "Expected rule to be present")

	assert.Equal(t, RuleEffectSHOW, control.Rule.Effect)

	condition, ok := control.Rule.Condition.(*SchemaBasedCondition)
	require.True(t, ok, "Expected SchemaBasedCondition, got %T", control.Rule.Condition)

	assert.Equal(t, "#/properties/subscribe", condition.Scope)
}

func TestParseRuleWithLeafCondition(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/phone",
		"rule": {
			"effect": "ENABLE",
			"condition": {
				"type": "LEAF",
				"scope": "#/properties/contactMethod",
				"expectedValue": "phone"
			}
		}
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok, "Expected Control, got %T", result.UISchema)

	require.NotNil(t, control.Rule, "Expected rule to be present")

	condition, ok := control.Rule.Condition.(*LeafCondition)
	require.True(t, ok, "Expected LeafCondition, got %T", control.Rule.Condition)

	assert.Equal(t, "phone", condition.ExpectedValue)
}

func TestParseRuleWithAndCondition(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/newsletter",
		"rule": {
			"effect": "SHOW",
			"condition": {
				"type": "AND",
				"conditions": [
					{
						"type": "LEAF",
						"scope": "#/properties/subscribe",
						"expectedValue": true
					},
					{
						"type": "LEAF",
						"scope": "#/properties/email",
						"expectedValue": ""
					}
				]
			}
		}
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok, "Expected Control, got %T", result.UISchema)

	require.NotNil(t, control.Rule, "Expected rule to be present")

	andCondition, ok := control.Rule.Condition.(*AndCondition)
	require.True(t, ok, "Expected AndCondition, got %T", control.Rule.Condition)

	assert.Len(t, andCondition.Conditions, 2)
}

func TestParseWithOptions(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/description",
		"options": {
			"multi": true,
			"trim": true
		}
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok, "Expected Control, got %T", result.UISchema)

	require.NotNil(t, control.Options, "Expected options to be present")

	assert.Equal(t, true, control.Options["multi"])
	assert.Equal(t, true, control.Options["trim"])
}

func TestParseWithI18n(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/name",
		"i18n": "person.name"
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok, "Expected Control, got %T", result.UISchema)

	require.NotNil(t, control.I18n, "Expected i18n to be present")

	assert.Equal(t, "person.name", *control.I18n)
}

func TestParseSchema(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/name"
	}`)

	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string"
			}
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	require.NotNil(t, result.Schema, "Expected data schema to be present")

	// Check if it's a valid map structure
	schemaMap, ok := result.Schema.(map[string]any)
	require.True(t, ok, "Expected data schema to be a map, got %T", result.Schema)

	assert.Equal(t, "object", schemaMap["type"])
}

func TestParseInvalidJSON(t *testing.T) {
	uiSchema := []byte(`{invalid json}`)

	_, err := Parse(uiSchema, nil)
	assert.Error(t, err, "Expected error for invalid JSON")
}

func TestParseMissingType(t *testing.T) {
	uiSchema := []byte(`{
		"scope": "#/properties/name"
	}`)

	_, err := Parse(uiSchema, nil)
	assert.Error(t, err, "Expected error for missing type field")
}

func TestParseMissingRequiredField(t *testing.T) {
	// Control without scope
	uiSchema := []byte(`{
		"type": "Control"
	}`)

	_, err := Parse(uiSchema, nil)
	assert.Error(t, err, "Expected error for Control without scope")
}

func TestParseCustomElement(t *testing.T) {
	uiSchema := []byte(`{
		"type": "VerticalLayout",
		"elements": [
			{
				"type": "Control",
				"scope": "#/properties/name"
			},
			{
				"type": "Notice",
				"options": {
					"bg": "brand-blue"
				},
				"elements": [
					{
						"type": "Markdown",
						"options": {
							"copy": "This is a custom notice"
						}
					}
				]
			},
			{
				"type": "Control",
				"scope": "#/properties/email"
			}
		]
	}`)

	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"email": {"type": "string"}
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	// Check that we got a VerticalLayout
	layout, ok := result.UISchema.(*VerticalLayout)
	require.True(t, ok, "Expected VerticalLayout, got %T", result.UISchema)

	// Should have 3 elements
	assert.Len(t, layout.Elements, 3)

	// First should be Control
	_, ok = layout.Elements[0].(*Control)
	assert.True(t, ok, "Expected first element to be Control, got %T", layout.Elements[0])

	// Second should be CustomElement
	customElem, ok := layout.Elements[1].(*CustomElement)
	require.True(t, ok, "Expected second element to be CustomElement, got %T", layout.Elements[1])

	// Check custom element type
	assert.Equal(t, "Notice", customElem.GetType())

	// Check custom element has child elements
	assert.Len(t, customElem.Elements, 1)

	// Child should also be a CustomElement (Markdown)
	childCustom, ok := customElem.Elements[0].(*CustomElement)
	if assert.True(t, ok, "Expected child to be CustomElement, got %T", customElem.Elements[0]) {
		assert.Equal(t, "Markdown", childCustom.GetType())
	}

	// Third should be Control
	_, ok = layout.Elements[2].(*Control)
	assert.True(t, ok, "Expected third element to be Control, got %T", layout.Elements[2])
}

func TestParseNestedLayouts(t *testing.T) {
	uiSchema := []byte(`{
		"type": "VerticalLayout",
		"elements": [
			{
				"type": "Group",
				"label": "Section 1",
				"elements": [
					{
						"type": "HorizontalLayout",
						"elements": [
							{
								"type": "Control",
								"scope": "#/properties/field1"
							},
							{
								"type": "Control",
								"scope": "#/properties/field2"
							}
						]
					}
				]
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	layout, ok := result.UISchema.(*VerticalLayout)
	require.True(t, ok, "Expected VerticalLayout, got %T", result.UISchema)

	assert.Len(t, layout.Elements, 1)

	group, ok := layout.Elements[0].(*Group)
	require.True(t, ok, "Expected Group, got %T", layout.Elements[0])

	assert.Len(t, group.Elements, 1)

	horizontalLayout, ok := group.Elements[0].(*HorizontalLayout)
	require.True(t, ok, "Expected HorizontalLayout, got %T", group.Elements[0])

	assert.Len(t, horizontalLayout.Elements, 2)
}

func TestParseCategory(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Category",
		"label": "Personal Information",
		"elements": [
			{
				"type": "Control",
				"scope": "#/properties/firstName"
			},
			{
				"type": "Control",
				"scope": "#/properties/lastName"
			},
			{
				"type": "Label",
				"text": "Please provide your full name"
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	category, ok := result.UISchema.(*Category)
	require.True(t, ok, "Expected Category, got %T", result.UISchema)

	assert.Equal(t, "Personal Information", category.Label)
	assert.Len(t, category.Elements, 3)

	// Check first element is Control
	control1, ok := category.Elements[0].(*Control)
	if assert.True(t, ok, "Expected first element to be Control, got %T", category.Elements[0]) {
		assert.Equal(t, "#/properties/firstName", control1.Scope)
	}

	// Check second element is Control
	control2, ok := category.Elements[1].(*Control)
	if assert.True(t, ok, "Expected second element to be Control, got %T", category.Elements[1]) {
		assert.Equal(t, "#/properties/lastName", control2.Scope)
	}

	// Check third element is Label
	label, ok := category.Elements[2].(*Label)
	if assert.True(t, ok, "Expected third element to be Label, got %T", category.Elements[2]) {
		assert.Equal(t, "Please provide your full name", label.Text)
	}
}

func TestParseNestedCategorization(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Categorization",
		"elements": [
			{
				"type": "Category",
				"label": "Main",
				"elements": [
					{
						"type": "Control",
						"scope": "#/properties/mainField"
					}
				]
			},
			{
				"type": "Categorization",
				"label": "Nested Tabs",
				"elements": [
					{
						"type": "Category",
						"label": "Sub Tab 1",
						"elements": [
							{
								"type": "Control",
								"scope": "#/properties/subField1"
							}
						]
					},
					{
						"type": "Category",
						"label": "Sub Tab 2",
						"elements": [
							{
								"type": "Control",
								"scope": "#/properties/subField2"
							}
						]
					}
				]
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	categorization, ok := result.UISchema.(*Categorization)
	require.True(t, ok, "Expected Categorization, got %T", result.UISchema)

	assert.Len(t, categorization.Elements, 2)

	// First element should be a Category
	category, ok := categorization.Elements[0].(*Category)
	require.True(t, ok, "Expected first element to be Category, got %T", categorization.Elements[0])

	assert.Equal(t, "Main", category.Label)

	// Second element should be a nested Categorization
	nestedCat, ok := categorization.Elements[1].(*Categorization)
	require.True(t, ok, "Expected second element to be Categorization, got %T", categorization.Elements[1])

	label := "Nested Tabs"
	if assert.NotNil(t, nestedCat.Label, "Expected nested categorization label to be present") {
		assert.Equal(t, label, *nestedCat.Label)
	}

	assert.Len(t, nestedCat.Elements, 2)

	// Check nested categories
	subCat1, ok := nestedCat.Elements[0].(*Category)
	require.True(t, ok, "Expected first nested element to be Category, got %T", nestedCat.Elements[0])

	assert.Equal(t, "Sub Tab 1", subCat1.Label)

	subCat2, ok := nestedCat.Elements[1].(*Category)
	require.True(t, ok, "Expected second nested element to be Category, got %T", nestedCat.Elements[1])

	assert.Equal(t, "Sub Tab 2", subCat2.Label)

	// Verify the controls in the nested categories
	assert.Len(t, subCat1.Elements, 1)

	if control, ok := subCat1.Elements[0].(*Control); ok {
		assert.Equal(t, "#/properties/subField1", control.Scope)
	}

	assert.Len(t, subCat2.Elements, 1)

	if control, ok := subCat2.Elements[0].(*Control); ok {
		assert.Equal(t, "#/properties/subField2", control.Scope)
	}
}

func TestControlSchemaPropertyResolved(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/name"
	}`)
	schema := []byte(`{
		"type": "object",
		"required": ["name"],
		"properties": {
			"name": {
				"type": "string",
				"minLength": 2,
				"maxLength": 50
			}
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	control, ok := result.UISchema.(*Control)
	require.True(t, ok)
	require.NotNil(t, control.SchemaProperty, "Expected SchemaProperty to be resolved")

	assert.Equal(t, "string", control.SchemaProperty.Type)
	assert.Equal(t, 2, *control.SchemaProperty.MinLength)
	assert.Equal(t, 50, *control.SchemaProperty.MaxLength)
	assert.True(t, control.SchemaProperty.Required)
}

func TestControlSchemaPropertyEnum(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/title"
	}`)
	schema := []byte(`{
		"type": "object",
		"properties": {
			"title": {
				"type": "string",
				"enum": ["Mr", "Mrs", "Miss", "Ms"]
			}
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	control := result.UISchema.(*Control)
	require.NotNil(t, control.SchemaProperty)

	assert.Equal(t, "string", control.SchemaProperty.Type)
	assert.Equal(t, []any{"Mr", "Mrs", "Miss", "Ms"}, control.SchemaProperty.Enum)
	assert.False(t, control.SchemaProperty.Required)
}

func TestControlSchemaPropertyBoolean(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/acceptTerms"
	}`)
	schema := []byte(`{
		"type": "object",
		"properties": {
			"acceptTerms": {
				"type": "boolean",
				"default": false,
				"const": true
			}
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	control := result.UISchema.(*Control)
	require.NotNil(t, control.SchemaProperty)

	assert.Equal(t, "boolean", control.SchemaProperty.Type)
	assert.Equal(t, false, control.SchemaProperty.Default)
	assert.Equal(t, true, control.SchemaProperty.Const)
}

func TestControlSchemaPropertyNestedScope(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/personalDetails/properties/address/properties/postcode"
	}`)
	schema := []byte(`{
		"type": "object",
		"properties": {
			"personalDetails": {
				"type": "object",
				"properties": {
					"address": {
						"type": "object",
						"required": ["postcode"],
						"properties": {
							"postcode": {
								"type": "string",
								"pattern": "^[A-Z]{1,2}[0-9][0-9A-Z]?\\s?[0-9][A-Z]{2}$"
							}
						}
					}
				}
			}
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	control := result.UISchema.(*Control)
	require.NotNil(t, control.SchemaProperty)

	assert.Equal(t, "string", control.SchemaProperty.Type)
	assert.Equal(t, `^[A-Z]{1,2}[0-9][0-9A-Z]?\s?[0-9][A-Z]{2}$`, control.SchemaProperty.Pattern)
	assert.True(t, control.SchemaProperty.Required)
}

func TestControlSchemaPropertyNilSchema(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/name"
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	control := result.UISchema.(*Control)
	assert.Nil(t, control.SchemaProperty)
}

func TestControlSchemaPropertyUnresolvableScope(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/nonexistent/properties/field"
	}`)
	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": { "type": "string" }
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	control := result.UISchema.(*Control)
	assert.Nil(t, control.SchemaProperty)
}

func TestControlSchemaPropertyWithFormat(t *testing.T) {
	uiSchema := []byte(`{
		"type": "Control",
		"scope": "#/properties/email"
	}`)
	schema := []byte(`{
		"type": "object",
		"properties": {
			"email": {
				"type": "string",
				"format": "email"
			}
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	control := result.UISchema.(*Control)
	require.NotNil(t, control.SchemaProperty)

	assert.Equal(t, "string", control.SchemaProperty.Type)
	assert.Equal(t, "email", control.SchemaProperty.Format)
}

func TestLayoutWithMultipleControlsSchemaResolved(t *testing.T) {
	uiSchema := []byte(`{
		"type": "VerticalLayout",
		"elements": [
			{ "type": "Control", "scope": "#/properties/name" },
			{ "type": "Control", "scope": "#/properties/age" }
		]
	}`)
	schema := []byte(`{
		"type": "object",
		"required": ["name"],
		"properties": {
			"name": { "type": "string" },
			"age": { "type": "integer", "minimum": 0, "maximum": 150 }
		}
	}`)

	result, err := Parse(uiSchema, schema)
	require.NoError(t, err)

	layout := result.UISchema.(*VerticalLayout)
	require.Len(t, layout.Elements, 2)

	nameCtrl := layout.Elements[0].(*Control)
	require.NotNil(t, nameCtrl.SchemaProperty)
	assert.Equal(t, "string", nameCtrl.SchemaProperty.Type)
	assert.True(t, nameCtrl.SchemaProperty.Required)

	ageCtrl := layout.Elements[1].(*Control)
	require.NotNil(t, ageCtrl.SchemaProperty)
	assert.Equal(t, "integer", ageCtrl.SchemaProperty.Type)
	assert.InDelta(t, float64(0), *ageCtrl.SchemaProperty.Minimum, 0)
	assert.InDelta(t, float64(150), *ageCtrl.SchemaProperty.Maximum, 0)
	assert.False(t, ageCtrl.SchemaProperty.Required)
}

// countingVisitor counts all element types encountered during a walk
type countingVisitor struct {
	BaseVisitor
	ControlCount          int
	VerticalLayoutCount   int
	HorizontalLayoutCount int
	GroupCount            int
	CategorizationCount   int
	CategoryCount         int
	LabelCount            int
	CustomElementCount    int
}

func (v *countingVisitor) VisitControl(c *Control) error {
	v.ControlCount++
	return nil
}

func (v *countingVisitor) VisitVerticalLayout(vl *VerticalLayout) error {
	v.VerticalLayoutCount++
	return nil
}

func (v *countingVisitor) VisitHorizontalLayout(hl *HorizontalLayout) error {
	v.HorizontalLayoutCount++
	return nil
}

func (v *countingVisitor) VisitGroup(g *Group) error {
	v.GroupCount++
	return nil
}

func (v *countingVisitor) VisitCategorization(c *Categorization) error {
	v.CategorizationCount++
	return nil
}

func (v *countingVisitor) VisitCategory(c *Category) error {
	v.CategoryCount++
	return nil
}

func (v *countingVisitor) VisitLabel(l *Label) error {
	v.LabelCount++
	return nil
}

func (v *countingVisitor) VisitCustomElement(ce *CustomElement) error {
	v.CustomElementCount++
	return nil
}

func TestVisitorAllElements(t *testing.T) {
	// Parse complex schema with all element types and walk with visitor
	uiSchema := []byte(`{
		"type": "VerticalLayout",
		"elements": [
			{
				"type": "Control",
				"scope": "#/properties/field1"
			},
			{
				"type": "HorizontalLayout",
				"elements": [
					{
						"type": "Control",
						"scope": "#/properties/field2"
					},
					{
						"type": "Control",
						"scope": "#/properties/field3"
					}
				]
			},
			{
				"type": "Group",
				"label": "Personal Info",
				"elements": [
					{
						"type": "Label",
						"text": "Enter your details"
					},
					{
						"type": "Control",
						"scope": "#/properties/firstName"
					}
				]
			},
			{
				"type": "Categorization",
				"elements": [
					{
						"type": "Category",
						"label": "Basic",
						"elements": [
							{
								"type": "Control",
								"scope": "#/properties/basic1"
							}
						]
					},
					{
						"type": "Category",
						"label": "Advanced",
						"elements": [
							{
								"type": "Control",
								"scope": "#/properties/advanced1"
							}
						]
					}
				]
			},
			{
				"type": "Notice",
				"elements": [
					{
						"type": "Markdown",
						"text": "**Important**"
					}
				]
			}
		]
	}`)

	result, err := Parse(uiSchema, nil)
	require.NoError(t, err)

	visitor := &countingVisitor{}

	err = Walk(result.UISchema, visitor)
	require.NoError(t, err)

	// Verify all element types were visited
	assert.Equal(t, 1, visitor.VerticalLayoutCount)
	assert.Equal(t, 1, visitor.HorizontalLayoutCount)
	assert.Equal(t, 1, visitor.GroupCount)
	assert.Equal(t, 1, visitor.CategorizationCount)
	assert.Equal(t, 2, visitor.CategoryCount)
	assert.Equal(t, 6, visitor.ControlCount)
	assert.Equal(t, 1, visitor.LabelCount)
	assert.Equal(t, 2, visitor.CustomElementCount)
}
