package layout

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

//     Line breaking and layout for inline-level boxes.

type lineBoxe struct {
	line *bo.LineBox
	resumeAt *float32
}

// Return an iterator of ``(line, resumeAt)``.
// 
// ``line`` is a laid-out LineBox with as much content as possible that
// fits in the available width.
// 
// :param box: a non-laid-out :class:`LineBox`
// :param positionY: vertical top position of the line box on the page
// :param skipStack: ``None`` to start at the beginning of ``linebox``,
// 				   or a ``resumeAt`` value to continue just after an
// 				   already laid-out line.
func iterLineBoxes(context LayoutContext, box *bo.LineBox, positionY, skipStack, containingBlock bo.Box,
                    absoluteBoxes []AbsolutePlaceholder, fixedBoxes []bo.Box, firstLetterStyle) []lineBoxe {
    resolvePercentages(box, containingBlock)
    if skipStack == nil {
        // TODO: wrong, see https://github.com/Kozea/WeasyPrint/issues/679
        box.TextIndent = resolveOnePercentage(box, "textIndent", containingBlock.width)
    } else {
        box.TextIndent = 0
	} 
	var out []lineBoxe
	for {
        line, resumeAt := getNextLinebox( context, box, positionY, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes, firstLetterStyle)
        if line {
            positionY = line.positionY + line.height
		}
		 if line == nil {
            return out
		} 
		out = append(out, lineBoxe{line: line, resumeAt: resumeAt})
        if resumeAt == nil {
            return out
		} 
		skipStack = resumeAt
        box.TextIndent = 0
        firstLetterStyle = None
	}
}


func getNextLinebox(context LayoutContext, linebox *bo.LineBox, positionY, skipStack,
                     containingBlock, absoluteBoxes, fixedBoxes,
                     firstLetterStyle pr.Properties)  lineBoxe {
                     
    skipStack , cont := skipFirstWhitespace(linebox, skipStack)
    if cont {
        return lineBoxe{}
    }

    skipStack = firstLetterToBox(linebox, skipStack, firstLetterStyle)

    linebox.PositionY = positionY

    if len(context.excludedShapes) != 0 {
        // Width && height must be calculated to avoid floats
        linebox.Width = inlineMinContentWidth( context, linebox, skipStack=skipStack, firstLine=true)
        linebox.Height, _ = strutLayout(linebox.style, context)
    } else {
        // No float, width && height will be set by the lines
		linebox.Width = 0
		linebox.Height = 0
	} 
	positionX, positionY, availableWidth := avoidCollisions(context, linebox, containingBlock, outer=false)

    candidateHeight := linebox.Height

	excludedShapes := make([]shape, len(context.excludedShapes))
	for i,v := range context.excludedShapes {
		excludedShapes[i] = v
	}

    for {
        linebox.PositionX = positionX
        linebox.PositionY = positionY
        maxX := positionX + availableWidth
        positionX += linebox.TextIndent
    
        linePlaceholders = []
        lineAbsolutes = []
        lineFixed = []
        waitingFloats = []

        (line, resumeAt, preservedLineBreak, firstLetter,
         lastLetter, floatWidth) = splitInlineBox(
             context, linebox, positionX, maxX, skipStack, containingBlock,
             lineAbsolutes, lineFixed, linePlaceholders, waitingFloats,
             lineChildren=[])
        linebox.Width, linebox.Height = line.width, line.height

        if isPhantomLinebox(line) && not preservedLineBreak {
            line.height = 0
            break
        }

        removeLastWhitespace(context, line)

        newPositionX, _, newAvailableWidth = avoidCollisions(
            context, linebox, containingBlock, outer=false)
        // TODO: handle rtl
        newAvailableWidth -= floatWidth["right"]
        alignmentAvailableWidth = (
            newAvailableWidth + newPositionX - linebox.PositionX)
        offsetX = textAlign(
            context, line, alignmentAvailableWidth,
            last=(resumeAt == nil || preservedLineBreak))

        bottom, top = lineBoxVerticality(line)
        assert top is not None
        assert bottom is not None
        line.baseline = -top
        line.positionY = top
        line.height = bottom - top
        offsetY = positionY - top
        line.marginTop = 0
        line.marginBottom = 0

        line.translate(offsetX, offsetY)
        // Avoid floating point errors, as positionY - top + top != positionY
        // Removing this line breaks the position == linebox.Position test below
        // See https://github.com/Kozea/WeasyPrint/issues/583
        line.positionY = positionY

        if line.height <= candidateHeight {
            break
        } candidateHeight = line.height

        newExcludedShapes = context.excludedShapes
        context.excludedShapes = excludedShapes
        positionX, positionY, availableWidth = avoidCollisions(
            context, line, containingBlock, outer=false)
        if (positionX, positionY) == (
                linebox.PositionX, linebox.PositionY) {
                }
            context.excludedShapes = newExcludedShapes
            break

    absoluteBoxes.extend(lineAbsolutes)
    fixedBoxes.extend(lineFixed)

    for placeholder := range linePlaceholders {
        if placeholder.style["WeasySpecifiedDisplay"].startswith("inline") {
            // Inline-level static position {
            } placeholder.translate(0, positionY - placeholder.positionY)
        } else {
            // Block-level static position: at the start of the next line
            placeholder.translate(
                line.positionX - placeholder.positionX,
                positionY + line.height - placeholder.positionY)
        }
    }

    floatChildren = []
    waitingFloatsY = line.positionY + line.height
    for waitingFloat := range waitingFloats {
        waitingFloat.positionY = waitingFloatsY
        waitingFloat = floatLayout(
            context, waitingFloat, containingBlock, absoluteBoxes,
            fixedBoxes)
        floatChildren.append(waitingFloat)
    } if floatChildren {
        line.children += tuple(floatChildren)
    }

    return line, resumeAt


// Return the ``skipStack`` to start just after the remove spaces
//     at the beginning of the line.
//     See http://www.w3.org/TR/CSS21/text.html#white-space-model
func skipFirstWhitespace(box bo.Box, skipStack *bo.SkipStack) (ss *bo.SkipStack, continue_ bool){
	var (
		index int 
		nextSkipStack *bo.SkipStack
	)
	if skipStack != nil {
        index, nextSkipStack = skipStack.Skip, skipStack.Stack
    }

    if textBox, ok :=box.(*bo.TextBox) ; ok {
        if nextSkipStack != nil {
			log.Fatalf("expected nil nextSkipStack, got %v", nextSkipStack)
		}
        whiteSpace := textBox.Style.GetWhitespace()
		text := []rune(textBox.Text)
		length := len(text)
        if index == length {
            // Starting a the end of the TextBox, no text to see: Continue
            return nil, true
		} 
		if whiteSpace == "normal" || whiteSpace == "nowrap" || whiteSpace == "pre-line" {
            for index < length && text[index] == ' ' {
                index += 1
            }
		} 
		if index != 0{
			return &bo.SkipStack{Skip:index}, false
		}
		return nil, false
    }

    if bo.TypeLineBox.IsInstance(box) ||bo.TypeInlineBox.IsInstance(box) {
		children := box.Box().Children
		if index == 0 && len(children) == 0{
            return nil, false
		} 
		result, cont := skipFirstWhitespace(children[index], nextSkipStack)
        if cont {
            index += 1
            if index >= len(children) {
                return nil, true
			} 
			result = skipFirstWhitespace(children[index], nil)
		} 
		if index != 0 || result != nil {
			return &bo.SkipStack{Skip:index, Stack: result}, false
		}
		return nil, false
    }
	if skipStack != nil {
		log.Fatalf("unexpected skip inside %s", box)
	} 
    return nil, false
}

// Remove := range place space characters at the end of a line.
//     This also reduces the width of the inline parents of the modified text.
//     
func removeLastWhitespace(context, box) {
    ancestors = []
    while isinstance(box, (boxes.LineBox, boxes.InlineBox)) {
        ancestors.append(box)
        if not box.children {
            return
        } box = box.children[-1]
    } if not (isinstance(box, boxes.TextBox) and
            box.style["whiteSpace"] := range ("normal", "nowrap", "pre-line")) {
            }
        return
    newText = box.text.rstrip(" ")
    if newText {
        if len(newText) == len(box.text) {
            return
        } box.text = newText
        newBox, resume, _ = splitTextBox(context, box, None, 0)
        assert newBox is not None
        assert resume == nil
        spaceWidth = box.Width - newBox.Width
        box.Width = newBox.Width
    } else {
        spaceWidth = box.Width
        box.Width = 0
        box.text = ""
    }
} 
    for ancestor := range ancestors {
        ancestor.width -= spaceWidth
    }

    // TODO: All tabs (U+0009) are rendered as a horizontal shift that
    // lines up the start edge of the next glyph with the next tab stop.
    // Tab stops occur at points that are multiples of 8 times the width
    // of a space (U+0020) rendered in the block"s font from the block"s
    // starting content edge.

    // TODO: If spaces (U+0020) or tabs (U+0009) at the end of a line have
    // "white-space" set to "pre-wrap", UAs may visually collapse them.
					 }

