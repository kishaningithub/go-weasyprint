// This package defines the types needed to handle the various CSS properties.
// There are 3 groups of types for a property, separated by 2 steps : cascading and computation.
// Thus the need of 3 types (see below).
// Schematically, the style computation is :
//		ValidatedProperty (ComputedFromCascaded)-> CascadedPropery (Compute)-> CssProperty
package properties

import "github.com/benoitkugler/go-weasyprint/style/parser"

// CssProperty is final form of a css input :
// "var()", "attr()" and custom properties have been resolved.
type CssProperty interface {
	isCssProperty()
}

// CascadedProperty may contain either a classic CSS property
// or one the 3 special values var(), attr() or custom properties.
// "initial" and "inherited" values have been resolved
type CascadedProperty struct {
	prop            CssProperty
	SpecialProperty specialProperty
}

// AsCss will panic if c.SpecialProperty is not nil.
func (c CascadedProperty) AsCss() CssProperty {
	if c.SpecialProperty != nil {
		panic("attempted to bypass the SpecialProperty of a CascadedProperty")
	}
	return c.prop
}

func (c CascadedProperty) IsNone() bool {
	return c.prop == nil && c.SpecialProperty == nil
}

// ValidatedProperty is valid css input, so it may contain
// a classic property, a special one, or one of the keyword "inherited" or "initial".
type ValidatedProperty struct {
	prop    CascadedProperty
	Default DefaultKind
}

func (v ValidatedProperty) IsNone() bool {
	return v.prop.IsNone() && v.Default == 0
}

// AsCascaded will panic if c.Default is not zero.
func (c ValidatedProperty) AsCascaded() CascadedProperty {
	if c.Default != 0 {
		panic("attempted to bypass the Default of a ValidatedProperty")
	}
	return c.prop
}

type specialProperty interface {
	isSpecialProperty()
}

type DefaultKind uint8

const (
	Inherit DefaultKind = iota + 1
	Initial
)

func (d DefaultKind) ToV() ValidatedProperty {
	return ValidatedProperty{Default: d}
}

type VarData struct {
	Name        string // name of a custom property
	Declaration CustomProperty
}

func (v VarData) IsNone() bool {
	return v.Name == "" && v.Declaration == nil
}

func (v VarData) isSpecialProperty()        {}
func (v CustomProperty) isSpecialProperty() {}

// AttrData is actually only supported inside other properties,
// and for anchor.
func (v AttrData) isSpecialProperty() {}

// ---------- Convenience constructor -------------------------------
// Note than a CssProperty can naturally be seen as a CascadedProperty, but not the other way around.

func ToC(prop CssProperty) CascadedProperty {
	return CascadedProperty{prop: prop}
}

func ToC2(spe specialProperty) CascadedProperty {
	return CascadedProperty{SpecialProperty: spe}
}

func (c CascadedProperty) ToV() ValidatedProperty {
	return ValidatedProperty{prop: c}
}

// Properties is the general container for validated, cascaded and computed properties.
// In addition to the generic acces, an attempt to provide a "type safe" way is provided through the
// GetXXX and SetXXX methods. It relies on the convention than all the keys should be present,
// and values never be nil.
// "None" values are then encoded by the zero value of the concrete type.
type Properties map[string]CssProperty

func (p Properties) Keys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	return keys
}

// Copy return a shallow copy.
func (p Properties) Copy() Properties {
	out := make(Properties, len(p))
	for name, v := range p {
		out[name] = v
	}
	return out
}

// ResolveColor return the color for `key`, replacing
// `currentColor` with p["color"]
// replace Python getColor function
func (p Properties) ResolveColor(key string) Color {
	value := p[key].(Color)
	if value.Type == parser.ColorCurrentColor {
		return p.GetColor()
	}
	return value
}

// UpdateWith merge the entries from `other` to `p`.
func (p Properties) UpdateWith(other Properties) {
	for k, v := range other {
		p[k] = v
	}
}
