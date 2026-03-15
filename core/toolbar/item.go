package toolbar

import (
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// ItemKind identifies the type of toolbar item.
type ItemKind uint8

const (
	// ItemButton is a clickable icon button with optional text label.
	ItemButton ItemKind = iota

	// ItemSeparator is a vertical divider line between button groups.
	ItemSeparator

	// ItemSpacer is a flexible gap that pushes subsequent items to the right.
	ItemSpacer

	// ItemCustom embeds an arbitrary widget in the toolbar.
	ItemCustom
)

// itemKindNames maps each ItemKind to its human-readable name.
var itemKindNames = [...]string{
	ItemButton:    "Button",
	ItemSeparator: "Separator",
	ItemSpacer:    "Spacer",
	ItemCustom:    "Custom",
}

// unknownStr is the fallback string for unknown item kinds.
const unknownStr = "Unknown"

// String returns a human-readable name for the item kind.
func (k ItemKind) String() string {
	if int(k) < len(itemKindNames) {
		return itemKindNames[k]
	}
	return unknownStr
}

// Item represents a single element within a toolbar.
//
// Use the convenience constructors [IconButton], [TextButton], [Separator],
// [Spacer], and [Custom] to create items.
type Item struct {
	// Kind identifies the type of this item.
	Kind ItemKind

	// Label is the human-readable text for button items. Used for
	// accessibility announcements and optional visual display.
	Label string

	// Icon is the vector icon data for button items.
	Icon icon.IconData

	// OnClick is the callback invoked when a button item is activated.
	OnClick func()

	// Widget is the embedded widget for ItemCustom items.
	Widget widget.Widget

	// Enabled controls whether a button item responds to interaction.
	// Separators and spacers ignore this field.
	Enabled bool

	// ShowLabel controls whether the text label is displayed next to the icon.
	// When false (default for IconButton), only the icon is shown.
	ShowLabel bool
}

// IconButton creates a button item with an icon and label.
// The label is used for accessibility but not displayed by default.
// Use [TextIconButton] to show both icon and label.
func IconButton(label string, ic icon.IconData, onClick func()) Item {
	return Item{
		Kind:    ItemButton,
		Label:   label,
		Icon:    ic,
		OnClick: onClick,
		Enabled: true,
	}
}

// TextIconButton creates a button item that displays both icon and label text.
func TextIconButton(label string, ic icon.IconData, onClick func()) Item {
	return Item{
		Kind:      ItemButton,
		Label:     label,
		Icon:      ic,
		OnClick:   onClick,
		Enabled:   true,
		ShowLabel: true,
	}
}

// Separator creates a vertical divider item.
func Separator() Item {
	return Item{Kind: ItemSeparator, Enabled: true}
}

// Spacer creates a flexible gap item that pushes subsequent items to the right.
func Spacer() Item {
	return Item{Kind: ItemSpacer, Enabled: true}
}

// Custom creates an item that embeds an arbitrary widget.
func Custom(w widget.Widget) Item {
	return Item{Kind: ItemCustom, Widget: w, Enabled: true}
}