// Create a box for the ::first-letter selector.
func firstLetterToBox(box bo.Box, skipStack bo.SkipStack, firstLetterStyle pr.Properties) {
    if len(firstLetterStyle) != 0 && len(box.Box().Children) != 0 {
        // Some properties must be ignored :in first-letter boxes.
        // https://drafts.csswg.org/selectors-3/#application-in-css
        // At least, position is ignored to avoid layout troubles.
        firstLetterStyle.SetPosition(pr.String("static"))
    }
	tables := []*unicode.RangeTable{unicode.Ps, unicode.Pe, unicode.Pi, unicode.Pf, unicode.Po}

        firstLetter := ""
        child := box.Box().children[0]
        if textBox, ok := child.(*bo.TextBox); ok {
            letterStyle := tree.ComputedFromCascaded(cascaded={}, parentStyle=firstLetterStyle, element=None)
            if strings.HasSuffix(textBox.ElementTag, "::first-letter") {
                letterBox := bo.NewInlineBox(textBox.ElementTag + "::first-letter", letterStyle, []bo.Box{child})
                box.Box().Children[0] = letterBox
            } else if textBox.Text != "" {
				text := []rune(textBox.Text)
                characterFound = false
                if skipStack != nil {
                    childSkipStack := skipStack.Stack
                    if childSkipStack != nil {
                        index = childSkipStack.Skip
                        text = text[index:]
                        skipStack = nil
                    }
				}
				for len(text) != 0 {
					nextLetter := text[0]
						isPunc := unicode.In(nextLetter, tables...)
						if !isPunc {
							if characterFound {
								break
							}
							characterFound = true
						}
					firstLetter += string(nextLetter)
                    text = text[1:]
				} 
				textBox.Text = string(text)
				if strings.TrimLeft(firstLetter, "\n") != "" {
                    // "This type of initial letter is similar to an
                    // inline-level element if its "float" property is "none",
                    // otherwise it is similar to a floated element."
					if firstLetterStyle.GetFloat() == "none" {
                        letterBox := bo.NewInlineBox(textBox.ElementTag + "::first-letter", firstLetterStyle, nil)
                        textBox_ := bo.NewTextBox(textBox.ElementTag + "::first-letter", letterStyle, firstLetter)
                        letterBox.Children = []bo.Box{&textBox_}
                        textBox.Children = append([]bo.Box{&letterBox}, textBox.Children...)
                    } else {
                        letterBox := bo.NewBlockBox(textBox.ElementTag + "::first-letter", firstLetterStyle, nil)
                        letterBox.FirstLetterStyle = nil
                        lineBox := bo.NewLineBox(textBox.ElementTag + "::first-letter", firstLetterStyle, nil)
                        letterBox.Children = []bo.Box{&lineBox}
                        textBox_ := bo.NewTextBox(textBox.ElementTag + "::first-letter", letterStyle, firstLetter)
                        lineBox.Children = []bo.Box{&textBox_}
                        textBox.Children = append([]bo.Box{&letterBox}, textBox.Children...)
						} 
					if skipStack  != nil && childSkipStack != nil {
                        skipStack = bo.SkipStack{Skip: skipStack.Skip, Stack: &bo.SkipStack{
							Skip: childSkipStack.Skip + 1,
							Stack: childSkipStack,
						}}
                    }
                }
            }
        } else if bo.IsParentBox(child) {
            if skipStack {
                childSkipStack = skipStack[1]
            } else {
                childSkipStack = None
            } childSkipStack = firstLetterToBox(
                child, childSkipStack, firstLetterStyle)
            if skipStack {
                skipStack = (skipStack[0], childSkipStack)
            }
        }
    return skipStack
	}

