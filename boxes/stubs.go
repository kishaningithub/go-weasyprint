package boxes

import (
	"fmt"

	"github.com/benoitkugler/go-weasyprint/images"
	"github.com/benoitkugler/go-weasyprint/style/tree"
)

// autogenerated from source_box.py

func (TableRowGroupBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func (TableRowBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableRowGroupBox:
		return true
	default:
		return false
	}
}

func (TableColumnGroupBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func (TableColumnBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableColumnGroupBox:
		return true
	default:
		return false
	}
}

func (TableCaptionBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

// A box that has children.
type instanceParentBox interface {
	isParentBox()
}

func IsParentBox(box Box) bool {
	_, is := box.(instanceParentBox)
	return is
}

// A box that participates in an block formatting context.
// An element with a ``display`` value of ``block``, ``list-item`` or
// ``table`` generates a block-level box.
type instanceBlockLevelBox interface {
	isBlockLevelBox()
}

func IsBlockLevelBox(box Box) bool {
	_, is := box.(instanceBlockLevelBox)
	return is
}

// A box that contains only block-level boxes or only line boxes.
// A box that either contains only block-level boxes or establishes an inline
// formatting context and thus contains only line boxes.
// A non-replaced element with a ``display`` value of ``block``,
// ``list-item``, ``inline-block`` or 'table-cell' generates a block container
// box.
type instanceBlockContainerBox interface {
	isBlockContainerBox()
	isParentBox()
}

func IsBlockContainerBox(box Box) bool {
	_, is := box.(instanceBlockContainerBox)
	return is
}

// A block-level box that is also a block container.
// A non-replaced element with a ``display`` value of ``block``, ``list-item``
// generates a block box.
type instanceBlockBox interface {
	isBlockBox()
	isParentBox()
	isBlockContainerBox()
	isBlockLevelBox()
}

func (BlockBox) isBlockBox()          {}
func (BlockBox) isParentBox()         {}
func (BlockBox) isBlockContainerBox() {}
func (BlockBox) isBlockLevelBox()     {}
func (b *BlockBox) Box() *BoxFields   { return &b.BoxFields }

// Copy is a shallow copy
func (b BlockBox) Copy() Box { return &b }

func (b BlockBox) String() string {
	return fmt.Sprintf("<BlockBox %s>", b.BoxFields.elementTag)
}

func BlockBoxAnonymousFrom(parent Box, children []Box) *BlockBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewBlockBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeBlockBox) IsInstance(box Box) bool {
	_, is := box.(instanceBlockBox)
	return is
}

type typeBlockBox struct{}

func (t typeBlockBox) AnonymousFrom(parent Box, children []Box) Box {
	return BlockBoxAnonymousFrom(parent, children)
}

// A box that represents a line in an inline formatting context.
// Can only contain inline-level boxes.
// In early stages of building the box tree a single line box contains many
// consecutive inline boxes. Later, during layout phase, each line boxes will
// be split into multiple line boxes, one for each actual line.
type instanceLineBox interface {
	isLineBox()
	isParentBox()
}

func (LineBox) isLineBox()         {}
func (LineBox) isParentBox()       {}
func (b *LineBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b LineBox) Copy() Box { return &b }

func (b LineBox) String() string {
	return fmt.Sprintf("<LineBox %s>", b.BoxFields.elementTag)
}

func (t typeLineBox) IsInstance(box Box) bool {
	_, is := box.(instanceLineBox)
	return is
}

type typeLineBox struct{}

func (t typeLineBox) AnonymousFrom(parent Box, children []Box) Box {
	return LineBoxAnonymousFrom(parent, children)
}

// A box that participates in an inline formatting context.
// An inline-level box that is not an inline box is said to be "atomic". Such
// boxes are inline blocks, replaced elements and inline tables.
// An element with a ``display`` value of ``inline``, ``inline-table``, or
// ``inline-block`` generates an inline-level box.
type instanceInlineLevelBox interface {
	isInlineLevelBox()
}

