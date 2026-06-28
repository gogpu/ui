// IDE-like demo — GoLand/JetBrains-inspired layout with DevTools theme.
//
// Layout:
//
//	TitleBar (40px, frameless)
//	SplitView (horizontal)
//	  Left: TreeView (project files)
//	  Right: SplitView (vertical)
//	    Top: TabView (code editor tabs)
//	    Bottom: TabView (terminal/output tabs)
//	StatusBar (24px)
package main

import (
	"log"

	_ "github.com/gogpu/gg/gpu"

	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/core/stripe"
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/core/titlebar"
	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/core/treeview"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/devtools"
)

func main() {
	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("GoLand — myproject").
		WithSize(1024, 700).
		WithFrameless(true))

	dt := devtools.NewDarkTheme()
	cs := dt.Colors
	p := devtools.NewPainters(dt)

	// --- Title Bar with embedded toolbar (GoLand-style) ---
	noop := func() {} // placeholder click handler
	mainToolbar := toolbar.New(
		toolbar.Items(
			toolbar.IconButton("Menu", icon.ExpMenu, noop),
			toolbar.Separator(),
			toolbar.IconButton("Project", icon.ExpFolder, noop),
			toolbar.IconButton("Branch", icon.ExpRefresh, noop),
			toolbar.Separator(),
			toolbar.IconButton("Save", icon.ExpSave, noop),
			toolbar.IconButton("Back", icon.ExpArrowLeft, noop),
			toolbar.IconButton("Forward", icon.ExpArrowRight, noop),
			toolbar.Spacer(),
			toolbar.IconButton("Run", icon.ExpRun, noop),
			toolbar.IconButton("Debug", icon.ExpDebug, noop),
			toolbar.Separator(),
			toolbar.IconButton("Search", icon.ExpSearch, noop),
			toolbar.IconButton("Settings", icon.ExpSettings, noop),
		),
		toolbar.Height(40),     // titlebar height
		toolbar.ButtonSize(30), // JB: 30x30 toolbar buttons
		toolbar.Gap(10),        // JB: 10px between items
		toolbar.PainterOpt(p.Toolbar),
	)
	tb := titlebar.New(
		titlebar.Title("gogpu — main.go"),
		titlebar.Height(40),
		titlebar.PainterOpt(p.TitleBar),
		titlebar.Chrome(gogpuApp),
		titlebar.Focused(true),
		titlebar.Leading(mainToolbar),
	)

	// --- Project Tree (left panel) ---
	projectTree := buildProjectTree(p)

	// --- Editor Tabs (top right) ---
	editorTabs := buildEditorTabs(cs, p)

	// --- Terminal Tabs (bottom right) ---
	terminalTabs := buildTerminalTabs(cs, p)

	// --- Right panel: vertical split (editor top, terminal bottom) ---
	rightPanel := splitview.New(
		splitview.First(editorTabs),
		splitview.Second(terminalTabs),
		splitview.OrientationOpt(splitview.Vertical),
		splitview.InitialRatio(0.65),
		splitview.MinFirst(100),
		splitview.MinSecond(80),
		splitview.DividerWidth(2),
		splitview.PainterOpt(p.SplitView),
	)

	// --- Main split: tree left, editor+terminal right ---
	mainSplit := splitview.New(
		splitview.First(
			primitives.VBox(projectTree).Background(cs.Surface),
		),
		splitview.Second(rightPanel),
		splitview.OrientationOpt(splitview.Horizontal),
		splitview.FixedFirst(220),
		splitview.MinFirst(150),
		splitview.MinSecond(300),
		splitview.DividerWidth(2),
		splitview.CollapsibleOpt(true),
		splitview.PainterOpt(p.SplitView),
	)

	// --- Tool Window Strips (left + right sidebars) ---
	leftStrip := buildLeftStrip(p)
	rightStrip := buildRightStrip(p)

	// --- Three-column layout: fixed strips + expanded center ---
	middleRow := primitives.HBox(
		leftStrip,
		primitives.Expanded(mainSplit),
		rightStrip,
	).Background(cs.Background)

	// --- Status Bar ---
	statusBar := buildStatusBar(cs)

	// --- Root layout: titlebar + expanded middle + fixed statusbar ---
	root := primitives.VBox(tb, primitives.Expanded(middleRow), statusBar).
		Background(cs.Background)

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(dt.AsTheme()),
	)
	uiApp.SetRoot(root)

	registerHitTest(gogpuApp, tb)

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

