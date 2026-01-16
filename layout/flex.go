package layout

import (
	"github.com/gogpu/ui/geometry"
)

// flexLayoutName is the constant name for flex layout.
const flexLayoutName = "flex"

// FlexLayout implements CSS Flexbox-style layout.
//
// FlexLayout arranges children along a main axis (row or column),
// distributing space according to flex properties from the node styles.
//
// # Properties Used
//
// From parent style:
//   - FlexDirection: main axis direction
//   - FlexWrap: whether items wrap
//   - JustifyContent: main axis space distribution
//   - AlignItems: cross axis alignment
//   - Gap: space between items
//
// From child styles:
//   - FlexGrow: how much item grows to fill space
//   - FlexShrink: how much item shrinks when space is tight
//   - FlexBasis: initial size before grow/shrink
type FlexLayout struct{}

// Name returns "flex".
func (f *FlexLayout) Name() string {
	return flexLayoutName
}

// Compute performs flexbox layout on the tree starting at root.
func (f *FlexLayout) Compute(tree LayoutTree, root NodeID, available geometry.Size) Result {
	style := tree.Style(root)
	childCount := tree.ChildCount(root)

	if childCount == 0 {
		return Result{Size: geometry.Size{}}
	}

	isHorizontal := style.FlexDirection.IsHorizontal()
	isReversed := style.FlexDirection.IsReversed()

	// Get main/cross axis constraints
	mainMax, crossMax := f.getAxisSizes(available, isHorizontal)

	// Collect children and measure them
	items := make([]flexItem, childCount)
	for i := 0; i < childCount; i++ {
		childID := tree.ChildAt(root, i)
		childStyle := tree.Style(childID)

		// Measure child with loose constraints
		constraints := f.createMeasureConstraints(crossMax, isHorizontal)
		childSize := tree.Measure(childID, constraints)

		items[i] = flexItem{
			id:        childID,
			grow:      childStyle.FlexGrow,
			shrink:    childStyle.FlexShrink,
			basis:     childStyle.FlexBasis.Resolve(mainMax, 0),
			mainSize:  f.mainSize(childSize, isHorizontal),
			crossSize: f.crossSize(childSize, isHorizontal),
		}

		// Use basis if specified, otherwise use measured size
		if !childStyle.FlexBasis.IsAuto() {
			items[i].mainSize = items[i].basis
		}
	}

	// Calculate totals
	totalMain := f.calculateTotalMain(items, style.Gap)
	totalGrow := f.calculateTotalGrow(items)
	totalShrink := f.calculateTotalShrink(items)
	maxCross := f.calculateMaxCross(items)

	// Distribute space
	freeSpace := mainMax - totalMain
	if freeSpace > 0 && totalGrow > 0 {
		f.growItems(items, freeSpace, totalGrow)
	} else if freeSpace < 0 && totalShrink > 0 {
		f.shrinkItems(items, -freeSpace, totalShrink)
	}

	// Calculate positioning
	numGaps := childCount - 1
	if numGaps < 0 {
		numGaps = 0
	}
	contentMain := f.sumMainSizes(items) + style.Gap*float32(numGaps)
	mainStart, mainSpacing := f.calculateJustification(style, mainMax-contentMain, childCount)

	// Position children
	currentMain := mainStart
	if isReversed {
		currentMain = mainMax - mainStart
	}

	for i := range items {
		item := &items[i]

		// Calculate main position
		var mainPos float32
		if isReversed {
			mainPos = currentMain - item.mainSize
			currentMain -= item.mainSize + mainSpacing
		} else {
			mainPos = currentMain
			currentMain += item.mainSize + mainSpacing
		}

		// Calculate cross position
		crossPos := f.calculateCrossPosition(item, maxCross, style.AlignItems)

		// Set layout
		var pos geometry.Point
		var size geometry.Size
		if isHorizontal {
			pos = geometry.Point{X: mainPos, Y: crossPos}
			size = geometry.Size{Width: item.mainSize, Height: item.crossSize}
		} else {
			pos = geometry.Point{X: crossPos, Y: mainPos}
			size = geometry.Size{Width: item.crossSize, Height: item.mainSize}
		}
		tree.SetLayout(item.id, NodeLayout{Position: pos, Size: size})
	}

	// Calculate final size
	var resultSize geometry.Size
	if isHorizontal {
		resultSize = geometry.Size{Width: contentMain, Height: maxCross}
	} else {
		resultSize = geometry.Size{Width: maxCross, Height: contentMain}
	}

	return Result{
		Size:     resultSize,
		Overflow: contentMain > mainMax,
	}
}