@handleMinMaxWidth
// 
//     Compute && set the used width for replaced boxes (inline- || block-level)
//     
func replacedBoxWidth(box, containingBlock) {
    from .blocks import blockLevelWidth
} 
    intrinsicWidth, intrinsicHeight = box.replacement.getIntrinsicSize(
        box.style["imageResolution"], box.style["fontSize"])

    // This algorithm simply follows the different points of the specification {
    } // http://www.w3.org/TR/CSS21/visudet.html#inline-replaced-width
    if box.Height == "auto" && box.Width == "auto" {
        if intrinsicWidth is not None {
            // Point #1
            box.Width = intrinsicWidth
        } else if box.replacement.intrinsicRatio is not None {
            if intrinsicHeight is not None {
                // Point #2 first part
                box.Width = intrinsicHeight * box.replacement.intrinsicRatio
            } else {
                // Point #3
                blockLevelWidth(box, containingBlock)
            }
        }
    }

    if box.Width == "auto" {
        if box.replacement.intrinsicRatio is not None {
            // Point #2 second part
            box.Width = box.Height * box.replacement.intrinsicRatio
        } else if intrinsicWidth is not None {
            // Point #4
            box.Width = intrinsicWidth
        } else {
            // Point #5
            // It"s pretty useless to rely on device size to set width.
            box.Width = 300
        }
    }


@handleMinMaxHeight
// 
//     Compute && set the used height for replaced boxes (inline- || block-level)
//     
func replacedBoxHeight(box) {
    // http://www.w3.org/TR/CSS21/visudet.html#inline-replaced-height
    intrinsicWidth, intrinsicHeight = box.replacement.getIntrinsicSize(
        box.style["imageResolution"], box.style["fontSize"])
    intrinsicRatio = box.replacement.intrinsicRatio
} 
    // Test "auto" on the computed width, not the used width
    if box.Height == "auto" && box.Width == "auto" {
        box.Height = intrinsicHeight
    } else if box.Height == "auto" && intrinsicRatio {
        box.Height = box.Width / intrinsicRatio
    }

    if (box.Height == "auto" && box.Width == "auto" and
            intrinsicHeight is not None) {
            }
        box.Height = intrinsicHeight
    else if intrinsicRatio is not None && box.Height == "auto" {
        box.Height = box.Width / intrinsicRatio
    } else if box.Height == "auto" && intrinsicHeight is not None {
        box.Height = intrinsicHeight
    } else if box.Height == "auto" {
        // It"s pretty useless to rely on device size to set width.
        box.Height = 150
    }


// Lay out an inline :class:`boxes.ReplacedBox` ``box``.
func inlineReplacedBoxLayout(box, containingBlock) {
    for side := range ["top", "right", "bottom", "left"] {
        if getattr(box, "margin" + side) == "auto" {
            setattr(box, "margin" + side, 0)
        }
	} 
	inlineReplacedBoxWidthHeight(box, containingBlock)
} 

func inlineReplacedBoxWidthHeight(box bo.Box, containingBlock block) {
	if style := box.Box().Style; style.GetWidth().String == "auto" && style.GetHeight().String == "auto" {
		replacedBoxWidth.withoutMinMax(box, containingBlock)
		replacedBoxHeight.withoutMinMax(box)
		minMaxAutoReplaced(box)
	} else {
		replacedBoxWidth(box, containingBlock)
		replacedBoxHeight(box)
	}
}
// Resolve {min,max}-{width,height} constraints on replaced elements
//     that have "auto" width && heights.
//     
func minMaxAutoReplaced(box) {
    width = box.Width
    height = box.Height
    minWidth = box.minWidth
    minHeight = box.minHeight
    maxWidth = max(minWidth, box.maxWidth)
    maxHeight = max(minHeight, box.maxHeight)
} 
    // (violationWidth, violationHeight)
    violations = (
        "min" if width < minWidth else "max" if width > maxWidth else "",
        "min" if height < minHeight else "max" if height > maxHeight else "")

    // Work around divisions by zero. These are pathological cases anyway.
    // TODO: is there a cleaner way?
    if width == 0 {
        width = 1e-6
    } if height == 0 {
        height = 1e-6
    }

    // ("", ""): nothing to do
    if violations == ("max", "") {
        box.Width = maxWidth
        box.Height = max(maxWidth * height / width, minHeight)
    } else if violations == ("min", "") {
        box.Width = minWidth
        box.Height = min(minWidth * height / width, maxHeight)
    } else if violations == ("", "max") {
        box.Width = max(maxHeight * width / height, minWidth)
        box.Height = maxHeight
    } else if violations == ("", "min") {
        box.Width = min(minHeight * width / height, maxWidth)
        box.Height = minHeight
    } else if violations == ("max", "max") {
        if maxWidth / width <= maxHeight / height {
            box.Width = maxWidth
            box.Height = max(minHeight, maxWidth * height / width)
        } else {
            box.Width = max(minWidth, maxHeight * width / height)
            box.Height = maxHeight
        }
    } else if violations == ("min", "min") {
        if minWidth / width <= minHeight / height {
            box.Width = min(maxWidth, minHeight * width / height)
            box.Height = minHeight
        } else {
            box.Width = minWidth
            box.Height = min(maxHeight, minWidth * height / width)
        }
    } else if violations == ("min", "max") {
        box.Width = minWidth
        box.Height = maxHeight
    } else if violations == ("max", "min") {
        box.Width = maxWidth
        box.Height = minHeight
    }


