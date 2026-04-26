package gitlog

// commit represents a git commit with essential metadata for display and filtering.
// All commit attributes are derived from git log output.
type commit struct {
	hash      string // hash is the full 40-character commit hash
	shortHash string // shortHash is the 7-character short form for display
	author    string // author is the commit author's name
	when      string // when is relative time string (e.g., "2 days ago")
	subject   string // subject is the first line of the commit message
}

// panel represents a UI panel for the primary display area.
type panel int

const (
	panelList panel = iota // panelList shows the commit list view
	panelDiff              // panelDiff shows the selected commit's diff
)

// lineKind represents the classification of a line in a unified diff.
// Used to apply appropriate styling and filtering.
type lineKind int

const (
	lineContext lineKind = iota // lineContext is an unchanged line in the diff
	lineAdded                    // lineAdded is a newly added line (prefix: +)
	lineRemoved                  // lineRemoved is a deleted line (prefix: -)
	lineHunk                     // lineHunk is a hunk header (prefix: @@)
	lineMeta                     // lineMeta is diff metadata (prefix: diff, index, etc.)
)

// diffLine represents a single line in a unified diff output.
// Each line is classified by kind for syntax highlighting and filtering.
type diffLine struct {
	kind lineKind // kind classifies the line type for rendering
	text string   // text is the actual line content without the diff prefix
}

// fileItem represents a file changed in a commit with its diff location.
// Used to navigate between modified files within a commit's diff.
type fileItem struct {
	path    string // path is the relative file path in the repository
	diffIdx int    // diffIdx is the starting line index in the diff output
}

// blameLine represents a single line annotated with blame information.
// Shows when and by whom each line was last modified.
type blameLine struct {
	shortHash string // shortHash is the commit that last modified this line
	author    string // author is the name of who made the change
	date      string // date is when the change was made (YYYY-MM-DD format)
	lineNum   int    // lineNum is the line number in the file (1-indexed)
	text      string // text is the actual source code line
}

// commitGroup represents a logical grouping of commits by a common criteria.
// Used for viewing commits organized by date, author, branch, or other dimensions.
type commitGroup struct {
	name           string   // name is the group key (date, author, branch name, etc.)
	commitHashes   []string // commitHashes are the commit identifiers in this group
}