// flexItem stores computed values during layout.
type flexItem struct {
	id        NodeID
	grow      float32
	shrink    float32
	basis     float32
	mainSize  float32
	crossSize float32
}

func (f *FlexLayout) getAxisSizes(available geometry.Size, isHorizontal bool) (mainMax, crossMax float32) {
	if isHorizontal {
		return available.Width, available.Height
	}
	return available.Height, available.Width
}

func (f *FlexLayout) mainSize(size geometry.Size, isHorizontal bool) float32 {
	if isHorizontal {
		return size.Width
	}
	return size.Height
}

func (f *FlexLayout) crossSize(size geometry.Size, isHorizontal bool) float32 {
	if isHorizontal {
		return size.Height
	}
	return size.Width
}

func (f *FlexLayout) createMeasureConstraints(crossMax float32, isHorizontal bool) geometry.Constraints {
	if isHorizontal {
		return geometry.BoxConstraints(0, geometry.Infinity, 0, crossMax)
	}
	return geometry.BoxConstraints(0, crossMax, 0, geometry.Infinity)
}

func (f *FlexLayout) calculateTotalMain(items []flexItem, gap float32) float32 {
	var total float32
	for i, item := range items {
		total += item.mainSize
		if i < len(items)-1 {
			total += gap
		}
	}
	return total
}

func (f *FlexLayout) calculateTotalGrow(items []flexItem) float32 {
	var total float32
	for _, item := range items {
		total += item.grow
	}
	return total
}

func (f *FlexLayout) calculateTotalShrink(items []flexItem) float32 {
	var total float32
	for _, item := range items {
		total += item.shrink * item.mainSize
	}
	return total
}

func (f *FlexLayout) calculateMaxCross(items []flexItem) float32 {
	var maxCross float32
	for _, item := range items {
		if item.crossSize > maxCross {
			maxCross = item.crossSize
		}
	}
	return maxCross
}

func (f *FlexLayout) growItems(items []flexItem, space, totalGrow float32) {
	for i := range items {
		if items[i].grow > 0 {
			items[i].mainSize += space * (items[i].grow / totalGrow)
		}
	}
}

func (f *FlexLayout) shrinkItems(items []flexItem, amount, totalShrink float32) {
	for i := range items {
		if items[i].shrink > 0 {
			scaled := items[i].shrink * items[i].mainSize
			items[i].mainSize -= amount * (scaled / totalShrink)
			if items[i].mainSize < 0 {
				items[i].mainSize = 0
			}
		}
	}
}

func (f *FlexLayout) sumMainSizes(items []flexItem) float32 {
	var total float32
	for _, item := range items {
		total += item.mainSize
	}
	return total
}

func (f *FlexLayout) calculateJustification(style *Style, freeSpace float32, childCount int) (start, spacing float32) {
	switch style.JustifyContent {
	case JustifyStart:
		return 0, style.Gap
	case JustifyEnd:
		return freeSpace, style.Gap
	case JustifyCenter:
		return freeSpace / 2, style.Gap
	case JustifySpaceBetween:
		if childCount > 1 {
			return 0, freeSpace / float32(childCount-1)
		}
		return 0, 0
	case JustifySpaceAround:
		if childCount > 0 {
			space := freeSpace / float32(childCount)
			return space / 2, space
		}
		return 0, 0
	case JustifySpaceEvenly:
		if childCount > 0 {
			space := freeSpace / float32(childCount+1)
			return space, space
		}
		return 0, 0
	default:
		return 0, style.Gap
	}
}

func (f *FlexLayout) calculateCrossPosition(item *flexItem, maxCross float32, align AlignItems) float32 {
	switch align {
	case AlignItemsStart, AlignItemsBaseline:
		return 0
	case AlignItemsEnd:
		return maxCross - item.crossSize
	case AlignItemsCenter:
		return (maxCross - item.crossSize) / 2
	case AlignItemsStretch:
		item.crossSize = maxCross
		return 0
	default:
		return 0
	}
}

func init() {
	Register(&FlexLayout{})
}
