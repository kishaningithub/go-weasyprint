package structure

import (
	"fmt"
	"log"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
)

type BlockLevelBox struct {
	clearance interface{}
}

type BlockBox struct {
	InstanceBlockBox
	BlockLevelBox
	BoxFields
}

type LineBox struct {
	InstanceLineBox
	BoxFields
	textOverflow string
}

type InlineLevelBox struct {
	InstanceInlineLevelBox
}

type InlineBox struct {
	InstanceInlineBox

	BoxFields
	InlineLevelBox
}

type TextBox struct {
	InstanceTextBox

	BoxFields
	InlineLevelBox

	text string
}

type AtomicInlineLevelBox struct {
	InstanceAtomicInlineLevelBox

	InlineLevelBox
}

type InlineBlockBox struct {
	InstanceBlockBox

	BoxFields
	AtomicInlineLevelBox
}

type ReplacedBox struct {
	InstanceReplacedBox

	BoxFields
	replacement pr.Image
}

type BlockReplacedBox struct {
	InstanceBlockReplacedBox

	ReplacedBox
}

type InlineReplacedBox struct {
	InstanceInlineReplacedBox

	ReplacedBox
	AtomicInlineLevelBox
}

type TableBox struct {
	InstanceTableBox

	BoxFields
}

type InlineTableBox struct {
	InstanceInlineTableBox

	TableBox
}

type TableRowGroupBox struct {
	InstanceTableRowGroupBox

	BoxFields
}

type TableRowBox struct {
	InstanceTableRowBox

	BoxFields
}

type TableColumnGroupBox struct {
	InstanceTableColumnGroupBox

	BoxFields
}

type TableColumnBox struct {
	InstanceTableColumnBox

	BoxFields
}

type TableCellBox struct {
	InstanceTableCellBox

	BoxFields
}

type TableCaptionBox struct {
	InstanceTableCaptionBox

	BlockBox
}

type PageBox struct {
	InstancePageBox

	BoxFields
	pageType utils.PageElement
}

type MarginBox struct {
	InstanceMarginBox

	BoxFields
	atKeyword string
}

type FlexBox struct {
	InstanceFlexBox

	BoxFields
}

type InlineFlexBox struct {
	InstanceInlineFlexBox

	InlineLevelBox
	BoxFields
}

func IsParentBox(box Box) bool {
	_, is := box.(InstanceParentBox)
	return is
}