// registerHitTest sets up the frameless window hit-test callback for resize,
// drag, and client areas.
func registerHitTest(gogpuApp *gogpu.App, tb *titlebar.Widget) {
	const titleBarH = 40.0
	const controlW = 46.0
	const resizeW = 6.0

	gogpuApp.SetHitTestCallback(func(x, y float64) gpucontext.HitTestResult {
		w, h := gogpuApp.Size()
		fw, fh := float64(w), float64(h)

		onLeft := x < resizeW
		onRight := x >= fw-resizeW
		onTop := y < resizeW
		onBottom := y >= fh-resizeW

		// Corners first.
		if onTop && onLeft {
			return gpucontext.HitTestResizeNW
		}
		if onTop && onRight {
			return gpucontext.HitTestResizeNE
		}
		if onBottom && onLeft {
			return gpucontext.HitTestResizeSW
		}
		if onBottom && onRight {
			return gpucontext.HitTestResizeSE
		}
		if onTop {
			return gpucontext.HitTestResizeN
		}
		if onBottom {
			return gpucontext.HitTestResizeS
		}
		if onLeft {
			return gpucontext.HitTestResizeW
		}
		if onRight {
			return gpucontext.HitTestResizeE
		}

		// Below title bar -> client.
		if y >= titleBarH {
			return gpucontext.HitTestClient
		}

		// Window control buttons (right side) -> client.
		if x >= fw-controlW*3 {
			return gpucontext.HitTestClient
		}

		// Ask TitleBar widget for hit-test (toolbar buttons = client, gaps = caption).
		result := tb.HitTest(x, y)
		if result == titlebar.HitTestClient {
			return gpucontext.HitTestClient
		}
		return gpucontext.HitTestCaption
	})
}

// buildProjectTree creates the left-panel project file tree.
func buildProjectTree(p devtools.Painters) *treeview.Widget {
	root := &treeview.TreeNode{
		ID:       "root",
		Label:    "myproject",
		Expanded: true,
		Children: []*treeview.TreeNode{
			{
				ID:       "cmd",
				Label:    "cmd",
				Expanded: true,
				Children: []*treeview.TreeNode{
					{ID: "cmd-main", Label: "main.go"},
				},
			},
			{
				ID:       "pkg",
				Label:    "pkg",
				Expanded: true,
				Children: []*treeview.TreeNode{
					{
						ID:       "pkg-server",
						Label:    "server",
						Expanded: false,
						Children: []*treeview.TreeNode{
							{ID: "pkg-server-go", Label: "server.go"},
							{ID: "pkg-server-test", Label: "server_test.go"},
						},
					},
					{
						ID:       "pkg-config",
						Label:    "config",
						Expanded: false,
						Children: []*treeview.TreeNode{
							{ID: "pkg-config-go", Label: "config.go"},
						},
					},
				},
			},
			{
				ID:       "internal",
				Label:    "internal",
				Expanded: false,
				Children: []*treeview.TreeNode{
					{ID: "int-handler", Label: "handler.go"},
					{ID: "int-middleware", Label: "middleware.go"},
				},
			},
			{ID: "go-mod", Label: "go.mod"},
			{ID: "go-sum", Label: "go.sum"},
			{ID: "readme", Label: "README.md"},
			{ID: "license", Label: "LICENSE"},
			{ID: "makefile", Label: "Makefile"},
		},
	}

	return treeview.New(
		treeview.Root(root),
		treeview.ItemHeight(24),
		treeview.IndentWidth(16),
		treeview.SelectionModeOpt(treeview.SelectionSingle),
		treeview.SelectedNodeID("cmd-main"),
		treeview.ShowLines(false),
		treeview.PainterOpt(p.TreeView),
		treeview.OnSelect(func(node *treeview.TreeNode) {
			log.Printf("Selected: %s", node.Label)
		}),
	)
}