func atomicBox(context, box, positionX, skipStack, containingBlock,
               absoluteBoxes, fixedBoxes) {
               }
    """Compute the width && the height of the atomic ``box``."""
    if isinstance(box, boxes.ReplacedBox) {
        box = box.copy()
        inlineReplacedBoxLayout(box, containingBlock)
        box.baseline = box.marginHeight()
    } else if isinstance(box, boxes.InlineBlockBox) {
        if box.isTableWrapper {
            tableWrapperWidth(
                context, box,
                (containingBlock.width, containingBlock.height))
        } box = inlineBlockBoxLayout(
            context, box, positionX, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes)
    } else:  // pragma: no cover
        raise TypeError("Layout for %s not handled yet" % type(box)._Name_)
    return box


func inlineBlockBoxLayout(context, box, positionX, skipStack,
                            containingBlock, absoluteBoxes, fixedBoxes) {
                            }
    // Avoid a circular import
    from .blocks import blockContainerLayout

    resolvePercentages(box, containingBlock)

    // http://www.w3.org/TR/CSS21/visudet.html#inlineblock-width
    if box.marginLeft == "auto" {
        box.marginLeft = 0
    } if box.marginRight == "auto" {
        box.marginRight = 0
    } // http://www.w3.org/TR/CSS21/visudet.html#block-root-margin
    if box.marginTop == "auto" {
        box.marginTop = 0
    } if box.marginBottom == "auto" {
        box.marginBottom = 0
    }

    inlineBlockWidth(box, context, containingBlock)

    box.PositionX = positionX
    box.PositionY = 0
    box, _, _, _, _ = blockContainerLayout(
        context, box, maxPositionY=float("inf"), skipStack=skipStack,
        pageIsEmpty=true, absoluteBoxes=absoluteBoxes,
        fixedBoxes=fixedBoxes)
    box.baseline = inlineBlockBaseline(box)
    return box


// 
//     Return the y position of the baseline for an inline block
//     from the top of its margin box.
//     http://www.w3.org/TR/CSS21/visudet.html#propdef-vertical-align
//     
func inlineBlockBaseline(box) {
    if box.isTableWrapper {
        // Inline table"s baseline is its first row"s baseline
        for child := range box.children {
            if isinstance(child, boxes.TableBox) {
                if child.children && child.children[0].children {
                    firstRow = child.children[0].children[0]
                    return firstRow.baseline
                }
            }
        }
    } else if box.style["overflow"] == "visible" {
        result = findInFlowBaseline(box, last=true)
        if result {
            return result
        }
    } return box.PositionY + box.marginHeight()
} 

@handleMinMaxWidth
func inlineBlockWidth(box, context, containingBlock) {
    if box.Width == "auto" {
        box.Width = shrinkToFit(context, box, containingBlock.width)
    }
} 