func IsInlineLevelBox(box Box) bool {
	_, is := box.(instanceInlineLevelBox)
	return is
}

// An inline box with inline children.
// A box that participates in an inline formatting context and whose content
// also participates in that inline formatting context.
// A non-replaced element with a ``display`` value of ``inline`` generates an
// inline box.
type instanceInlineBox interface {
	isInlineBox()
	isParentBox()
	isInlineLevelBox()
}

func (InlineBox) isInlineBox()       {}
func (InlineBox) isParentBox()       {}
func (InlineBox) isInlineLevelBox()  {}
func (b *InlineBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b InlineBox) Copy() Box { return &b }

func (b InlineBox) String() string {
	return fmt.Sprintf("<InlineBox %s>", b.BoxFields.elementTag)
}

func InlineBoxAnonymousFrom(parent Box, children []Box) *InlineBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewInlineBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeInlineBox) IsInstance(box Box) bool {
	_, is := box.(instanceInlineBox)
	return is
}

type typeInlineBox struct{}

func (t typeInlineBox) AnonymousFrom(parent Box, children []Box) Box {
	return InlineBoxAnonymousFrom(parent, children)
}

// A box that contains only text and has no box children.
// Any text in the document ends up in a text box. What CSS calls "anonymous
// inline boxes" are also text boxes.
type instanceTextBox interface {
	isTextBox()
	isInlineLevelBox()
}

func (TextBox) isTextBox()         {}
func (TextBox) isInlineLevelBox()  {}
func (b *TextBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b TextBox) Copy() Box { return &b }

func (b TextBox) String() string {
	return fmt.Sprintf("<TextBox %s>", b.BoxFields.elementTag)
}

func TextBoxAnonymousFrom(parent Box, text string) *TextBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTextBox(parent.Box().elementTag, style, text)
	return &out

}

func IsTextBox(box Box) bool {
	_, is := box.(instanceTextBox)
	return is
}

// An atomic box in an inline formatting context.
// This inline-level box cannot be split for line breaks.
type instanceAtomicInlineLevelBox interface {
	isAtomicInlineLevelBox()
	isInlineLevelBox()
}

func IsAtomicInlineLevelBox(box Box) bool {
	_, is := box.(instanceAtomicInlineLevelBox)
	return is
}

// A box that is both inline-level and a block container.
// It behaves as inline on the outside and as a block on the inside.
// A non-replaced element with a 'display' value of 'inline-block' generates
// an inline-block box.
type instanceInlineBlockBox interface {
	isInlineBlockBox()
	isAtomicInlineLevelBox()
	isParentBox()
	isBlockContainerBox()
	isInlineLevelBox()
}

func (InlineBlockBox) isInlineBlockBox()       {}
func (InlineBlockBox) isAtomicInlineLevelBox() {}
func (InlineBlockBox) isParentBox()            {}
func (InlineBlockBox) isBlockContainerBox()    {}
func (InlineBlockBox) isInlineLevelBox()       {}
func (b *InlineBlockBox) Box() *BoxFields      { return &b.BoxFields }

// Copy is a shallow copy
func (b InlineBlockBox) Copy() Box { return &b }

func (b InlineBlockBox) String() string {
	return fmt.Sprintf("<InlineBlockBox %s>", b.BoxFields.elementTag)
}

func InlineBlockBoxAnonymousFrom(parent Box, children []Box) *InlineBlockBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewInlineBlockBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeInlineBlockBox) IsInstance(box Box) bool {
	_, is := box.(instanceInlineBlockBox)
	return is
}

type typeInlineBlockBox struct{}

func (t typeInlineBlockBox) AnonymousFrom(parent Box, children []Box) Box {
	return InlineBlockBoxAnonymousFrom(parent, children)
}

