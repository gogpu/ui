package icon

// File type icons for DevTools and IDE-style interfaces.
//
// Each icon uses [MultiColorIcon] with two groups:
//   - "primary" for the document outline/base shape
//   - A file-type-specific key for the indicator (language logo, symbol, etc.)

// docOutlineOps returns the standard document outline path ops shared by
// most file type icons. The outline is a page with a folded top-right corner.
func docOutlineOps() []PathOp {
	return []PathOp{
		Move(6, 3), Line(15, 3), Line(18, 6), Line(18, 21),
		Line(6, 21), ClosePath(),
		// Corner fold
		Move(15, 3), Line(15, 6), Line(18, 6),
	}
}

// FileGo is a Go source file icon with "Go" text indicator.
var FileGo = MultiColorIcon{
	Name:    "file_go",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeyGo, Ops: []PathOp{
			// "Go" text approximation
			Move(8, 13), Line(12, 13), Line(12, 11), Line(10, 11),
			Move(11, 12), Line(12, 12),
			// "o"
			Move(13, 13), Line(13, 11), Line(16, 11), Line(16, 13), ClosePath(),
		}},
	},
}

// FileJSON is a JSON file icon with curly braces indicator.
var FileJSON = MultiColorIcon{
	Name:    "file_json",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeyJSON, Ops: []PathOp{
			// Left brace "{"
			Move(10, 10), Line(9, 10), Line(9, 12), Line(8, 12),
			Line(9, 12), Line(9, 14), Line(10, 14),
			// Right brace "}"
			Move(14, 10), Line(15, 10), Line(15, 12), Line(16, 12),
			Line(15, 12), Line(15, 14), Line(14, 14),
		}},
	},
}

// FileYAML is a YAML file icon with "Y" indicator.
var FileYAML = MultiColorIcon{
	Name:    "file_yaml",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeyYAML, Ops: []PathOp{
			// "Y" shape
			Move(9, 10), Line(12, 13),
			Move(15, 10), Line(12, 13),
			Move(12, 13), Line(12, 17),
		}},
	},
}

// FileMD is a Markdown file icon with "M" and down-arrow indicator.
var FileMD = MultiColorIcon{
	Name:    "file_md",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeyMarkdown, Ops: []PathOp{
			// "M" shape
			Move(8, 16), Line(8, 10), Line(10, 13), Line(12, 10), Line(12, 16),
			// Down arrow
			Move(15, 11), Line(15, 15),
			Move(13, 13), Line(15, 15), Line(17, 13),
		}},
	},
}

// FileTest is a test file icon with checkmark indicator.
var FileTest = MultiColorIcon{
	Name:    "file_test",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeySuccess, Ops: []PathOp{
			// Checkmark
			Move(9, 13), Line(11, 15), Line(15, 10),
		}},
	},
}

// FileConfig is a configuration file icon with gear indicator.
var FileConfig = MultiColorIcon{
	Name:    "file_config",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeySecondary, Ops: []PathOp{
			// Simplified gear
			Move(12, 9), Line(13, 10), Line(14, 10), Line(15, 11),
			Line(14, 12), Line(15, 13), Line(14, 14), Line(13, 14),
			Line(12, 15), Line(11, 14), Line(10, 14), Line(9, 13),
			Line(10, 12), Line(9, 11), Line(10, 10), Line(11, 10),
			ClosePath(),
		}},
	},
}

// FileImage is an image file icon with landscape indicator.
var FileImage = MultiColorIcon{
	Name:    "file_image",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeyAccent, Ops: []PathOp{
			// Mountain/landscape shape
			Move(8, 16), Line(11, 11), Line(13, 14), Line(15, 12), Line(17, 16),
			ClosePath(),
			// Sun dot
			Move(15, 10), Line(16, 10), Line(16, 11), Line(15, 11), ClosePath(),
		}},
	},
}

// FileGeneric is a plain document outline with no type indicator.
var FileGeneric = MultiColorIcon{
	Name:    "file_generic",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: docOutlineOps()},
		{ColorKey: KeySecondary, Ops: []PathOp{
			// Three horizontal lines (text placeholder)
			Move(8, 11), Line(16, 11),
			Move(8, 14), Line(16, 14),
			Move(8, 17), Line(13, 17),
		}},
	},
}
