package plugin

import (
	"fmt"
	"strconv"
	"strings"
)

// dependencyGraph represents plugins and their dependencies for resolution.
type dependencyGraph struct {
	// nodes maps plugin names to their dependencies
	nodes map[string][]string
}

// newDependencyGraph creates a new empty dependency graph.
func newDependencyGraph() *dependencyGraph {
	return &dependencyGraph{
		nodes: make(map[string][]string),
	}
}

// addNode adds a plugin to the graph with its dependencies.
func (g *dependencyGraph) addNode(name string, dependencies []string) {
	g.nodes[name] = dependencies
}

// topologicalSort returns plugins in dependency order (dependencies first).
//
// Returns an error if a circular dependency is detected.
func (g *dependencyGraph) topologicalSort() ([]string, error) {
	inDegree := g.buildInDegreeMap()
	queue := g.findNoDependencyNodes(inDegree)
	result := g.processQueue(queue, inDegree)

	if err := g.checkForCycles(result, inDegree); err != nil {
		return nil, err
	}

	return result, nil
}

// buildInDegreeMap creates a map of node names to their dependency count.
func (g *dependencyGraph) buildInDegreeMap() map[string]int {
	inDegree := make(map[string]int, len(g.nodes))
	for name := range g.nodes {
		inDegree[name] = len(g.nodes[name])
	}
	return inDegree
}

// findNoDependencyNodes returns nodes with no dependencies.
func (g *dependencyGraph) findNoDependencyNodes(inDegree map[string]int) []string {
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}
	return queue
}

// processQueue processes nodes in dependency order using BFS.
func (g *dependencyGraph) processQueue(queue []string, inDegree map[string]int) []string {
	var result []string
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		result = append(result, name)
		queue = g.updateDependents(name, queue, inDegree)
	}
	return result
}

// updateDependents decreases in-degree for nodes that depend on the processed node.
func (g *dependencyGraph) updateDependents(processedName string, queue []string, inDegree map[string]int) []string {
	for nodeName, deps := range g.nodes {
		for _, dep := range deps {
			if dep == processedName {
				inDegree[nodeName]--
				if inDegree[nodeName] == 0 {
					queue = append(queue, nodeName)
				}
			}
		}
	}
	return queue
}

// checkForCycles detects if not all nodes were processed (indicates a cycle).
func (g *dependencyGraph) checkForCycles(result []string, inDegree map[string]int) error {
	if len(result) == len(g.nodes) {
		return nil
	}

	var cycleNodes []string
	for name, degree := range inDegree {
		if degree > 0 {
			cycleNodes = append(cycleNodes, name)
		}
	}
	return fmt.Errorf("%w: involving plugins %v", ErrCircularDependency, cycleNodes)
}

// semver represents a parsed semantic version.
type semver struct {
	major int
	minor int
	patch int
}

// parseSemver parses a semantic version string like "1.2.3".
func parseSemver(version string) (semver, error) {
	// Remove leading 'v' if present
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return semver{}, fmt.Errorf("invalid version format: %s", version)
	}

	var v semver
	var err error

	// Parse major
	v.major, err = strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// Parse minor (default 0)
	if len(parts) > 1 {
		v.minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return semver{}, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}

	// Parse patch (default 0)
	if len(parts) > 2 {
		// Handle pre-release suffixes like "1.0.0-alpha"
		patchPart := strings.Split(parts[2], "-")[0]
		v.patch, err = strconv.Atoi(patchPart)
		if err != nil {
			return semver{}, fmt.Errorf("invalid patch version: %s", patchPart)
		}
	}

	return v, nil
}

// compare returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v semver) compare(other semver) int {
	if v.major != other.major {
		if v.major < other.major {
			return -1
		}
		return 1
	}
	if v.minor != other.minor {
		if v.minor < other.minor {
			return -1
		}
		return 1
	}
	if v.patch != other.patch {
		if v.patch < other.patch {
			return -1
		}
		return 1
	}
	return 0
}

// versionConstraint represents a single version constraint.
type versionConstraint struct {
	operator string // "", ">=", "<=", ">", "<", "="
	version  semver
}

// parseConstraint parses a single version constraint like ">=1.0.0".
func parseConstraint(constraint string) (versionConstraint, error) {
	constraint = strings.TrimSpace(constraint)

	if constraint == "" {
		return versionConstraint{operator: ""}, nil
	}

	var vc versionConstraint

	// Check operators (order matters - check longer operators first)
	switch {
	case strings.HasPrefix(constraint, ">="):
		vc.operator = ">="
		constraint = strings.TrimPrefix(constraint, ">=")
	case strings.HasPrefix(constraint, "<="):
		vc.operator = "<="
		constraint = strings.TrimPrefix(constraint, "<=")
	case strings.HasPrefix(constraint, ">"):
		vc.operator = ">"
		constraint = strings.TrimPrefix(constraint, ">")
	case strings.HasPrefix(constraint, "<"):
		vc.operator = "<"
		constraint = strings.TrimPrefix(constraint, "<")
	case strings.HasPrefix(constraint, "="):
		vc.operator = "="
		constraint = strings.TrimPrefix(constraint, "=")
	default:
		// No operator means exact match
		vc.operator = "="
	}

	var err error
	vc.version, err = parseSemver(strings.TrimSpace(constraint))
	if err != nil {
		return versionConstraint{}, err
	}

	return vc, nil
}

// satisfies checks if a version satisfies the constraint.
func (vc versionConstraint) satisfies(v semver) bool {
	if vc.operator == "" {
		// Empty constraint matches any version
		return true
	}

	cmp := v.compare(vc.version)

	switch vc.operator {
	case "=":
		return cmp == 0
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	default:
		return false
	}
}

// checkVersionConstraint checks if a version string satisfies a constraint string.
//
// The constraint can be:
//   - "" - any version
//   - "1.0.0" - exact version
//   - ">=1.0.0" - minimum version
//   - "<=1.0.0" - maximum version
//   - ">1.0.0" - greater than
//   - "<1.0.0" - less than
//   - ">=1.0.0,<2.0.0" - range (AND)
func checkVersionConstraint(version, constraint string) (bool, error) {
	if constraint == "" {
		return true, nil
	}

	v, err := parseSemver(version)
	if err != nil {
		return false, fmt.Errorf("invalid version %q: %w", version, err)
	}

	// Split by comma for AND conditions
	parts := strings.Split(constraint, ",")
	for _, part := range parts {
		vc, err := parseConstraint(part)
		if err != nil {
			return false, fmt.Errorf("invalid constraint %q: %w", part, err)
		}

		if !vc.satisfies(v) {
			return false, nil
		}
	}

	return true, nil
}