// buildEditorTabs creates the top tab view with simulated code content.
func buildEditorTabs(cs devtools.ColorScheme, p devtools.Painters) *tabview.Widget {
	mainGoContent := primitives.VBox(
		codeLine(cs, "package main"),
		codeLine(cs, ""),
		codeLine(cs, "import ("),
		codeLineSecondary(cs, "    \"fmt\""),
		codeLineSecondary(cs, "    \"net/http\""),
		codeLine(cs, ")"),
		codeLine(cs, ""),
		codeLine(cs, "func main() {"),
		codeLineSecondary(cs, "    http.HandleFunc(\"/\", handler)"),
		codeLineSecondary(cs, "    fmt.Println(\"Server starting on :8080\")"),
		codeLineSecondary(cs, "    http.ListenAndServe(\":8080\", nil)"),
		codeLine(cs, "}"),
		codeLine(cs, ""),
		codeLine(cs, "func handler(w http.ResponseWriter, r *http.Request) {"),
		codeLineSecondary(cs, "    fmt.Fprintf(w, \"Hello, gogpu!\")"),
		codeLine(cs, "}"),
	).Padding(12).Gap(2).Background(cs.Background)

	serverGoContent := primitives.VBox(
		codeLine(cs, "package server"),
		codeLine(cs, ""),
		codeLine(cs, "import ("),
		codeLineSecondary(cs, "    \"context\""),
		codeLineSecondary(cs, "    \"log\""),
		codeLine(cs, ")"),
		codeLine(cs, ""),
		codeLine(cs, "type Server struct {"),
		codeLineSecondary(cs, "    addr string"),
		codeLineSecondary(cs, "    log  *log.Logger"),
		codeLine(cs, "}"),
		codeLine(cs, ""),
		codeLine(cs, "func New(addr string) *Server {"),
		codeLineSecondary(cs, "    return &Server{addr: addr}"),
		codeLine(cs, "}"),
		codeLine(cs, ""),
		codeLine(cs, "func (s *Server) Run(ctx context.Context) error {"),
		codeLineSecondary(cs, "    s.log.Printf(\"listening on %s\", s.addr)"),
		codeLineSecondary(cs, "    return nil"),
		codeLine(cs, "}"),
	).Padding(12).Gap(2).Background(cs.Background)

	goModContent := primitives.VBox(
		codeLine(cs, "module github.com/user/myproject"),
		codeLine(cs, ""),
		codeLine(cs, "go 1.25"),
		codeLine(cs, ""),
		codeLine(cs, "require ("),
		codeLineSecondary(cs, "    github.com/gogpu/gg v0.37.4"),
		codeLineSecondary(cs, "    github.com/gogpu/ui v0.1.3"),
		codeLine(cs, ")"),
	).Padding(12).Gap(2).Background(cs.Background)

	return tabview.New(
		[]tabview.Tab{
			{Label: "main.go", Content: mainGoContent},
			{Label: "server.go", Content: serverGoContent},
			{Label: "go.mod", Content: goModContent},
		},
		tabview.PositionOpt(tabview.Top),
		tabview.SelectedIndex(0),
		tabview.Closeable(true),
		tabview.PainterOpt(p.TabView),
	)
}

// buildTerminalTabs creates the bottom tab view with terminal/output content.
func buildTerminalTabs(cs devtools.ColorScheme, p devtools.Painters) *tabview.Widget {
	terminalContent := primitives.VBox(
		termLine(cs, "PS D:\\projects\\myproject> go build ./..."),
		termLine(cs, "PS D:\\projects\\myproject> go test ./... -count=1"),
		termLineSuccess(cs, "ok  \tgithub.com/user/myproject/pkg/server\t0.042s"),
		termLineSuccess(cs, "ok  \tgithub.com/user/myproject/internal\t0.038s"),
		termLine(cs, "PS D:\\projects\\myproject> go run cmd/main.go"),
		termLineHighlight(cs, "Server starting on :8080"),
		termLine(cs, "PS D:\\projects\\myproject> _"),
	).Padding(8).Gap(1).Background(cs.Background)

	outputContent := primitives.VBox(
		outputLine(cs, "[INFO]  2026-03-15 14:32:01  Build started..."),
		outputLine(cs, "[INFO]  2026-03-15 14:32:02  Compiling cmd/main.go"),
		outputLine(cs, "[INFO]  2026-03-15 14:32:02  Compiling pkg/server/server.go"),
		outputLineSuccess(cs, "[OK]    2026-03-15 14:32:03  Build succeeded (1.2s)"),
		outputLine(cs, "[INFO]  2026-03-15 14:32:03  Running tests..."),
		outputLineSuccess(cs, "[PASS]  2026-03-15 14:32:04  All 12 tests passed"),
	).Padding(8).Gap(1).Background(cs.Background)

	return tabview.New(
		[]tabview.Tab{
			{Label: "Terminal", Content: terminalContent},
			{Label: "Output", Content: outputContent},
		},
		tabview.PositionOpt(tabview.Top),
		tabview.SelectedIndex(0),
		tabview.PainterOpt(p.TabView),
	)
}

