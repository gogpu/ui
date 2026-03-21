package app

import (
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// --- Tracking Canvas ---

// trackingCanvas records all DrawText calls for first-frame rendering verification.
type trackingCanvas struct {
	drawTextCalls []drawTextCall
	drawRectCount int
}

// drawTextCall records the parameters of a single DrawText invocation.
type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	bold     bool
}

func (c *trackingCanvas) Clear(widget.Color)                                       {}
func (c *trackingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                 { c.drawRectCount++ }
func (c *trackingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)    {}
func (c *trackingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {}
func (c *trackingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
}
func (c *trackingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *trackingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *trackingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}

func (c *trackingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, _ widget.Color, bold bool, _ widget.TextAlign) {
	c.drawTextCalls = append(c.drawTextCalls, drawTextCall{
		text:     text,
		bounds:   bounds,
		fontSize: fontSize,
		bold:     bold,
	})
}

func (c *trackingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *trackingCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *trackingCanvas) PushClip(_ geometry.Rect)                     {}
func (c *trackingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *trackingCanvas) PopClip()                                     {}
func (c *trackingCanvas) PushTransform(_ geometry.Point)               {}
func (c *trackingCanvas) PopTransform()                                {}
func (c *trackingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// Compile-time check.
var _ widget.Canvas = (*trackingCanvas)(nil)

// --- Tree Walker ---

// boundsOf returns the Bounds() of a widget, or an empty Rect if
// the widget does not embed WidgetBase.
func boundsOf(w widget.Widget) geometry.Rect {
	if b, ok := w.(interface{ Bounds() geometry.Rect }); ok {
		return b.Bounds()
	}
	return geometry.Rect{}
}

// walkTree recursively collects all widgets in the tree via Children().
func walkTree(w widget.Widget) []widget.Widget {
	if w == nil {
		return nil
	}
	result := []widget.Widget{w}
	for _, child := range w.Children() {
		result = append(result, walkTree(child)...)
	}
	return result
}

// collectUnselectedTabContent walks the tree and returns a set of widgets
// that are unselected TabView content. TabView uses lazy rendering: only
// the selected tab's content is laid out and drawn.
func collectUnselectedTabContent(root widget.Widget) map[widget.Widget]bool {
	result := make(map[widget.Widget]bool)
	walkForTabViews(root, result)
	return result
}

// walkForTabViews recursively finds TabView widgets and marks their
// unselected children as unselected content.
func walkForTabViews(w widget.Widget, unselected map[widget.Widget]bool) {
	if w == nil {
		return
	}
	if tv, ok := w.(*tabview.Widget); ok {
		children := tv.Children()
		selected := tv.SelectedIndex()
		for i, child := range children {
			if i != selected {
				// Mark this child and all its descendants as unselected.
				for _, desc := range walkTree(child) {
					unselected[desc] = true
				}
			}
		}
	}
	for _, child := range w.Children() {
		walkForTabViews(child, unselected)
	}
}

// collectTextWidgets filters the tree for *primitives.TextWidget instances.
func collectTextWidgets(all []widget.Widget) []*primitives.TextWidget {
	var texts []*primitives.TextWidget
	for _, w := range all {
		if tw, ok := w.(*primitives.TextWidget); ok {
			texts = append(texts, tw)
		}
	}
	return texts
}

// --- Test Helpers ---

// buildIDELayout builds a widget tree similar to the IDE example:
//
//	VBox
//	  SplitView(Horizontal)
//	    first:  Text("Explorer")
//	    second: SplitView(Vertical)
//	      first:  TabView [Tab("main.go", Text("package main")), Tab("go.mod", Text("module app"))]
//	      second: TabView [Tab("Output", Text("Build succeeded")), Tab("Terminal", Text("$ go run ."))]
func buildIDELayout() widget.Widget {
	// Left panel: simple text simulating a tree view.
	leftPanel := primitives.Text("Explorer")

	// Top-right: code editor tabs.
	topTabView := tabview.New([]tabview.Tab{
		{Label: "main.go", Content: primitives.Text("package main")},
		{Label: "go.mod", Content: primitives.Text("module app")},
	})

	// Bottom-right: output/terminal tabs.
	bottomTabView := tabview.New([]tabview.Tab{
		{Label: "Output", Content: primitives.Text("Build succeeded")},
		{Label: "Terminal", Content: primitives.Text("$ go run .")},
	})

	// Right panel: vertical split between editor and terminal.
	rightSplit := splitview.New(
		splitview.First(topTabView),
		splitview.Second(bottomTabView),
		splitview.OrientationOpt(splitview.Vertical),
		splitview.InitialRatio(0.6),
	)

	// Main: horizontal split between explorer and editor area.
	mainSplit := splitview.New(
		splitview.First(leftPanel),
		splitview.Second(rightSplit),
		splitview.OrientationOpt(splitview.Horizontal),
		splitview.InitialRatio(0.25),
	)

	return primitives.VBox(mainSplit)
}

// setupFirstFrameWindow creates a headless window with the given root,
// sets size, runs Frame() once (layout), and returns window + canvas.
func setupFirstFrameWindow(root widget.Widget) (*Window, *trackingCanvas) {
	a := New()
	w := a.Window()
	w.SetRoot(root)
	w.HandleResize(1024, 700)
	w.Frame()

	canvas := &trackingCanvas{}
	return w, canvas
}

// --- Tests ---

func TestFirstFrame_AllWidgetsGetNonZeroBounds(t *testing.T) {
	root := buildIDELayout()
	w, _ := setupFirstFrameWindow(root)

	all := walkTree(w.Root())
	if len(all) == 0 {
		t.Fatal("walkTree returned empty tree")
	}

	// Collect unselected TabView content widgets. TabView uses lazy layout:
	// only the selected tab's content gets laid out and receives bounds.
	// Unselected content having empty bounds is correct behavior.
	unselectedContent := collectUnselectedTabContent(w.Root())

	var zeroBoundsWidgets []string
	for i, wgt := range all {
		if unselectedContent[wgt] {
			continue // Skip: unselected tab content is expected to have zero bounds.
		}
		bounds := boundsOf(wgt)
		if bounds.IsEmpty() {
			name := fmt.Sprintf("widget[%d] type=%T", i, wgt)
			if tw, ok := wgt.(*primitives.TextWidget); ok {
				name = fmt.Sprintf("TextWidget(%q)", tw.Content())
			}
			zeroBoundsWidgets = append(zeroBoundsWidgets, name)
		}
	}

	if len(zeroBoundsWidgets) > 0 {
		t.Errorf("widgets with zero/empty bounds after single Frame():\n  %s",
			strings.Join(zeroBoundsWidgets, "\n  "))
	}

	t.Logf("total widgets in tree: %d", len(all))
}

func TestFirstFrame_AllVisibleTextWidgetsDrawn(t *testing.T) {
	root := buildIDELayout()
	w, canvas := setupFirstFrameWindow(root)

	// Collect all text widgets in the tree, excluding unselected tab content.
	// TabView only draws the selected tab's content (lazy rendering).
	all := walkTree(w.Root())
	unselected := collectUnselectedTabContent(w.Root())
	textWidgets := collectTextWidgets(all)

	if len(textWidgets) == 0 {
		t.Fatal("no TextWidgets found in IDE layout tree")
	}

	// Draw the tree to the tracking canvas.
	w.DrawTo(canvas)

	// Build a set of expected text content (only from visible/selected widgets).
	expectedTexts := make(map[string]bool)
	for _, tw := range textWidgets {
		if unselected[tw] {
			continue // Skip unselected tab content — not expected to be drawn.
		}
		content := tw.Content()
		if content != "" {
			expectedTexts[content] = false
		}
	}

	// Mark which texts were actually drawn.
	for _, call := range canvas.drawTextCalls {
		if _, exists := expectedTexts[call.text]; exists {
			expectedTexts[call.text] = true
		}
	}

	// Check for undrawn text widgets.
	var missing []string
	for text, drawn := range expectedTexts {
		if !drawn {
			missing = append(missing, text)
		}
	}

	if len(missing) > 0 {
		t.Errorf("visible text widgets not drawn on first frame:\n  %s\n\nDrawText calls received: %d\nExpected visible text widgets: %d",
			strings.Join(missing, "\n  "),
			len(canvas.drawTextCalls),
			len(expectedTexts))

		// Diagnostic: show what WAS drawn.
		if len(canvas.drawTextCalls) > 0 {
			drawn := make([]string, 0, len(canvas.drawTextCalls))
			for _, call := range canvas.drawTextCalls {
				drawn = append(drawn, fmt.Sprintf("%q bounds=%v", call.text, call.bounds))
			}
			t.Logf("actually drawn:\n  %s", strings.Join(drawn, "\n  "))
		}
	}

	t.Logf("visible text widgets: %d, DrawText calls: %d", len(expectedTexts), len(canvas.drawTextCalls))
}

func TestFirstFrame_NestedSplitViewTabViewText(t *testing.T) {
	// This test specifically targets the bug scenario:
	// VBox > SplitView(H) > [left: Text, right: SplitView(V) > [top: TabView, bottom: TabView]]
	// where text inside both TabViews must render on the first frame.

	root := buildIDELayout()
	w, canvas := setupFirstFrameWindow(root)

	w.DrawTo(canvas)

	// We expect these specific strings to be drawn (the selected tab content
	// for each TabView, plus the Explorer text).
	requiredTexts := []string{
		"Explorer",        // Left panel text.
		"package main",    // Top TabView, first tab content (selected by default).
		"Build succeeded", // Bottom TabView, first tab content (selected by default).
	}

	drawnSet := make(map[string]bool)
	for _, call := range canvas.drawTextCalls {
		drawnSet[call.text] = true
	}

	for _, text := range requiredTexts {
		if !drawnSet[text] {
			t.Errorf("required text %q not drawn on first frame", text)
		}
	}

	// Also verify that the tab labels were drawn.
	tabLabels := []string{"main.go", "go.mod", "Output", "Terminal"}
	for _, label := range tabLabels {
		if !drawnSet[label] {
			t.Errorf("tab label %q not drawn on first frame", label)
		}
	}
}

func TestFirstFrame_TextBoundsNonZero(t *testing.T) {
	// Verify that text drawn on the first frame has non-empty bounds,
	// which would indicate the layout pass correctly propagated sizes.

	root := buildIDELayout()
	w, canvas := setupFirstFrameWindow(root)

	w.DrawTo(canvas)

	for _, call := range canvas.drawTextCalls {
		if call.bounds.IsEmpty() {
			t.Errorf("DrawText(%q) called with empty bounds %v", call.text, call.bounds)
		}
	}
}

func TestFirstFrame_DrawTextCountMatchesVisibleTextWidgets(t *testing.T) {
	// Only selected tab content should be drawn (lazy rendering).
	// With 2 TabViews each with 2 tabs (index 0 selected), we expect:
	//   - "Explorer" (left panel)
	//   - "package main" (top TabView, tab 0 content)
	//   - "Build succeeded" (bottom TabView, tab 0 content)
	//   - Tab labels: "main.go", "go.mod", "Output", "Terminal" (4 labels)
	// Total: at least 3 content texts + 4 tab labels = 7 DrawText calls minimum.

	root := buildIDELayout()
	w, canvas := setupFirstFrameWindow(root)

	w.DrawTo(canvas)

	// We should have at least 7 DrawText calls for content + labels.
	const minExpected = 7
	if len(canvas.drawTextCalls) < minExpected {
		t.Errorf("DrawText call count = %d, want >= %d", len(canvas.drawTextCalls), minExpected)
		for i, call := range canvas.drawTextCalls {
			t.Logf("  call[%d]: %q bounds=%v", i, call.text, call.bounds)
		}
	}
}

func TestFirstFrame_WalkTreeCoverage(t *testing.T) {
	root := buildIDELayout()
	a := New()
	w := a.Window()
	w.SetRoot(root)
	w.HandleResize(1024, 700)
	w.Frame()

	all := walkTree(w.Root())

	// Verify we find all expected widget types.
	var (
		boxCount     int
		splitCount   int
		tabViewCount int
		textCount    int
	)

	for _, wgt := range all {
		switch wgt.(type) {
		case *primitives.BoxWidget:
			boxCount++
		case *splitview.Widget:
			splitCount++
		case *tabview.Widget:
			tabViewCount++
		case *primitives.TextWidget:
			textCount++
		}
	}

	if boxCount < 1 {
		t.Errorf("BoxWidget count = %d, want >= 1", boxCount)
	}
	if splitCount < 2 {
		t.Errorf("SplitView count = %d, want >= 2", splitCount)
	}
	if tabViewCount < 2 {
		t.Errorf("TabView count = %d, want >= 2", tabViewCount)
	}
	if textCount < 5 {
		// "Explorer" + 4 tab content texts.
		t.Errorf("TextWidget count = %d, want >= 5", textCount)
	}

	t.Logf("tree: %d widgets (box=%d, split=%d, tabview=%d, text=%d)",
		len(all), boxCount, splitCount, tabViewCount, textCount)
}

func TestFirstFrame_SecondFrameNoLayoutIfClean(t *testing.T) {
	// Verify that a second Frame() does NOT re-layout when nothing changed.
	// This ensures the first frame is complete and self-sufficient.

	root := buildIDELayout()
	a := New()
	w := a.Window()
	w.SetRoot(root)
	w.HandleResize(1024, 700)

	// First frame.
	w.Frame()

	firstCanvas := &trackingCanvas{}
	w.DrawTo(firstCanvas)
	firstDrawCount := len(firstCanvas.drawTextCalls)

	// Second frame: nothing changed.
	w.Frame()

	secondCanvas := &trackingCanvas{}
	w.DrawTo(secondCanvas)
	secondDrawCount := len(secondCanvas.drawTextCalls)

	// Both frames should produce the same number of DrawText calls.
	if firstDrawCount != secondDrawCount {
		t.Errorf("DrawText count changed between frames: first=%d, second=%d",
			firstDrawCount, secondDrawCount)
	}

	// Verify same content was drawn.
	if firstDrawCount > 0 && secondDrawCount > 0 {
		firstTexts := make(map[string]bool)
		for _, call := range firstCanvas.drawTextCalls {
			firstTexts[call.text] = true
		}
		for _, call := range secondCanvas.drawTextCalls {
			if !firstTexts[call.text] {
				t.Errorf("second frame drew text %q not in first frame", call.text)
			}
		}
	}
}

func TestFirstFrame_SimpleTextWidget(t *testing.T) {
	// Baseline test: a single TextWidget should render on first frame.
	// If this fails, the issue is fundamental to TextWidget rendering.

	txt := primitives.Text("Hello World").FontSize(16)
	root := primitives.VBox(txt)

	w, canvas := setupFirstFrameWindow(root)
	w.DrawTo(canvas)

	found := false
	for _, call := range canvas.drawTextCalls {
		if call.text == "Hello World" {
			found = true
			if call.bounds.IsEmpty() {
				t.Error("Hello World drawn with empty bounds")
			}
			break
		}
	}

	if !found {
		t.Error("simple TextWidget not drawn on first frame")
	}
}

func TestFirstFrame_TabViewContentBoundsSet(t *testing.T) {
	// Verify that TabView sets bounds on its selected content widget
	// during the first Layout pass.

	contentText := primitives.Text("Tab Content")

	tv := tabview.New([]tabview.Tab{
		{Label: "Tab1", Content: contentText},
	})

	root := primitives.VBox(tv)

	a := New()
	w := a.Window()
	w.SetRoot(root)
	w.HandleResize(800, 600)
	w.Frame()

	bounds := contentText.Bounds()
	if bounds.IsEmpty() {
		t.Error("TabView content text has empty bounds after first Frame()")
	}

	canvas := &trackingCanvas{}
	w.DrawTo(canvas)

	found := false
	for _, call := range canvas.drawTextCalls {
		if call.text == "Tab Content" {
			found = true
			if call.bounds.IsEmpty() {
				t.Error("Tab Content drawn with empty bounds")
			}
			break
		}
	}

	if !found {
		t.Errorf("TabView content text not drawn; DrawText calls: %d", len(canvas.drawTextCalls))
	}
}

func TestFirstFrame_SplitViewChildBoundsSet(t *testing.T) {
	// Verify that SplitView sets bounds on both children during first Layout.

	leftText := primitives.Text("Left Panel")
	rightText := primitives.Text("Right Panel")

	sv := splitview.New(
		splitview.First(leftText),
		splitview.Second(rightText),
		splitview.OrientationOpt(splitview.Horizontal),
	)

	root := primitives.VBox(sv)

	a := New()
	w := a.Window()
	w.SetRoot(root)
	w.HandleResize(800, 600)
	w.Frame()

	if leftText.Bounds().IsEmpty() {
		t.Error("SplitView first child has empty bounds after first Frame()")
	}
	if rightText.Bounds().IsEmpty() {
		t.Error("SplitView second child has empty bounds after first Frame()")
	}

	canvas := &trackingCanvas{}
	w.DrawTo(canvas)

	drawnSet := make(map[string]bool)
	for _, call := range canvas.drawTextCalls {
		drawnSet[call.text] = true
	}

	if !drawnSet["Left Panel"] {
		t.Error("Left Panel not drawn on first frame")
	}
	if !drawnSet["Right Panel"] {
		t.Error("Right Panel not drawn on first frame")
	}
}
