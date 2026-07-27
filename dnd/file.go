package dnd

// KindFile identifies OS file drag-and-drop data.
//
// When files are dragged from the OS file manager (Explorer, Finder, Nautilus)
// and dropped on the application window, the DragData.Kind is set to KindFile
// and the Payload is a [FilePayload] containing the file paths.
const KindFile = "file"

// FilePayload carries file paths from OS drag-and-drop events.
//
// The Paths field contains absolute paths to the dropped files or directories,
// as provided by the operating system's drag-and-drop protocol (WM_DROPFILES
// on Windows, NSDraggingDestination on macOS, XDND on X11, wl_data_device
// on Wayland).
type FilePayload struct {
	// Paths contains the absolute file system paths of the dropped items.
	Paths []string
}
