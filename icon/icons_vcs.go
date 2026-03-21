package icon

// VCS (Version Control System) icons for DevTools and IDE-style interfaces.
//
// Each icon uses [MultiColorIcon] with two groups:
//   - "primary" for lines and structural elements
//   - "accent" for dots, indicators, and highlights

// GitBranch is a branch icon with a line forking into two.
var GitBranch = MultiColorIcon{
	Name:    "git_branch",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: []PathOp{
			// Main line
			Move(8, 6), Line(8, 14),
			// Branch line
			Move(8, 10), Line(14, 6),
		}},
		{ColorKey: KeyAccent, Ops: []PathOp{
			// Top dot (main)
			Move(7, 5), Line(9, 5), Line(9, 7), Line(7, 7), ClosePath(),
			// Bottom dot (main)
			Move(7, 14), Line(9, 14), Line(9, 16), Line(7, 16), ClosePath(),
			// Branch dot
			Move(13, 5), Line(15, 5), Line(15, 7), Line(13, 7), ClosePath(),
		}},
	},
}

// GitCommit is a commit icon — a circle on a vertical line.
var GitCommit = MultiColorIcon{
	Name:    "git_commit",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: []PathOp{
			// Vertical line (top)
			Move(12, 4), Line(12, 9),
			// Vertical line (bottom)
			Move(12, 15), Line(12, 20),
		}},
		{ColorKey: KeyAccent, Ops: []PathOp{
			// Circle (octagon approximation)
			Move(14, 9), Line(15, 10), Line(15, 14), Line(14, 15),
			Line(10, 15), Line(9, 14), Line(9, 10), Line(10, 9),
			ClosePath(),
		}},
	},
}

// GitMerge is a merge icon — two branches converging.
var GitMerge = MultiColorIcon{
	Name:    "git_merge",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: []PathOp{
			// Main line
			Move(8, 6), Line(8, 18),
			// Merge line (from branch to main)
			Move(16, 8), Line(12, 12), Line(8, 14),
		}},
		{ColorKey: KeyAccent, Ops: []PathOp{
			// Top dot (main)
			Move(7, 5), Line(9, 5), Line(9, 7), Line(7, 7), ClosePath(),
			// Bottom dot (main)
			Move(7, 17), Line(9, 17), Line(9, 19), Line(7, 19), ClosePath(),
			// Branch dot
			Move(15, 6), Line(17, 6), Line(17, 8), Line(15, 8), ClosePath(),
		}},
	},
}

// GitPR is a pull request icon — two dots with arrows indicating direction.
var GitPR = MultiColorIcon{
	Name:    "git_pr",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: []PathOp{
			// Left vertical line
			Move(8, 6), Line(8, 18),
			// Right vertical line
			Move(16, 6), Line(16, 18),
			// Arrow from right to left (top)
			Move(16, 8), Line(12, 8),
			Move(14, 6), Line(12, 8), Line(14, 10),
		}},
		{ColorKey: KeyAccent, Ops: []PathOp{
			// Top-left dot
			Move(7, 4), Line(9, 4), Line(9, 6), Line(7, 6), ClosePath(),
			// Bottom-left dot
			Move(7, 18), Line(9, 18), Line(9, 20), Line(7, 20), ClosePath(),
			// Top-right dot
			Move(15, 4), Line(17, 4), Line(17, 6), Line(15, 6), ClosePath(),
			// Bottom-right dot
			Move(15, 18), Line(17, 18), Line(17, 20), Line(15, 20), ClosePath(),
		}},
	},
}

// GitModified is a modified file indicator — a filled dot.
var GitModified = MultiColorIcon{
	Name:    "git_modified",
	ViewBox: defaultViewBox,
	Groups: []PathGroup{
		{ColorKey: KeyPrimary, Ops: []PathOp{
			// Outer ring (octagon)
			Move(14, 7), Line(17, 10), Line(17, 14), Line(14, 17),
			Line(10, 17), Line(7, 14), Line(7, 10), Line(10, 7),
			ClosePath(),
		}},
		{ColorKey: KeyWarning, Ops: []PathOp{
			// Inner dot (smaller octagon)
			Move(13, 10), Line(14, 11), Line(14, 13), Line(13, 14),
			Line(11, 14), Line(10, 13), Line(10, 11), Line(11, 10),
			ClosePath(),
		}},
	},
}
