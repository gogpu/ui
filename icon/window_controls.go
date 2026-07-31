package icon

// Window control icons for title bar buttons.
// Based on JetBrains IntelliJ Platform icons (Apache 2.0).
// Uses filled rect/path SVGs for pixel-perfect rendering at all DPI scales.

// WindowMinimize is a horizontal filled rectangle icon for the minimize button.
var WindowMinimize = FromSVGXML("window_minimize", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<rect x="3" y="8" width="10" height="1" fill="#CED0D6"/>
</svg>`))

// WindowMaximize is a stroked square icon for the maximize button.
var WindowMaximize = FromSVGXML("window_maximize", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<rect x="3.5" y="3.5" width="9" height="9" stroke="#CED0D6"/>
</svg>`))

// WindowRestore is two overlapping rectangles icon for the restore button.
// Uses filled paths with even-odd fill rule (not strokes) for crisp edges.
var WindowRestore = FromSVGXML("window_restore", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path fill-rule="evenodd" clip-rule="evenodd" d="M5 3H13V11H10V10H12V4H6V6H5V3Z" fill="#CED0D6"/>
<path fill-rule="evenodd" clip-rule="evenodd" d="M11 5H3V13H11V5ZM10 6H4V12H10V6Z" fill="#CED0D6"/>
</svg>`))

// WindowClose is an X shape icon for the close button.
// Uses two rotated filled rectangles for crisp diagonal lines.
var WindowClose = FromSVGXML("window_close", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<rect x="3.75781" y="3.05029" width="13" height="1" transform="rotate(45 3.75781 3.05029)" fill="#CED0D6"/>
<rect width="13" height="1" transform="matrix(-0.707107 0.707107 0.707107 0.707107 12.2432 3.05029)" fill="#CED0D6"/>
</svg>`))
