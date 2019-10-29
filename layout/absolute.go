package layout

import (
	"fmt"
	"log"
	"math"

	"github.com/benoitkugler/go-weasyprint/style/tree"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

type _Box = bo.Box

// AbsolutePlaceholder is left where an absolutely-positioned box was taken out of the flow.
type AbsolutePlaceholder struct {
	_Box
	layoutDone bool
	index      int
	resumeAt   *tree.SkipStack
	// equals false for raw Box
	// allow to merge Box and AbsolutePlaceholder
	isProperAbsolutePlaceholder bool
}

func NewAbsolutePlaceholder(box Box) *AbsolutePlaceholder {
	out := AbsolutePlaceholder{_Box: box, layoutDone: false, isProperAbsolutePlaceholder: true}
	return &out
}

func (abs *AbsolutePlaceholder) setLaidOutBox(newBox Box) {
	abs._Box = newBox
	abs.layoutDone = true
}

func (abs *AbsolutePlaceholder) Translate(box Box, dx, dy pr.Float, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	if abs.layoutDone {
		abs._Box.Translate(box, dx, dy, ignoreFloats)
	} else {
		// Descendants do not have a position yet.
		abs._Box.Box().PositionX += dx
		abs._Box.Box().PositionY += dy
	}
}

func (abs AbsolutePlaceholder) Copy() Box {
	out := abs
	out._Box = abs._Box.Copy()
	return &out
}

func (abs AbsolutePlaceholder) String() string {
	return fmt.Sprintf("<Placeholder %s>", abs.Box)
}

var absoluteWidth = handleMinMaxWidth(_absoluteWidth)

// @handleMinMaxWidth
func _absoluteWidth(box_ Box, context *LayoutContext, containingBlock block) (bool, pr.Float) {
	// http://www.w3.org/TR/CSS2/visudet.html#abs-replaced-width
	box := box_.Box()
	// These names are waaay too long
	marginL := box.MarginLeft
	marginR := box.MarginRight
	paddingL := box.PaddingLeft
	paddingR := box.PaddingRight
	borderL := box.BorderLeftWidth
	borderR := box.BorderRightWidth
	width := box.Width
	left := box.Left
	right := box.Right

	cbX, _, cbWidth, _ := containingBlock.unpack()

	// TODO: handle bidi
	paddingPlusBordersX := paddingL.V() + paddingR.V() + borderL.V() + borderR.V()
	var translateX pr.Float = 0
	translateBoxWidth := false
	defaultTranslateX := cbX - box.PositionX
	if left == pr.Auto && right == pr.Auto && width == pr.Auto {
		if marginL == pr.Auto {
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.Auto {
			box.MarginRight = pr.Float(0)
		}
		availableWidth := cbWidth - (paddingPlusBordersX + box.MarginLeft.V() + box.MarginRight.V())
		box.Width = shrinkToFit(context, box_, availableWidth)
	} else if left != pr.Auto && right != pr.Auto && width != pr.Auto {
		widthForMargins := cbWidth - (right.V() + left.V() + paddingPlusBordersX)
		if marginL == pr.Auto && marginR == pr.Auto {
			if width.V()+paddingPlusBordersX+right.V()+left.V() <= cbWidth {
				box.MarginLeft = widthForMargins / 2
				box.MarginRight = box.MarginLeft
			} else {
				box.MarginLeft = pr.Float(0)
				box.MarginRight = widthForMargins
			}
		} else if marginL == pr.Auto {
			box.MarginLeft = widthForMargins
		} else if marginR == pr.Auto {
			box.MarginRight = widthForMargins
		} else {
			box.MarginRight = widthForMargins
		}
		translateX = left.V() + defaultTranslateX
	} else {
		if marginL == pr.Auto {
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.Auto {
			box.MarginRight = pr.Float(0)
		}
		spacing := paddingPlusBordersX + box.MarginLeft.V() + box.MarginRight.V()
		if left == pr.Auto && width == pr.Auto {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-right.V())
			translateX = cbWidth - right.V() - spacing + defaultTranslateX
			translateBoxWidth = true
		} else if left == pr.Auto && right == pr.Auto {
			// Keep the static position
		} else if width == pr.Auto && right == pr.Auto {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-left.V())
			translateX = left.V() + defaultTranslateX
		} else if left == pr.Auto {
			translateX = (cbWidth + defaultTranslateX - right.V() - spacing - width.V())
		} else if width == pr.Auto {
			box.Width = cbWidth.V() - right.V() - left.V() - spacing
			translateX = left.V() + defaultTranslateX
		} else if right == pr.Auto {
			translateX = left.V() + defaultTranslateX
		}
	}
	return translateBoxWidth, translateX
}

func absoluteHeight(box_ Box, context LayoutContext, containingBlock block) (bool, pr.Float) {
	box := box_.Box()
	// These names are waaay too long
	marginT := box.MarginTop
	marginB := box.MarginBottom
	paddingT := box.PaddingTop
	paddingB := box.PaddingBottom
	borderT := box.BorderTopWidth
	borderB := box.BorderBottomWidth
	height := box.Height
	top := box.Top
	bottom := box.Bottom

	_, cbY, _, cbHeight := containingBlock.unpack()

	// http://www.w3.org/TR/CSS2/visudet.html#abs-non-replaced-height

	paddingsPlusBordersY := paddingT.V() + paddingB.V() + borderT.V() + borderB.V()
	var translateY pr.Float = 0
	translateBoxHeight := false
	defaultTranslateY := cbY - box.PositionY
	if top == pr.Auto && bottom == pr.Auto && height == pr.Auto {
		// Keep the static position
		if marginT == pr.Auto {
			box.MarginTop = pr.Float(0)
		}
		if marginB == pr.Auto {
			box.MarginBottom = pr.Float(0)
		}
	} else if top != pr.Auto && bottom != pr.Auto && height != pr.Auto {
		heightForMargins := cbHeight - (top.V() + bottom.V() + paddingsPlusBordersY)
		if marginT == pr.Auto && marginB == pr.Auto {
			box.MarginTop = heightForMargins / 2
			box.MarginBottom = box.MarginTop
		} else if marginT == pr.Auto {
			box.MarginTop = heightForMargins
		} else if marginB == pr.Auto {
			box.MarginBottom = heightForMargins
		} else {
			box.MarginBottom = heightForMargins
		}
		translateY = top.V() + defaultTranslateY
	} else {
		if marginT == pr.Auto {
			box.MarginTop = pr.Float(0)
		}
		if marginB == pr.Auto {
			box.MarginBottom = pr.Float(0)
		}
		spacing := paddingsPlusBordersY + box.MarginTop.V() + box.MarginBottom.V()
		if top == pr.Auto && height == pr.Auto {
			translateY = cbHeight.V() - bottom.V() - spacing + defaultTranslateY
			translateBoxHeight = true
		} else if top == pr.Auto && bottom == pr.Auto {
			// Keep the static position
		} else if height == pr.Auto && bottom == pr.Auto {
			translateY = top.V() + defaultTranslateY
		} else if top == pr.Auto {
			translateY = (cbHeight.V() + defaultTranslateY - bottom.V() - spacing - height.V())
		} else if height == pr.Auto {
			box.Height = cbHeight.V() - bottom.V() - top.V() - spacing
			translateY = top.V() + defaultTranslateY
		} else if bottom == pr.Auto {
			translateY = top.V() + defaultTranslateY
		}
	}
	return translateBoxHeight, translateY
}

func absoluteBlock(context *LayoutContext, box_ Box, containingBlock block, fixedBoxes *[]*AbsolutePlaceholder) Box {
	box := box_.Box()
	_, _, cbWidth, cbHeight := containingBlock.unpack()

	translateBoxWidth, translateX := absoluteWidth(box_, context, containingBlock)
	translateBoxHeight, translateY := absoluteHeight(box_, *context, containingBlock)

	// This box is the containing block for absolute descendants.
	var absoluteBoxes []*AbsolutePlaceholder

	if box.IsTableWrapper {
		tableWrapperWidth(context, box, bo.MaybePoint{cbWidth, cbHeight})
	}

	newBox := blockContainerLayout(context, box_, pr.Inf, nil, false, &absoluteBoxes, fixedBoxes, nil).newBox

	for _, childPlaceholder := range absoluteBoxes {
		absoluteLayout(context, childPlaceholder, newBox, fixedBoxes)
	}

	if translateBoxWidth {
		translateX -= newBox.Box().Width.V()
	}
	if translateBoxHeight {
		translateY -= newBox.Box().Height.V()
	}

	newBox.Translate(newBox, translateX, translateY, false)

	return newBox
}

// FIXME: waiting for weasyprint update
func absoluteFlex(context *LayoutContext, box_ Box, containingBlock block, fixedBoxes *[]*AbsolutePlaceholder) Box {
	//     // Avoid a circular import
	//     from .flex import flexLayout

	//     // TODO: this function is really close to absoluteBlock, we should have
	//     // only one function.
	//     // TODO: having containingBlockSizes && containingBlock is stupid.
	//     cbX, cbY, cbWidth, cbHeight = containingBlockSizes

	//     translateBoxWidth, translateX = absoluteWidth(
	//         box, context, containingBlockSizes)
	//     translateBoxHeight, translateY = absoluteHeight(
	//         box, context, containingBlockSizes)

	//     // This box is the containing block for absolute descendants.
	//     absoluteBoxes = []

	//     if box.isTableWrapper {
	//         tableWrapperWidth(context, box, (cbWidth, cbHeight))
	//     }

	//     newBox, _, _, _, _ = flexLayout(
	//         context, box, maxPositionY=float("inf"), skipStack=None,
	//         containingBlock=containingBlock, pageIsEmpty=false,
	//         absoluteBoxes=absoluteBoxes, fixedBoxes=fixedBoxes)

	//     for childPlaceholder := range absoluteBoxes {
	//         absoluteLayout(context, childPlaceholder, newBox, fixedBoxes)
	//     }

	//     if translateBoxWidth {
	//         translateX -= newBox.Width
	//     } if translateBoxHeight {
	//         translateY -= newBox.Height
	//     }

	//     newBox.translate(translateX, translateY)

	// return newBox
	return nil
}

// Set the width of absolute positioned ``box``.
func absoluteLayout(context *LayoutContext, placeholder *AbsolutePlaceholder, containingBlock Box, fixedBoxes *[]*AbsolutePlaceholder) {
	if placeholder.layoutDone {
		log.Fatalf("placeholder can't have its layout done.")
	}
	box := placeholder._Box
	placeholder.setLaidOutBox(absoluteBoxLayout(context, box, containingBlock, fixedBoxes))
}

func absoluteBoxLayout(context *LayoutContext, box Box, cb_ Box, fixedBoxes *[]*AbsolutePlaceholder) Box {
	// TODO: handle inline boxes (point 10.1.4.1)
	// http://www.w3.org/TR/CSS2/visudet.html#containing-block-details
	var containingBlock block
	cb := cb_.Box()
	if _, isPageBox := cb_.(*bo.PageBox); isPageBox {
		containingBlock.X = cb.ContentBoxX()
		containingBlock.Y = cb.ContentBoxY()
		containingBlock.Width = cb.Width.V()
		containingBlock.Height = cb.Height.V()
	} else {
		containingBlock.X = cb.PaddingBoxX()
		containingBlock.Y = cb.PaddingBoxY()
		containingBlock.Width = cb.PaddingWidth()
		containingBlock.Height = cb.PaddingHeight()
	}

	resolvePercentages(box, bo.MaybePoint{containingBlock.Width, containingBlock.Height}, "")
	resolvePositionPercentages(box.Box(), bo.Point{containingBlock.Width, containingBlock.Height})

	context.createBlockFormattingContext()
	// Absolute tables are wrapped into block boxes
	var newBox Box
	if bo.TypeBlockBox.IsInstance(box) {
		newBox = absoluteBlock(context, box, containingBlock, fixedBoxes)
	} else if bo.IsFlexContainerBox(box) {
		newBox = absoluteFlex(context, box, containingBlock, fixedBoxes)
	} else {
		if !bo.IsBlockReplacedBox(box) {
			log.Fatalf("box should be a BlockReplaced, got %s", box)
		}
		newBox = absoluteReplaced(context, box, containingBlock)
	}
	context.finishBlockFormattingContext(newBox)
	return newBox
}

func intDiv(a pr.Float, b int) pr.Float {
	return pr.Float(int(math.Floor(float64(a))) / b)
}

func absoluteReplaced(context *LayoutContext, box_ Box, containingBlock block) Box {
	inlineReplacedBoxWidthHeight(box_, containingBlock)
	box := box_.Box()
	cbX, cbY, cbWidth, cbHeight := containingBlock.unpack()
	ltr := box.Style.GetDirection() == "ltr"

	// http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-width
	if box.Left == pr.Auto && box.Right == pr.Auto {
		// static position:
		if ltr {
			box.Left = box.PositionX - cbX
		} else {
			box.Right = cbX + cbWidth - box.PositionX
		}
	}
	if box.Left == pr.Auto || box.Right == pr.Auto {
		if box.MarginLeft == pr.Auto {
			box.MarginLeft = pr.Float(0)
		}
		if box.MarginRight == pr.Auto {
			box.MarginRight = pr.Float(0)
		}
		remaining := cbWidth - box.MarginWidth()
		if box.Left == pr.Auto {
			box.Left = remaining - box.Right.V()
		}
		if box.Right == pr.Auto {
			box.Right = remaining - box.Left.V()
		}
	} else if pr.Auto == box.MarginLeft || pr.Auto == box.MarginRight {
		remaining := cbWidth - (box.BorderWidth() + box.Left.V() + box.Right.V())
		if box.MarginLeft == pr.Auto && box.MarginRight == pr.Auto {
			if remaining >= 0 {
				box.MarginLeft = intDiv(remaining, 2)
				box.MarginRight = box.MarginLeft
			} else if ltr {
				box.MarginLeft = pr.Float(0)
				box.MarginRight = remaining
			} else {
				box.MarginLeft = remaining
				box.MarginRight = pr.Float(0)
			}
		} else if box.MarginLeft == pr.Auto {
			box.MarginLeft = remaining
		} else {
			box.MarginRight = remaining
		}
	} else {
		// Over-constrained
		if ltr {
			box.Right = cbWidth - (box.MarginWidth() + box.Left.V())
		} else {
			box.Left = cbWidth - (box.MarginWidth() + box.Right.V())
		}
	}

	// http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-height
	if box.Top == pr.Auto && box.Bottom == pr.Auto {
		box.Top = box.PositionY - cbY
	}
	if box.Top == pr.Auto || box.Bottom == pr.Auto {
		if box.MarginTop == pr.Auto {
			box.MarginTop = pr.Float(0)
		}
		if box.MarginBottom == pr.Auto {
			box.MarginBottom = pr.Float(0)
		}
		remaining := cbHeight - box.MarginHeight()
		if box.Top == pr.Auto {
			box.Top = remaining
		}
		if box.Bottom == pr.Auto {
			box.Bottom = remaining
		}
	} else if box.MarginTop == pr.Auto || box.MarginBottom == pr.Auto {
		remaining := cbHeight - (box.BorderHeight() + box.Top.V() + box.Bottom.V())
		if box.MarginTop == pr.Auto && box.MarginBottom == pr.Auto {
			box.MarginTop = intDiv(remaining, 2)
			box.MarginBottom = box.MarginTop
		} else if box.MarginTop == pr.Auto {
			box.MarginTop = remaining
		} else {
			box.MarginBottom = remaining
		}
	} else {
		// Over-constrained
		box.Bottom = cbHeight - (box.MarginHeight() + box.Top.V())
	}

	// No children for replaced boxes, no need to .translate()
	box.PositionX = cbX + box.Left.V()
	box.PositionY = cbY + box.Top.V()
	return box_
}