// A box whose content is replaced.
// For example, ``<img>`` are replaced: their content is rendered externally
// and is opaque from CSS’s point of view.
type instanceReplacedBox interface {
	isReplacedBox()
}

func (ReplacedBox) isReplacedBox()     {}
func (b *ReplacedBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b ReplacedBox) Copy() Box { return &b }

func (b ReplacedBox) String() string {
	return fmt.Sprintf("<ReplacedBox %s>", b.BoxFields.elementTag)
}

func ReplacedBoxAnonymousFrom(parent Box, replacement images.Image) *ReplacedBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewReplacedBox(parent.Box().elementTag, style, replacement)
	return &out

}

func IsReplacedBox(box Box) bool {
	_, is := box.(instanceReplacedBox)
	return is
}

// A box that is both replaced and block-level.
// A replaced element with a ``display`` value of ``block``, ``liste-item`` or
// ``table`` generates a block-level replaced box.
type instanceBlockReplacedBox interface {
	isBlockReplacedBox()
	isReplacedBox()
	isBlockLevelBox()
}

func (BlockReplacedBox) isBlockReplacedBox() {}
func (BlockReplacedBox) isReplacedBox()      {}
func (BlockReplacedBox) isBlockLevelBox()    {}
func (b *BlockReplacedBox) Box() *BoxFields  { return &b.BoxFields }

// Copy is a shallow copy
func (b BlockReplacedBox) Copy() Box { return &b }

func (b BlockReplacedBox) String() string {
	return fmt.Sprintf("<BlockReplacedBox %s>", b.BoxFields.elementTag)
}

func BlockReplacedBoxAnonymousFrom(parent Box, replacement images.Image) *BlockReplacedBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewBlockReplacedBox(parent.Box().elementTag, style, replacement)
	return &out

}

func IsBlockReplacedBox(box Box) bool {
	_, is := box.(instanceBlockReplacedBox)
	return is
}

// A box that is both replaced and inline-level.
// A replaced element with a ``display`` value of ``inline``,
// ``inline-table``, or ``inline-block`` generates an inline-level replaced
// box.
type instanceInlineReplacedBox interface {
	isInlineReplacedBox()
	isAtomicInlineLevelBox()
	isReplacedBox()
	isInlineLevelBox()
}

func (InlineReplacedBox) isInlineReplacedBox()    {}
func (InlineReplacedBox) isAtomicInlineLevelBox() {}
func (InlineReplacedBox) isReplacedBox()          {}
func (InlineReplacedBox) isInlineLevelBox()       {}
func (b *InlineReplacedBox) Box() *BoxFields      { return &b.BoxFields }

// Copy is a shallow copy
func (b InlineReplacedBox) Copy() Box { return &b }

func (b InlineReplacedBox) String() string {
	return fmt.Sprintf("<InlineReplacedBox %s>", b.BoxFields.elementTag)
}

func InlineReplacedBoxAnonymousFrom(parent Box, replacement images.Image) *InlineReplacedBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewInlineReplacedBox(parent.Box().elementTag, style, replacement)
	return &out

}

func IsInlineReplacedBox(box Box) bool {
	_, is := box.(instanceInlineReplacedBox)
	return is
}

// Box for elements with ``display: table``
type instanceTableBox interface {
	isTableBox()
	isBlockLevelBox()
	isParentBox()
}

func (TableBox) isTableBox()        {}
func (TableBox) isBlockLevelBox()   {}
func (TableBox) isParentBox()       {}
func (b *TableBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b TableBox) Copy() Box { return &b }

func (b TableBox) String() string {
	return fmt.Sprintf("<TableBox %s>", b.BoxFields.elementTag)
}

func TableBoxAnonymousFrom(parent Box, children []Box) *TableBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTableBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeTableBox) IsInstance(box Box) bool {
	_, is := box.(instanceTableBox)
	return is
}

type typeTableBox struct{}

func (t typeTableBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableBoxAnonymousFrom(parent, children)
}