func splitInlineLevel(context, box, positionX, maxX, skipStack,
                       containingBlock, absoluteBoxes, fixedBoxes,
                       linePlaceholders, waitingFloats, lineChildren) {
    """Fit as much content as possible from an inline-level box := range a width.

    Return ``(newBox, resumeAt, preservedLineBreak, firstLetter,
    lastLetter)``. ``resumeAt`` is ``None`` if all of the content
    fits. Otherwise it can be passed as a ``skipStack`` parameter to resume
    where we left off.

    ``newBox`` is non-empty (unless the box is empty) && as big as possible
    while being narrower than ``availableWidth``, if possible (may overflow
    is no split is possible.)

                       }
    """
    resolvePercentages(box, containingBlock)
    floatWidths = {"left": 0, "right": 0}
    if isinstance(box, boxes.TextBox) {
        box.PositionX = positionX
        if skipStack == nil {
            skip = 0
        } else {
            skip, skipStack = skipStack
            skip = skip || 0
            assert skipStack == nil
        }
    }

        newBox, skip, preservedLineBreak = splitTextBox(
            context, box, maxX - positionX, skip)

        if skip == nil {
            resumeAt = None
        } else {
            resumeAt = (skip, None)
        } if box.text {
            firstLetter = box.text[0]
            if skip == nil {
                lastLetter = box.text[-1]
            } else {
                lastLetter = box.text[skip - 1]
            }
        } else {
            firstLetter = lastLetter = None
        }
    else if isinstance(box, boxes.InlineBox) {
        if box.marginLeft == "auto" {
            box.marginLeft = 0
        } if box.marginRight == "auto" {
            box.marginRight = 0
        } (newBox, resumeAt, preservedLineBreak, firstLetter,
         lastLetter, floatWidths) = splitInlineBox(
            context, box, positionX, maxX, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes, linePlaceholders, waitingFloats,
             lineChildren)
    } else if isinstance(box, boxes.AtomicInlineLevelBox) {
        newBox = atomicBox(
            context, box, positionX, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes)
        newBox.PositionX = positionX
        resumeAt = None
        preservedLineBreak = false
        // See https://www.w3.org/TR/css-text-3/#line-breaking
        // Atomic inlines behave like ideographic characters.
        firstLetter = "\u2e80"
        lastLetter = "\u2e80"
    } else if isinstance(box, boxes.InlineFlexBox) {
        box.PositionX = positionX
        box.PositionY = 0
        for side := range ["top", "right", "bottom", "left"] {
            if getattr(box, "margin" + side) == "auto" {
                setattr(box, "margin" + side, 0)
            }
        } newBox, resumeAt, _, _, _ = flexLayout(
            context, box, float("inf"), skipStack, containingBlock,
            false, absoluteBoxes, fixedBoxes)
        preservedLineBreak = false
        firstLetter = "\u2e80"
        lastLetter = "\u2e80"
    } else:  // pragma: no cover
        raise TypeError("Layout for %s not handled yet" % type(box)._Name_)
    return (
        newBox, resumeAt, preservedLineBreak, firstLetter, lastLetter,
        floatWidths)


func splitInlineBox(context, box, positionX, maxX, skipStack,
                     containingBlock, absoluteBoxes, fixedBoxes,
                     linePlaceholders, waitingFloats, lineChildren) {
                     }
    """Same behavior as splitInlineLevel."""

    // In some cases (shrink-to-fit result being the preferred width)
    // maxX is coming from Pango itself,
    // but floating point errors have accumulated {
    } //   width2 = (width + X) - X   // := range some cases, width2 < width
    // Increase the value a bit to compensate && not introduce
    // an unexpected line break. The 1e-9 value comes from PEP 485.
    maxX *= 1 + 1e-9

    isStart = skipStack == nil
    initialPositionX = positionX
    initialSkipStack = skipStack
    assert isinstance(box, (boxes.LineBox, boxes.InlineBox))
    leftSpacing = (box.PaddingLeft + box.marginLeft +
                    box.borderLeftWidth)
    rightSpacing = (box.PaddingRight + box.marginRight +
                     box.borderRightWidth)
    contentBoxLeft = positionX

    children = []
    waitingChildren = []
    preservedLineBreak = false
    firstLetter = lastLetter = None
    floatWidths = {"left": 0, "right": 0}
    floatResumeAt = 0

    if box.style["position"] == "relative" {
        absoluteBoxes = []
    }

    if isStart {
        skip = 0
    } else {
        skip, skipStack = skipStack
    }

    for i, child := range enumerate(box.children[skip:]) {
        index = i + skip
        child.positionY = box.PositionY
        if child.isAbsolutelyPositioned() {
            child.positionX = positionX
            placeholder = AbsolutePlaceholder(child)
            linePlaceholders.append(placeholder)
            waitingChildren.append((index, placeholder))
            if child.style["position"] == "absolute" {
                absoluteBoxes.append(placeholder)
            } else {
                fixedBoxes.append(placeholder)
            } continue
        } else if child.isFloated() {
            child.positionX = positionX
            floatWidth = shrinkToFit(context, child, containingBlock.width)
        }
    }

            // To retrieve the real available space for floats, we must remove
            // the trailing whitespaces from the line
            nonFloatingChildren = [
                child_ for _, child_ := range (children + waitingChildren)
                if not child.isFloated()]
            if nonFloatingChildren {
                floatWidth -= trailingWhitespaceSize(
                    context, nonFloatingChildren[-1])
            }

            if floatWidth > maxX - positionX || waitingFloats {
                // TODO: the absolute && fixed boxes := range the floats must be
                // added here, && not := range iterLineBoxes
                waitingFloats.append(child)
            } else {
                child = floatLayout(
                    context, child, containingBlock, absoluteBoxes,
                    fixedBoxes)
                waitingChildren.append((index, child))
            }

                // Translate previous line children
                dx = max(child.marginWidth(), 0)
                floatWidths[child.style["float"]] += dx
                if child.style["float"] == "left" {
                    if isinstance(box, boxes.LineBox) {
                        // The parent is the line, update the current position
                        // for the next child. When the parent is not the line
                        // (it is an inline block), the current position of the
                        // line is updated by the box itself (see next
                        // splitInlineLevel call).
                        positionX += dx
                    }
                } else if child.style["float"] == "right" {
                    // Update the maximum x position for the next children
                    maxX -= dx
                } for _, oldChild := range lineChildren {
                    if not oldChild.isInNormalFlow() {
                        continue
                    } if ((child.style["float"] == "left" and
                            box.style["direction"] == "ltr") or
                        (child.style["float"] == "right" and
                            box.style["direction"] == "rtl")) {
                            }
                        oldChild.translate(dx=dx)
                }
            floatResumeAt = index + 1
            continue

        lastChild = (index == len(box.children) - 1)
        availableWidth = maxX
        childWaitingFloats = []
        newChild, resumeAt, preserved, first, last, newFloatWidths = (
            splitInlineLevel(
                context, child, positionX, availableWidth, skipStack,
                containingBlock, absoluteBoxes, fixedBoxes,
                linePlaceholders, childWaitingFloats, lineChildren))
        if lastChild && rightSpacing && resumeAt == nil {
            // TODO: we should take care of children added into absoluteBoxes,
            // fixedBoxes && other lists.
            if box.style["direction"] == "rtl" {
                availableWidth -= leftSpacing
            } else {
                availableWidth -= rightSpacing
            } newChild, resumeAt, preserved, first, last, newFloatWidths = (
                splitInlineLevel(
                    context, child, positionX, availableWidth, skipStack,
                    containingBlock, absoluteBoxes, fixedBoxes,
                    linePlaceholders, childWaitingFloats, lineChildren))
        }

        if box.style["direction"] == "rtl" {
            maxX -= newFloatWidths["left"]
        } else {
            maxX -= newFloatWidths["right"]
        }

        skipStack = None
        if preserved {
            preservedLineBreak = true
        }

        canBreak = None
        if lastLetter is true {
            lastLetter = " "
        } else if lastLetter is false {
            lastLetter = " "  // no-break space
        } else if box.style["whiteSpace"] := range ("pre", "nowrap") {
            canBreak = false
        } if canBreak == nil {
            if None := range (lastLetter, first) {
                canBreak = false
            } else {
                canBreak = canBreakText(
                    lastLetter + first, child.style["lang"])
            }
        }

        if canBreak {
            children.extend(waitingChildren)
            waitingChildren = []
        }

        if firstLetter == nil {
            firstLetter = first
        } if child.trailingCollapsibleSpace {
            lastLetter = true
        } else {
            lastLetter = last
        }

        if newChild == nil {
            // May be None where we have an empty TextBox.
            assert isinstance(child, boxes.TextBox)
        } else {
            if isinstance(box, boxes.LineBox) {
                lineChildren.append((index, newChild))
            } // TODO: we should try to find a better condition here.
            trailingWhitespace = (
                isinstance(newChild, boxes.TextBox) and
                not newChild.text.strip())
        }

            marginWidth = newChild.marginWidth()
            newPositionX = newChild.positionX + marginWidth

            if newPositionX > maxX && not trailingWhitespace {
                if waitingChildren {
                    // Too wide, let"s try to cut inside waiting children,
                    // starting from the end.
                    // TODO: we should take care of children added into
                    // absoluteBoxes, fixedBoxes && other lists.
                    waitingChildrenCopy = waitingChildren[:]
                    breakFound = false
                    while waitingChildrenCopy {
                        childIndex, child = waitingChildrenCopy.pop()
                        // TODO: should we also accept relative children?
                        if (child.isInNormalFlow() and
                                canBreakInside(child)) {
                                }
                            // We break the waiting child at its last possible
                            // breaking point.
                            // TODO: The dirty solution chosen here is to
                            // decrease the actual size by 1 && render the
                            // waiting child again with this constraint. We may
                            // find a better way.
                            maxX = child.positionX + child.marginWidth() - 1
                            childNewChild, childResumeAt, _, _, _, _ = (
                                splitInlineLevel(
                                    context, child, child.positionX, maxX,
                                    None, box, absoluteBoxes, fixedBoxes,
                                    linePlaceholders, waitingFloats,
                                    lineChildren))
                    }
                }
            }

                            // As PangoLayout && PangoLogAttr don"t always
                            // agree, we have to rely on the actual split to
                            // know whether the child was broken.
                            // https://github.com/Kozea/WeasyPrint/issues/614
                            breakFound = childResumeAt is not None
                            if childResumeAt == nil {
                                // PangoLayout decided not to break the child
                                childResumeAt = (0, None)
                            } // TODO: use this when Pango is always 1.40.13+ {
                            } // breakFound = true

                            children = children + waitingChildrenCopy
                            if childNewChild == nil {
                                // May be None where we have an empty TextBox.
                                assert isinstance(child, boxes.TextBox)
                            } else {
                                children += [(childIndex, childNewChild)]
                            }

                            // As this child has already been broken
                            // following the original skip stack, we have to
                            // add the original skip stack to the partial
                            // skip stack we get after the new rendering.

                            // We have to do {
                            } // resumeAt + initialSkipStack
                            // but adding skip stacks is a bit complicated
                            currentSkipStack = initialSkipStack
                            currentResumeAt = (childIndex, childResumeAt)
                            stack = []
                            while currentSkipStack && currentResumeAt {
                                skip, currentSkipStack = (
                                    currentSkipStack)
                                resume, currentResumeAt = (
                                    currentResumeAt)
                                stack.append(skip + resume)
                                if resume != 0 {
                                    break
                                }
                            } resumeAt = currentResumeAt
                            while stack {
                                resumeAt = (stack.pop(), resumeAt)
                            } break
                    if breakFound {
                        break
                    }
                if children {
                    // Too wide, can"t break waiting children && the inline is
                    // non-empty: put child entirely on the next line.
                    resumeAt = (children[-1][0] + 1, None)
                    childWaitingFloats = []
                    break
                }

            positionX = newPositionX
            waitingChildren.append((index, newChild))

        waitingFloats.extend(childWaitingFloats)
        if resumeAt is not None {
            children.extend(waitingChildren)
            resumeAt = (index, resumeAt)
            break
        }
    else {
        children.extend(waitingChildren)
        resumeAt = None
    }

    isEnd = resumeAt == nil
    newBox = box.copyWithChildren(
        [boxChild for index, boxChild := range children],
        isStart=isStart, isEnd=isEnd)
    if isinstance(box, boxes.LineBox) {
        // We must reset line box width according to its new children
        inFlowChildren = [
            boxChild for boxChild := range newBox.children
            if boxChild.isInNormalFlow()]
        if inFlowChildren {
            newBox.Width = (
                inFlowChildren[-1].positionX +
                inFlowChildren[-1].marginWidth() -
                newBox.PositionX)
        } else {
            newBox.Width = 0
        }
    } else {
        newBox.PositionX = initialPositionX
        if box.style["boxDecorationBreak"] == "clone" {
            translationNeeded = true
        } else {
            translationNeeded = (
                isStart if box.style["direction"] == "ltr" else isEnd)
        } if translationNeeded {
            for child := range newBox.children {
                child.translate(dx=leftSpacing)
            }
        } newBox.Width = positionX - contentBoxLeft
        newBox.translate(dx=floatWidths["left"], ignoreFloats=true)
    }

    lineHeight, newBox.baseline = strutLayout(box.style, context)
    newBox.Height = box.style["fontSize"]
    halfLeading = (lineHeight - newBox.Height) / 2.
    // Set margins to the half leading but also compensate for borders and
    // paddings. We want marginHeight() == lineHeight
    newBox.marginTop = (halfLeading - newBox.borderTopWidth -
                          newBox.PaddingTop)
    newBox.marginBottom = (halfLeading - newBox.borderBottomWidth -
                             newBox.PaddingBottom)

    if newBox.style["position"] == "relative" {
        for absoluteBox := range absoluteBoxes {
            absoluteLayout(context, absoluteBox, newBox, fixedBoxes)
        }
    }

    if resumeAt is not None {
        if resumeAt[0] < floatResumeAt {
            resumeAt = (floatResumeAt, None)
        }
    }

    return (
        newBox, resumeAt, preservedLineBreak, firstLetter, lastLetter,
        floatWidths)


