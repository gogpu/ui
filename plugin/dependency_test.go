package plugin

import (
	"testing"
)

// TestParseSemver tests semantic version parsing.
func TestParseSemver(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    semver
		wantErr bool
	}{
		{
			name:  "full version",
			input: "1.2.3",
			want:  semver{major: 1, minor: 2, patch: 3},
		},
		{
			name:  "major only",
			input: "1",
			want:  semver{major: 1, minor: 0, patch: 0},
		},
		{
			name:  "major.minor",
			input: "1.2",
			want:  semver{major: 1, minor: 2, patch: 0},
		},
		{
			name:  "with v prefix",
			input: "v1.2.3",
			want:  semver{major: 1, minor: 2, patch: 3},
		},
		{
			name:  "with prerelease",
			input: "1.0.0-alpha",
			want:  semver{major: 1, minor: 0, patch: 0},
		},
		{
			name:  "zeros",
			input: "0.0.0",
			want:  semver{major: 0, minor: 0, patch: 0},
		},
		{
			name:  "large numbers",
			input: "100.200.300",
			want:  semver{major: 100, minor: 200, patch: 300},
		},
		{
			name:    "invalid major",
			input:   "abc.1.2",
			wantErr: true,
		},
		{
			name:    "invalid minor",
			input:   "1.abc.2",
			wantErr: true,
		},
		{
			name:    "invalid patch",
			input:   "1.2.abc",
			wantErr: true,
		},
		{
			name:    "too many parts",
			input:   "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSemver(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSemver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseSemver() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSemverCompare tests version comparison.
func TestSemverCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   semver
		v2   semver
		want int
	}{
		{"equal", semver{1, 2, 3}, semver{1, 2, 3}, 0},
		{"major less", semver{1, 0, 0}, semver{2, 0, 0}, -1},
		{"major greater", semver{2, 0, 0}, semver{1, 0, 0}, 1},
		{"minor less", semver{1, 1, 0}, semver{1, 2, 0}, -1},
		{"minor greater", semver{1, 2, 0}, semver{1, 1, 0}, 1},
		{"patch less", semver{1, 0, 1}, semver{1, 0, 2}, -1},
		{"patch greater", semver{1, 0, 2}, semver{1, 0, 1}, 1},
		{"zeros equal", semver{0, 0, 0}, semver{0, 0, 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.compare(tt.v2)
			if got != tt.want {
				t.Errorf("compare() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestParseConstraint tests constraint parsing.
func TestParseConstraint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
		version  semver
		wantErr  bool
	}{
		{"empty", "", "", semver{}, false},
		{"exact", "1.0.0", "=", semver{1, 0, 0}, false},
		{"gte", ">=1.0.0", ">=", semver{1, 0, 0}, false},
		{"lte", "<=1.0.0", "<=", semver{1, 0, 0}, false},
		{"gt", ">1.0.0", ">", semver{1, 0, 0}, false},
		{"lt", "<1.0.0", "<", semver{1, 0, 0}, false},
		{"eq explicit", "=1.0.0", "=", semver{1, 0, 0}, false},
		{"with spaces", " >= 1.0.0 ", ">=", semver{1, 0, 0}, false},
		{"invalid version", ">=abc", "", semver{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConstraint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.operator != tt.operator {
					t.Errorf("operator = %q, want %q", got.operator, tt.operator)
				}
				if got.version != tt.version {
					t.Errorf("version = %v, want %v", got.version, tt.version)
				}
			}
		})
	}
}

// TestVersionConstraintSatisfies tests constraint satisfaction.
func TestVersionConstraintSatisfies(t *testing.T) {
	tests := []struct {
		name       string
		constraint versionConstraint
		version    semver
		want       bool
	}{
		{"empty matches all", versionConstraint{operator: ""}, semver{1, 2, 3}, true},
		{"exact match", versionConstraint{"=", semver{1, 0, 0}}, semver{1, 0, 0}, true},
		{"exact no match", versionConstraint{"=", semver{1, 0, 0}}, semver{1, 0, 1}, false},
		{"gte equal", versionConstraint{">=", semver{1, 0, 0}}, semver{1, 0, 0}, true},
		{"gte greater", versionConstraint{">=", semver{1, 0, 0}}, semver{1, 1, 0}, true},
		{"gte less", versionConstraint{">=", semver{1, 0, 0}}, semver{0, 9, 0}, false},
		{"lte equal", versionConstraint{"<=", semver{1, 0, 0}}, semver{1, 0, 0}, true},
		{"lte less", versionConstraint{"<=", semver{1, 0, 0}}, semver{0, 9, 0}, true},
		{"lte greater", versionConstraint{"<=", semver{1, 0, 0}}, semver{1, 1, 0}, false},
		{"gt greater", versionConstraint{">", semver{1, 0, 0}}, semver{1, 0, 1}, true},
		{"gt equal", versionConstraint{">", semver{1, 0, 0}}, semver{1, 0, 0}, false},
		{"lt less", versionConstraint{"<", semver{1, 0, 0}}, semver{0, 9, 9}, true},
		{"lt equal", versionConstraint{"<", semver{1, 0, 0}}, semver{1, 0, 0}, false},
		{"unknown operator", versionConstraint{"~", semver{1, 0, 0}}, semver{1, 0, 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.constraint.satisfies(tt.version)
			if got != tt.want {
				t.Errorf("satisfies() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckVersionConstraint tests the high-level constraint checking.
func TestCheckVersionConstraint(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		want       bool
		wantErr    bool
	}{
		{"empty constraint", "1.0.0", "", true, false},
		{"exact match", "1.0.0", "1.0.0", true, false},
		{"exact no match", "1.0.1", "1.0.0", false, false},
		{"gte satisfied", "1.1.0", ">=1.0.0", true, false},
		{"gte not satisfied", "0.9.0", ">=1.0.0", false, false},
		{"lt satisfied", "0.9.0", "<1.0.0", true, false},
		{"range satisfied", "1.5.0", ">=1.0.0,<2.0.0", true, false},
		{"range not satisfied low", "0.9.0", ">=1.0.0,<2.0.0", false, false},
		{"range not satisfied high", "2.0.0", ">=1.0.0,<2.0.0", false, false},
		{"invalid version", "abc", ">=1.0.0", false, true},
		{"invalid constraint", "1.0.0", ">=abc", false, true},
		{"multiple constraints AND", "1.5.0", ">=1.0.0,<=2.0.0", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkVersionConstraint(tt.version, tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkVersionConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkVersionConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDependencyGraphBasic tests basic dependency graph operations.
func TestDependencyGraphBasic(t *testing.T) {
	graph := newDependencyGraph()

	// Simple chain: A -> B -> C (A depends on B, B depends on C)
	graph.addNode("A", []string{"B"})
	graph.addNode("B", []string{"C"})
	graph.addNode("C", nil)

	order, err := graph.topologicalSort()
	if err != nil {
		t.Fatalf("topologicalSort failed: %v", err)
	}

	// C must come before B, B must come before A
	cIdx, bIdx, aIdx := -1, -1, -1
	for i, name := range order {
		switch name {
		case "A":
			aIdx = i
		case "B":
			bIdx = i
		case "C":
			cIdx = i
		}
	}

	if cIdx > bIdx {
		t.Error("C should come before B")
	}
	if bIdx > aIdx {
		t.Error("B should come before A")
	}
}

// TestDependencyGraphNoDependencies tests plugins with no dependencies.
func TestDependencyGraphNoDependencies(t *testing.T) {
	graph := newDependencyGraph()

	graph.addNode("A", nil)
	graph.addNode("B", nil)
	graph.addNode("C", nil)

	order, err := graph.topologicalSort()
	if err != nil {
		t.Fatalf("topologicalSort failed: %v", err)
	}

	if len(order) != 3 {
		t.Errorf("Expected 3 items, got %d", len(order))
	}
}

// TestDependencyGraphDiamond tests diamond dependency pattern.
func TestDependencyGraphDiamond(t *testing.T) {
	graph := newDependencyGraph()

	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	graph.addNode("A", []string{"B", "C"})
	graph.addNode("B", []string{"D"})
	graph.addNode("C", []string{"D"})
	graph.addNode("D", nil)

	order, err := graph.topologicalSort()
	if err != nil {
		t.Fatalf("topologicalSort failed: %v", err)
	}

	// Find positions
	positions := make(map[string]int)
	for i, name := range order {
		positions[name] = i
	}

	// D must come before B and C
	if positions["D"] > positions["B"] {
		t.Error("D should come before B")
	}
	if positions["D"] > positions["C"] {
		t.Error("D should come before C")
	}

	// B and C must come before A
	if positions["B"] > positions["A"] {
		t.Error("B should come before A")
	}
	if positions["C"] > positions["A"] {
		t.Error("C should come before A")
	}
}

// TestDependencyGraphCircular tests circular dependency detection.
func TestDependencyGraphCircular(t *testing.T) {
	graph := newDependencyGraph()

	// A -> B -> C -> A (circular)
	graph.addNode("A", []string{"B"})
	graph.addNode("B", []string{"C"})
	graph.addNode("C", []string{"A"})

	_, err := graph.topologicalSort()
	if err == nil {
		t.Error("Expected circular dependency error")
	}
}

// TestDependencyGraphSelfLoop tests self-referencing dependency.
func TestDependencyGraphSelfLoop(t *testing.T) {
	graph := newDependencyGraph()

	// A depends on itself
	graph.addNode("A", []string{"A"})

	_, err := graph.topologicalSort()
	if err == nil {
		t.Error("Expected circular dependency error for self-loop")
	}
}

// TestDependencyGraphEmpty tests empty graph.
func TestDependencyGraphEmpty(t *testing.T) {
	graph := newDependencyGraph()

	order, err := graph.topologicalSort()
	if err != nil {
		t.Fatalf("topologicalSort failed: %v", err)
	}

	if len(order) != 0 {
		t.Errorf("Expected empty order, got %v", order)
	}
}

// TestDependencyGraphMissingDep tests dependency on non-existent node.
func TestDependencyGraphMissingDep(t *testing.T) {
	graph := newDependencyGraph()

	// A depends on B, but B is not in the graph
	graph.addNode("A", []string{"B"})

	// This should work - A's dependency on B just means
	// A waits for B, but B doesn't exist so A never gets unblocked
	_, err := graph.topologicalSort()
	// Actually, in our implementation, if B is not in the graph,
	// A will never have its in-degree decremented to 0
	// So we should get a circular dependency error
	if err == nil {
		t.Error("Expected error for missing dependency")
	}
}
