package icon

// expui toolwindow icons from JetBrains IntelliJ (Apache 2.0, 16x16 viewBox).
// These are the New UI (expui) outline-style icons used in tool window sidebars.
// Full SVG XML rendered via gg/svg.RenderWithColor for pixel-perfect quality.

// ToolProject is the Project tool window icon (folder outline).
var ToolProject = FromSVGXML("tool_project", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M8.15132 4.35836L8.29689 4.5H8.5H13C13.8284 4.5 14.5 5.17157 14.5 6V12.1333C14.5 12.919 13.9104 13.5 13.25 13.5H2.75C2.08955 13.5 1.5 12.919 1.5 12.1333V3.86667C1.5 3.08099 2.08955 2.5 2.75 2.5H6.03823C6.16847 2.5 6.29357 2.55082 6.38691 2.64164L8.15132 4.35836Z" stroke="#6C707E"/>
</svg>`))

// ToolCommit is the Commit tool window icon (circle on horizontal line).
var ToolCommit = FromSVGXML("tool_commit", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path fill-rule="evenodd" clip-rule="evenodd" d="M8 10C9.10457 10 10 9.10457 10 8C10 6.89543 9.10457 6 8 6C6.89543 6 6 6.89543 6 8C6 9.10457 6.89543 10 8 10ZM10.9585 7.5C10.7205 6.08114 9.4865 5 8 5C6.5135 5 5.27952 6.08114 5.04148 7.5H0.5C0.223858 7.5 0 7.72386 0 8C0 8.27614 0.223858 8.5 0.5 8.5H5.04148C5.27952 9.91886 6.5135 11 8 11C9.4865 11 10.7205 9.91886 10.9585 8.5H15.5C15.7761 8.5 16 8.27614 16 8C16 7.72386 15.7761 7.5 15.5 7.5H10.9585Z" fill="#6C707E"/>
</svg>`))

// ToolStructure is the Structure tool window icon (tree with two blocks).
var ToolStructure = FromSVGXML("tool_structure", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path fill-rule="evenodd" clip-rule="evenodd" d="M3 1.5C3 1.22386 2.77614 1 2.5 1C2.22386 1 2 1.22386 2 1.5V14.5C2 14.7761 2.22386 15 2.5 15C2.77614 15 3 14.7761 3 14.5V12H5V13C5 14.1046 5.89543 15 7 15H12C13.1046 15 14 14.1046 14 13V11C14 9.89543 13.1046 9 12 9H7C5.89543 9 5 9.89543 5 11H3V4H5V5C5 6.10457 5.89543 7 7 7H12C13.1046 7 14 6.10457 14 5V3C14 1.89543 13.1046 1 12 1H7C5.89543 1 5 1.89543 5 3L3 3V1.5ZM6 5C6 5.55228 6.44772 6 7 6H12C12.5523 6 13 5.55228 13 5V3C13 2.44772 12.5523 2 12 2H7C6.44772 2 6 2.44772 6 3V5ZM6 13C6 13.5523 6.44772 14 7 14H12C12.5523 14 13 13.5523 13 13V11C13 10.4477 12.5523 10 12 10H7C6.44772 10 6 10.4477 6 11V13Z" fill="#6C707E"/>
</svg>`))

// ToolServices is the Services tool window icon (hexagon with play).
var ToolServices = FromSVGXML("tool_services", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M15.2117 7.50028C15.3901 7.80954 15.3901 8.19046 15.2117 8.49972L12.0386 13.9997C11.86 14.3093 11.5298 14.5 11.1724 14.5L4.82756 14.5C4.47018 14.5 4.13997 14.3093 3.96138 13.9997L0.788301 8.49972C0.609882 8.19046 0.609882 7.80954 0.788301 7.50028L3.96138 2.00028C4.13997 1.69072 4.47018 1.5 4.82756 1.5L11.1724 1.5C11.5298 1.5 11.86 1.69072 12.0386 2.00028L15.2117 7.50028Z" stroke="#6C707E"/>
<path d="M10.5 8.43301L6.75 10.5981C6.41667 10.7905 6 10.55 6 10.1651L6 5.83493C6 5.45004 6.41667 5.20947 6.75 5.40192L10.5 7.56699C10.8333 7.75944 10.8333 8.24056 10.5 8.43301Z" stroke="#6C707E"/>
</svg>`))

// ToolProblems is the Problems tool window icon (circle with exclamation).
var ToolProblems = FromSVGXML("tool_problems", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<circle cx="8" cy="8" r="6.5" stroke="#6C707E"/>
<path d="M8 4.59998L8 8.39998" stroke="#6C707E" stroke-width="1.2" stroke-linecap="round"/>
<circle cx="8.00078" cy="10.7" r="0.5" fill="#6C707E" stroke="#6C707E" stroke-width="0.4"/>
</svg>`))

// ToolTerminal is the Terminal tool window icon.
var ToolTerminal = Terminal

// ToolGit is the Git/VCS tool window icon (branch with circles).
var ToolGit = FromSVGXML("tool_git", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<circle cx="4.5" cy="4" r="2" stroke="#6C707E"/>
<path d="M4.5 11.5H8.5C9.60457 11.5 10.5 10.6046 10.5 9.5V9.5V8" stroke="#6C707E"/>
<path d="M4.5 6.5L4.5 14.5" stroke="#6C707E" stroke-linecap="round" stroke-linejoin="round"/>
<circle cx="10.5" cy="6" r="2" stroke="#6C707E"/>
</svg>`))

// ToolNotifications is the Notifications tool window icon (bell).
var ToolNotifications = FromSVGXML("tool_notifications", []byte(`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M6.95674 14.4579H9.04941C8.99287 14.5871 8.91235 14.7059 8.81068 14.8076C8.59649 15.0218 8.30599 15.1421 8.00308 15.1421C7.70016 15.1421 7.40966 15.0218 7.19547 14.8076C7.0938 14.7059 7.01329 14.5871 6.95674 14.4579Z" stroke="#6C707E" stroke-width="0.915741"/>
<path d="M3.9472 8.2236C3.98191 8.15418 3.99999 8.07762 3.99999 8V6C3.99999 4.29256 4.57108 3.18513 5.32407 2.50006C6.0881 1.80496 7.08663 1.50196 8.00113 1.50001L8.00388 1.5L8.00449 1.5L8.00462 1.5H8.00543L8.00617 1.5L8.00892 1.50001C8.92327 1.50196 9.91901 1.80486 10.6804 2.49958C11.431 3.1844 12 4.2919 12 6V8C12 8.07762 12.0181 8.15418 12.0528 8.2236L13.8492 11.8164C14.0062 12.1305 13.7779 12.5 13.4267 12.5H2.57326C2.22214 12.5 1.99376 12.1305 2.15079 11.8164L3.9472 8.2236Z" stroke="#6C707E" stroke-linejoin="round"/>
</svg>`))