// Keep as much text as possible from a TextBox in a limited width.
//
// Try not to overflow but always have some text in ``new_box``
//
// Return ``(new_box, skip, preserved_line_break)``. ``skip`` is the number of
// UTF-8 bytes to skip form the start of the TextBox for the next line, or
// ``None`` if all of the text fits.
//
// Also break on preserved line breaks.
func splitTextBox(context LayoutContext, box bo.TextBox, availableWidth float32, skip *int) (*bo.TextBox, *int, bool) {}  

func splitTextBox(context, box, availableWidth, skip) {
    assert isinstance(box, boxes.TextBox)
    fontSize = box.style["fontSize"]
    text = box.text[skip:]
    if fontSize == 0 || not text {
        return None, None, false
    } layout, length, resumeAt, width, height, baseline = splitFirstLine(
        text, box.style, context, availableWidth, box.justificationSpacing)
    assert resumeAt != 0
} 
    // Convert ``length`` && ``resumeAt`` from UTF-8 indexes := range text
    // to Unicode indexes.
    // No need to encode what’s after resumeAt (if set) || length (if
    // resumeAt is not set). One code point is one || more byte, so
    // UTF-8 indexes are always bigger || equal to Unicode indexes.
    newText = layout.text
    encoded = text.encode("utf8")
    if resumeAt is not None {
        between = encoded[length:resumeAt].decode("utf8")
        resumeAt = len(encoded[:resumeAt].decode("utf8"))
    } length = len(encoded[:length].decode("utf8"))

    if length > 0 {
        box = box.copyWithText(newText)
        box.Width = width
        box.PangoLayout = layout
        // "The height of the content area should be based on the font,
        //  but this specification does not specify how."
        // http://www.w3.org/TR/CSS21/visudet.html#inline-non-replaced
        // We trust Pango && use the height of the LayoutLine.
        box.Height = height
        // "only the "line-height" is used when calculating the height
        //  of the line box."
        // Set margins so that marginHeight() == lineHeight
        lineHeight, _ = strutLayout(box.style, context)
        halfLeading = (lineHeight - height) / 2.
        box.marginTop = halfLeading
        box.marginBottom = halfLeading
        // form the top of the content box
        box.baseline = baseline
        // form the top of the margin box
        box.baseline += box.marginTop
    } else {
        box = None
    }

    if resumeAt == nil {
        preservedLineBreak = false
    } else {
        preservedLineBreak = (length != resumeAt) && between.strip(" ")
        if preservedLineBreak {
            // See http://unicode.org/reports/tr14/
            // \r is already handled by processWhitespace
            lineBreaks = ("\n", "\t", "\f", "\u0085", "\u2028", "\u2029")
            assert between := range lineBreaks, (
                "Got %r between two lines. "
                "Expected nothing || a preserved line break" % (between,))
        } resumeAt += skip
    }

    return box, resumeAt, preservedLineBreak