func NewBlockBox(elementTag string, style pr.Properties, children []Box) BlockBox {
	out := BlockBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func LineBoxAnonymousFrom(parent Box, children []Box) *LineBox {
	parentBox := parent.Box()
	style := tree.ComputedFromCascaded(nil, nil, parentBox.style, nil, "", "", nil)
	out := NewLineBox(parentBox.elementTag, style, children)
	if parentBox.style.GetOverflow() != "visible" {
		out.textOverflow = string(parentBox.style.GetTextOverflow())
	}
	return &out
}

func NewLineBox(elementTag string, style pr.Properties, children []Box) LineBox {
	out := LineBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.textOverflow = "clip"
	return out
}

func (InlineLevelBox) removeDecoration(box *BoxFields, start, end bool) {
	ltr := box.style.GetDirection() == "ltr"
	if start {
		side := "right"
		if ltr {
			side = "left"
		}
		box.resetSpacing(side)
	}
	if end {
		side := "left"
		if ltr {
			side = "right"
		}
		box.resetSpacing(side)
	}
}

func NewInlineBox(elementTag string, style pr.Properties, children []Box) InlineBox {
	out := InlineBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

// Return the (x, y, w, h) rectangle where the box is clickable.
func (b *InlineBox) hitArea() (x float32, y float32, w float32, h float32) {
	return b.Box().borderBoxX(), b.Box().positionY, b.Box().borderWidth(), b.Box().marginHeight()
}

func NewTextBox(elementTag string, style pr.Properties, text string) TextBox {
	if len(text) == 0 {
		log.Fatalf("empty text")
	}
	textTransform := style.GetTextTransform()
	if textTransform != "none" {
		switch textTransform {
		case "uppercase":
			text = strings.ToUpper(text)
		case "lowercase":
			text = strings.ToLower(text)
		// Python’s unicode.captitalize is not the same.
		case "capitalize":
			text = strings.ToTitle(text)
		case "full-width":
			text = strings.Map(func(u rune) rune {
				rep, in := asciiToWide[u]
				if !in {
					return -1
				}
				return rep
			}, text)
		}
	}
	if style.GetHyphens() == "none" {
		text = strings.ReplaceAll(text, "\u00AD", "") //  U+00AD SOFT HYPHEN (SHY)
	}
	box := newBoxFields(elementTag, style, nil)
	out := TextBox{BoxFields: box, text: text}
	return out
}

// Return a new TextBox identical to this one except for the text.
func (b TextBox) copyWithText(text string) TextBox {
	if len(text) == 0 {
		log.Fatal("empty text")
	}
	newBox := b
	newBox.text = text
	return newBox
}

func (u TextBox) removeDecoration(b *BoxFields, start, end bool) {
	u.InlineLevelBox.removeDecoration(b, start, end)
}

func NewInlineBlockBox(elementTag string, style pr.Properties, children []Box) InlineBlockBox {
	out := InlineBlockBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func (u InlineBox) removeDecoration(b *BoxFields, start, end bool) {
	u.InlineLevelBox.removeDecoration(b, start, end)
}

func NewReplacedBox(elementTag string, style pr.Properties, replacement pr.Image) ReplacedBox {
	out := ReplacedBox{BoxFields: newBoxFields(elementTag, style, nil)}
	out.replacement = replacement
	return out
}

func NewBlockReplacedBox(elementTag string, style pr.Properties, replacement pr.Image) BlockReplacedBox {
	out := BlockReplacedBox{ReplacedBox: NewReplacedBox(elementTag, style, replacement)}
	return out
}

func NewInlineReplacedBox(elementTag string, style pr.Properties, replacement pr.Image) InlineReplacedBox {
	out := InlineReplacedBox{ReplacedBox: NewReplacedBox(elementTag, style, replacement)}
	return out
}

func (u InlineReplacedBox) removeDecoration(b *BoxFields, start, end bool) {
	u.ReplacedBox.removeDecoration(b, start, end)
}

func NewTableBox(elementTag string, style pr.Properties, children []Box) TableBox {
	out := TableBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.tabularContainer = true
	return out
}

func (b *TableBox) allChildren() []Box {
	return append(b.Box().children, b.columnGroups...)
}

func (b *TableBox) translate(box Box, dx float32, dy float32, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	table := b.Box()
	for index, position := range table.columnPositions {
		table.columnPositions[index] = position + dx
	}
	table.translate(box, dx, dy, ignoreFloats)
}

func (b *TableBox) pageValues() (pr.Page, pr.Page) {
	s := b.Box().style
	return s.GetPage(), s.GetPage()
}

func NewInlineTableBox(elementTag string, style pr.Properties, children []Box) InlineTableBox {
	out := InlineTableBox{TableBox: NewTableBox(elementTag, style, children)}
	return out
}

func NewTableRowGroupBox(elementTag string, style pr.Properties, children []Box) TableRowGroupBox {
	out := TableRowGroupBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.tabularContainer = true
	out.isHeader = true
	out.isFooter = true
	return out
}

func NewTableRowBox(elementTag string, style pr.Properties, children []Box) TableRowBox {
	out := TableRowBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func NewTableColumnGroupBox(elementTag string, style pr.Properties, children []Box) TableColumnGroupBox {
	out := TableColumnGroupBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.span = 1
	return out
}

type withCells interface {
	getCells() []Box
}

// Return cells that originate in the group's columns.
func (b *TableColumnGroupBox) getCells() []Box {
	var out []Box
	for _, column := range b.Box().children {
		for _, cell := range column.(withCells).getCells() {
			out = append(out, cell)
		}
	}
	return out
}

func NewTableColumnBox(elementTag string, style pr.Properties, children []Box) TableColumnBox {
	out := TableColumnBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.span = 1
	return out
}

// Return cells that originate in the column.
// May be overriden on instances.
func (b *TableColumnBox) getCells() []Box {
	return []Box{}
}

func NewTableCellBox(elementTag string, style pr.Properties, children []Box) TableCellBox {
	out := TableCellBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.internalTableOrCaption = true
	out.colspan = 1
	out.rowspan = 1
	return out
}

func NewTableCaptionBox(elementTag string, style pr.Properties, children []Box) TableCaptionBox {
	out := TableCaptionBox{BlockBox: NewBlockBox(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	return out
}

func NewPageBox(pageType utils.PageElement, style pr.Properties) PageBox {
	fields := newBoxFields("", style, nil)
	return PageBox{BoxFields: fields, pageType: pageType}
}

func (b *PageBox) String() string {
	return fmt.Sprintf("<PageBox %v>", b.pageType)
}

func NewMarginBox(atKeyword string, style pr.Properties) MarginBox {
	fields := newBoxFields("", style, nil)
	return MarginBox{BoxFields: fields, atKeyword: atKeyword}
}

func (b *MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", b.atKeyword)
}

func NewFlexBox(elementTag string, style pr.Properties, children []Box) FlexBox {
	out := FlexBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func NewInlineFlexBox(elementTag string, style pr.Properties, children []Box) InlineFlexBox {
	out := InlineFlexBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func (u InlineFlexBox) removeDecoration(b *BoxFields, start, end bool) {
	u.BoxFields.removeDecoration(b, start, end)
}