// Box for elements with ``display: inline-table``
type instanceInlineTableBox interface {
	isInlineTableBox()
	isTableBox()
	isBlockLevelBox()
	isParentBox()
}

func (InlineTableBox) isInlineTableBox()  {}
func (InlineTableBox) isTableBox()        {}
func (InlineTableBox) isBlockLevelBox()   {}
func (InlineTableBox) isParentBox()       {}
func (b *InlineTableBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b InlineTableBox) Copy() Box { return &b }

func (b InlineTableBox) String() string {
	return fmt.Sprintf("<InlineTableBox %s>", b.BoxFields.elementTag)
}

func InlineTableBoxAnonymousFrom(parent Box, children []Box) *InlineTableBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewInlineTableBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeInlineTableBox) IsInstance(box Box) bool {
	_, is := box.(instanceInlineTableBox)
	return is
}

type typeInlineTableBox struct{}

func (t typeInlineTableBox) AnonymousFrom(parent Box, children []Box) Box {
	return InlineTableBoxAnonymousFrom(parent, children)
}

// Box for elements with ``display: table-row-group``
type instanceTableRowGroupBox interface {
	isTableRowGroupBox()
	isParentBox()
}

func (TableRowGroupBox) isTableRowGroupBox() {}
func (TableRowGroupBox) isParentBox()        {}
func (b *TableRowGroupBox) Box() *BoxFields  { return &b.BoxFields }

// Copy is a shallow copy
func (b TableRowGroupBox) Copy() Box { return &b }

func (b TableRowGroupBox) String() string {
	return fmt.Sprintf("<TableRowGroupBox %s>", b.BoxFields.elementTag)
}

func TableRowGroupBoxAnonymousFrom(parent Box, children []Box) *TableRowGroupBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTableRowGroupBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeTableRowGroupBox) IsInstance(box Box) bool {
	_, is := box.(instanceTableRowGroupBox)
	return is
}

type typeTableRowGroupBox struct{}

func (t typeTableRowGroupBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableRowGroupBoxAnonymousFrom(parent, children)
}

// Box for elements with ``display: table-row``
type instanceTableRowBox interface {
	isTableRowBox()
	isParentBox()
}

func (TableRowBox) isTableRowBox()     {}
func (TableRowBox) isParentBox()       {}
func (b *TableRowBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b TableRowBox) Copy() Box { return &b }

func (b TableRowBox) String() string {
	return fmt.Sprintf("<TableRowBox %s>", b.BoxFields.elementTag)
}

func TableRowBoxAnonymousFrom(parent Box, children []Box) *TableRowBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTableRowBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeTableRowBox) IsInstance(box Box) bool {
	_, is := box.(instanceTableRowBox)
	return is
}

type typeTableRowBox struct{}

func (t typeTableRowBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableRowBoxAnonymousFrom(parent, children)
}

// Box for elements with ``display: table-column-group``
type instanceTableColumnGroupBox interface {
	isTableColumnGroupBox()
	isParentBox()
}

func (TableColumnGroupBox) isTableColumnGroupBox() {}
func (TableColumnGroupBox) isParentBox()           {}
func (b *TableColumnGroupBox) Box() *BoxFields     { return &b.BoxFields }

// Copy is a shallow copy
func (b TableColumnGroupBox) Copy() Box { return &b }

func (b TableColumnGroupBox) String() string {
	return fmt.Sprintf("<TableColumnGroupBox %s>", b.BoxFields.elementTag)
}

func TableColumnGroupBoxAnonymousFrom(parent Box, children []Box) *TableColumnGroupBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTableColumnGroupBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeTableColumnGroupBox) IsInstance(box Box) bool {
	_, is := box.(instanceTableColumnGroupBox)
	return is
}

type typeTableColumnGroupBox struct{}

func (t typeTableColumnGroupBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableColumnGroupBoxAnonymousFrom(parent, children)
}