// Handle ``vertical-align`` within an :class:`LineBox` (or of a
//     non-align sub-tree).
//     Place all boxes vertically assuming that the baseline of ``box``
//     is at `y = 0`.
//     Return ``(maxY, minY)``, the maximum && minimum vertical position
//     of margin boxes.
//     
func lineBoxVerticality(box) {
    topBottomSubtrees = []
    maxY, minY = alignedSubtreeVerticality(
        box, topBottomSubtrees, baselineY=0)
    subtreesWithMinMax = [
        (subtree, subMaxY, subMinY)
        for subtree := range topBottomSubtrees
        for subMaxY, subMinY := range [
            (None, None) if subtree.isFloated()
            else alignedSubtreeVerticality(
                subtree, topBottomSubtrees, baselineY=0)
        ]
    ]
} 
    if subtreesWithMinMax {
        subPositions = [
            subMaxY - subMinY
            for subtree, subMaxY, subMinY := range subtreesWithMinMax
            if not subtree.isFloated()]
        if subPositions {
            highestSub = max(subPositions)
            maxY = max(maxY, minY + highestSub)
        }
    }

    for subtree, subMaxY, subMinY := range subtreesWithMinMax {
        if subtree.isFloated() {
            dy = minY - subtree.positionY
        } else if subtree.style["verticalAlign"] == "top" {
            dy = minY - subMinY
        } else {
            assert subtree.style["verticalAlign"] == "bottom"
            dy = maxY - subMaxY
        } translateSubtree(subtree, dy)
    } return maxY, minY


func translateSubtree(box, dy) {
    if isinstance(box, boxes.InlineBox) {
        box.PositionY += dy
        if box.style["verticalAlign"] := range ("top", "bottom") {
            for child := range box.children {
                translateSubtree(child, dy)
            }
        }
    } else {
        // Text || atomic boxes
        box.translate(dy=dy)
    }
} 

func alignedSubtreeVerticality(box, topBottomSubtrees, baselineY) {
    maxY, minY = inlineBoxVerticality(box, topBottomSubtrees, baselineY)
    // Account for the line box itself {
    } top = baselineY - box.baseline
    bottom = top + box.marginHeight()
    if minY == nil || top < minY {
        minY = top
    } if maxY == nil || bottom > maxY {
        maxY = bottom
    }
} 
    return maxY, minY