// buildStatusBar creates the bottom status bar.
// JetBrains StatusBar uses JBFont.label() = 13px, no offset.
func buildStatusBar(cs devtools.ColorScheme) *primitives.BoxWidget {
	const statusFontSize float32 = 12 // slightly smaller than JB 13px for better fit

	left := primitives.Text("myproject").
		Color(cs.OnSurfaceSecondary).FontSize(statusFontSize)

	branch := primitives.Text("main").
		Color(cs.Primary).FontSize(statusFontSize)

	spacer := primitives.Box().Width(0).Height(1)

	goVer := primitives.Text("Go 1.25").
		Color(cs.OnSurfaceSecondary).FontSize(statusFontSize)

	encoding := primitives.Text("UTF-8").
		Color(cs.OnSurfaceSecondary).FontSize(statusFontSize)

	lineCol := primitives.Text("Ln 8, Col 12").
		Color(cs.OnSurfaceSecondary).FontSize(statusFontSize)

	version := primitives.Text("v0.1.3").
		Color(cs.OnSurfaceSecondary).FontSize(statusFontSize)

	return primitives.HBox(left, branch, spacer, lineCol, encoding, goVer, version).
		Background(cs.Surface).
		PaddingXY(8, 4).
		Gap(16).
		Height(28).
		BorderStyle(1, cs.Border)
}

// --- Helper functions for code/terminal text lines ---

// buildLeftStrip creates the left tool window strip using the stripe widget.
func buildLeftStrip(p devtools.Painters) *stripe.Widget {
	noop := func() {}
	return stripe.New(
		stripe.TopItems(
			stripe.Button{ID: "project", Label: "Project", Icon: icon.ToolProject, OnClick: noop},
			stripe.Button{ID: "commit", Label: "Commit", Icon: icon.ToolCommit, OnClick: noop},
			stripe.Button{ID: "structure", Label: "Structure", Icon: icon.ToolStructure, OnClick: noop},
		),
		stripe.BottomItems(
			stripe.Button{ID: "services", Label: "Services", Icon: icon.ToolServices, OnClick: noop},
			stripe.Button{ID: "terminal", Label: "Terminal", Icon: icon.ToolTerminal, OnClick: noop},
			stripe.Button{ID: "problems", Label: "Problems", Icon: icon.ToolProblems, OnClick: noop},
			stripe.Button{ID: "git", Label: "Git", Icon: icon.ToolGit, OnClick: noop},
		),
		stripe.ActiveID("terminal"),
		stripe.ShowLabels(true),
		stripe.Width(64),
		stripe.PainterOpt(p.Stripe),
	)
}

// buildRightStrip creates the right tool window strip (icons only, no labels).
// buildRightStrip creates the right tool window strip (icons only).
func buildRightStrip(p devtools.Painters) *stripe.Widget {
	noop := func() {}
	return stripe.New(
		stripe.TopItems(
			stripe.Button{ID: "notifications", Label: "Notifi", Icon: icon.ToolNotifications, OnClick: noop},
			stripe.Button{ID: "ai", Label: "AI", Icon: icon.ExpSearch, OnClick: noop},
			stripe.Button{ID: "db", Label: "DB", Icon: icon.ExpSettings, OnClick: noop},
			stripe.Button{ID: "nx", Label: "Nx", Icon: icon.ExpChevronRight, OnClick: noop},
		),
		stripe.ShowLabels(false),
		stripe.Width(40),
		stripe.PainterOpt(p.Stripe),
	)
}

func codeLine(cs devtools.ColorScheme, text string) *primitives.TextWidget {
	return primitives.Text(text).
		Color(cs.OnSurface).
		FontSize(13)
}

func codeLineSecondary(cs devtools.ColorScheme, text string) *primitives.TextWidget {
	return primitives.Text(text).
		Color(cs.OnSurfaceSecondary).
		FontSize(13)
}

func termLine(cs devtools.ColorScheme, text string) *primitives.TextWidget {
	return primitives.Text(text).
		Color(cs.OnSurfaceSecondary).
		FontSize(12)
}

func termLineSuccess(cs devtools.ColorScheme, text string) *primitives.TextWidget {
	return primitives.Text(text).
		Color(cs.Success).
		FontSize(12)
}

func termLineHighlight(cs devtools.ColorScheme, text string) *primitives.TextWidget {
	return primitives.Text(text).
		Color(cs.Primary).
		FontSize(12)
}

func outputLine(cs devtools.ColorScheme, text string) *primitives.TextWidget {
	return primitives.Text(text).
		Color(cs.OnSurfaceSecondary).
		FontSize(11)
}

func outputLineSuccess(cs devtools.ColorScheme, text string) *primitives.TextWidget {
	return primitives.Text(text).
		Color(cs.Success).
		FontSize(11)
}
