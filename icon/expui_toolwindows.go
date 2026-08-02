package icon

// expui toolwindow icons from JetBrains IntelliJ (Apache 2.0, @20x20 viewBox).
// These are the New UI (expui) outline-style icons used in tool window sidebars.
// Native @20x20 variants from JetBrains reference — no fractional scaling needed.
// Full SVG XML rendered via gg/svg.RenderToSceneWithColor for vector quality.
//
// Stroke width: JetBrains originals use stroke-width="1.5" for @20x20 icons.
// We use stroke-width="1" because our current SVG renderer (gg/svg) produces
// ~2px visual strokes for 1.5px due to AA spreading, while Skia (used by
// JetBrains) renders 1.5px crisply via stroke hinting. With stroke-width="1",
// our renderer produces crisp 1px strokes matching visual quality of JetBrains.
// When gg SVG renderer reaches Skia-level stroke quality (gg#463), restore
// to original stroke-width="1.5" from JetBrains @20x20 SVGs.

// ToolProject is the Project tool window icon (folder outline).
var ToolProject = FromSVGXML("tool_project", []byte(`<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M10.5199 5.57617L10.7285 5.75H11H17C17.6904 5.75 18.25 6.30964 18.25 7V15.1667C18.25 16.0671 17.553 16.75 16.75 16.75H3.25C2.44705 16.75 1.75 16.0671 1.75 15.1667V4.83333C1.75 3.93294 2.44705 3.25 3.25 3.25H7.63795C7.69643 3.25 7.75307 3.2705 7.798 3.30794L10.5199 5.57617Z" stroke="#6C707E" stroke-width="1"/>
</svg>`))

// ToolCommit is the Commit tool window icon (circle on horizontal line).
var ToolCommit = FromSVGXML("tool_commit", []byte(`<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
<path fill-rule="evenodd" clip-rule="evenodd" d="M19.25 9.25C19.6642 9.25 20 9.58579 20 10C20 10.4142 19.6642 10.75 19.25 10.75L13.9298 10.75C13.5787 12.6006 11.9528 14 10 14C8.04721 14 6.42125 12.6006 6.0702 10.75L0.75 10.75C0.335787 10.75 -1.92951e-07 10.4142 -1.74846e-07 10C-1.5674e-07 9.58579 0.335787 9.25 0.75 9.25L6.0702 9.25C6.42125 7.39935 8.04721 6 10 6C11.9528 6 13.5787 7.39935 13.9298 9.25L19.25 9.25ZM10 12.5C11.3807 12.5 12.5 11.3807 12.5 10C12.5 8.61929 11.3807 7.5 10 7.5C8.61929 7.5 7.5 8.61929 7.5 10C7.5 11.3807 8.61929 12.5 10 12.5Z" fill="#6C707E"/>
</svg>`))

// ToolStructure is the Structure tool window icon (tree with two blocks).
var ToolStructure = FromSVGXML("tool_structure", []byte(`<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
<path fill-rule="evenodd" clip-rule="evenodd" d="M2.75 1C3.16421 1 3.5 1.33579 3.5 1.75V4H6V3C6 1.89543 6.89543 1 8 1H16C17.1046 1 18 1.89543 18 3V7C18 8.10457 17.1046 9 16 9H8C6.89543 9 6 8.10457 6 7V5.5H3.5V14H6V13C6 11.8954 6.89543 11 8 11H16C17.1046 11 18 11.8954 18 13V17C18 18.1046 17.1046 19 16 19H8C6.89543 19 6 18.1046 6 17V15.5H3.5V18.25C3.5 18.6642 3.16421 19 2.75 19C2.33579 19 2 18.6642 2 18.25V1.75C2 1.33579 2.33579 1 2.75 1ZM16 12.5H8C7.72386 12.5 7.5 12.7239 7.5 13V17C7.5 17.2761 7.72386 17.5 8 17.5H16C16.2761 17.5 16.5 17.2761 16.5 17V13C16.5 12.7239 16.2761 12.5 16 12.5ZM8 2.5H16C16.2761 2.5 16.5 2.72386 16.5 3V7C16.5 7.27614 16.2761 7.5 16 7.5H8C7.72386 7.5 7.5 7.27614 7.5 7V3C7.5 2.72386 7.72386 2.5 8 2.5Z" fill="#6C707E"/>
</svg>`))

// ToolServices is the Services tool window icon (hexagon with play).
var ToolServices = FromSVGXML("tool_services", []byte(`<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M18.6913 9.4855C18.8813 9.80219 18.8813 10.1978 18.6913 10.5145L14.7913 17.0145C14.6106 17.3157 14.2851 17.5 13.9338 17.5L6.06619 17.5C5.71493 17.5 5.38942 17.3157 5.2087 17.0145L1.3087 10.5145C1.11869 10.1978 1.11869 9.80219 1.3087 9.4855L5.2087 2.9855C5.38942 2.6843 5.71493 2.5 6.06619 2.5L13.9338 2.5C14.2851 2.5 14.6106 2.6843 14.7913 2.9855L18.6913 9.4855Z" stroke="#6C707E" stroke-width="1"/>
<path d="M14.125 10.2165L8.125 13.6806C7.95833 13.7768 7.75 13.6566 7.75 13.4641L7.75 6.5359C7.75 6.34345 7.95833 6.22317 8.125 6.31939L14.125 9.78349C14.2917 9.87972 14.2917 10.1203 14.125 10.2165Z" stroke="#6C707E" stroke-width="1"/>
</svg>`))