// Box for elements with ``display: table-column``
type instanceTableColumnBox interface {
	isTableColumnBox()
	isParentBox()
}

func (TableColumnBox) isTableColumnBox()  {}
func (TableColumnBox) isParentBox()       {}
func (b *TableColumnBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b TableColumnBox) Copy() Box { return &b }

func (b TableColumnBox) String() string {
	return fmt.Sprintf("<TableColumnBox %s>", b.BoxFields.elementTag)
}

func TableColumnBoxAnonymousFrom(parent Box, children []Box) *TableColumnBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTableColumnBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeTableColumnBox) IsInstance(box Box) bool {
	_, is := box.(instanceTableColumnBox)
	return is
}

type typeTableColumnBox struct{}

func (t typeTableColumnBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableColumnBoxAnonymousFrom(parent, children)
}

// Box for elements with ``display: table-cell``
type instanceTableCellBox interface {
	isTableCellBox()
	isBlockContainerBox()
	isParentBox()
}

func (TableCellBox) isTableCellBox()      {}
func (TableCellBox) isBlockContainerBox() {}
func (TableCellBox) isParentBox()         {}
func (b *TableCellBox) Box() *BoxFields   { return &b.BoxFields }

// Copy is a shallow copy
func (b TableCellBox) Copy() Box { return &b }

func (b TableCellBox) String() string {
	return fmt.Sprintf("<TableCellBox %s>", b.BoxFields.elementTag)
}

func TableCellBoxAnonymousFrom(parent Box, children []Box) *TableCellBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTableCellBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeTableCellBox) IsInstance(box Box) bool {
	_, is := box.(instanceTableCellBox)
	return is
}

type typeTableCellBox struct{}

func (t typeTableCellBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableCellBoxAnonymousFrom(parent, children)
}

// Box for elements with ``display: table-caption``
type instanceTableCaptionBox interface {
	isTableCaptionBox()
	isBlockBox()
	isParentBox()
	isBlockContainerBox()
	isBlockLevelBox()
}

func (TableCaptionBox) isTableCaptionBox()   {}
func (TableCaptionBox) isBlockBox()          {}
func (TableCaptionBox) isParentBox()         {}
func (TableCaptionBox) isBlockContainerBox() {}
func (TableCaptionBox) isBlockLevelBox()     {}
func (b *TableCaptionBox) Box() *BoxFields   { return &b.BoxFields }

// Copy is a shallow copy
func (b TableCaptionBox) Copy() Box { return &b }

func (b TableCaptionBox) String() string {
	return fmt.Sprintf("<TableCaptionBox %s>", b.BoxFields.elementTag)
}

func TableCaptionBoxAnonymousFrom(parent Box, children []Box) *TableCaptionBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewTableCaptionBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeTableCaptionBox) IsInstance(box Box) bool {
	_, is := box.(instanceTableCaptionBox)
	return is
}

type typeTableCaptionBox struct{}

func (t typeTableCaptionBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableCaptionBoxAnonymousFrom(parent, children)
}

// Box for a page.
// Initially the whole document will be in the box for the root element.
// During layout a new page box is created after every page break.
type instancePageBox interface {
	isPageBox()
	isParentBox()
}

func (PageBox) isPageBox()         {}
func (PageBox) isParentBox()       {}
func (b *PageBox) Box() *BoxFields { return &b.BoxFields }

// Copy is a shallow copy
func (b PageBox) Copy() Box { return &b }

func IsPageBox(box Box) bool {
	_, is := box.(instancePageBox)
	return is
}

// Box in page margins, as defined in CSS3 Paged Media
type instanceMarginBox interface {
	isMarginBox()
	isBlockContainerBox()
	isParentBox()
}

func (MarginBox) isMarginBox()         {}
func (MarginBox) isBlockContainerBox() {}
func (MarginBox) isParentBox()         {}
func (b *MarginBox) Box() *BoxFields   { return &b.BoxFields }