// Handle ``vertical-align`` within an :class:`InlineBox`.
//     Place all boxes vertically assuming that the baseline of ``box``
//     is at `y = baselineY`.
//     Return ``(maxY, minY)``, the maximum && minimum vertical position
//     of margin boxes.
//     
func inlineBoxVerticality(box, topBottomSubtrees, baselineY) {
    maxY = None
    minY = None
    if not isinstance(box, (boxes.LineBox, boxes.InlineBox)) {
        return maxY, minY
    }
} 
    for child := range box.children {
        if not child.isInNormalFlow() {
            if child.isFloated() {
                topBottomSubtrees.append(child)
            } continue
        } verticalAlign = child.style["verticalAlign"]
        if verticalAlign == "baseline" {
            childBaselineY = baselineY
        } else if verticalAlign == "middle" {
            oneEx = box.style["fontSize"] * exRatio(box.style)
            top = baselineY - (oneEx + child.marginHeight()) / 2.
            childBaselineY = top + child.baseline
        } else if verticalAlign == "text-top" {
            // align top with the top of the parent’s content area
            top = (baselineY - box.baseline + box.marginTop +
                   box.borderTopWidth + box.PaddingTop)
            childBaselineY = top + child.baseline
        } else if verticalAlign == "text-bottom" {
            // align bottom with the bottom of the parent’s content area
            bottom = (baselineY - box.baseline + box.marginTop +
                      box.borderTopWidth + box.PaddingTop + box.Height)
            childBaselineY = bottom - child.marginHeight() + child.baseline
        } else if verticalAlign := range ("top", "bottom") {
            // TODO: actually implement vertical-align: top && bottom
            // Later, we will assume for this subtree that its baseline
            // is at y=0.
            childBaselineY = 0
        } else {
            // Numeric value: The child’s baseline is `verticalAlign` above
            // (lower y) the parent’s baseline.
            childBaselineY = baselineY - verticalAlign
        }
    }

        // the child’s `top` is `child.baseline` above (lower y) its baseline.
        top = childBaselineY - child.baseline
        if isinstance(child, (boxes.InlineBlockBox, boxes.InlineFlexBox)) {
            // This also includes table wrappers for inline tables.
            child.translate(dy=top - child.positionY)
        } else {
            child.positionY = top
            // grand-children for inline boxes are handled below
        }

        if verticalAlign := range ("top", "bottom") {
            // top || bottom are special, they need to be handled in
            // a later pass.
            topBottomSubtrees.append(child)
            continue
        }

        bottom = top + child.marginHeight()
        if minY == nil || top < minY {
            minY = top
        } if maxY == nil || bottom > maxY {
            maxY = bottom
        } if isinstance(child, boxes.InlineBox) {
            childrenMaxY, childrenMinY = inlineBoxVerticality(
                child, topBottomSubtrees, childBaselineY)
            if childrenMinY is not None && childrenMinY < minY {
                minY = childrenMinY
            } if childrenMaxY is not None && childrenMaxY > maxY {
                maxY = childrenMaxY
            }
        }
    return maxY, minY


// Return how much the line should be moved horizontally according to
//     the `text-align` property.
//     
func textAlign(context, line, availableWidth, last) {
    // "When the total width of the inline-level boxes on a line is less than
    // the width of the line box containing them, their horizontal distribution
    // within the line box is determined by the "text-align" property."
    if line.width >= availableWidth {
        return 0
    }
} 
    align = line.style["textAlign"]
    spaceCollapse = line.style["whiteSpace"] := range (
        "normal", "nowrap", "pre-line")
    if align := range ("-weasy-start", "-weasy-end") {
        if (align == "-weasy-start") ^ (line.style["direction"] == "rtl") {
            align = "left"
        } else {
            align = "right"
        }
    } if align == "justify" && last {
        align = "right" if line.style["direction"] == "rtl" else "left"
    } if align == "left" {
        return 0
    } offset = availableWidth - line.width
    if align == "justify" {
        if spaceCollapse {
            // Justification of texts where white space is not collapsing is
            // - forbidden by CSS 2, and
            // - not required by CSS 3 Text.
            justifyLine(context, line, offset)
        } return 0
    } if align == "center" {
        return offset / 2
    } else {
        assert align == "right"
        return offset
    }


func justifyLine(context, line, extraWidth) {
    // TODO: We should use a better alorithm here, see
    // https://www.w3.org/TR/css-text-3/#justify-algos
    nbSpaces = countSpaces(line)
    if nbSpaces == 0 {
        return
    } addWordSpacing(context, line, extraWidth / nbSpaces, 0)
} 

func countSpaces(box) {
    if isinstance(box, boxes.TextBox) {
        // TODO: remove trailing spaces correctly
        return box.text.count(" ")
    } else if isinstance(box, (boxes.LineBox, boxes.InlineBox)) {
        return sum(countSpaces(child) for child := range box.children)
    } else {
        return 0
    }
} 

func addWordSpacing(context, box, justificationSpacing, xAdvance) {
    if isinstance(box, boxes.TextBox) {
        box.justificationSpacing = justificationSpacing
        box.PositionX += xAdvance
        nbSpaces = countSpaces(box)
        if nbSpaces > 0 {
            layout = createLayout(
                box.text, box.style, context, float("inf"),
                box.justificationSpacing)
            layout.deactivate()
            extraSpace = justificationSpacing * nbSpaces
            xAdvance += extraSpace
            box.Width += extraSpace
            box.PangoLayout = layout
        }
    } else if isinstance(box, (boxes.LineBox, boxes.InlineBox)) {
        box.PositionX += xAdvance
        previousXAdvance = xAdvance
        for child := range box.children {
            if child.isInNormalFlow() {
                xAdvance = addWordSpacing(
                    context, child, justificationSpacing, xAdvance)
            }
        } box.Width += xAdvance - previousXAdvance
    } else {
        // Atomic inline-level box
        box.translate(xAdvance, 0)
    } return xAdvance
} 

// http://www.w3.org/TR/CSS21/visuren.html#phantom-line-box
func isPhantomLinebox(linebox) {
    for child := range linebox.children {
        if isinstance(child, boxes.InlineBox) {
            if not isPhantomLinebox(child) {
                return false
            } for side := range ("top", "right", "bottom", "left") {
                if (getattr(child.style["margin%s" % side], "value", None) or
                        child.style["border%sWidth" % side] or
                        child.style["padding%s" % side].value) {
                        }
                    return false
            }
        } else if child.isInNormalFlow() {
            return false
        }
    } return true
} 

func canBreakInside(box) {
    // See https://www.w3.org/TR/css-text-3/#white-space-property
    textWrap = box.style["whiteSpace"] := range ("normal", "pre-wrap", "pre-line")
    if isinstance(box, boxes.AtomicInlineLevelBox) {
        return false
    } else if isinstance(box, boxes.TextBox) {
        if textWrap {
            return canBreakText(box.text, box.style["lang"])
        } else {
            return false
        }
    } else if isinstance(box, boxes.ParentBox) {
        if textWrap {
            return any(canBreakInside(child) for child := range box.children)
        } else {
            return false
        }
    } return false
