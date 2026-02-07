package a11y

// Role represents the semantic purpose of a UI element for assistive technology.
//
// Roles are derived from the WAI-ARIA specification and AccessKit. They tell
// screen readers and other assistive technology how to present and interact
// with each element. Every accessible node must have exactly one role.
//
// The role determines what properties, states, and actions are valid for a node.
// For example, a [RoleSlider] node is expected to have numeric value properties,
// while a [RoleButton] node supports a click action.
type Role uint8

// unknownStr is the string representation for unknown/unrecognized values.
const unknownStr = "Unknown"

// Structural roles define the layout and organization of the UI.
const (
	// RoleUnknown indicates an element whose role is not known.
	// This should only be used as a fallback when no other role applies.
	RoleUnknown Role = iota

	// RoleWindow represents a top-level application window.
	RoleWindow

	// RoleGroup represents a generic grouping of related elements.
	RoleGroup

	// RoleSeparator represents a visual or logical divider between sections.
	RoleSeparator

	// RoleToolbar represents a collection of commonly used function buttons
	// or controls.
	RoleToolbar

	// RoleStatusBar represents a bar that displays status information,
	// typically at the bottom of a window.
	RoleStatusBar

	// RoleMenuBar represents a horizontal bar containing menu items.
	RoleMenuBar

	// RoleGenericContainer represents a container with no specific semantics.
	// Use [RoleGroup] instead when elements are logically related.
	RoleGenericContainer
)

// Input roles define interactive elements that accept user input.
const (
	// RoleButton represents a clickable button that triggers an action.
	RoleButton Role = iota + 20

	// RoleCheckbox represents a control with checked, unchecked, or mixed state.
	RoleCheckbox

	// RoleRadio represents a radio button within a group where only one
	// can be selected at a time.
	RoleRadio

	// RoleTextField represents a single-line text input field.
	RoleTextField

	// RoleTextArea represents a multi-line text input field.
	RoleTextArea

	// RoleSlider represents a control for selecting a value from a continuous range.
	RoleSlider

	// RoleSwitch represents a toggle control with on/off state.
	RoleSwitch

	// RoleComboBox represents a composite widget combining a text field with
	// a popup list of choices.
	RoleComboBox

	// RoleSpinButton represents a numeric input with increment/decrement controls.
	RoleSpinButton

	// RoleRadioGroup represents a group of radio buttons where only one
	// can be selected.
	RoleRadioGroup

	// RoleSearchBox represents a text field specifically for search input.
	RoleSearchBox

	// RoleToggleButton represents a button that can be toggled on or off.
	RoleToggleButton

	// RoleColorWell represents a control for selecting a color value.
	RoleColorWell
)

// Display roles define elements that present information to the user.
const (
	// RoleLabel represents a text label, often associated with a form control.
	RoleLabel Role = iota + 50

	// RoleImage represents a graphical image.
	RoleImage

	// RoleProgressBar represents a progress indicator showing completion
	// of a long-running operation.
	RoleProgressBar

	// RoleTooltip represents a small popup that provides additional context
	// when hovering over an element.
	RoleTooltip

	// RoleAlert represents an important message that demands the user's attention.
	RoleAlert

	// RoleBadge represents a small status descriptor, such as a notification count.
	RoleBadge

	// RoleHeading represents a heading that labels a section of content.
	RoleHeading

	// RoleMeter represents a scalar measurement within a known range,
	// such as disk usage or a gauge.
	RoleMeter

	// RoleStaticText represents non-interactive text content.
	RoleStaticText
)

// Container roles define elements that contain and organize other elements.
const (
	// RoleDialog represents a modal or non-modal dialog window.
	RoleDialog Role = iota + 70

	// RoleAlertDialog represents a dialog that conveys an urgent message
	// and requires user response.
	RoleAlertDialog

	// RoleMenu represents a popup menu offering a list of choices.
	RoleMenu

	// RoleMenuItem represents a single item within a menu.
	RoleMenuItem

	// RoleMenuItemCheckbox represents a checkable menu item.
	RoleMenuItemCheckbox

	// RoleMenuItemRadio represents a radio-selectable menu item within a group.
	RoleMenuItemRadio

	// RoleList represents an ordered or unordered list of items.
	RoleList

	// RoleListItem represents a single item within a list.
	RoleListItem

	// RoleTree represents a hierarchical tree view.
	RoleTree

	// RoleTreeItem represents a single item within a tree view.
	RoleTreeItem

	// RoleTab represents a selectable tab within a tab list.
	RoleTab

	// RoleTabList represents a list of tabs for switching between panels.
	RoleTabList

	// RoleTabPanel represents the content panel associated with a tab.
	RoleTabPanel

	// RoleGrid represents a two-dimensional grid of interactive cells.
	RoleGrid

	// RoleGridCell represents a single cell within a grid.
	RoleGridCell

	// RoleTable represents a data table with rows and columns.
	RoleTable

	// RoleRow represents a row within a table or grid.
	RoleRow

	// RoleCell represents a single cell within a table row.
	RoleCell

	// RoleColumnHeader represents a header cell for a table column.
	RoleColumnHeader

	// RoleRowHeader represents a header cell for a table row.
	RoleRowHeader

	// RoleListBox represents a list widget from which the user can select
	// one or more items.
	RoleListBox

	// RoleScrollView represents a scrollable container.
	RoleScrollView

	// RoleApplication represents a region declared as a web application
	// (used for complex interactive widgets).
	RoleApplication

	// RoleDocument represents a document content region.
	RoleDocument

	// RoleFeed represents a scrollable list of articles that grows dynamically.
	RoleFeed
)