// Copy is a shallow copy
func (b MarginBox) Copy() Box { return &b }

func IsMarginBox(box Box) bool {
	_, is := box.(instanceMarginBox)
	return is
}

// A box that contains only flex-items.
type instanceFlexContainerBox interface {
	isFlexContainerBox()
	isParentBox()
}

func IsFlexContainerBox(box Box) bool {
	_, is := box.(instanceFlexContainerBox)
	return is
}

// A box that is both block-level and a flex container.
// It behaves as block on the outside and as a flex container on the inside.
type instanceFlexBox interface {
	isFlexBox()
	isParentBox()
	isBlockLevelBox()
	isFlexContainerBox()
}

func (FlexBox) isFlexBox()          {}
func (FlexBox) isParentBox()        {}
func (FlexBox) isBlockLevelBox()    {}
func (FlexBox) isFlexContainerBox() {}
func (b *FlexBox) Box() *BoxFields  { return &b.BoxFields }

// Copy is a shallow copy
func (b FlexBox) Copy() Box { return &b }

func (b FlexBox) String() string {
	return fmt.Sprintf("<FlexBox %s>", b.BoxFields.elementTag)
}

func FlexBoxAnonymousFrom(parent Box, children []Box) *FlexBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewFlexBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeFlexBox) IsInstance(box Box) bool {
	_, is := box.(instanceFlexBox)
	return is
}

type typeFlexBox struct{}

func (t typeFlexBox) AnonymousFrom(parent Box, children []Box) Box {
	return FlexBoxAnonymousFrom(parent, children)
}

// A box that is both inline-level and a flex container.
// It behaves as inline on the outside and as a flex container on the inside.
type instanceInlineFlexBox interface {
	isInlineFlexBox()
	isParentBox()
	isFlexContainerBox()
	isInlineLevelBox()
}

func (InlineFlexBox) isInlineFlexBox()    {}
func (InlineFlexBox) isParentBox()        {}
func (InlineFlexBox) isFlexContainerBox() {}
func (InlineFlexBox) isInlineLevelBox()   {}
func (b *InlineFlexBox) Box() *BoxFields  { return &b.BoxFields }

// Copy is a shallow copy
func (b InlineFlexBox) Copy() Box { return &b }

func (b InlineFlexBox) String() string {
	return fmt.Sprintf("<InlineFlexBox %s>", b.BoxFields.elementTag)
}

func InlineFlexBoxAnonymousFrom(parent Box, children []Box) *InlineFlexBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
	out := NewInlineFlexBox(parent.Box().elementTag, style, children)
	return &out

}

func (t typeInlineFlexBox) IsInstance(box Box) bool {
	_, is := box.(instanceInlineFlexBox)
	return is
}

type typeInlineFlexBox struct{}

func (t typeInlineFlexBox) AnonymousFrom(parent Box, children []Box) Box {
	return InlineFlexBoxAnonymousFrom(parent, children)
}

var (
	TypeBlockBox            BoxType = typeBlockBox{}
	TypeLineBox             BoxType = typeLineBox{}
	TypeInlineBox           BoxType = typeInlineBox{}
	TypeInlineBlockBox      BoxType = typeInlineBlockBox{}
	TypeTableBox            BoxType = typeTableBox{}
	TypeInlineTableBox      BoxType = typeInlineTableBox{}
	TypeTableRowGroupBox    BoxType = typeTableRowGroupBox{}
	TypeTableRowBox         BoxType = typeTableRowBox{}
	TypeTableColumnGroupBox BoxType = typeTableColumnGroupBox{}
	TypeTableColumnBox      BoxType = typeTableColumnBox{}
	TypeTableCellBox        BoxType = typeTableCellBox{}
	TypeTableCaptionBox     BoxType = typeTableCaptionBox{}
	TypeFlexBox             BoxType = typeFlexBox{}
	TypeInlineFlexBox       BoxType = typeInlineFlexBox{}
)