// ToolProblems is the Problems tool window icon (circle with exclamation).
var ToolProblems = FromSVGXML("tool_problems", []byte(`<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M10 18.25C14.5563 18.25 18.25 14.5563 18.25 10C18.25 5.44365 14.5563 1.75 10 1.75C5.44365 1.75 1.75 5.44365 1.75 10C1.75 14.5563 5.44365 18.25 10 18.25Z" stroke="#6C707E" stroke-width="1"/>
<path d="M10 6V10" stroke="#6C707E" stroke-width="2" stroke-linecap="round"/>
<path d="M9.99999 15.2C10.6627 15.2 11.2 14.6627 11.2 14C11.2 13.3372 10.6627 12.8 9.99999 12.8C9.33725 12.8 8.79999 13.3372 8.79999 14C8.79999 14.6627 9.33725 15.2 9.99999 15.2Z" fill="#6C707E"/>
</svg>`))

// ToolTerminal is the Terminal tool window icon (prompt >_ symbol, 20x20).
// No @20x20 variant in JetBrains reference — coordinates scaled from 16x16
// by 1.25 to match the 20x20 viewBox used by all other tool window icons.
var ToolTerminal = FromSVGXML("tool_terminal", []byte(`<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M3.75 6.25L8.75 10L3.75 13.75" stroke="#6C707E" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"/>
<path d="M11.25 13.75H16.25" stroke="#6C707E" stroke-width="1" stroke-linecap="round"/>
</svg>`))

// ToolGit is the Git/VCS tool window icon (branch with circles).
var ToolGit = FromSVGXML("tool_git", []byte(`<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
<path fill-rule="evenodd" clip-rule="evenodd" d="M7.5 4.5C7.5 5.60457 6.60457 6.5 5.5 6.5C4.39543 6.5 3.5 5.60457 3.5 4.5C3.5 3.39543 4.39543 2.5 5.5 2.5C6.60457 2.5 7.5 3.39543 7.5 4.5ZM9 4.5C9 6.17556 7.82259 7.57612 6.25 7.91946L6.25 13.25H9.5C10.2092 13.25 10.7035 13.2496 11.0904 13.2232C11.4706 13.1973 11.692 13.1487 11.861 13.0787C12.4124 12.8504 12.8504 12.4124 13.0787 11.861C13.1487 11.692 13.1973 11.4706 13.2232 11.0904C13.244 10.7848 13.2487 10.4121 13.2497 9.91939C11.6773 9.57594 10.5 8.17546 10.5 6.5C10.5 4.567 12.067 3 14 3C15.933 3 17.5 4.567 17.5 6.5C17.5 8.17565 16.3225 9.57626 14.7498 9.91951C14.7488 10.4185 14.7439 10.8379 14.7197 11.1925C14.6886 11.6491 14.6229 12.0528 14.4645 12.4351C14.0839 13.3539 13.3539 14.0839 12.4351 14.4645C12.0528 14.6229 11.6491 14.6886 11.1925 14.7197C10.7485 14.75 10.203 14.75 9.52664 14.75H9.5H6.25V18C6.25 18.4142 5.91421 18.75 5.5 18.75C5.08579 18.75 4.75 18.4142 4.75 18L4.75 7.91946C3.17741 7.57612 2 6.17556 2 4.5C2 2.567 3.567 1 5.5 1C7.433 1 9 2.567 9 4.5ZM16 6.5C16 7.60457 15.1046 8.5 14 8.5C12.8954 8.5 12 7.60457 12 6.5C12 5.39543 12.8954 4.5 14 4.5C15.1046 4.5 16 5.39543 16 6.5Z" fill="#6C707E"/>
</svg>`))

// ToolNotifications is the Notifications tool window icon (bell).
var ToolNotifications = FromSVGXML("tool_notifications", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M6.95674 14.4579H9.04941C8.99287 14.5871 8.91235 14.7059 8.81068 14.8076C8.59649 15.0218 8.30599 15.1421 8.00308 15.1421C7.70016 15.1421 7.40966 15.0218 7.19547 14.8076C7.0938 14.7059 7.01329 14.5871 6.95674 14.4579Z" stroke="#6C707E" stroke-width="0.915741"/>
<path d="M3.9472 8.2236C3.98191 8.15418 3.99999 8.07762 3.99999 8V6C3.99999 4.29256 4.57108 3.18513 5.32407 2.50006C6.0881 1.80496 7.08663 1.50196 8.00113 1.50001L8.00388 1.5L8.00449 1.5L8.00462 1.5H8.00543L8.00617 1.5L8.00892 1.50001C8.92327 1.50196 9.91901 1.80486 10.6804 2.49958C11.431 3.1844 12 4.2919 12 6V8C12 8.07762 12.0181 8.15418 12.0528 8.2236L13.8492 11.8164C14.0062 12.1305 13.7779 12.5 13.4267 12.5H2.57326C2.22214 12.5 1.99376 12.1305 2.15079 11.8164L3.9472 8.2236Z" stroke="#6C707E" stroke-linejoin="round"/>
</svg>`))