// Navigation roles define elements that help the user navigate the UI.
const (
	// RoleLink represents a navigational hyperlink.
	RoleLink Role = iota + 110

	// RoleScrollBar represents a scrollbar control.
	RoleScrollBar

	// RoleNavigation represents a navigation landmark region.
	RoleNavigation

	// RoleBanner represents a banner landmark region, typically site-wide.
	RoleBanner

	// RoleMain represents the main content landmark region.
	RoleMain

	// RoleContentInfo represents informational content about the page,
	// typically a footer.
	RoleContentInfo

	// RoleComplementary represents a complementary landmark region
	// that supports the main content.
	RoleComplementary

	// RoleRegion represents a generic landmark region of significance.
	RoleRegion

	// RoleForm represents a form landmark region.
	RoleForm

	// RoleSearch represents a search landmark region.
	RoleSearch
)

// roleNames maps each Role to its human-readable name.
var roleNames = map[Role]string{
	// Structural
	RoleUnknown:          unknownStr,
	RoleWindow:           "Window",
	RoleGroup:            "Group",
	RoleSeparator:        "Separator",
	RoleToolbar:          "Toolbar",
	RoleStatusBar:        "StatusBar",
	RoleMenuBar:          "MenuBar",
	RoleGenericContainer: "GenericContainer",

	// Input
	RoleButton:       "Button",
	RoleCheckbox:     "Checkbox",
	RoleRadio:        "Radio",
	RoleTextField:    "TextField",
	RoleTextArea:     "TextArea",
	RoleSlider:       "Slider",
	RoleSwitch:       "Switch",
	RoleComboBox:     "ComboBox",
	RoleSpinButton:   "SpinButton",
	RoleRadioGroup:   "RadioGroup",
	RoleSearchBox:    "SearchBox",
	RoleToggleButton: "ToggleButton",
	RoleColorWell:    "ColorWell",

	// Display
	RoleLabel:       "Label",
	RoleImage:       "Image",
	RoleProgressBar: "ProgressBar",
	RoleTooltip:     "Tooltip",
	RoleAlert:       "Alert",
	RoleBadge:       "Badge",
	RoleHeading:     "Heading",
	RoleMeter:       "Meter",
	RoleStaticText:  "StaticText",

	// Container
	RoleDialog:           "Dialog",
	RoleAlertDialog:      "AlertDialog",
	RoleMenu:             "Menu",
	RoleMenuItem:         "MenuItem",
	RoleMenuItemCheckbox: "MenuItemCheckbox",
	RoleMenuItemRadio:    "MenuItemRadio",
	RoleList:             "List",
	RoleListItem:         "ListItem",
	RoleTree:             "Tree",
	RoleTreeItem:         "TreeItem",
	RoleTab:              "Tab",
	RoleTabList:          "TabList",
	RoleTabPanel:         "TabPanel",
	RoleGrid:             "Grid",
	RoleGridCell:         "GridCell",
	RoleTable:            "Table",
	RoleRow:              "Row",
	RoleCell:             "Cell",
	RoleColumnHeader:     "ColumnHeader",
	RoleRowHeader:        "RowHeader",
	RoleListBox:          "ListBox",
	RoleScrollView:       "ScrollView",
	RoleApplication:      "Application",
	RoleDocument:         "Document",
	RoleFeed:             "Feed",

	// Navigation
	RoleLink:          "Link",
	RoleScrollBar:     "ScrollBar",
	RoleNavigation:    "Navigation",
	RoleBanner:        "Banner",
	RoleMain:          "Main",
	RoleContentInfo:   "ContentInfo",
	RoleComplementary: "Complementary",
	RoleRegion:        "Region",
	RoleForm:          "Form",
	RoleSearch:        "Search",
}

// String returns a human-readable name for the role.
//
// The returned string matches the role constant name without the "Role" prefix.
// For example, RoleButton returns "Button" and RoleCheckbox returns "Checkbox".
func (r Role) String() string {
	if name, ok := roleNames[r]; ok {
		return name
	}
	return unknownStr
}

// IsInteractive returns true if the role represents an element that accepts
// user input, such as buttons, text fields, and sliders.
func (r Role) IsInteractive() bool {
	return r >= RoleButton && r <= RoleColorWell
}

// IsContainer returns true if the role represents an element that contains
// other accessible elements, such as dialogs, lists, and tables.
func (r Role) IsContainer() bool {
	return r >= RoleDialog && r <= RoleFeed
}

// IsLandmark returns true if the role represents a navigational landmark region,
// such as main content, navigation, or search.
func (r Role) IsLandmark() bool {
	switch r {
	case RoleNavigation, RoleBanner, RoleMain, RoleContentInfo,
		RoleComplementary, RoleRegion, RoleForm, RoleSearch:
		return true
	default:
		return false
	}
}
