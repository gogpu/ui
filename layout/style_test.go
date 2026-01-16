package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestDisplay_String(t *testing.T) {
	tests := []struct {
		d    Display
		want string
	}{
		{DisplayFlex, "Flex"},
		{DisplayGrid, "Grid"},
		{DisplayBlock, "Block"},
		{DisplayNone, "None"},
		{Display(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("Display(%d).String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFlexDirection_String(t *testing.T) {
	tests := []struct {
		d    FlexDirection
		want string
	}{
		{FlexRow, "Row"},
		{FlexRowReverse, "RowReverse"},
		{FlexColumn, "Column"},
		{FlexColumnReverse, "ColumnReverse"},
		{FlexDirection(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("FlexDirection(%d).String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFlexDirection_IsHorizontal(t *testing.T) {
	tests := []struct {
		d    FlexDirection
		want bool
	}{
		{FlexRow, true},
		{FlexRowReverse, true},
		{FlexColumn, false},
		{FlexColumnReverse, false},
	}

	for _, tt := range tests {
		if got := tt.d.IsHorizontal(); got != tt.want {
			t.Errorf("FlexDirection(%d).IsHorizontal() = %v, want %v", tt.d, got, tt.want)
		}
	}
}

func TestFlexDirection_IsReversed(t *testing.T) {
	tests := []struct {
		d    FlexDirection
		want bool
	}{
		{FlexRow, false},
		{FlexRowReverse, true},
		{FlexColumn, false},
		{FlexColumnReverse, true},
	}

	for _, tt := range tests {
		if got := tt.d.IsReversed(); got != tt.want {
			t.Errorf("FlexDirection(%d).IsReversed() = %v, want %v", tt.d, got, tt.want)
		}
	}
}

func TestFlexWrap_String(t *testing.T) {
	tests := []struct {
		w    FlexWrap
		want string
	}{
		{FlexNoWrap, "NoWrap"},
		{FlexWrapOn, "Wrap"},
		{FlexWrapReverse, "WrapReverse"},
		{FlexWrap(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.w.String(); got != tt.want {
			t.Errorf("FlexWrap(%d).String() = %q, want %q", tt.w, got, tt.want)
		}
	}
}

func TestJustifyContent_String(t *testing.T) {
	tests := []struct {
		j    JustifyContent
		want string
	}{
		{JustifyStart, "Start"},
		{JustifyEnd, "End"},
		{JustifyCenter, "Center"},
		{JustifySpaceBetween, "SpaceBetween"},
		{JustifySpaceAround, "SpaceAround"},
		{JustifySpaceEvenly, "SpaceEvenly"},
		{JustifyContent(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.j.String(); got != tt.want {
			t.Errorf("JustifyContent(%d).String() = %q, want %q", tt.j, got, tt.want)
		}
	}
}

func TestAlignItems_String(t *testing.T) {
	tests := []struct {
		a    AlignItems
		want string
	}{
		{AlignItemsStart, "Start"},
		{AlignItemsEnd, "End"},
		{AlignItemsCenter, "Center"},
		{AlignItemsStretch, "Stretch"},
		{AlignItemsBaseline, "Baseline"},
		{AlignItems(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.a.String(); got != tt.want {
			t.Errorf("AlignItems(%d).String() = %q, want %q", tt.a, got, tt.want)
		}
	}
}

func TestAlignContent_String(t *testing.T) {
	tests := []struct {
		a    AlignContent
		want string
	}{
		{AlignContentStart, "Start"},
		{AlignContentEnd, "End"},
		{AlignContentCenter, "Center"},
		{AlignContentStretch, "Stretch"},
		{AlignContentSpaceBetween, "SpaceBetween"},
		{AlignContentSpaceAround, "SpaceAround"},
		{AlignContent(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.a.String(); got != tt.want {
			t.Errorf("AlignContent(%d).String() = %q, want %q", tt.a, got, tt.want)
		}
	}
}

func TestDimension_Constructors(t *testing.T) {
	auto := Auto()
	if !auto.IsAuto() {
		t.Error("Auto() should create auto dimension")
	}

	px := Px(100)
	if px.Unit != DimensionPixels || px.Value != 100 {
		t.Errorf("Px(100) = {%v, %v}, want {DimensionPixels, 100}", px.Unit, px.Value)
	}

	pct := Pct(50)
	if pct.Unit != DimensionPercent || pct.Value != 50 {
		t.Errorf("Pct(50) = {%v, %v}, want {DimensionPercent, 50}", pct.Unit, pct.Value)
	}
}

func TestDimension_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		dim       Dimension
		reference float32
		fallback  float32
		want      float32
	}{
		{"auto uses fallback", Auto(), 200, 50, 50},
		{"pixels returns value", Px(100), 200, 50, 100},
		{"percent calculates", Pct(50), 200, 50, 100},
		{"percent 25%", Pct(25), 400, 0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dim.Resolve(tt.reference, tt.fallback)
			if got != tt.want {
				t.Errorf("Dimension.Resolve(%v, %v) = %v, want %v", tt.reference, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestDefaultStyle(t *testing.T) {
	s := DefaultStyle()

	if s.Display != DisplayFlex {
		t.Errorf("Display = %v, want DisplayFlex", s.Display)
	}
	if s.FlexDirection != FlexRow {
		t.Errorf("FlexDirection = %v, want FlexRow", s.FlexDirection)
	}
	if s.FlexShrink != 1 {
		t.Errorf("FlexShrink = %v, want 1", s.FlexShrink)
	}
	if !s.Width.IsAuto() {
		t.Error("Width should be auto")
	}
}

func TestStyle_WithMethods(t *testing.T) {
	s := DefaultStyle()

	s = s.WithDisplay(DisplayGrid)
	if s.Display != DisplayGrid {
		t.Errorf("after WithDisplay: Display = %v", s.Display)
	}

	s = s.WithFlexDirection(FlexColumn)
	if s.FlexDirection != FlexColumn {
		t.Errorf("after WithFlexDirection: FlexDirection = %v", s.FlexDirection)
	}

	s = s.WithJustifyContent(JustifyCenter)
	if s.JustifyContent != JustifyCenter {
		t.Errorf("after WithJustifyContent: JustifyContent = %v", s.JustifyContent)
	}

	s = s.WithAlignItems(AlignItemsCenter)
	if s.AlignItems != AlignItemsCenter {
		t.Errorf("after WithAlignItems: AlignItems = %v", s.AlignItems)
	}

	s = s.WithFlex(2, 3, Px(100))
	if s.FlexGrow != 2 || s.FlexShrink != 3 || s.FlexBasis.Value != 100 {
		t.Errorf("after WithFlex: Grow=%v, Shrink=%v, Basis=%v", s.FlexGrow, s.FlexShrink, s.FlexBasis)
	}

	s = s.WithSize(Px(200), Px(150))
	if s.Width.Value != 200 || s.Height.Value != 150 {
		t.Errorf("after WithSize: Width=%v, Height=%v", s.Width, s.Height)
	}

	margin := geometry.UniformInsets(10)
	s = s.WithMargin(margin)
	if s.Margin != margin {
		t.Errorf("after WithMargin: Margin=%v", s.Margin)
	}

	padding := geometry.UniformInsets(5)
	s = s.WithPadding(padding)
	if s.Padding != padding {
		t.Errorf("after WithPadding: Padding=%v", s.Padding)
	}

	s = s.WithGap(8)
	if s.Gap != 8 {
		t.Errorf("after WithGap: Gap=%v", s.Gap)
	}
}
