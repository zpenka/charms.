package gitlog

// Core git types - fundamental data structures for the git log tool

// commit represents a git commit with essential metadata
type commit struct {
	hash      string // Full 40-character hash
	shortHash string // 7-character short hash
	author    string // Commit author name
	when      string // Relative time (e.g., "2 days ago")
	subject   string // Commit message subject
}

// panel represents a UI panel (list or diff view)
type panel int

const (
	panelList panel = iota
	panelDiff
)

// lineKind represents the type of a diff line
type lineKind int

const (
	lineContext lineKind = iota
	lineAdded
	lineRemoved
	lineHunk
	lineMeta
)

// diffLine represents a single line in a unified diff
type diffLine struct {
	kind lineKind // Type of line (added, removed, context, etc.)
	text string   // The actual line content
}

// fileItem represents a file changed in a commit
type fileItem struct {
	path    string // File path
	diffIdx int    // Index in diff lines
}

// blameLine represents a single line with blame information
type blameLine struct {
	shortHash string // Commit that last changed this line
	author    string // Author who made the change
	date      string // When the change was made
	lineNum   int    // Line number in the file
	text      string // Line content
}
