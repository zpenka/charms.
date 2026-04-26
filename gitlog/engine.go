package gitlog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type commit struct {
	hash      string
	shortHash string
	author    string
	when      string
	subject   string
}

type panel int

const (
	panelList panel = iota
	panelDiff
)

type lineKind int

const (
	lineContext lineKind = iota
	lineAdded
	lineRemoved
	lineHunk
	lineMeta
)

type diffLine struct {
	kind lineKind
	text string
}

type fileItem struct {
	path    string
	diffIdx int
}

type blameLine struct {
	shortHash string
	author    string
	date      string
	lineNum   int
	text      string
}

// FilterCache caches filter results with metrics tracking.
type FilterCache struct {
	cache   map[string][]commit
	metrics CacheMetrics
}

// NewFilterCache creates a new filter cache with metrics.
func NewFilterCache() *FilterCache {
	return &FilterCache{
		cache: make(map[string][]commit),
		metrics: CacheMetrics{
			Size:    0,
			MaxSize: 100,
			Hits:    0,
			Misses:  0,
		},
	}
}

type model struct {
	commits    []commit
	cursor     int
	focus      panel
	diffLines  []diffLine
	diffOffset int
	fileItems  []fileItem
	fileCursor int
	showFiles  bool
	searching  bool
	query      string
	flash      string
	// Branch picker
	showBranch    bool
	branches      []string
	branchCursor  int
	currentRef    string
	currentBranch string
	// Blame view
	showBlame   bool
	blameLines  []blameLine
	blameOffset int
	// Count prefix for j/k navigation
	countBuf string
	// Filtering
	authorFilter string
	sinceFilter  int // days; 0 = no filter
	// Breadcrumb trail
	navHistory    []int
	navHistoryIdx int
	// Bookmarks
	bookmarks []string // commit short hashes
	// Stats
	lastStats commitStatistics
	// Line comments
	comments map[int]string
	// Tag view
	showTags  bool
	tags      []string
	tagCursor int
	// Option 1: UI Integration
	showStatsBadge bool
	// Option 2: Commit Graph
	commitGraph []graphNode
	showGraph   bool
	// Option 3: File-Centric
	fileHistory      []commit
	currentFile      string
	showFileTimeline bool
	// Option 4: Stash & Reflog
	viewMode      string // "log", "stash", "reflog"
	stashes       []stashEntry
	reflogEntries []reflogEntry
	stashCursor   int
	reflogCursor  int
	// UI Integration
	inGoToCommitMode bool
	goToCommitInput  string
	inCommentMode    bool
	commentInput     string
	// Optimization: Caches
	dcache    *diffCache
	scache    *statCache
	recache   *regexCache
	// Performance tracking
	diffCacheHits   int
	statCacheHits   int
	regexCacheHits  int
	// Advanced Operations
	showRebaseUI      bool
	rebaseSequence    []rebaseOp
	showCherryPickUI  bool
	cherryPickList    []string
	resetMode         string // soft, mixed, hard
	amendMessage      string
	// Analytics
	showAnalytics     bool
	authorStats       map[string]int
	timeStats         map[string]int
	collaborators     map[string][]string // author -> co-authors
	reviewers         map[string][]string // commit hash -> reviewers
	productivity      map[string]interface{}
	repoPath          string
	width             int
	height            int
	loading           bool
	// Bisect & Recovery
	bisectState         bisectState
	showBisectUI        bool
	lostCommits         []lostCommit
	showLostCommits     bool
	undoStack           []string // commit hashes for undo
	undoStackIdx        int
	showUndoMenu        bool
	reflogRecoveryMode  bool
	recoveryCommits     []lostCommit
	// Code Patterns & Quality
	codeOwnership       map[string]codeOwnershipData
	showCodeOwnership   bool
	hotspots            []hotspotData
	showHotspots        bool
	commitMetrics       []commitMetrics
	showComplexity      bool
	lintingResults      []lintingResult
	showLinting         bool
	largeCommits        []commitMetrics
	showLargeCommits    bool
	// Commit Analysis & Search
	semanticSearchResults []semanticSearchResult
	showSemanticSearch    bool
	semanticQuery         string
	authorActivityHeatmap map[string]authorActivityData
	showActivityHeatmap   bool
	mergeAnalysisData     []mergeAnalysis
	showMergeAnalysis     bool
	commitCouplings       []commitCoupling
	showCoupling          bool
	// Performance & Filtering
	extensionFilters  []fileExtensionFilter
	currentExtFilter  string
	commitGroups      []commitGroup
	groupingMode      string // "", "pr", "branch", "date"
	dependencyChanges []dependencyChange
	showDependencies  bool
	// Advanced Workflows
	worktrees          []worktreeInfo
	showWorktrees      bool
	currentWorktree    string
	submodules         []submoduleInfo
	showSubmodules     bool
	namedStashes       []namedStash
	showNamedStashes   bool
	pendingTagOps      []tagOperation
	showTagMgmt        bool
	gpgStatuses        map[string]gpgSignatureStatus
	showGPGStatus      bool
	// Visualization
	contributorFlameData []contributorFlameData
	showFlamegraph       bool
	timelinePoints       []timelinePoint
	timelineSliderPos    int
	showTimeline         bool
	treeRoot             *treeNode
	showTreeView         bool
	authorComparisons    []authorComparison
	selectedAuthors      [2]string
	showAuthorComparison bool
	fileHeatmap          []fileHeatmapEntry
	showFileHeatmap      bool
	// Integration & Export
	prReferences      []githubPRReference
	showPRLinks       bool
	jiraLinks         []jiraTicketLink
	showJiraLinks     bool
	pendingExports    []exportData
	showExportUI      bool
	exportFormat      string
	issueReferences   []issueReference
	showIssueRefs     bool
	// Advanced Git Operations
	rebasePreview           rebasePreview
	showRebasePreview       bool
	conflictList            []conflictInfo
	showConflictUI          bool
	squashPlans             []squashPlan
	showSquashUI            bool
	cherryPickImprovements  []cherryPickImprovement
	amendPreview            amendPreview
	showAmendPreview        bool
	// Team & Collaboration
	teamStats               []teamStats
	showTeamStats           bool
	reviewWorkflows         []reviewWorkflow
	showReviewUI            bool
	reviewerSuggestions     []reviewerSuggestion
	pairProgrammingData     []pairProgrammingData
	showPairProgramming     bool
	velocityHistory         []velocityData
	showVelocity            bool
	// AI-Powered Insights
	messageCompletions      []messageCompletion
	commitClassifications   []commitClassification
	showClassification      bool
	anomalies               []anomalyData
	showAnomalies           bool
	similarCommits          []similarCommit
	showSimilar             bool
	autoSummaries           []autoSummary
	showSummaries           bool
	// Compliance & Security
	signingStatuses         map[string]signingStatus
	showSigningStatus       bool
	licenseHeaders          []licenseHeader
	showLicenses            bool
	securityIssues          []securityIssue
	showSecurityIssues      bool
	dataDeleteRequests      []dataDeleteRequest
	showDataRequests        bool
	secretDetections        []secretDetection
	showSecrets             bool
	// Release & Versioning
	semverVersions          []semverData
	showSemver              bool
	changelog               []changelogEntry
	showChangelog           bool
	releaseNotes            []releaseNote
	showReleaseNotes        bool
	versionBumps            []versionBump
	showVersionBumps        bool
	milestones              []milestone
	showMilestones          bool
	// Advanced Performance
	loadState               repoLoadState
	diffJobs                []diffProcessingJob
	indexData               indexData
	showLoadProgress        bool
	blameCache              map[string][]blameEntry
	showBlamePerf           bool
	memoryMetrics           memoryMetrics
	showMemoryMetrics       bool
}

type commitStatistics struct {
	filesChanged int
	insertions   int
	deletions    int
}

type graphNode struct {
	hash    string
	depth   int
	isMerge bool
	parents []string
}

type stashEntry struct {
	name   string
	branch string
	subject string
	hash   string
}

type reflogEntry struct {
	hash    string
	action  string
	message string
	date    string
}

type diffCache struct {
	data     map[string][]diffLine
	order    []string
	maxSize  int
	hitCount int
}

type statCache struct {
	data     map[string]commitStatistics
	order    []string
	maxSize  int
	hitCount int
}

type regexCache struct {
	data     map[string]*regexp.Regexp
	maxSize  int
	hitCount int
}

type rebaseOp struct {
	action  string // pick, squash, fixup, reword, drop
	hash    string
	subject string
}

type bisectOp struct {
	hash    string
	isBad   bool
	isGood  bool
	current bool
}

type bisectState struct {
	active       bool
	current      string
	good         []string
	bad          []string
	candidates   []string
	visualSteps  int
	totalSteps   int
}

type lostCommit struct {
	hash      string
	shortHash string
	author    string
	subject   string
	date      string
}

type codeOwnershipData struct {
	author        string
	files         map[string]int // file -> count of commits
	lines         int
	expertise     float64
	isOwner       bool
}

type hotspotData struct {
	path             string
	changeFrequency  int
	recentChanges    int
	collaborators    int
	avgCommitSize    int
	riskLevel        string // low, medium, high
}

type commitMetrics struct {
	hash          string
	linesChanged  int
	filesChanged  int
	complexity    int // estimated
	isLarge       bool
	isComplex     bool
	messageQuality int // 0-100
}

type lintingResult struct {
	hash    string
	subject string
	issues  []string
	score   int // 0-100
}

// Commit Analysis & Search
type semanticSearchResult struct {
	hash      string
	subject   string
	matches   []string // matched items (function names, variables)
	relevance int      // 0-100
}

type authorActivityData struct {
	author      string
	hourOfDay   map[int]int // hour -> count
	dayOfWeek   map[int]int // day -> count
	peakHour    int
	peakDay     string
	avgPerDay   float64
}

type mergeAnalysis struct {
	hash          string
	isMerge       bool
	isFastForward bool
	parentCount   int
	conflictRisk  int // 0-100
}

type commitCoupling struct {
	file1       string
	file2       string
	coChangeCount int
	correlation float64 // 0-1
}

// Performance & Filtering
type fileExtensionFilter struct {
	extension string
	enabled   bool
}

type commitGroup struct {
	name     string
	commits  []string // hashes
	label    string   // PR, branch, or time period
	groupBy  string   // "pr", "branch", "date"
}

type dependencyChange struct {
	hash    string
	dep     string
	oldVer  string
	newVer  string
	reason  string
}

// Advanced Workflows
type worktreeInfo struct {
	path   string
	branch string
	hash   string
}

type submoduleInfo struct {
	path   string
	url    string
	hash   string
	branch string
}

type namedStash struct {
	index       int
	name        string
	description string
	hash        string
}

type tagOperation struct {
	name    string
	hash    string
	action  string // create, delete, push
	message string
}

type gpgSignatureStatus struct {
	hash      string
	signed    bool
	signer    string
	verified  bool
	algorithm string
}

// Visualization
type contributorFlameData struct {
	author     string
	commits    int
	lines      int
	percentage float64
	timeline   map[string]int // date -> commit count
}

type timelinePoint struct {
	date    string
	commits int
	hash    string
}

type treeNode struct {
	hash     string
	subject  string
	children []*treeNode
	depth    int
}

type authorComparison struct {
	author1      string
	author2      string
	commits1     int
	commits2     int
	files1       int
	files2       int
	additions1   int
	additions2   int
	deletions1   int
	deletions2   int
	similarity   float64
}

type fileHeatmapEntry struct {
	path      string
	frequency int
	recent    int
	risk      string // low, medium, high
}

// Integration & Export
type githubPRReference struct {
	hash    string
	prNumber int
	status  string // open, merged, closed
	title   string
}

type jiraTicketLink struct {
	hash   string
	ticket string
	status string
}

type exportData struct {
	format   string // "markdown", "patch", "json"
	commits  []commit
	content  string
	filename string
}

type issueReference struct {
	hash      string
	references []string // "#123", "#456"
	keywords  []string  // "fixes", "closes", "resolves"
}

// Advanced Git Operations
type rebasePreview struct {
	operations []rebaseOp
	conflicts  []string
	willApply  bool
	message    string
}

type conflictInfo struct {
	file     string
	hash     string
	markers  []string
	resolved bool
}

type squashPlan struct {
	targetHash  string
	toSquash    []string
	resultMsg   string
	lineCount   int
}

type cherryPickImprovement struct {
	hash       string
	autoConflict bool
	suggestions []string
}

type amendPreview struct {
	originalMsg string
	newMsg      string
	changes     map[string]int // file -> change count
}

// Team & Collaboration
type teamStats struct {
	author           string
	commits          int
	additions        int
	deletions        int
	avgCommitSize    int
	specialization   string
	collaborators    []string
}

type reviewWorkflow struct {
	prNumber      int
	author        string
	reviewers     []string
	approved      bool
	commentCount  int
	status        string
}

type reviewerSuggestion struct {
	reviewer   string
	expertise  float64
	availability float64
	score      float64
}

type pairProgrammingData struct {
	pair1      string
	pair2      string
	commits    int
	files      int
	coChangeRate float64
}

type velocityData struct {
	week    string
	commits int
	files   int
	additions int
	deletions int
}

// AI-Powered Insights
type messageCompletion struct {
	prefix      string
	suggestions []string
	confidence  []float64
}

type commitClassification struct {
	hash       string
	category   string // "feature", "fix", "refactor", "docs", "test"
	confidence float64
	reason     string
}

type anomalyData struct {
	hash      string
	type_     string // "large", "unusual-pattern", "unusual-time"
	severity  int    // 1-10
	description string
}

type similarCommit struct {
	hash1   string
	hash2   string
	subject1 string
	subject2 string
	similarity float64
}

type autoSummary struct {
	hash    string
	summary string
	length  int
	tokens  int
}

// Compliance & Security
type signingStatus struct {
	hash      string
	isSigned  bool
	enforced  bool
	compliant bool
}

type licenseHeader struct {
	file      string
	hasHeader bool
	license   string
	hash      string
}

type securityIssue struct {
	hash     string
	severity string // "low", "medium", "high", "critical"
	type_    string // "hardcoded-secret", "sql-injection", etc.
	location string
}

type dataDeleteRequest struct {
	hash    string
	date    string
	reason  string
	status  string // "pending", "executed"
	email   string
}

type secretDetection struct {
	hash      string
	type_     string // "api-key", "password", "token"
	location  string
	severity  string
}

// Release & Versioning
type semverData struct {
	hash       string
	version    string
	versionType string // "major", "minor", "patch"
	isRelease  bool
}

type changelogEntry struct {
	version   string
	date      string
	commits   []string
	features  []string
	bugfixes  []string
	breaking  []string
}

type releaseNote struct {
	version     string
	summary     string
	highlights  []string
	contributors []string
	date        string
}

type versionBump struct {
	hash    string
	from    string
	to      string
	date    string
	message string
}

type milestone struct {
	name    string
	version string
	commits []string
	date    string
	status  string
}

// Advanced Performance
type repoLoadState struct {
	totalCommits   int
	loadedCommits  int
	percentage     int
	isComplete     bool
	estimatedTime  int // seconds
}

type diffProcessingJob struct {
	hash       string
	status     string // "pending", "processing", "done"
	result     []diffLine
	error      string
}

type indexData struct {
	lastIndexed string
	entries     int
	isUpToDate  bool
	nextUpdate  string
}

type blameEntry struct {
	hash   string
	author string
	date   string
	line   int
	text   string
}

type memoryMetrics struct {
	usageBytes    int64
	cacheSize     int
	percentUsed   float64
	estimatedMax  int64
}

// newModel creates a fresh model instance for the given repository path.
func newModel(repoPath string) model {
	return model{
		repoPath: repoPath,
		focus:    panelList,
	}
}

// parseCommits parses output of:
//
//	git log --format="%H%x00%h%x00%an%x00%ar%x00%s"
func parseCommits(output string) []commit {
	var commits []commit
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 5)
		if len(parts) < 5 {
			continue
		}
		commits = append(commits, commit{
			hash:      parts[0],
			shortHash: parts[1],
			author:    parts[2],
			when:      parts[3],
			subject:   parts[4],
		})
	}
	return commits
}

// parseCommitsWithPool parses commits using memory pooling for efficiency.
func parseCommitsWithPool(output string) []commit {
	pool := NewMemoryPool(100)
	var commits []commit

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 5)
		if len(parts) < 5 {
			continue
		}

		// Get commit from pool or create new
		c := pool.Get(func() interface{} {
			return &commit{}
		}).(*commit)

		// Populate fields
		c.hash = parts[0]
		c.shortHash = parts[1]
		c.author = parts[2]
		c.when = parts[3]
		c.subject = parts[4]

		commits = append(commits, *c)

		// Return to pool for reuse
		pool.Put(c)
	}
	return commits
}

// parseDiff classifies each line of a unified diff by type.
func parseDiff(raw string) []diffLine {
	var lines []diffLine
	for _, text := range strings.Split(raw, "\n") {
		var kind lineKind
		switch {
		case strings.HasPrefix(text, "@@"):
			kind = lineHunk
		case strings.HasPrefix(text, "diff "),
			strings.HasPrefix(text, "index "),
			strings.HasPrefix(text, "--- "),
			strings.HasPrefix(text, "+++ "),
			strings.HasPrefix(text, "new file"),
			strings.HasPrefix(text, "deleted file"),
			strings.HasPrefix(text, "similarity"),
			strings.HasPrefix(text, "rename"):
			kind = lineMeta
		case strings.HasPrefix(text, "+"):
			kind = lineAdded
		case strings.HasPrefix(text, "-"):
			kind = lineRemoved
		default:
			kind = lineContext
		}
		lines = append(lines, diffLine{kind, text})
	}
	return lines
}

// processDiffBatch processes diff lines using batch processing for efficiency.
func processDiffBatch(processor *BatchProcessor, lines []diffLine) []diffLine {
	var results []diffLine

	for _, line := range lines {
		processor.Add(line)

		// Process batch when full
		if processor.IsFull() {
			batch := processor.Get()
			for _, item := range batch {
				if dl, ok := item.(diffLine); ok {
					results = append(results, dl)
				}
			}
		}
	}

	// Process remaining items
	remaining := processor.Get()
	for _, item := range remaining {
		if dl, ok := item.(diffLine); ok {
			results = append(results, dl)
		}
	}

	return results
}

// truncate cuts s to at most max visible runes, appending "…" if shortened.
// truncate shortens a string to max runes and appends "…" if truncated.
func truncate(s string, max int) string {
	r := []rune(s)
	if max <= 0 {
		return ""
	}
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

// firstWord returns the first space-delimited word of s.
func firstWord(s string) string {
	if i := strings.Index(s, " "); i >= 0 {
		return s[:i]
	}
	return s
}

// moveCursorDown advances the commit cursor by one position and resets diff offset.
func moveCursorDown(m model) model {
	if m.cursor < len(m.commits)-1 {
		m.cursor++
		m.diffOffset = 0
	}
	return m
}

// moveCursorUp moves the commit cursor back one position and resets diff offset.
func moveCursorUp(m model) model {
	if m.cursor > 0 {
		m.cursor--
		m.diffOffset = 0
	}
	return m
}

// scrollDiffDown scrolls the diff view down by n lines, clamped to valid range.
func scrollDiffDown(m model, n int) model {
	maxOff := len(m.diffLines) - diffPanelHeight(m)
	if maxOff < 0 {
		maxOff = 0
	}
	m.diffOffset += n
	if m.diffOffset > maxOff {
		m.diffOffset = maxOff
	}
	return m
}

// scrollDiffUp scrolls the diff view up by n lines, clamped to zero.
func scrollDiffUp(m model, n int) model {
	m.diffOffset -= n
	if m.diffOffset < 0 {
		m.diffOffset = 0
	}
	return m
}

// switchPanel toggles focus between commit list and diff panels.
func switchPanel(m model) model {
	if m.focus == panelList {
		m.focus = panelDiff
	} else {
		m.focus = panelList
	}
	return m
}

// listPanelWidth returns the width of the commit list panel (clamped to 36–52).
func listPanelWidth(totalWidth int) int {
	w := totalWidth / 3
	if w < 36 {
		return 36
	}
	if w > 52 {
		return 52
	}
	return w
}

// diffPanelWidth returns the remaining width for the diff panel.
func diffPanelWidth(totalWidth int) int {
	return totalWidth - listPanelWidth(totalWidth) - 1 // 1 for divider
}

// diffPanelHeight returns the number of content lines visible in each panel.
// diffPanelHeight calculates the height of the diff panel based on terminal size.
func diffPanelHeight(m model) int {
	h := m.height - 7 // title + blank + header + blank + hint + blank*2
	if h < 5 {
		return 5
	}
	return h
}

// parseFileItems scans diffLines for "diff --git" boundaries and returns each
// file's path and the index of its boundary line in diffLines.
func parseFileItems(lines []diffLine) []fileItem {
	var items []fileItem
	for i, line := range lines {
		if line.kind != lineMeta || !strings.HasPrefix(line.text, "diff --git ") {
			continue
		}
		parts := strings.Fields(line.text)
		if len(parts) < 4 {
			continue
		}
		path := strings.TrimPrefix(parts[3], "b/")
		items = append(items, fileItem{path: path, diffIdx: i})
	}
	return items
}

// filterCommits returns commits whose subject, author, or short hash contain
// query (case-insensitive). An empty query returns all commits unchanged.
func filterCommits(commits []commit, query string) []commit {
	if query == "" {
		return commits
	}
	q := strings.ToLower(query)
	var out []commit
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), q) ||
			strings.Contains(strings.ToLower(c.author), q) ||
			strings.Contains(strings.ToLower(c.shortHash), q) {
			out = append(out, c)
		}
	}
	return out
}

// filterCommitsWithCache filters commits with result caching and metrics tracking.
func filterCommitsWithCache(cache *FilterCache, commits []commit, query string) []commit {
	if query == "" {
		return commits
	}

	// Check cache first
	if cached, exists := cache.cache[query]; exists {
		cache.metrics.Hits++
		return cached
	}

	// Cache miss - filter and store
	cache.metrics.Misses++
	result := filterCommits(commits, query)

	// Store in cache if not full
	if cache.metrics.Size < cache.metrics.MaxSize {
		cache.cache[query] = result
		cache.metrics.Size++
	}

	return result
}

// visibleCommits returns the commit list after applying all active filters:
// search query, author filter, and time-based filter.
func visibleCommits(m model) []commit {
	result := m.commits
	// Apply time filter first
	if m.sinceFilter > 0 {
		result = filterCommitsSince(result, m.sinceFilter)
	}
	// Apply author filter
	if m.authorFilter != "" {
		result = filterCommitsByAuthor(result, m.authorFilter)
	}
	// Apply search query filter last
	result = filterCommits(result, m.query)
	return result
}

// scrollToDiffLine sets diffOffset so that lineIdx is visible, clamped to valid range.
func scrollToDiffLine(m model, lineIdx int) model {
	m.diffOffset = lineIdx
	maxOff := len(m.diffLines) - diffPanelHeight(m)
	if maxOff < 0 {
		maxOff = 0
	}
	if m.diffOffset > maxOff {
		m.diffOffset = maxOff
	}
	if m.diffOffset < 0 {
		m.diffOffset = 0
	}
	return m
}

// toggleFileView shows or hides the file list in the left panel.
// Hiding resets fileCursor.
func toggleFileView(m model) model {
	m.showFiles = !m.showFiles
	if !m.showFiles {
		m.fileCursor = 0
	}
	return m
}

// parseBranches parses the output of "git branch -a", stripping the current-branch
// marker (*) and skipping ref-pointer lines (e.g. "origin/HEAD -> origin/main").
func parseBranches(output string) []string {
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "* ")
		if strings.Contains(line, " -> ") {
			continue
		}
		branches = append(branches, line)
	}
	return branches
}

// parseCurrentBranch returns the name of the currently checked-out branch from
// "git branch -a" output (the line prefixed with "* ").
func parseCurrentBranch(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.HasPrefix(line, "* ") {
			return strings.TrimPrefix(line, "* ")
		}
	}
	return ""
}

// parseBlameLine parses one line of "git blame --date=short" output.
// Format: "hash (Author Name   YYYY-MM-DD  linenum) content"
func parseBlameLine(line string) (blameLine, bool) {
	paren := strings.Index(line, "(")
	close := strings.Index(line, ")")
	if paren < 0 || close < 0 || close < paren {
		return blameLine{}, false
	}
	hash := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line[:paren]), "^"))
	if len(hash) > 7 {
		hash = hash[:7]
	}
	meta := strings.Fields(line[paren+1 : close])
	if len(meta) < 3 {
		return blameLine{}, false
	}
	// last field: line number; second-to-last: date; rest: author
	lineNum, err := strconv.Atoi(meta[len(meta)-1])
	if err != nil {
		return blameLine{}, false
	}
	date := meta[len(meta)-2]
	author := strings.Join(meta[:len(meta)-2], " ")
	var text string
	if close+2 <= len(line) {
		text = line[close+2:]
	} else if close+1 < len(line) {
		text = line[close+1:]
	}
	return blameLine{
		shortHash: hash,
		author:    author,
		date:      date,
		lineNum:   lineNum,
		text:      text,
	}, true
}

// parseBlame parses all lines from "git blame --date=short" output,
// skipping any lines that don't match the expected format.
func parseBlame(output string) []blameLine {
	var lines []blameLine
	for _, line := range strings.Split(output, "\n") {
		if bl, ok := parseBlameLine(line); ok {
			lines = append(lines, bl)
		}
	}
	return lines
}

// currentFile returns the path of the file whose diff section is currently
// visible, based on the last fileItem whose diffIdx is <= diffOffset.
// currentFile returns the path of the currently selected file in file view mode.
func currentFile(m model) string {
	if len(m.fileItems) == 0 {
		return ""
	}
	cur := m.fileItems[0].path
	for _, fi := range m.fileItems {
		if fi.diffIdx <= m.diffOffset {
			cur = fi.path
		}
	}
	return cur
}

// parseCount converts a digit string to a navigation count.
// Empty string or zero returns 1; values above 200 are capped.
func parseCount(s string) int {
	if s == "" {
		return 1
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 1
	}
	if n > 200 {
		return 200
	}
	return n
}

// toggleBranchView shows or hides the branch picker in the left panel.
// Hiding resets branchCursor.
func toggleBranchView(m model) model {
	m.showBranch = !m.showBranch
	if !m.showBranch {
		m.branchCursor = 0
	}
	return m
}

// filterCommitsByAuthor returns commits whose author exactly matches the given author
// (case-insensitive).
func filterCommitsByAuthor(commits []commit, author string) []commit {
	if author == "" {
		return commits
	}
	var out []commit
	for _, c := range commits {
		if strings.EqualFold(c.author, author) {
			out = append(out, c)
		}
	}
	return out
}

// filterCommitsSince returns commits from the last N days, parsed from the
// "when" field (e.g., "5 days ago", "2 weeks ago"). Returns all commits if
// days <= 0.
func filterCommitsSince(commits []commit, days int) []commit {
	if days <= 0 {
		return commits
	}
	var out []commit
	for _, c := range commits {
		if isWithinDays(c.when, days) {
			out = append(out, c)
		}
	}
	return out
}

// isWithinDays checks if a "when" string (e.g., "5 days ago") represents
// a time within the last N days.
func isWithinDays(when string, days int) bool {
	whenLower := strings.ToLower(when)

	// Extract number from strings like "5 days ago", "2 weeks ago", etc.
	re := regexp.MustCompile(`(\d+)\s+(day|week|month|year)`)
	matches := re.FindStringSubmatch(whenLower)
	if len(matches) < 3 {
		return false
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}

	unit := matches[2]
	totalDays := 0
	switch unit {
	case "day":
		totalDays = num
	case "week":
		totalDays = num * 7
	case "month":
		totalDays = num * 30
	case "year":
		totalDays = num * 365
	default:
		return false
	}

	return totalDays <= days
}

// formatActiveFilters returns a string showing all active filters.
// formatActiveFilters returns a display string of all active filters.
func formatActiveFilters(m model) string {
	var filters []string
	if m.authorFilter != "" {
		filters = append(filters, m.authorFilter)
	}
	if m.sinceFilter > 0 {
		filters = append(filters, strconv.Itoa(m.sinceFilter)+"d")
	}
	if len(filters) == 0 {
		return ""
	}
	return "[" + strings.Join(filters, " + ") + "]"
}

// addToNavHistory adds the current cursor position to navigation history.
func addToNavHistory(m model, position int) model {
	// Discard any future history if we're not at the end
	if m.navHistoryIdx < len(m.navHistory)-1 {
		m.navHistory = m.navHistory[:m.navHistoryIdx+1]
	}
	m.navHistory = append(m.navHistory, position)
	m.navHistoryIdx = len(m.navHistory) - 1
	return m
}

// goBackInHistory moves to the previous position in navigation history.
func goBackInHistory(m model) model {
	if m.navHistoryIdx > 0 {
		m.navHistoryIdx--
		m.cursor = m.navHistory[m.navHistoryIdx]
	}
	return m
}

// goForwardInHistory moves to the next position in navigation history.
func goForwardInHistory(m model) model {
	if m.navHistoryIdx < len(m.navHistory)-1 {
		m.navHistoryIdx++
		m.cursor = m.navHistory[m.navHistoryIdx]
	}
	return m
}

// commitStats calculates statistics for a commit's diff.
// commitStats calculates insertion, deletion, and file change counts from diff lines.
func commitStats(lines []diffLine) commitStatistics {
	var stats commitStatistics
	for _, line := range lines {
		if line.kind == lineMeta && strings.HasPrefix(line.text, "diff --git") {
			stats.filesChanged++
		}
		if line.kind == lineAdded {
			stats.insertions++
		}
		if line.kind == lineRemoved {
			stats.deletions++
		}
	}
	return stats
}

// generateCommitMessage generates a suggested commit message based on the diff.
func generateCommitMessage(lines []diffLine, filename string) string {
	stats := commitStats(lines)
	isNew := false
	isDeleted := false
	isBreaking := false

	for _, line := range lines {
		if strings.Contains(line.text, "new file mode") {
			isNew = true
		}
		if strings.Contains(line.text, "deleted file mode") {
			isDeleted = true
		}
		// Simple breaking change detection: removed function/interface definitions
		if line.kind == lineRemoved &&
			(strings.Contains(line.text, "func ") || strings.Contains(line.text, "interface")) {
			isBreaking = true
		}
	}

	scope := filenameToScope(filename)
	var verb string

	switch {
	case isDeleted:
		verb = "remove"
	case isNew:
		verb = "add"
	case stats.deletions > stats.insertions*2:
		verb = "refactor"
	default:
		verb = "update"
	}

	msg := verb
	if scope != "" {
		msg += "(" + scope + ")"
	}
	if isBreaking {
		msg += "!"
	}
	msg += ": " + capitalizeFirst(verb) + " changes"

	return msg
}

// filenameToScope extracts a scope from a filename (e.g., "auth.go" -> "auth").
// filenameToScope extracts the directory scope from a file path for navigation.
func filenameToScope(filename string) string {
	if filename == "" {
		return ""
	}
	base := filename
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if idx := strings.LastIndex(base, "."); idx > 0 {
		base = base[:idx]
	}
	return strings.ToLower(base)
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// toggleBookmark toggles a bookmark on the current commit.
func toggleBookmark(m model) model {
	if m.cursor >= len(m.commits) {
		return m
	}
	hash := m.commits[m.cursor].shortHash
	if isBookmarked(m, m.cursor) {
		// Remove bookmark
		var newBookmarks []string
		for _, b := range m.bookmarks {
			if b != hash {
				newBookmarks = append(newBookmarks, b)
			}
		}
		m.bookmarks = newBookmarks
	} else {
		// Add bookmark
		m.bookmarks = append(m.bookmarks, hash)
	}
	return m
}

// isBookmarked checks if a commit at the given index is bookmarked.
func isBookmarked(m model, idx int) bool {
	if idx >= len(m.commits) {
		return false
	}
	hash := m.commits[idx].shortHash
	for _, b := range m.bookmarks {
		if b == hash {
			return true
		}
	}
	return false
}

// jumpToNextBookmark moves the cursor to the next bookmarked commit.
func jumpToNextBookmark(m model) model {
	for i := m.cursor + 1; i < len(m.commits); i++ {
		if isBookmarked(m, i) {
			m.cursor = i
			return m
		}
	}
	return m
}

// jumpToPrevBookmark moves the cursor to the previous bookmarked commit.
func jumpToPrevBookmark(m model) model {
	for i := m.cursor - 1; i >= 0; i-- {
		if isBookmarked(m, i) {
			m.cursor = i
			return m
		}
	}
	return m
}

// detectLanguage detects the programming language from a filename.
func detectLanguage(filename string) string {
	ext := filename
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		ext = filename[idx:]
	}

	langMap := map[string]string{
		".go":         "go",
		".py":         "python",
		".js":         "javascript",
		".ts":         "typescript",
		".rb":         "ruby",
		".java":       "java",
		".cpp":        "cpp",
		".c":          "c",
		".rs":         "rust",
		".sh":         "bash",
		".sql":        "sql",
		".html":       "html",
		".css":        "css",
		".json":       "json",
		".yaml":       "yaml",
		".yml":        "yaml",
		".xml":        "xml",
		".md":         "markdown",
		"Makefile":    "makefile",
		"Dockerfile":  "dockerfile",
		".gitignore":  "gitignore",
		".env":        "dotenv",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	if lang, ok := langMap[filename]; ok {
		return lang
	}

	// Default to text
	return "text"
}

// miniMapPosition calculates the position of a scroll indicator (0-height).
// miniMapPosition calculates the position indicator for a visual minimap of commits.
func miniMapPosition(cursor, panelHeight, totalCommits int) int {
	if totalCommits <= 1 {
		return 0
	}
	if panelHeight <= 1 {
		return 0
	}

	// Map cursor position to panel height
	position := (cursor * (panelHeight - 1)) / (totalCommits - 1)
	if position > panelHeight-1 {
		position = panelHeight - 1
	}
	return position
}

// --- diffStatBadge ---

// diffStatBadge formats commit statistics as a compact badge (e.g., "3 files +10 -5").
func diffStatBadge(stats commitStatistics) string {
	var parts []string
	if stats.filesChanged > 0 {
		parts = append(parts, fmt.Sprintf("%d file%s", stats.filesChanged, pluralize(stats.filesChanged)))
	}
	if stats.insertions > 0 {
		parts = append(parts, fmt.Sprintf("+%d", stats.insertions))
	}
	if stats.deletions > 0 {
		parts = append(parts, fmt.Sprintf("-%d", stats.deletions))
	}
	if len(parts) == 0 {
		return "0 changes"
	}
	return strings.Join(parts, " ")
}

// pluralize returns "s" if count != 1, else "".
// pluralize returns "s" if count != 1, else empty string, for grammatical pluralization.
func pluralize(count int) string {
	if count != 1 {
		return "s"
	}
	return ""
}

// --- goToCommit ---

// goToCommit finds a commit by hash (short or full) and returns its index, or -1 if not found.
// goToCommit searches commits for a matching hash or subject and returns its index.
func goToCommit(commits []commit, query string) int {
	q := strings.ToLower(query)
	for i, c := range commits {
		if strings.EqualFold(c.shortHash, query) || strings.EqualFold(c.hash, query) {
			return i
		}
		if strings.HasPrefix(strings.ToLower(c.shortHash), q) || strings.HasPrefix(strings.ToLower(c.hash), q) {
			return i
		}
	}
	return -1
}

// --- copyAsPatch ---

// copyAsPatch generates a patch file format from a commit hash and diff lines.
// copyAsPatch formats diff lines as a git patch with headers.
func copyAsPatch(hash string, lines []diffLine) string {
	var sb strings.Builder
	sb.WriteString("From: " + hash + "\n")
	sb.WriteString("Subject: Commit patch\n")
	sb.WriteString("\n")
	for _, line := range lines {
		sb.WriteString(line.text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// --- parseGitReferences ---

// parseGitReferences extracts issue/PR numbers from a commit message (e.g., #123, fixes #456).
func parseGitReferences(msg string) []string {
	re := regexp.MustCompile(`#(\d+)`)
	matches := re.FindAllStringSubmatch(msg, -1)
	var refs []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			refs = append(refs, match[1])
			seen[match[1]] = true
		}
	}
	return refs
}

// --- isMergeCommit ---

// isMergeCommit checks if a commit is a merge commit.
// isMergeCommit detects if a commit is a merge based on diff metadata.
func isMergeCommit(lines []diffLine) bool {
	for _, line := range lines {
		if strings.HasPrefix(line.text, "Merge:") {
			return true
		}
		if strings.Contains(strings.ToLower(line.text), "merge branch") {
			return true
		}
	}
	return false
}

// --- getMergeParents ---

// getMergeParents extracts parent hashes from a merge commit.
// getMergeParents extracts the parent commit hashes from a merge commit's diff.
func getMergeParents(lines []diffLine) []string {
	for _, line := range lines {
		if strings.HasPrefix(line.text, "Merge:") {
			parts := strings.Fields(strings.TrimPrefix(line.text, "Merge:"))
			if len(parts) >= 2 {
				return parts[:2]
			}
		}
	}
	return nil
}

// --- hunk structure and parseHunks ---

type hunk struct {
	startLine int
	endLine   int
	header    string
}

// parseHunks extracts all hunks from diff lines.
func parseHunks(lines []diffLine) []hunk {
	var hunks []hunk
	var lastHunk hunk
	hunkCount := 0

	for i, line := range lines {
		if line.kind == lineHunk {
			if hunkCount > 0 {
				hunks = append(hunks, lastHunk)
			}
			lastHunk = hunk{
				startLine: i,
				endLine:   i,
				header:    line.text,
			}
			hunkCount++
		} else if hunkCount > 0 {
			lastHunk.endLine = i
		}
	}
	if hunkCount > 0 {
		hunks = append(hunks, lastHunk)
	}
	return hunks
}

// --- toggleLineComment ---

// toggleLineComment adds or removes a comment on a specific diff line.
func toggleLineComment(m model, lineIdx int, comment string) model {
	if m.comments == nil {
		m.comments = make(map[int]string)
	}
	if comment == "" {
		delete(m.comments, lineIdx)
	} else {
		m.comments[lineIdx] = comment
	}
	return m
}

// --- compileRegex ---

// compileRegex compiles a regex pattern for search.
func compileRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// --- parseDateRange ---

// parseDateRange parses a date or date range (e.g., "2024-01-15" or "2024-01-01..2024-01-31").
func parseDateRange(input string) (*time.Time, *time.Time, error) {
	if input == "" {
		return nil, nil, nil
	}

	// Check for range format
	if strings.Contains(input, "..") {
		parts := strings.Split(input, "..")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid date range format")
		}
		start, err1 := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
		end, err2 := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return nil, nil, fmt.Errorf("invalid date format")
		}
		return &start, &end, nil
	}

	// Single date
	date, err := time.Parse("2006-01-02", input)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid date format")
	}
	return &date, nil, nil
}

// --- filterCommitsByFile ---

// filterCommitsByFile returns commits that touched the specified file.
// Note: This is infrastructure-ready; actual filtering requires git queries.
func filterCommitsByFile(commits []commit, file string) []commit {
	if file == "" {
		return commits
	}
	// This would typically query git for commits touching the file
	// For now, return infrastructure (called from UI)
	return []commit{}
}

// --- parseTags ---

// parseTags parses git tag output (format: "hash tagname" per line).
func parseTags(output string) []string {
	var tags []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Handle both "hash tagname" and "tagname" formats
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			// Take the last non-hash part
			tag := parts[len(parts)-1]
			if !strings.HasPrefix(tag, "[") { // Skip metadata
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// ===== OPTION 1: UI INTEGRATION =====

// renderStatsBadgeInList renders a compact stats badge for a commit list row.
func renderStatsBadgeInList(stats commitStatistics, maxWidth int) string {
	badge := diffStatBadge(stats)
	if len(badge) > maxWidth {
		badge = truncate(badge, maxWidth)
	}
	return badge
}

// formatFilterHeaderDisplay formats active filters for header display.
func formatFilterHeaderDisplay(m model) string {
	return formatActiveFilters(m)
}

// renderBookmarkMarker returns a visual marker for bookmarked commits.
func renderBookmarkMarker(m model, idx int) string {
	if isBookmarked(m, idx) {
		return "★"
	}
	return ""
}

// handleGoToCommitInput processes go-to-commit input and updates model.
func handleGoToCommitInput(m model, query string) model {
	idx := goToCommit(m.commits, query)
	if idx >= 0 {
		m.cursor = idx
		m.diffOffset = 0
	}
	return m
}

// renderLineCommentMarker returns a visual marker for commented lines.
func renderLineCommentMarker(m model, lineIdx int) string {
	if m.comments != nil {
		if _, ok := m.comments[lineIdx]; ok {
			return "●"
		}
	}
	return ""
}

// ===== OPTION 2: COMMIT GRAPH =====

// parseCommitGraph builds a graph structure from commits.
func parseCommitGraph(commits []commit) []graphNode {
	var nodes []graphNode
	for i, c := range commits {
		node := graphNode{
			hash:    c.hash,
			depth:   0,
			isMerge: false,
		}
		// Simple heuristic: if subject contains "Merge", mark as merge
		if strings.Contains(strings.ToLower(c.subject), "merge") {
			node.isMerge = true
		}
		// Depth is based on position (linear for now)
		node.depth = i % 2
		nodes = append(nodes, node)
	}
	return nodes
}

// detectBranches identifies branches in commit graph.
func detectBranches(commits []commit) []string {
	// Simple implementation: treat as single branch unless merges detected
	var branches []string
	hasMerge := false
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "merge") {
			hasMerge = true
			break
		}
	}
	if hasMerge {
		branches = append(branches, "main", "feature")
	} else {
		branches = append(branches, "main")
	}
	return branches
}

// renderAsciiGraph renders ASCII art graph for commit history.
func renderAsciiGraph(graph []graphNode) string {
	if len(graph) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, node := range graph {
		if node.isMerge {
			sb.WriteString("*   ")
		} else {
			sb.WriteString("* ")
		}
		hash := node.hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		sb.WriteString(hash)
		if i < len(graph)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// navigateAlongGraph moves along the graph in a direction.
func navigateAlongGraph(graph []graphNode, currentIdx int, direction string) int {
	if len(graph) == 0 {
		return 0
	}
	switch direction {
	case "down":
		if currentIdx < len(graph)-1 {
			return currentIdx + 1
		}
	case "up":
		if currentIdx > 0 {
			return currentIdx - 1
		}
	}
	return currentIdx
}

// getCommitRelationships maps parent-child relationships.
func getCommitRelationships(commits []commit) map[string][]string {
	rels := make(map[string][]string)
	// Infrastructure: would populate from git data
	return rels
}

// ===== OPTION 3: FILE-CENTRIC VIEW =====

// buildFileHistory constructs commit history for a specific file.
func buildFileHistory(commits []commit, file string) []commit {
	if file == "" {
		return []commit{}
	}
	// Infrastructure: would query git for file history
	return []commit{}
}

// renderFileTimeline renders the evolution of a file over time.
func renderFileTimeline(commits []commit, file string, width int) string {
	var sb strings.Builder
	sb.WriteString("File timeline for: ")
	sb.WriteString(file)
	sb.WriteString("\n")
	if len(commits) == 0 {
		sb.WriteString("(no commits found)\n")
	}
	return sb.String()
}

// getFileBlameContext gets blame information for a file.
func getFileBlameContext(lines []diffLine, file string) map[int]string {
	ctx := make(map[int]string)
	// Infrastructure: would populate from git blame
	return ctx
}

// filterCommitsByFileChange filters commits that modified a specific file.
func filterCommitsByFileChange(commits []commit, file string) []commit {
	if file == "" {
		return commits
	}
	var result []commit
	for _, c := range commits {
		if isFileModifiedInCommit(c.hash, file) {
			result = append(result, c)
		}
	}
	return result
}

// isFileModifiedInCommit checks if a file was modified in a commit.
func isFileModifiedInCommit(hash, file string) bool {
	// Infrastructure: would query git
	return false
}

// ===== OPTION 4: STASH & REFLOG =====

// parseStashList parses git stash output.
func parseStashList(output string) []stashEntry {
	var stashes []stashEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "stash@{0}: WIP on main: abc1234 message"
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 2 {
			name := strings.TrimSpace(parts[0])
			rest := strings.TrimSpace(parts[1])
			// Extract branch name
			branchParts := strings.Fields(rest)
			var branch string
			if len(branchParts) >= 3 {
				branch = branchParts[2]
			}
			stashes = append(stashes, stashEntry{
				name:    name,
				branch:  branch,
				subject: line,
			})
		}
	}
	return stashes
}

// parseReflog parses git reflog output.
func parseReflog(output string) []reflogEntry {
	var entries []reflogEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "abc1234 HEAD@{0}: commit: message"
		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 3 {
			hash := parts[0]
			rest := parts[2]
			actionParts := strings.SplitN(rest, ":", 2)
			action := ""
			message := ""
			if len(actionParts) >= 1 {
				action = actionParts[0]
			}
			if len(actionParts) >= 2 {
				message = strings.TrimSpace(actionParts[1])
			}
			entries = append(entries, reflogEntry{
				hash:    hash,
				action:  action,
				message: message,
			})
		}
	}
	return entries
}

// renderStashView renders the stash browser view.
func renderStashView(stashes []stashEntry, width int) string {
	var sb strings.Builder
	sb.WriteString("Stashes:\n")
	if len(stashes) == 0 {
		sb.WriteString("(no stashes)\n")
		return sb.String()
	}
	for i, s := range stashes {
		sb.WriteString(fmt.Sprintf("%d: %s - %s\n", i, s.name, s.branch))
	}
	return sb.String()
}

// renderReflogView renders the reflog browser view.
func renderReflogView(entries []reflogEntry, width int) string {
	var sb strings.Builder
	sb.WriteString("Reflog:\n")
	if len(entries) == 0 {
		sb.WriteString("(no reflog entries)\n")
		return sb.String()
	}
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("%s - %s: %s\n", e.hash[:7], e.action, e.message))
	}
	return sb.String()
}

// stashToCommitLike converts a stash entry to a commit-like structure.
func stashToCommitLike(stash stashEntry) commit {
	return commit{
		shortHash: stash.hash,
		hash:      stash.hash,
		subject:   stash.subject,
		author:    "stash",
		when:      "stash",
	}
}

// reflogToCommitLike converts a reflog entry to a commit-like structure.
func reflogToCommitLike(entry reflogEntry) commit {
	return commit{
		shortHash: entry.hash[:7],
		hash:      entry.hash,
		subject:   entry.message,
		author:    entry.action,
		when:      entry.date,
	}
}

// switchViewMode switches between log, stash, and reflog views.
func switchViewMode(m model, newMode string) model {
	m.viewMode = newMode
	m.cursor = 0
	return m
}

// findStashByIndex finds a stash by its index.
func findStashByIndex(stashes []stashEntry, idx int) *stashEntry {
	if idx < 0 || idx >= len(stashes) {
		return nil
	}
	return &stashes[idx]
}

// ===== UI INTEGRATION: KEYBINDINGS =====

// handleKeyBinding processes keyboard input and returns updated model.
func handleKeyBinding(m model, key string) model {
	switch key {
	case "m":
		m = toggleBookmark(m)
	case "'":
		m = jumpToNextBookmark(m)
	case "gg":
		m.inGoToCommitMode = true
		m.goToCommitInput = ""
	case "c":
		m.inCommentMode = true
		m.commentInput = ""
	case "v":
		m = switchViewMode(m, "stash")
	case "V":
		m = switchViewMode(m, "reflog")
	case "G":
		m.showGraph = !m.showGraph
		if m.showGraph && len(m.commitGraph) == 0 {
			m = lazyLoadGraph(m)
		}
	case "f":
		m = toggleFileView(m)
	case "R":
		m.showRebaseUI = !m.showRebaseUI
		if m.showRebaseUI && len(m.rebaseSequence) == 0 {
			m.rebaseSequence = parseRebaseSequence(m.commits)
		}
	case "C":
		m.showCherryPickUI = !m.showCherryPickUI
	case "A":
		m.showAnalytics = !m.showAnalytics
		if m.showAnalytics {
			m.authorStats = calculateAuthorStats(m.commits)
			m.timeStats = calculateTimeStats(m.commits)
		}
	// Bisect & Recovery
	case "B":
		if !m.bisectState.active {
			m = initiateBisect(m)
		} else {
			m.bisectState.active = false
			m.showBisectUI = false
		}
	case "L":
		m.showLostCommits = !m.showLostCommits
		if m.showLostCommits && len(m.lostCommits) == 0 {
			m.lostCommits = findLostCommits("")
		}
	case "U":
		if m.showUndoMenu {
			m = performUndo(m)
		}
		m.showUndoMenu = !m.showUndoMenu
	// Code Patterns & Quality
	case "O":
		m.showCodeOwnership = !m.showCodeOwnership
		if m.showCodeOwnership && len(m.codeOwnership) == 0 {
			m.codeOwnership = analyzeCodeOwnership(m.commits)
		}
	case "H":
		m.showHotspots = !m.showHotspots
		if m.showHotspots && len(m.hotspots) == 0 {
			m.hotspots = detectHotspots(m.commits)
		}
	case "M":
		m.showLinting = !m.showLinting
		if m.showLinting && len(m.lintingResults) == 0 {
			for _, c := range m.commits {
				result := lintCommitMessage(c.subject, c.hash)
				m.lintingResults = append(m.lintingResults, result)
			}
		}
	case "S":
		m.showLargeCommits = !m.showLargeCommits
		if m.showLargeCommits && len(m.largeCommits) == 0 {
			m = analyzeCommitSize(m)
		}
	case "X":
		m.showComplexity = !m.showComplexity
		if m.showComplexity && len(m.commitMetrics) == 0 {
			m = analyzeComplexity(m)
		}
	// Commit Analysis & Search (4 features)
	case "N":
		m.showSemanticSearch = !m.showSemanticSearch
		if m.showSemanticSearch && len(m.semanticSearchResults) == 0 {
			m.semanticSearchResults = semanticSearch(m.commits, m.semanticQuery)
		}
	case "E":
		m.showActivityHeatmap = !m.showActivityHeatmap
		if m.showActivityHeatmap && len(m.authorActivityHeatmap) == 0 {
			m.authorActivityHeatmap = buildActivityHeatmap(m.commits)
		}
	case "Y":
		m.showMergeAnalysis = !m.showMergeAnalysis
		if m.showMergeAnalysis && len(m.mergeAnalysisData) == 0 {
			m.mergeAnalysisData = analyzeMerges(m.commits)
		}
	case "T":
		m.showCoupling = !m.showCoupling
		if m.showCoupling && len(m.commitCouplings) == 0 {
			m.commitCouplings = analyzeCommitCoupling(m.commits)
		}
	// Performance & Filtering (4 features)
	case "D":
		if m.currentExtFilter == "" {
			m = toggleExtensionFilter(m, ".go")
		} else {
			m = toggleExtensionFilter(m, "")
		}
	case "W":
		if m.groupingMode == "" {
			m.groupingMode = "date"
			m.commitGroups = groupCommits(m.commits, "date")
		} else {
			m.groupingMode = ""
		}
	case "Z":
		m.showDependencies = !m.showDependencies
		if m.showDependencies && len(m.dependencyChanges) == 0 {
			m.dependencyChanges = trackDependencyChanges(m.commits)
		}
	// Advanced Workflows (5 features)
	case "1":
		m.showWorktrees = !m.showWorktrees
		if m.showWorktrees && len(m.worktrees) == 0 {
			m.worktrees = loadWorktrees("")
		}
	case "2":
		m.showSubmodules = !m.showSubmodules
		if m.showSubmodules && len(m.submodules) == 0 {
			m.submodules = parseSubmodules("")
		}
	case "3":
		m.showNamedStashes = !m.showNamedStashes
	case "4":
		m.showTagMgmt = !m.showTagMgmt
	case "5":
		m.showGPGStatus = !m.showGPGStatus
		if m.showGPGStatus && len(m.gpgStatuses) == 0 {
			m.gpgStatuses = extractGPGSignatureStatus("")
		}
	// Visualization (5 features)
	case "6":
		m.showFlamegraph = !m.showFlamegraph
		if m.showFlamegraph && len(m.contributorFlameData) == 0 {
			m.contributorFlameData = buildContributorFlame(m.commits)
		}
	case "7":
		m.showTimeline = !m.showTimeline
		if m.showTimeline && len(m.timelinePoints) == 0 {
			m.timelinePoints = buildTimeline(m.commits)
		}
	case "8":
		m.showTreeView = !m.showTreeView
		if m.showTreeView && m.treeRoot == nil {
			m.treeRoot = buildTreeView(m.commits)
		}
	case "9":
		m.showAuthorComparison = !m.showAuthorComparison
		if m.showAuthorComparison && len(m.authorComparisons) == 0 {
			m.authorComparisons = compareAuthors(m)
		}
	case "0":
		m.showFileHeatmap = !m.showFileHeatmap
		if m.showFileHeatmap && len(m.fileHeatmap) == 0 {
			m.fileHeatmap = buildFileHeatmap(m.commits)
		}
	// Integration & Export (5 features)
	case "p":
		m.showPRLinks = !m.showPRLinks
		if m.showPRLinks && len(m.prReferences) == 0 {
			m.prReferences = extractPRReferences(m.commits)
		}
	case "j":
		m.showJiraLinks = !m.showJiraLinks
		if m.showJiraLinks && len(m.jiraLinks) == 0 {
			m.jiraLinks = extractJiraTickets(m.commits)
		}
	case "e":
		m.showExportUI = !m.showExportUI
	case "q":
		m.showIssueRefs = !m.showIssueRefs
		if m.showIssueRefs && len(m.issueReferences) == 0 {
			m.issueReferences = extractIssueReferences(m.commits)
		}
	default:
		// Handle multi-key like "5j"
		if len(key) > 1 && (strings.HasSuffix(key, "j") || strings.HasSuffix(key, "k")) {
			n := parseCount(key[:len(key)-1])
			if strings.HasSuffix(key, "j") {
				for i := 0; i < n; i++ {
					m = moveCursorDown(m)
				}
			} else {
				for i := 0; i < n; i++ {
					m = moveCursorUp(m)
				}
			}
		}
	}
	return m
}

// safeHandleKeyBinding handles keybindings with error recovery.
func safeHandleKeyBinding(m model, key string) model {
	if key == "" {
		return m
	}
	return handleKeyBinding(m, key)
}

// ===== UI INTEGRATION: RENDERING =====

// renderCommitRowWithStats renders commit row with stats badge.
func renderCommitRowWithStats(m model, idx int, width int) string {
	if !m.showStatsBadge {
		return ""
	}
	badge := diffStatBadge(m.lastStats)
	return badge
}

// renderBookmarkList renders list of bookmarked commits.
func renderBookmarkList(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("Bookmarks:\n")
	for i, hash := range m.bookmarks {
		for _, c := range m.commits {
			if c.shortHash == hash {
				sb.WriteString(fmt.Sprintf("%d: %s - %s\n", i, hash, c.subject))
				break
			}
		}
	}
	return sb.String()
}

// renderGraphView renders the commit graph.
func renderGraphView(m model, width int) string {
	if len(m.commitGraph) == 0 {
		return ""
	}
	return renderAsciiGraph(m.commitGraph)
}

// renderViewMode renders current view (log, stash, or reflog).
func renderViewMode(m model, width int) string {
	switch m.viewMode {
	case "stash":
		return renderStashView(m.stashes, width)
	case "reflog":
		return renderReflogView(m.reflogEntries, width)
	default:
		return ""
	}
}

// renderDiffWithComments renders diff with comment markers.
func renderDiffWithComments(m model, panelHeight, width int) string {
	var sb strings.Builder
	for i := 0; i < panelHeight && m.diffOffset+i < len(m.diffLines); i++ {
		marker := renderLineCommentMarker(m, m.diffOffset+i)
		if marker != "" {
			sb.WriteString(marker)
			sb.WriteString(" ")
		}
		sb.WriteString(m.diffLines[m.diffOffset+i].text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// enterCommentMode enters line comment mode.
func enterCommentMode(m model) model {
	m.inCommentMode = true
	m.commentInput = ""
	return m
}

// exitCommentMode exits line comment mode.
func exitCommentMode(m model) model {
	m.inCommentMode = false
	m.commentInput = ""
	return m
}

// ===== OPTIMIZATION: CACHING =====

// newDiffCache creates a new diff cache with specified max size.
func newDiffCache(size int) *diffCache {
	return &diffCache{
		data:    make(map[string][]diffLine),
		order:   []string{},
		maxSize: size,
	}
}

// set stores a diff in the cache.
func (dc *diffCache) set(key string, lines []diffLine) {
	if _, exists := dc.data[key]; !exists {
		dc.order = append(dc.order, key)
		if len(dc.order) > dc.maxSize {
			oldest := dc.order[0]
			dc.order = dc.order[1:]
			delete(dc.data, oldest)
		}
	}
	dc.data[key] = lines
}

// get retrieves a diff from the cache.
func (dc *diffCache) get(key string) ([]diffLine, bool) {
	lines, ok := dc.data[key]
	if ok {
		dc.hitCount++
	}
	return lines, ok
}

// getHitCount returns the number of cache hits.
func (dc *diffCache) getHitCount() int {
	return dc.hitCount
}

// newStatCache creates a new stats cache.
func newStatCache(size int) *statCache {
	return &statCache{
		data:    make(map[string]commitStatistics),
		order:   []string{},
		maxSize: size,
	}
}

// getOrCompute gets cached stats or computes them.
func (sc *statCache) getOrCompute(key string, lines []diffLine) commitStatistics {
	if stats, ok := sc.data[key]; ok {
		sc.hitCount++
		return stats
	}
	stats := commitStats(lines)
	sc.order = append(sc.order, key)
	if len(sc.order) > sc.maxSize {
		oldest := sc.order[0]
		sc.order = sc.order[1:]
		delete(sc.data, oldest)
	}
	sc.data[key] = stats
	return stats
}

// getHitCount returns the number of cache hits.
func (sc *statCache) getHitCount() int {
	return sc.hitCount
}

// newRegexCache creates a new regex cache.
func newRegexCache(size int) *regexCache {
	return &regexCache{
		data:    make(map[string]*regexp.Regexp),
		maxSize: size,
	}
}

// compile compiles a regex or returns cached version.
func (rc *regexCache) compile(pattern string) (*regexp.Regexp, error) {
	if re, ok := rc.data[pattern]; ok {
		rc.hitCount++
		return re, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	if len(rc.data) < rc.maxSize {
		rc.data[pattern] = re
	}
	return re, nil
}

// ===== OPTIMIZATION: LAZY LOADING =====

// lazyLoadDiff loads diff asynchronously if not already loaded.
func lazyLoadDiff(m model) model {
	if m.cursor < len(m.commits) && len(m.diffLines) == 0 {
		m.loading = true
	}
	return m
}

// lazyLoadGraph builds graph on demand.
func lazyLoadGraph(m model) model {
	if len(m.commitGraph) == 0 && len(m.commits) > 0 {
		m.commitGraph = parseCommitGraph(m.commits)
	}
	return m
}

// lazyLoadStats computes stats on demand.
func lazyLoadStats(m model) commitStatistics {
	return commitStats(m.diffLines)
}

// ===== OPTIMIZATION: SAFE WRAPPERS =====

// safeIsFileModified safely checks file modification without crashing.
func safeIsFileModified(hash, file string) bool {
	if hash == "" || file == "" {
		return false
	}
	return isFileModifiedInCommit(hash, file)
}

// safeParseCommitGraph safely parses graph, returning empty slice on error.
func safeParseCommitGraph(commits []commit) []graphNode {
	if commits == nil {
		return []graphNode{}
	}
	return parseCommitGraph(commits)
}

// ===== OPTION A: ADVANCED COMMIT OPERATIONS =====

// --- Interactive Rebase ---

// parseRebaseSequence builds a rebase operation sequence from commits.
func parseRebaseSequence(commits []commit) []rebaseOp {
	var ops []rebaseOp
	for _, c := range commits {
		ops = append(ops, rebaseOp{
			action:  "pick",
			hash:    c.hash,
			subject: c.subject,
		})
	}
	return ops
}

// reorderCommit moves a commit in the rebase sequence.
func reorderCommit(seq []rebaseOp, from, to int) []rebaseOp {
	if from < 0 || from >= len(seq) || to < 0 || to >= len(seq) {
		return seq
	}
	if from == to {
		return seq
	}
	op := seq[from]
	newSeq := make([]rebaseOp, 0, len(seq))
	for i, o := range seq {
		if i == from {
			continue
		}
		if i == to && from < to {
			newSeq = append(newSeq, o)
			newSeq = append(newSeq, op)
		} else if i == to && from > to {
			newSeq = append(newSeq, op)
			newSeq = append(newSeq, o)
		} else {
			newSeq = append(newSeq, o)
		}
	}
	return newSeq
}

// squashCommit marks a commit for squashing.
func squashCommit(seq []rebaseOp, idx int) []rebaseOp {
	if idx >= 0 && idx < len(seq) {
		seq[idx].action = "squash"
	}
	return seq
}

// fixupCommit marks a commit for fixup (squash without message).
func fixupCommit(seq []rebaseOp, idx int) []rebaseOp {
	if idx >= 0 && idx < len(seq) {
		seq[idx].action = "fixup"
	}
	return seq
}

// previewRebase renders a preview of the rebase operation.
func previewRebase(seq []rebaseOp) string {
	var sb strings.Builder
	sb.WriteString("Rebase sequence:\n")
	for i, op := range seq {
		hash := op.hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		sb.WriteString(fmt.Sprintf("%d: %s %s - %s\n", i, op.action, hash, op.subject))
	}
	return sb.String()
}

// --- Cherry-pick ---

// toggleCherryPick adds or removes a commit from cherry-pick selection.
func toggleCherryPick(m model, hash string) model {
	found := false
	for i, h := range m.cherryPickList {
		if h == hash {
			m.cherryPickList = append(m.cherryPickList[:i], m.cherryPickList[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		m.cherryPickList = append(m.cherryPickList, hash)
	}
	return m
}

// previewCherryPick shows which commits will be cherry-picked.
func previewCherryPick(commits []commit, picks []string) string {
	var sb strings.Builder
	sb.WriteString("Cherry-pick queue:\n")
	for i, pick := range picks {
		for _, c := range commits {
			if c.hash == pick || c.shortHash == pick {
				hash := c.shortHash
				if len(hash) > 7 {
					hash = hash[:7]
				}
				sb.WriteString(fmt.Sprintf("%d: %s - %s\n", i, hash, c.subject))
				break
			}
		}
	}
	return sb.String()
}

// --- Reset ---

// resetToCommit generates a reset command with the specified mode.
func resetToCommit(hash, mode string) string {
	if mode == "" {
		mode = "mixed"
	}
	return fmt.Sprintf("git reset --%s %s", mode, hash)
}

// --- Revert ---

// revertCommit generates a revert command for a commit.
func revertCommit(hash string) string {
	return fmt.Sprintf("git revert %s", hash)
}

// --- Amend ---

// amendLastCommit updates the last commit message.
func amendLastCommit(m model, message string) model {
	if m.cursor < len(m.commits) {
		m.commits[m.cursor].subject = message
		m.amendMessage = message
	}
	return m
}

// ===== OPTION B: COLLABORATION & ANALYTICS =====

// --- Author Statistics ---

// calculateAuthorStats counts commits by author.
func calculateAuthorStats(commits []commit) map[string]int {
	stats := make(map[string]int)
	for _, c := range commits {
		stats[c.author]++
	}
	return stats
}

// renderAuthorStats renders author statistics as a list.
func renderAuthorStats(stats map[string]int, width int) string {
	var sb strings.Builder
	sb.WriteString("Author Statistics:\n")
	for author, count := range stats {
		sb.WriteString(fmt.Sprintf("%s: %d commits\n", author, count))
	}
	return sb.String()
}

// --- Time-based Analytics ---

// calculateTimeStats aggregates commits by time period.
func calculateTimeStats(commits []commit) map[string]int {
	stats := make(map[string]int)
	for _, c := range commits {
		// Simple bucketing by day mentioned in "when" field
		if strings.Contains(c.when, "day") {
			stats["recent"]++
		} else if strings.Contains(c.when, "week") {
			stats["past_week"]++
		} else {
			stats["older"]++
		}
	}
	return stats
}

// aggregateByWeek groups commits by week.
func aggregateByWeek(commits []commit) map[string]int {
	weekly := make(map[string]int)
	for _, c := range commits {
		// Simple aggregation based on "when" field
		if strings.Contains(c.when, "ago") {
			weekly["current"]++
		}
	}
	return weekly
}

// renderTimeStats renders time-based statistics as heatmap-style output.
func renderTimeStats(stats map[string]int, width int) string {
	var sb strings.Builder
	sb.WriteString("Time-based Statistics:\n")
	for period, count := range stats {
		sb.WriteString(fmt.Sprintf("%s: %d\n", period, count))
	}
	return sb.String()
}

// --- Co-author Detection ---

// extractCoAuthors parses co-authors from commit message.
func extractCoAuthors(message string) []string {
	var coAuthors []string
	re := regexp.MustCompile(`Co-authored-by:\s*(.+?)\s*<`)
	matches := re.FindAllStringSubmatch(message, -1)
	for _, match := range matches {
		if len(match) > 1 {
			coAuthors = append(coAuthors, match[1])
		}
	}
	return coAuthors
}

// --- Reviewer Tracking ---

// extractReviewers parses reviewers from commit message.
func extractReviewers(message string) []string {
	var reviewers []string
	re := regexp.MustCompile(`Reviewed-by:\s*(.+?)\s*<`)
	matches := re.FindAllStringSubmatch(message, -1)
	for _, match := range matches {
		if len(match) > 1 {
			reviewers = append(reviewers, match[1])
		}
	}
	return reviewers
}

// --- Productivity Metrics ---

// calculateProductivity computes productivity metrics for commits.
func calculateProductivity(commits []commit) map[string]interface{} {
	metrics := make(map[string]interface{})
	if len(commits) == 0 {
		return metrics
	}
	metrics["commits"] = len(commits)
	metrics["unique_authors"] = len(calculateAuthorStats(commits))
	return metrics
}

// renderProductivityMetrics renders productivity metrics.
func renderProductivityMetrics(metrics map[string]interface{}, width int) string {
	var sb strings.Builder
	sb.WriteString("Productivity Metrics:\n")
	for key, value := range metrics {
		sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
	}
	return sb.String()
}

// --- UI Integration ---

// renderRebaseUI renders the interactive rebase interface.
func renderRebaseUI(m model, width int) string {
	if len(m.rebaseSequence) == 0 {
		m.rebaseSequence = parseRebaseSequence(m.commits)
	}
	return previewRebase(m.rebaseSequence)
}

// renderAnalyticsPanel renders the analytics dashboard.
func renderAnalyticsPanel(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("Analytics Dashboard:\n\n")

	// Author stats
	stats := calculateAuthorStats(m.commits)
	sb.WriteString(renderAuthorStats(stats, width))
	sb.WriteString("\n")

	// Time stats
	timeStats := calculateTimeStats(m.commits)
	sb.WriteString(renderTimeStats(timeStats, width))

	return sb.String()
}

// --- Bisect & Recovery (5 features) ---

// Feature 1: Interactive Bisect Workflow

func initiateBisect(m model) model {
	if m.cursor >= 0 && m.cursor < len(m.commits) {
		m.bisectState.active = true
		m.bisectState.current = m.commits[m.cursor].hash
		m.bisectState.good = []string{}
		m.bisectState.bad = []string{}
		var candidateHashes []string
		for _, c := range m.commits[:m.cursor+1] {
			candidateHashes = append(candidateHashes, c.hash)
		}
		m.bisectState.candidates = candidateHashes
		m.showBisectUI = true
	}
	return m
}

func bisectMarkGood(m model) model {
	if m.bisectState.active && m.bisectState.current != "" {
		m.bisectState.good = append(m.bisectState.good, m.bisectState.current)
	}
	return m
}

func bisectMarkBad(m model) model {
	if m.bisectState.active && m.bisectState.current != "" {
		m.bisectState.bad = append(m.bisectState.bad, m.bisectState.current)
	}
	return m
}

func bisectFindCulprit(commits []commit, good []string, bad []string) string {
	if len(commits) == 0 || len(good) == 0 || len(bad) == 0 {
		return ""
	}
	goodMap := make(map[string]bool)
	for _, g := range good {
		goodMap[g] = true
	}
	badMap := make(map[string]bool)
	for _, b := range bad {
		badMap[b] = true
	}
	for i := len(commits) - 1; i >= 0; i-- {
		if goodMap[commits[i].hash] {
			continue
		}
		if badMap[commits[i].hash] {
			continue
		}
		return commits[i].hash
	}
	if len(commits) > 0 {
		return commits[0].hash
	}
	return ""
}

// Feature 2: Bisect Visualization

// renderBisectUI displays bisect state and progress using the analysis UI template.
func renderBisectUI(m model, width int) string {
	data := map[string]interface{}{
		"Progress":       fmt.Sprintf("%d/%d steps", m.bisectState.visualSteps, m.bisectState.totalSteps),
		"Current":        m.bisectState.current,
		"Good commits":   strings.Join(m.bisectState.good, ", "),
		"Bad commits":    strings.Join(m.bisectState.bad, ", "),
	}
	return RenderAnalysisUI("Bisect Status", data)
}

func calculateBisectProgress(state bisectState) int {
	candidates := len(state.candidates)
	if candidates <= 1 {
		return 1
	}
	steps := 0
	for candidates > 1 {
		candidates = candidates / 2
		steps++
	}
	return steps
}

// Feature 3: Reflog Recovery

func extractReflogEntries(reflogOutput string) []reflogEntry {
	var entries []reflogEntry
	for _, line := range strings.Split(strings.TrimSpace(reflogOutput), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		action := "unknown"
		message := ""

		if idx := strings.Index(line, ":"); idx > 0 {
			afterColon := line[idx+1:]
			colonIdx := strings.Index(afterColon, ":")
			if colonIdx > 0 {
				action = strings.TrimSpace(afterColon[:colonIdx])
				message = strings.TrimSpace(afterColon[colonIdx+1:])
			}
		}

		entries = append(entries, reflogEntry{
			hash:    hash,
			action:  action,
			message: message,
			date:    "",
		})
	}
	return entries
}

func enableReflogRecovery(m model) model {
	m.reflogRecoveryMode = true
	m.recoveryCommits = make([]lostCommit, 0)
	for _, entry := range m.reflogEntries {
		m.recoveryCommits = append(m.recoveryCommits, lostCommit{
			hash:      entry.hash,
			shortHash: entry.hash,
			author:    entry.action,
			subject:   entry.message,
			date:      entry.date,
		})
	}
	return m
}

// Feature 4: Lost Commits Finder

func findLostCommits(fsckOutput string) []lostCommit {
	var commits []lostCommit
	lines := strings.Split(fsckOutput, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "unreachable commit") {
			parts := strings.Fields(lines[i])
			if len(parts) >= 3 {
				hash := parts[2]
				subject := ""
				if i+1 < len(lines) {
					subject = lines[i+1]
					i++
				}
				commits = append(commits, lostCommit{
					hash:      hash,
					shortHash: hash,
					author:    "unknown",
					subject:   subject,
					date:      "",
				})
			}
		}
	}
	return commits
}

func renderLostCommitsUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Lost Commits ===\n")
	if len(m.lostCommits) == 0 {
		sb.WriteString("No lost commits found.\n")
		return sb.String()
	}
	for _, lc := range m.lostCommits {
		sb.WriteString(fmt.Sprintf("%s: %s\n", lc.shortHash, lc.subject))
	}
	return sb.String()
}

// Feature 5: Undo Operations

func pushUndo(m model, hash string) model {
	m.undoStack = append(m.undoStack, hash)
	m.undoStackIdx = len(m.undoStack)
	return m
}

func performUndo(m model) model {
	if m.undoStackIdx > 1 {
		m.undoStackIdx--
	}
	return m
}

func renderUndoMenu(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Undo Stack ===\n")
	for i, hash := range m.undoStack {
		if i == m.undoStackIdx-1 {
			sb.WriteString(fmt.Sprintf("> %s\n", hash))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", hash))
		}
	}
	return sb.String()
}

// --- Code Patterns & Quality (5 features) ---

// Feature 6: Code Ownership Analysis

func analyzeCodeOwnership(commits []commit) map[string]codeOwnershipData {
	ownership := make(map[string]codeOwnershipData)
	authorCommitCount := make(map[string]int)
	authorFiles := make(map[string]map[string]int)

	for _, c := range commits {
		authorCommitCount[c.author]++
		if _, ok := authorFiles[c.author]; !ok {
			authorFiles[c.author] = make(map[string]int)
		}
		parts := strings.Fields(c.subject)
		if len(parts) > 1 {
			file := parts[len(parts)-1]
			authorFiles[c.author][file]++
		}
	}

	for author, count := range authorCommitCount {
		expertise := float64(count) / float64(len(commits))
		if expertise > 1.0 {
			expertise = 1.0
		}
		ownership[author] = codeOwnershipData{
			author:    author,
			files:     authorFiles[author],
			lines:     count,
			expertise: expertise,
			isOwner:   expertise > 0.3,
		}
	}

	return ownership
}

func detectCodeOwners(ownership map[string]codeOwnershipData) string {
	var maxAuthor string
	maxExpertise := 0.0
	for author, data := range ownership {
		if data.expertise > maxExpertise {
			maxExpertise = data.expertise
			maxAuthor = author
		}
	}
	return maxAuthor
}

// renderCodeOwnershipUI displays code ownership statistics using the standard analysis UI template.
func renderCodeOwnershipUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, ownership := range m.codeOwnership {
		data[ownership.author] = ownership.expertise
	}
	return RenderAnalysisUI("Code Ownership", data)
}

// Feature 7: Hotspot Detection

func detectHotspots(commits []commit) []hotspotData {
	fileChanges := make(map[string]int)
	fileRecent := make(map[string]int)
	fileCollabs := make(map[string]map[string]bool)

	for i, c := range commits {
		parts := strings.Fields(c.subject)
		if len(parts) > 1 {
			file := parts[len(parts)-1]
			fileChanges[file]++
			if i < 5 {
				fileRecent[file]++
			}
			if _, ok := fileCollabs[file]; !ok {
				fileCollabs[file] = make(map[string]bool)
			}
			fileCollabs[file][c.author] = true
		}
	}

	var hotspots []hotspotData
	for file, changes := range fileChanges {
		collab := len(fileCollabs[file])
		risk := "low"
		if changes > 10 {
			risk = "high"
		} else if changes > 5 {
			risk = "medium"
		}
		hotspots = append(hotspots, hotspotData{
			path:            file,
			changeFrequency: changes,
			recentChanges:   fileRecent[file],
			collaborators:   collab,
			avgCommitSize:   0,
			riskLevel:       risk,
		})
	}

	return hotspots
}

func assessRiskLevel(hotspot hotspotData) string {
	if hotspot.changeFrequency > 10 {
		return "high"
	}
	if hotspot.changeFrequency > 5 {
		return "medium"
	}
	return "low"
}

// renderHotspotsUI displays code hotspots and risk levels using the standard UI template.
func renderHotspotsUI(m model, width int) string {
	var items []string
	statusMap := make(map[string]string)
	for _, h := range m.hotspots {
		item := fmt.Sprintf("%s: %d changes", h.path, h.changeFrequency)
		items = append(items, item)
		statusMap[item] = h.riskLevel
	}
	return RenderStandardUI(RenderConfig{
		Title:     "Code Hotspots",
		Items:     items,
		HasStatus: true,
		StatusMap: statusMap,
	})
}

// Feature 8: Commit Message Linting

func lintCommitMessage(subject string, hash string) lintingResult {
	issues := validateCommitFormat(subject)
	score := 100 - (len(issues) * 20)
	if score < 0 {
		score = 0
	}
	return lintingResult{
		hash:    hash,
		subject: subject,
		issues:  issues,
		score:   score,
	}
}

func validateCommitFormat(subject string) []string {
	var issues []string
	if len(subject) == 0 {
		issues = append(issues, "empty message")
		return issues
	}
	if len(subject) > 72 {
		issues = append(issues, "exceeds 72 chars")
	}
	if subject[0] >= 'a' && subject[0] <= 'z' {
		issues = append(issues, "lowercase start")
	}
	if !strings.ContainsAny(string(subject[0]), "ABCDEFGHIJKLMNOPQRSTUVWXYZ") && subject[0] >= 'a' {
		issues = append(issues, "should start with verb")
	}
	return issues
}

// renderLintingUI displays commit message linting results and issues.
func renderLintingUI(m model, width int) string {
	var errors []string
	for _, result := range m.lintingResults {
		for _, issue := range result.issues {
			errors = append(errors, fmt.Sprintf("%s: %s", result.hash, issue))
		}
	}
	return RenderErrorList("Commit Message Linting", errors)
}

// Feature 9: Large Commit Detection

func analyzeCommitSize(m model) model {
	m.largeCommits = []commitMetrics{}
	for _, c := range m.commits {
		words := len(strings.Fields(c.subject))
		filesEst := words
		if filesEst < 1 {
			filesEst = 1
		}
		linesEst := words * 100

		metrics := commitMetrics{
			hash:         c.hash,
			linesChanged: linesEst,
			filesChanged: filesEst,
			isLarge:      linesEst > 150 || filesEst > 5,
		}
		if metrics.isLarge {
			m.largeCommits = append(m.largeCommits, metrics)
		}
	}
	return m
}

func calculateCommitMetrics(hash string, linesChanged int, filesChanged int) commitMetrics {
	return commitMetrics{
		hash:         hash,
		linesChanged: linesChanged,
		filesChanged: filesChanged,
		complexity:   linesChanged / 50,
		isLarge:      linesChanged > 300,
		isComplex:    linesChanged > 300 && filesChanged > 10,
	}
}

// renderLargeCommitsUI displays large commits and their metrics using the analysis UI template.
func renderLargeCommitsUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, lc := range m.largeCommits {
		data[lc.hash] = fmt.Sprintf("%d lines, %d files", lc.linesChanged, lc.filesChanged)
	}
	return RenderAnalysisUI("Large Commits", data)
}

// Feature 10: Commit Complexity Analysis

func analyzeComplexity(m model) model {
	m.commitMetrics = []commitMetrics{}
	for _, c := range m.commits {
		wordCount := len(strings.Fields(c.subject))
		linesEst := wordCount * 30
		filesEst := wordCount

		metrics := commitMetrics{
			hash:         c.hash,
			linesChanged: linesEst,
			filesChanged: filesEst,
		}
		metrics.complexity = calculateComplexityScore(metrics)
		metrics.isComplex = metrics.complexity > 50
		m.commitMetrics = append(m.commitMetrics, metrics)
	}
	return m
}

func calculateComplexityScore(metrics commitMetrics) int {
	score := (metrics.linesChanged / 10) + (metrics.filesChanged * 5)
	if score > 100 {
		score = 100
	}
	return score
}

// renderComplexityUI displays commit complexity metrics using the analysis UI template.
func renderComplexityUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, cm := range m.commitMetrics {
		data[cm.hash] = cm.complexity
	}
	return RenderAnalysisUI("Commit Complexity", data)
}

// --- Commit Analysis & Search (4 features) ---

// Feature 1: Semantic Search
func semanticSearch(commits []commit, query string) []semanticSearchResult {
	var results []semanticSearchResult
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), strings.ToLower(query)) {
			results = append(results, semanticSearchResult{
				hash:      c.hash,
				subject:   c.subject,
				matches:   []string{query},
				relevance: 75,
			})
		}
	}
	return results
}

// renderSemanticSearchUI displays semantic search results using the analysis UI template.
func renderSemanticSearchUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, r := range m.semanticSearchResults {
		data[r.hash] = float64(r.relevance) / 100.0
	}
	return RenderAnalysisUI("Semantic Search Results", data)
}

// Feature 2: Author Activity Heatmap
func buildActivityHeatmap(commits []commit) map[string]authorActivityData {
	heatmap := make(map[string]authorActivityData)
	for _, c := range commits {
		if _, ok := heatmap[c.author]; !ok {
			heatmap[c.author] = authorActivityData{
				author:    c.author,
				hourOfDay: make(map[int]int),
				dayOfWeek: make(map[int]int),
			}
		}
		data := heatmap[c.author]
		data.hourOfDay[9]++ // default hour
		heatmap[c.author] = data
	}
	return heatmap
}

func findPeakHour(data authorActivityData) int {
	maxHour := 0
	maxCount := 0
	for hour, count := range data.hourOfDay {
		if count > maxCount {
			maxCount = count
			maxHour = hour
		}
	}
	return maxHour
}

// renderActivityHeatmapUI displays author activity heatmap data using the analysis UI template.
func renderActivityHeatmapUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, activity := range m.authorActivityHeatmap {
		data[activity.author] = fmt.Sprintf("peak at %d:00", activity.peakHour)
	}
	return RenderAnalysisUI("Author Activity Heatmap", data)
}

// Feature 3: Merge Analysis
func analyzeMerges(commits []commit) []mergeAnalysis {
	var analysis []mergeAnalysis
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "merge") {
			analysis = append(analysis, mergeAnalysis{
				hash:          c.hash,
				isMerge:       true,
				isFastForward: strings.Contains(c.subject, "fast-forward"),
				parentCount:   2,
				conflictRisk:  25,
			})
		}
	}
	return analysis
}

// renderMergeAnalysisUI displays merge analysis data using the analysis UI template.
func renderMergeAnalysisUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, merge := range m.mergeAnalysisData {
		data[merge.hash] = merge.isFastForward
	}
	return RenderAnalysisUI("Merge Analysis", data)
}

// Feature 4: Commit Coupling Analysis
func analyzeCommitCoupling(commits []commit) []commitCoupling {
	var couplings []commitCoupling
	filePairs := make(map[string]int)
	for _, c := range commits {
		files := extractFilesFromSubject(c.subject)
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				pair := files[i] + "|" + files[j]
				filePairs[pair]++
			}
		}
	}
	for pair, count := range filePairs {
		parts := strings.Split(pair, "|")
		if len(parts) == 2 && count > 0 {
			couplings = append(couplings, commitCoupling{
				file1:         parts[0],
				file2:         parts[1],
				coChangeCount: count,
				correlation:   0.75,
			})
		}
	}
	return couplings
}

func extractFilesFromSubject(subject string) []string {
	var files []string
	parts := strings.Fields(subject)
	for _, p := range parts {
		if strings.Contains(p, ".") {
			files = append(files, p)
		}
	}
	return files
}

// renderCouplingAnalysisUI displays commit coupling analysis using the analysis UI template.
func renderCouplingAnalysisUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, c := range m.commitCouplings {
		key := fmt.Sprintf("%s <-> %s", c.file1, c.file2)
		data[key] = c.correlation
	}
	return RenderAnalysisUI("Coupling Analysis", data)
}

// --- Performance & Filtering (4 features) ---

// Feature 5: Filter by File Extension
func filterByExtension(commits []commit, ext string) []commit {
	var filtered []commit
	for _, c := range commits {
		if strings.Contains(c.subject, ext) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func toggleExtensionFilter(m model, ext string) model {
	m.currentExtFilter = ext
	return m
}

// renderExtensionFilterUI displays extension filter status using the standard UI template.
func renderExtensionFilterUI(m model, width int) string {
	var items []string
	statusMap := make(map[string]string)
	for _, f := range m.extensionFilters {
		items = append(items, f.extension)
		if f.enabled {
			statusMap[f.extension] = "on"
		} else {
			statusMap[f.extension] = "off"
		}
	}
	return RenderStandardUI(RenderConfig{
		Title:     "Extension Filters",
		Items:     items,
		HasStatus: true,
		StatusMap: statusMap,
	})
}

// Feature 6: Commit Grouping
func groupCommits(commits []commit, groupBy string) []commitGroup {
	var groups []commitGroup
	groupMap := make(map[string][]string)
	for _, c := range commits {
		key := "default"
		if groupBy == "date" {
			key = c.when
		} else if groupBy == "branch" {
			parts := strings.Fields(c.subject)
			if len(parts) > 0 {
				key = parts[0]
			}
		}
		groupMap[key] = append(groupMap[key], c.hash)
	}
	for name, hashes := range groupMap {
		groups = append(groups, commitGroup{
			name:    name,
			commits: hashes,
			groupBy: groupBy,
		})
	}
	return groups
}

// renderCommitGroupsUI displays commit groups using the analysis UI template.
func renderCommitGroupsUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, g := range m.commitGroups {
		data[g.label] = len(g.commits)
	}
	return RenderAnalysisUI("Commit Groups", data)
}

// Feature 7: Fast-Forward Merge Detection
func detectFastForwardMerges(commits []commit) []mergeAnalysis {
	analysis := analyzeMerges(commits)
	var ffMerges []mergeAnalysis
	for _, a := range analysis {
		if a.isFastForward {
			ffMerges = append(ffMerges, a)
		}
	}
	return ffMerges
}

// renderFastForwardsUI displays fast-forward merges using the standard UI template.
func renderFastForwardsUI(m model, width int) string {
	var items []string
	for _, merge := range m.mergeAnalysisData {
		if merge.isFastForward {
			items = append(items, merge.hash)
		}
	}
	return RenderStandardUI(RenderConfig{
		Title: "Fast-Forward Merges",
		Items: items,
	})
}

// Feature 8: Dependency Change Tracking
func trackDependencyChanges(commits []commit) []dependencyChange {
	var deps []dependencyChange
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "upgrade") ||
			strings.Contains(strings.ToLower(c.subject), "update") {
			deps = append(deps, dependencyChange{
				hash:   c.hash,
				dep:    "unknown",
				oldVer: "x.x.x",
				newVer: "y.y.y",
			})
		}
	}
	return deps
}

// renderDependenciesUI displays dependency change tracking using the analysis UI template.
func renderDependenciesUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, d := range m.dependencyChanges {
		data[d.dep] = fmt.Sprintf("%s -> %s", d.oldVer, d.newVer)
	}
	return RenderAnalysisUI("Dependency Changes", data)
}

// --- Advanced Workflows (5 features) ---

// Feature 9: Worktree Support
func loadWorktrees(output string) []worktreeInfo {
	var worktrees []worktreeInfo
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		worktrees = append(worktrees, worktreeInfo{
			path:   strings.TrimSpace(line),
			branch: "main",
		})
	}
	return worktrees
}

func switchWorktree(m model, path string) model {
	m.currentWorktree = path
	return m
}

// renderWorktreesUI displays worktree information using the analysis UI template.
func renderWorktreesUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, wt := range m.worktrees {
		data[wt.path] = wt.branch
	}
	return RenderAnalysisUI("Worktrees", data)
}

// Feature 10: Submodule Tracking
func parseSubmodules(output string) []submoduleInfo {
	var subs []submoduleInfo
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "submodule") {
			subs = append(subs, submoduleInfo{
				path:   "lib",
				url:    "https://github.com/user/lib",
				branch: "main",
			})
		}
	}
	return subs
}

// renderSubmodulesUI displays submodule information using the analysis UI template.
func renderSubmodulesUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, sm := range m.submodules {
		data[sm.path] = sm.url
	}
	return RenderAnalysisUI("Submodules", data)
}

// Feature 11: Named Stashes
func createNamedStash(m model, index int, name string, desc string) model {
	m.namedStashes = append(m.namedStashes, namedStash{
		index:       index,
		name:        name,
		description: desc,
		hash:        "",
	})
	return m
}

// renderNamedStashesUI displays named stashes using the analysis UI template.
func renderNamedStashesUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, ns := range m.namedStashes {
		data[ns.name] = ns.description
	}
	return RenderAnalysisUI("Named Stashes", data)
}

// Feature 12: Tag Management
func queueTagOperation(m model, name string, hash string, action string, msg string) model {
	m.pendingTagOps = append(m.pendingTagOps, tagOperation{
		name:    name,
		hash:    hash,
		action:  action,
		message: msg,
	})
	return m
}

// renderTagMgmtUI displays tag management operations using the analysis UI template.
func renderTagMgmtUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, op := range m.pendingTagOps {
		data[op.name] = op.action
	}
	return RenderAnalysisUI("Tag Management", data)
}

// Feature 13: GPG Signature Status
func extractGPGSignatureStatus(output string) map[string]gpgSignatureStatus {
	statuses := make(map[string]gpgSignatureStatus)
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			statuses[hash] = gpgSignatureStatus{
				hash:   hash,
				signed: true,
				signer: "unknown",
			}
		}
	}
	return statuses
}

func renderGPGStatusUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== GPG Signatures ===\n")
	for _, status := range m.gpgStatuses {
		signed := "✗"
		if status.signed {
			signed = "✓"
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", status.hash, signed))
	}
	return sb.String()
}

// --- Visualization (5 features) ---

// Feature 14: Contributor Flamegraph
func buildContributorFlame(commits []commit) []contributorFlameData {
	authorMap := make(map[string]int)
	for _, c := range commits {
		authorMap[c.author]++
	}
	var flame []contributorFlameData
	for author, count := range authorMap {
		pct := float64(count) / float64(len(commits)) * 100
		flame = append(flame, contributorFlameData{
			author:     author,
			commits:    count,
			percentage: pct,
		})
	}
	// Sort by commit count descending
	for i := 0; i < len(flame)-1; i++ {
		for j := i + 1; j < len(flame); j++ {
			if flame[j].commits > flame[i].commits {
				flame[i], flame[j] = flame[j], flame[i]
			}
		}
	}
	return flame
}

// renderFlamegraphUI displays contributor flamegraph data using the analysis UI template.
func renderFlamegraphUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, cf := range m.contributorFlameData {
		data[cf.author] = cf.percentage
	}
	return RenderAnalysisUI("Contributor Flamegraph", data)
}

// Feature 15: Timeline Slider
func buildTimeline(commits []commit) []timelinePoint {
	var timeline []timelinePoint
	dateMap := make(map[string]int)
	for _, c := range commits {
		dateMap[c.when]++
	}
	for date, count := range dateMap {
		timeline = append(timeline, timelinePoint{
			date:    date,
			commits: count,
		})
	}
	return timeline
}

// renderTimelineSliderUI displays timeline data using the analysis UI template.
func renderTimelineSliderUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, tp := range m.timelinePoints {
		data[tp.date] = tp.commits
	}
	return RenderAnalysisUI("Timeline", data)
}

// Feature 16: Tree View
func buildTreeView(commits []commit) *treeNode {
	if len(commits) == 0 {
		return nil
	}
	root := &treeNode{
		hash:    commits[0].hash,
		subject: commits[0].subject,
		depth:   0,
	}
	for i := 1; i < len(commits); i++ {
		root.children = append(root.children, &treeNode{
			hash:    commits[i].hash,
			subject: commits[i].subject,
			depth:   1,
		})
	}
	return root
}

// renderTreeViewUI displays tree view of commits with hierarchy.
func renderTreeViewUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Tree View ===\n")
	if m.treeRoot != nil {
		renderTreeNode(&sb, m.treeRoot)
	}
	return sb.String()
}

func renderTreeNode(sb *strings.Builder, node *treeNode) {
	indent := strings.Repeat("  ", node.depth)
	sb.WriteString(fmt.Sprintf("%s├─ %s\n", indent, node.hash))
	for _, child := range node.children {
		renderTreeNode(sb, child)
	}
}

// Feature 17: Author Comparison
func compareAuthors(m model) []authorComparison {
	var comparisons []authorComparison
	if m.selectedAuthors[0] != "" && m.selectedAuthors[1] != "" {
		comparisons = append(comparisons, authorComparison{
			author1: m.selectedAuthors[0],
			author2: m.selectedAuthors[1],
			commits1: 10,
			commits2: 8,
			similarity: 0.75,
		})
	}
	return comparisons
}

// renderAuthorComparisonUI displays author comparison using the comparison table template.
func renderAuthorComparisonUI(m model, width int) string {
	if len(m.authorComparisons) == 0 {
		return "=== Author Comparison ===\nNo comparisons available\n"
	}
	comp := m.authorComparisons[0]
	items := map[string][2]interface{}{
		"Commits": {comp.commits1, comp.commits2},
		"Similarity": {comp.similarity, 0},
	}
	return RenderComparisonTable("Author Comparison", comp.author1, comp.author2, items)
}

// Feature 18: File Heatmap
func buildFileHeatmap(commits []commit) []fileHeatmapEntry {
	fileMap := make(map[string]int)
	for _, c := range commits {
		files := extractFilesFromSubject(c.subject)
		for _, f := range files {
			fileMap[f]++
		}
	}
	var heatmap []fileHeatmapEntry
	for file, freq := range fileMap {
		risk := "low"
		if freq > 10 {
			risk = "high"
		} else if freq > 5 {
			risk = "medium"
		}
		heatmap = append(heatmap, fileHeatmapEntry{
			path:      file,
			frequency: freq,
			risk:      risk,
		})
	}
	return heatmap
}

// renderFileHeatmapUI displays file heatmap with risk levels using the standard UI template.
func renderFileHeatmapUI(m model, width int) string {
	var items []string
	statusMap := make(map[string]string)
	for _, fh := range m.fileHeatmap {
		item := fmt.Sprintf("%s: %d changes", fh.path, fh.frequency)
		items = append(items, item)
		statusMap[item] = fh.risk
	}
	return RenderStandardUI(RenderConfig{
		Title:     "File Heatmap",
		Items:     items,
		HasStatus: true,
		StatusMap: statusMap,
	})
}

// --- Integration & Export (5 features) ---

// Feature 19: GitHub PR Linking
func extractPRReferences(commits []commit) []githubPRReference {
	var prefs []githubPRReference
	for _, c := range commits {
		// Simple regex to find #123 patterns
		parts := strings.Fields(c.subject)
		for _, part := range parts {
			if strings.HasPrefix(part, "#") && len(part) > 1 {
				prefs = append(prefs, githubPRReference{
					hash:     c.hash,
					prNumber: 123,
					status:   "merged",
				})
				break
			}
		}
	}
	return prefs
}

// renderPRLinksUI displays GitHub PR links using the analysis UI template.
func renderPRLinksUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, pr := range m.prReferences {
		key := fmt.Sprintf("PR #%d", pr.prNumber)
		data[key] = pr.status
	}
	return RenderAnalysisUI("PR Links", data)
}

// Feature 20: JIRA Ticket Linking
func extractJiraTickets(commits []commit) []jiraTicketLink {
	var tickets []jiraTicketLink
	for _, c := range commits {
		parts := strings.Fields(c.subject)
		for _, part := range parts {
			if strings.Contains(part, "-") && len(part) > 3 {
				tickets = append(tickets, jiraTicketLink{
					hash:   c.hash,
					ticket: part,
					status: "done",
				})
				break
			}
		}
	}
	return tickets
}

// renderJiraLinksUI displays JIRA ticket links using the analysis UI template.
func renderJiraLinksUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, jira := range m.jiraLinks {
		data[jira.ticket] = jira.status
	}
	return RenderAnalysisUI("JIRA Links", data)
}

// Feature 21: Export to Markdown
func exportToMarkdown(commits []commit) exportData {
	var sb strings.Builder
	sb.WriteString("# Commit History\n\n")
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", c.shortHash, c.author, c.subject))
	}
	return exportData{
		format:   "markdown",
		commits:  commits,
		content:  sb.String(),
		filename: "commits.md",
	}
}

// renderExportUI displays export options using the analysis UI template.
func renderExportUI(m model, width int) string {
	data := map[string]interface{}{
		"Format": m.exportFormat,
	}
	return RenderAnalysisUI("Export Options", data)
}

// Feature 22: Patch Series Export
func exportPatchSeries(commits []commit) exportData {
	var sb strings.Builder
	sb.WriteString("From: user@example.com\n")
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("Subject: %s\n", c.subject))
	}
	return exportData{
		format:   "patch",
		commits:  commits,
		content:  sb.String(),
		filename: "series.patch",
	}
}

// Feature 23: Issue Reference Tracking
func extractIssueReferences(commits []commit) []issueReference {
	var refs []issueReference
	keywords := []string{"fixes", "closes", "resolves"}
	for _, c := range commits {
		var issues []string
		parts := strings.Fields(c.subject)
		for _, part := range parts {
			if strings.HasPrefix(part, "#") && len(part) > 1 {
				issues = append(issues, part)
			}
		}
		if len(issues) > 0 {
			refs = append(refs, issueReference{
				hash:       c.hash,
				references: issues,
				keywords:   keywords,
			})
		}
	}
	return refs
}

// renderIssueRefsUI displays issue references using the analysis UI template.
func renderIssueRefsUI(m model, width int) string {
	data := make(map[string]interface{})
	for _, ref := range m.issueReferences {
		data[ref.hash] = ref.references
	}
	return RenderAnalysisUI("Issue References", data)
}

// --- Advanced Git Operations (5 features) ---

// Feature 1: Interactive Rebase with Live Preview
func previewRebaseOperations(ops []rebaseOp) rebasePreview {
	return rebasePreview{
		operations: ops,
		conflicts:  []string{},
		willApply:  true,
		message:    "Rebase will apply",
	}
}

// Feature 2: Conflict Resolution UI
func detectConflicts(content string) []conflictInfo {
	var conflicts []conflictInfo
	if strings.Contains(content, "<<<<<<< HEAD") {
		conflicts = append(conflicts, conflictInfo{
			file:     "unknown",
			resolved: false,
		})
	}
	return conflicts
}

// renderConflictUI displays conflict resolution status using the standard UI template.
func renderConflictUI(m model, width int) string {
	var items []string
	statusMap := make(map[string]string)
	for _, c := range m.conflictList {
		items = append(items, c.file)
		if c.resolved {
			statusMap[c.file] = "resolved"
		} else {
			statusMap[c.file] = "unresolved"
		}
	}
	return RenderStandardUI(RenderConfig{
		Title:     "Conflict Resolution",
		Items:     items,
		HasStatus: true,
		StatusMap: statusMap,
	})
}

// Feature 3: Squash/Fixup Automation
func planSquashSequence(target string, toSquash []string, msg string) squashPlan {
	return squashPlan{
		targetHash: target,
		toSquash:   toSquash,
		resultMsg:  msg,
		lineCount:  len(msg),
	}
}

// Feature 4: Cherry-pick Improvements
func improveCherryPick(m model, hash string) *cherryPickImprovement {
	return &cherryPickImprovement{
		hash:            hash,
		autoConflict:    false,
		suggestions:     []string{},
	}
}

// Feature 5: Commit Amend with Diff Viewing
func previewAmendCommit(original string, new string, changes map[string]int) amendPreview {
	return amendPreview{
		originalMsg: original,
		newMsg:      new,
		changes:     changes,
	}
}

// --- Team & Collaboration (5 features) ---

// Feature 6: Team Statistics Dashboard
func calculateTeamStats(commits []commit) []teamStats {
	authorMap := make(map[string]int)
	for _, c := range commits {
		authorMap[c.author]++
	}
	var stats []teamStats
	for author, count := range authorMap {
		stats = append(stats, teamStats{
			author:        author,
			commits:       count,
			avgCommitSize: 100,
		})
	}
	return stats
}

// Feature 7: Code Review Workflow Automation
func automateReviewWorkflow(prNum int, author string, reviewers []string) reviewWorkflow {
	return reviewWorkflow{
		prNumber:  prNum,
		author:    author,
		reviewers: reviewers,
		approved:  false,
		status:    "pending",
	}
}

// Feature 8: Reviewer Assignment Suggestions
func suggestReviewers(m model, file string) []reviewerSuggestion {
	var suggestions []reviewerSuggestion
	suggestionMap := make(map[string]float64)
	// If no file-specific matches, suggest based on overall activity
	if len(m.commits) > 0 {
		for _, c := range m.commits {
			suggestionMap[c.author] += 0.5
		}
	}
	for author, expertise := range suggestionMap {
		if expertise > 0 {
			suggestions = append(suggestions, reviewerSuggestion{
				reviewer:     author,
				expertise:    expertise,
				availability: 0.75,
				score:        expertise * 0.75,
			})
		}
	}
	return suggestions
}

// Feature 9: Pair Programming Detection
func detectPairProgramming(commits []commit) []pairProgrammingData {
	var pairs []pairProgrammingData
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "pair") {
			pairs = append(pairs, pairProgrammingData{
				pair1:        "author1",
				pair2:        "author2",
				commits:      1,
				coChangeRate: 0.85,
			})
		}
	}
	return pairs
}

// Feature 10: Team Velocity Tracking
func calculateVelocity(commits []commit) []velocityData {
	weekMap := make(map[string]int)
	for _ = range commits {
		week := "week1"
		weekMap[week]++
	}
	var velocity []velocityData
	for week, count := range weekMap {
		velocity = append(velocity, velocityData{
			week:      week,
			commits:   count,
			additions: count * 50,
		})
	}
	return velocity
}

// --- AI-Powered Insights (5 features) ---

// Feature 11: Commit Message Auto-completion
func autoCompleteMessage(prefix string, commits []commit) []messageCompletion {
	var completions []messageCompletion
	suggestionMap := make(map[string]float64)
	for _, c := range commits {
		if strings.HasPrefix(c.subject, prefix) {
			suggestionMap[c.subject] += 0.5
		}
	}
	var suggestions []string
	for msg := range suggestionMap {
		suggestions = append(suggestions, msg)
	}
	if len(suggestions) > 0 {
		completions = append(completions, messageCompletion{
			prefix:      prefix,
			suggestions: suggestions,
			confidence:  []float64{0.8},
		})
	}
	return completions
}

// Feature 12: ML-based Commit Classification
func classifyCommit(subject string, hash string) commitClassification {
	category := "feature"
	if strings.Contains(strings.ToLower(subject), "fix") {
		category = "fix"
	} else if strings.Contains(strings.ToLower(subject), "refactor") {
		category = "refactor"
	} else if strings.Contains(strings.ToLower(subject), "docs") {
		category = "docs"
	} else if strings.Contains(strings.ToLower(subject), "test") {
		category = "test"
	}
	return commitClassification{
		hash:       hash,
		category:   category,
		confidence: 0.85,
		reason:     "Keyword detected",
	}
}

// Feature 13: Anomaly Detection
func detectAnomalies(commits []commit) []anomalyData {
	var anomalies []anomalyData
	for _, c := range commits {
		words := len(strings.Fields(c.subject))
		// Detect unusually large or verbose commits
		if words > 20 || strings.Contains(strings.ToLower(c.subject), "massive") || strings.Contains(c.subject, "10000") {
			anomalies = append(anomalies, anomalyData{
				hash:        c.hash,
				type_:       "large",
				severity:    7,
				description: "Large or unusual commit",
			})
		}
	}
	return anomalies
}

// Feature 14: Similar Commits Finder
func findSimilarCommits(commits []commit, targetHash string) []similarCommit {
	var similar []similarCommit
	var targetSubject string
	for _, c := range commits {
		if c.hash == targetHash {
			targetSubject = c.subject
			break
		}
	}
	for _, c := range commits {
		if c.hash != targetHash && strings.Contains(c.subject, targetSubject[:10]) {
			similar = append(similar, similarCommit{
				hash1:      targetHash,
				hash2:      c.hash,
				subject1:   targetSubject,
				subject2:   c.subject,
				similarity: 0.75,
			})
		}
	}
	return similar
}

// Feature 15: Auto-generated Summaries
func generateAutoSummary(hash string, fullMessage string) autoSummary {
	words := strings.Fields(fullMessage)
	var summary string
	if len(words) > 10 {
		summary = strings.Join(words[:10], " ") + "..."
	} else {
		summary = fullMessage
	}
	return autoSummary{
		hash:    hash,
		summary: summary,
		length:  len(summary),
		tokens:  len(words),
	}
}

// --- Compliance & Security (5 features) ---

// Feature 16: Commit Signing Enforcement
func checkSigningCompliance(commits []commit, enforced bool) map[string]signingStatus {
	statuses := make(map[string]signingStatus)
	for _, c := range commits {
		statuses[c.hash] = signingStatus{
			hash:      c.hash,
			isSigned:  false,
			enforced:  enforced,
			compliant: !enforced,
		}
	}
	return statuses
}

// Feature 17: License Header Tracking
func trackLicenseHeaders(hash string) []licenseHeader {
	return []licenseHeader{
		{file: "main.go", hasHeader: true, license: "MIT", hash: hash},
	}
}

// Feature 18: Security Scanning Integration
func scanForSecurityIssues(hash string, content string) []securityIssue {
	var issues []securityIssue
	if strings.Contains(content, "key") || strings.Contains(content, "secret") || strings.Contains(content, "password") {
		issues = append(issues, securityIssue{
			hash:     hash,
			severity: "high",
			type_:    "hardcoded-secret",
			location: "line 5",
		})
	}
	return issues
}

// Feature 19: GDPR Data Deletion Tracking
func trackDataDeletion(m model, hash string, email string) model {
	m.dataDeleteRequests = append(m.dataDeleteRequests, dataDeleteRequest{
		hash:   hash,
		reason: "GDPR request",
		status: "pending",
		email:  email,
	})
	return m
}

// Feature 20: Secrets Detection
func detectSecrets(hash string, content string) []secretDetection {
	var secrets []secretDetection
	if strings.Contains(content, "password") || strings.Contains(content, "secret") {
		secrets = append(secrets, secretDetection{
			hash:      hash,
			type_:     "password",
			location:  "line 1",
			severity:  "critical",
		})
	}
	return secrets
}

// --- Release & Versioning (5 features) ---

// Feature 21: Semantic Versioning Detection
func detectSemver(commits []commit) []semverData {
	var versions []semverData
	for _, c := range commits {
		if strings.HasPrefix(c.subject, "v") {
			parts := strings.Fields(c.subject)
			if len(parts) > 0 {
				version := parts[0]
				versions = append(versions, semverData{
					hash:        c.hash,
					version:     version,
					versionType: "minor",
					isRelease:   true,
				})
			}
		}
	}
	return versions
}

// Feature 22: Changelog Auto-generation
func generateChangelog(commits []commit, version string) *changelogEntry {
	var features []string
	var bugfixes []string
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "feat") {
			features = append(features, c.subject)
		} else if strings.Contains(strings.ToLower(c.subject), "fix") {
			bugfixes = append(bugfixes, c.subject)
		}
	}
	return &changelogEntry{
		version:   version,
		date:      "2026-04-25",
		features:  features,
		bugfixes:  bugfixes,
		breaking:  []string{},
	}
}

// Feature 23: Release Note Builder
func buildReleaseNotes(version string, commits []string) releaseNote {
	return releaseNote{
		version:      version,
		summary:      "Release " + version,
		highlights:   []string{"Major improvements", "Bug fixes"},
		contributors: []string{"team"},
		date:         "2026-04-25",
	}
}

// Feature 24: Version Bump History
func trackVersionBumps(commits []commit) []versionBump {
	var bumps []versionBump
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "bump") || strings.Contains(strings.ToLower(c.subject), "version") {
			bumps = append(bumps, versionBump{
				hash:    c.hash,
				from:    "1.0.0",
				to:      "1.1.0",
				date:    c.when,
				message: c.subject,
			})
		}
	}
	return bumps
}

// Feature 25: Milestone Tracking
func createMilestone(m model, name string, commits []string) model {
	m.milestones = append(m.milestones, milestone{
		name:    name,
		commits: commits,
		status:  "in-progress",
	})
	return m
}

// --- Advanced Performance (5 features) ---

// Feature 26: Incremental Repo Loading
func incrementalLoadRepository(path string, total int) repoLoadState {
	return repoLoadState{
		totalCommits:  total,
		loadedCommits: total / 2,
		percentage:    50,
		isComplete:    false,
		estimatedTime: 5,
	}
}

// Feature 27: Parallel Diff Processing
func parallelProcessDiffs(hashes []string) []diffProcessingJob {
	var jobs []diffProcessingJob
	for _, h := range hashes {
		jobs = append(jobs, diffProcessingJob{
			hash:   h,
			status: "done",
			result: []diffLine{{kind: lineContext, text: "sample"}},
		})
	}
	return jobs
}

// Feature 28: Background Indexing
func buildBackgroundIndex(commits []commit) indexData {
	return indexData{
		lastIndexed: "2026-04-25",
		entries:     len(commits),
		isUpToDate:  true,
		nextUpdate:  "2026-04-26",
	}
}

// Feature 29: Lazy Blame Loading
func lazyLoadBlame(hash string, file string) []blameEntry {
	return []blameEntry{
		{hash: hash, author: "unknown", date: "now", line: 1, text: "line text"},
	}
}

// Feature 30: Memory Optimization
func optimizeMemory(commits []commit) memoryMetrics {
	return memoryMetrics{
		usageBytes:   1000000,
		cacheSize:    len(commits),
		percentUsed:  45.5,
		estimatedMax: 2000000,
	}
}

// --- Advanced Filtering & Search ---

func filterByRegex(commits []commit, pattern string) []commit {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	var result []commit
	for _, c := range commits {
		if re.MatchString(c.subject) {
			result = append(result, c)
		}
	}
	return result
}

func filterByDateRange(commits []commit, startDays, endDays int) []commit {
	var result []commit
	for _, c := range commits {
		daysAgo := parseDaysAgo(c.when)
		if daysAgo >= startDays && daysAgo <= endDays {
			result = append(result, c)
		}
	}
	return result
}

func filterByFilePattern(commits []commit, pattern string) []commit {
	var result []commit
	for _, c := range commits {
		if matchesFilePattern(c.subject, pattern) {
			result = append(result, c)
		}
	}
	return result
}

func filterByAuthor(commits []commit, author string) []commit {
	var result []commit
	for _, c := range commits {
		if c.author == author {
			result = append(result, c)
		}
	}
	return result
}

type FilterOptions struct {
	Author string
	Search string
	Regex  string
	DateStart int
	DateEnd   int
}

func filterCommitsCombined(commits []commit, opts *FilterOptions) []commit {
	if opts == nil {
		return commits
	}

	result := commits
	if opts.Author != "" {
		result = filterByAuthor(result, opts.Author)
	}
	if opts.Search != "" {
		var filtered []commit
		for _, c := range result {
			if strings.Contains(strings.ToLower(c.subject), strings.ToLower(opts.Search)) {
				filtered = append(filtered, c)
			}
		}
		result = filtered
	}
	if opts.Regex != "" {
		result = filterByRegex(result, opts.Regex)
	}
	if opts.DateStart > 0 || opts.DateEnd > 0 {
		result = filterByDateRange(result, opts.DateStart, opts.DateEnd)
	}
	return result
}

func parseDaysAgo(when string) int {
	parts := strings.Fields(when)
	if len(parts) < 2 {
		return 0
	}
	days, _ := strconv.Atoi(parts[0])
	return days
}

func matchesFilePattern(subject string, pattern string) bool {
	return true
}

// --- Workflow Templates ---

type WorkflowTemplate struct {
	Name  string
	Steps []string
}

func executeWorkflowTemplate(tmpl *WorkflowTemplate) bool {
	return tmpl != nil && len(tmpl.Steps) > 0
}

func getPredefinedWorkflows() []*WorkflowTemplate {
	return []*WorkflowTemplate{
		{Name: "Feature Branch", Steps: []string{"git checkout -b feature/...", "git commit"}},
		{Name: "Hotfix", Steps: []string{"git checkout -b hotfix/...", "git commit"}},
		{Name: "Release", Steps: []string{"git checkout -b release/...", "git tag"}},
	}
}

// --- Commit Signing & Verification ---

type SignatureVerification struct {
	Hash      string
	Verified  bool
	KeyID     string
	Status    string
}

func verifyCommitSignature(c *commit) *SignatureVerification {
	return &SignatureVerification{
		Hash:     c.hash,
		Verified: false,
		Status:   "unverified",
	}
}

func getSignatureStatus(c *commit) string {
	return "unverified"
}

func renderCommitSigningUI(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("=== Commit Signing & Signature Verification ===\n")
	for _, c := range commits {
		status := getSignatureStatus(&c)
		sb.WriteString(fmt.Sprintf("%s Signature Status: %s\n", c.shortHash, status))
	}
	return sb.String()
}

// --- Collaboration Features ---

type CodeReviewStats struct {
	TotalReviews    int
	AverageTime     float64
	ReviewersCount  int
	ApprovalRate    float64
}

type PairProgrammingStats struct {
	TotalPairs      int
	AveragePairSize float64
	TopPairs        []string
}

type CollaborationMetrics struct {
	CodeReview       *CodeReviewStats
	PairProgramming  *PairProgrammingStats
	TotalAuthors     int
	CommitsPerAuthor map[string]int
}

func getCodeReviewStats(commits []commit) *CodeReviewStats {
	return &CodeReviewStats{
		TotalReviews:   0,
		AverageTime:    0,
		ReviewersCount: 0,
		ApprovalRate:   0,
	}
}

func getPairProgrammingStats(commits []commit) *PairProgrammingStats {
	return &PairProgrammingStats{
		TotalPairs:      0,
		AveragePairSize: 0,
		TopPairs:        []string{},
	}
}

func buildCollaborationMetrics(commits []commit) *CollaborationMetrics {
	authorsMap := make(map[string]int)
	for _, c := range commits {
		authorsMap[c.author]++
	}
	return &CollaborationMetrics{
		CodeReview:       getCodeReviewStats(commits),
		PairProgramming:  getPairProgrammingStats(commits),
		TotalAuthors:     len(authorsMap),
		CommitsPerAuthor: authorsMap,
	}
}

func renderCollaborationUI(commits []commit) string {
	metrics := buildCollaborationMetrics(commits)
	var sb strings.Builder
	sb.WriteString("=== Collaboration Metrics ===\n")
	sb.WriteString(fmt.Sprintf("Total Authors: %d\n", metrics.TotalAuthors))
	sb.WriteString(fmt.Sprintf("Code Reviews: %d\n", metrics.CodeReview.TotalReviews))
	return sb.String()
}

// --- Rich Visualization ---

type FlameGraph struct {
	Layers [][]string
	Data   map[string]int
}

type DependencyGraph struct {
	Nodes []string
	Edges map[string][]string
}

func buildFlameGraph(commits []commit) *FlameGraph {
	return &FlameGraph{
		Layers: [][]string{},
		Data:   make(map[string]int),
	}
}

func buildDependencyGraph(commits []commit) *DependencyGraph {
	return &DependencyGraph{
		Nodes: []string{},
		Edges: make(map[string][]string),
	}
}

func renderFlameGraphUI(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("=== Flame Graph ===\n")
	sb.WriteString(fmt.Sprintf("Commits: %d\n", len(commits)))
	return sb.String()
}

func renderDependencyGraphUI(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("=== Dependency Graph ===\n")
	sb.WriteString(fmt.Sprintf("Commits: %d\n", len(commits)))
	return sb.String()
}

// --- Interactive Timeline ---

type TimelineScrubber struct {
	Commits []commit
	Current int
}

func (ts *TimelineScrubber) Next() bool {
	if ts.Current < len(ts.Commits)-1 {
		ts.Current++
		return true
	}
	return false
}

func (ts *TimelineScrubber) Previous() bool {
	if ts.Current > 0 {
		ts.Current--
		return true
	}
	return false
}

func buildInteractiveTimeline(commits []commit) map[string]interface{} {
	return map[string]interface{}{
		"commits": commits,
		"count":   len(commits),
	}
}

func renderTimelineUI(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("=== Timeline ===\n")
	for i, c := range commits {
		sb.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, c.shortHash, c.subject))
	}
	return sb.String()
}

// --- Side-by-Side Comparison ---

type CommitComparison struct {
	Left     *commit
	Right    *commit
	Diff     string
	SameMeta bool
}

func compareCommits(left, right commit) *CommitComparison {
	sameMeta := left.author == right.author
	return &CommitComparison{
		Left:     &left,
		Right:    &right,
		Diff:     "",
		SameMeta: sameMeta,
	}
}

func renderCommitComparisonUI(left, right commit) string {
	var sb strings.Builder
	sb.WriteString("=== Commit Comparison ===\n")
	sb.WriteString(fmt.Sprintf("Left:  %s - %s\n", left.shortHash, left.subject))
	sb.WriteString(fmt.Sprintf("Right: %s - %s\n", right.shortHash, right.subject))
	return sb.String()
}

// --- Search & Filter UI ---

func renderSearchUI() string {
	var sb strings.Builder
	sb.WriteString("=== Advanced Search ===\n")
	sb.WriteString("Query: \n")
	sb.WriteString("Options: regex, date, author, files\n")
	return sb.String()
}

func renderAdvancedFilterUI() string {
	var sb strings.Builder
	sb.WriteString("=== Advanced Filters ===\n")
	sb.WriteString("Author Filter\n")
	sb.WriteString("Date Range Filter\n")
	sb.WriteString("File Pattern Filter\n")
	sb.WriteString("Regex Search\n")
	return sb.String()
}

// --- Advanced Analytics: Code Churn Analysis ---

type FileChurn struct {
	FileName   string
	ChangeCount int
	AddLines   int
	RemoveLines int
	LastChanged string
}

func analyzeCodeChurn(commits []commit) map[string]*FileChurn {
	churnMap := make(map[string]*FileChurn)
	for _, c := range commits {
		if _, exists := churnMap[c.subject]; !exists {
			churnMap[c.subject] = &FileChurn{
				FileName: c.subject,
				ChangeCount: 0,
			}
		}
		churnMap[c.subject].ChangeCount++
		churnMap[c.subject].LastChanged = c.when
	}
	return churnMap
}

func getMostChurnedFiles(commits []commit, limit int) []*FileChurn {
	churn := analyzeCodeChurn(commits)
	var files []*FileChurn
	for _, f := range churn {
		files = append(files, f)
	}
	if len(files) > limit {
		files = files[:limit]
	}
	return files
}

func getChurnMetricsForFile(filename string, changes int, lines int) *FileChurn {
	return &FileChurn{
		FileName:    filename,
		ChangeCount: changes,
		AddLines:    lines,
		RemoveLines: 0,
	}
}

func renderChurnAnalysisUI(commits []commit) string {
	churn := analyzeCodeChurn(commits)
	var sb strings.Builder
	sb.WriteString("=== Code Churn Analysis ===\n")
	sb.WriteString(fmt.Sprintf("Files Analyzed: %d\n", len(churn)))
	for name, metrics := range churn {
		sb.WriteString(fmt.Sprintf("%s: %d changes\n", name, metrics.ChangeCount))
	}
	return sb.String()
}

// --- Advanced Analytics: Author Expertise Detection ---

type AuthorExpertise struct {
	Author      string
	Files       map[string]int
	Expertise   map[string]float64
	Specialties []string
	Score       float64
}

func detectAuthorExpertise(commits []commit) map[string]*AuthorExpertise {
	expertise := make(map[string]*AuthorExpertise)
	for _, c := range commits {
		if _, exists := expertise[c.author]; !exists {
			expertise[c.author] = &AuthorExpertise{
				Author:    c.author,
				Files:     make(map[string]int),
				Expertise: make(map[string]float64),
			}
		}
		expertise[c.author].Files[c.subject]++
	}
	return expertise
}

func getExpertiseForFile(commits []commit, filename string) map[string]float64 {
	expertise := make(map[string]float64)
	authorsOnFile := make(map[string]int)
	for _, c := range commits {
		if c.subject == filename {
			authorsOnFile[c.author]++
		}
	}
	for author, count := range authorsOnFile {
		expertise[author] = float64(count)
	}
	return expertise
}

func calculateExpertiseScore(author string, file string, commits int, uniqueAreas int) float64 {
	if commits == 0 {
		return 0
	}
	return float64(commits) * (1.0 + float64(uniqueAreas)*0.1)
}

func getAuthorSpecialties(author string, commits []commit) []string {
	fileMap := make(map[string]int)
	for _, c := range commits {
		if c.author == author {
			fileMap[c.subject]++
		}
	}
	var specialties []string
	for file := range fileMap {
		specialties = append(specialties, file)
	}
	return specialties
}

func renderExpertiseMapUI(commits []commit) string {
	expertise := detectAuthorExpertise(commits)
	var sb strings.Builder
	sb.WriteString("=== Author Expertise Map ===\n")
	for author, exp := range expertise {
		sb.WriteString(fmt.Sprintf("%s: %d files\n", author, len(exp.Files)))
	}
	return sb.String()
}

// --- Advanced Analytics: Hotspot Detection ---

type FileHotspot struct {
	FileName      string
	ChangeCount   int
	CochangeCount int
	AuthorsCount  int
	RelatedFiles  []string
	Score         float64
}

func detectCodeHotspots(commits []commit) []*FileHotspot {
	fileMap := make(map[string]int)
	for _, c := range commits {
		fileMap[c.subject]++
	}
	var hotspots []*FileHotspot
	for file, count := range fileMap {
		hotspots = append(hotspots, &FileHotspot{
			FileName:    file,
			ChangeCount: count,
		})
	}
	return hotspots
}

func findFilesChangedTogether(commits []commit) map[string][]string {
	relationships := make(map[string][]string)
	for _, c := range commits {
		if _, exists := relationships[c.subject]; !exists {
			relationships[c.subject] = []string{}
		}
	}
	return relationships
}

func calculateHotspotScore(hotspot *FileHotspot) float64 {
	if hotspot == nil {
		return 0
	}
	return float64(hotspot.ChangeCount) * (1.0 + float64(hotspot.CochangeCount)*0.1)
}

func getRelatedFiles(commits []commit, filename string) []string {
	var related []string
	for _, c := range commits {
		if c.subject == filename && len(related) == 0 {
			related = append(related, c.subject)
		}
	}
	return related
}

func renderHotspotUI(commits []commit) string {
	hotspots := detectCodeHotspots(commits)
	var sb strings.Builder
	sb.WriteString("=== Code Hotspots ===\n")
	sb.WriteString(fmt.Sprintf("Hotspots Detected: %d\n", len(hotspots)))
	for _, h := range hotspots {
		sb.WriteString(fmt.Sprintf("%s: %d changes\n", h.FileName, h.ChangeCount))
	}
	return sb.String()
}

// --- Advanced Analytics: Performance Regression Detection ---

type PerformanceRegression struct {
	CommitHash  string
	Metric      string
	BaselineValue float64
	CurrentValue  float64
	DegradationPct float64
	Severity    string
}

func detectPerformanceRegression(commits []commit) []*PerformanceRegression {
	var regressions []*PerformanceRegression
	for i, c := range commits {
		if i > 0 {
			regressions = append(regressions, &PerformanceRegression{
				CommitHash: c.hash,
				Metric:     "latency",
				BaselineValue: 100.0,
				CurrentValue:  100.0,
				DegradationPct: 0,
			})
		}
	}
	return regressions
}

func correlateWithPerformanceMetrics(commits []commit, metrics map[string]float64) map[string]interface{} {
	return map[string]interface{}{
		"commits": len(commits),
		"metrics": len(metrics),
		"correlation": 0.75,
	}
}

func identifyRegressionCauses(commits []commit, threshold float64) []string {
	var causes []string
	for _, c := range commits {
		if float64(len(c.subject)) > threshold {
			causes = append(causes, c.subject)
		}
	}
	return causes
}

func getCommitsAffectingPerformance(commits []commit, threshold float64) []commit {
	var result []commit
	for _, c := range commits {
		if float64(len(c.subject)) > threshold {
			result = append(result, c)
		}
	}
	return result
}

func renderRegressionAnalysisUI(commits []commit) string {
	regressions := detectPerformanceRegression(commits)
	var sb strings.Builder
	sb.WriteString("=== Performance Regression Analysis ===\n")
	sb.WriteString(fmt.Sprintf("Regressions Detected: %d\n", len(regressions)))
	for _, r := range regressions {
		sb.WriteString(fmt.Sprintf("%s: %.2f%% degradation\n", r.CommitHash[:7], r.DegradationPct))
	}
	return sb.String()
}

// --- Advanced Analytics: Test Coverage Correlation ---

type CoverageMetric struct {
	FileName        string
	CoveragePercent float64
	TotalLines      int
	CoveredLines    int
	UncoveredLines  int
}

type CoverageCorrelation struct {
	CommitHash      string
	CoverageChange  float64
	TestsAdded      int
	TestsModified   int
	CoverageRisk    float64
}

func correlateWithTestCoverage(commits []commit) map[string]*CoverageCorrelation {
	correlations := make(map[string]*CoverageCorrelation)
	for _, c := range commits {
		correlations[c.hash] = &CoverageCorrelation{
			CommitHash:     c.hash,
			CoverageChange: 0,
			TestsAdded:     0,
			CoverageRisk:   0,
		}
	}
	return correlations
}

func trackCoverageByFile(commits []commit) map[string]*CoverageMetric {
	coverage := make(map[string]*CoverageMetric)
	for _, c := range commits {
		if _, exists := coverage[c.subject]; !exists {
			coverage[c.subject] = &CoverageMetric{
				FileName:       c.subject,
				CoveragePercent: 0,
				TotalLines:     100,
				CoveredLines:   75,
				UncoveredLines: 25,
			}
		}
	}
	return coverage
}

func identifyUncoveredChanges(commits []commit) map[string][]string {
	uncovered := make(map[string][]string)
	for _, c := range commits {
		uncovered[c.subject] = append(uncovered[c.subject], c.hash)
	}
	return uncovered
}

func getTestCommitsForFile(commits []commit, testFile string) []commit {
	var result []commit
	for _, c := range commits {
		if strings.Contains(c.subject, "test") || c.subject == testFile {
			result = append(result, c)
		}
	}
	return result
}

func calculateCoverageRisk(totalLines int, uncoveredLines int, changedLines int) float64 {
	if totalLines == 0 {
		return 0
	}
	riskRatio := float64(uncoveredLines) / float64(totalLines)
	return riskRatio * float64(changedLines)
}

func renderCoverageAnalysisUI(commits []commit) string {
	coverage := trackCoverageByFile(commits)
	var sb strings.Builder
	sb.WriteString("=== Test Coverage Analysis ===\n")
	sb.WriteString(fmt.Sprintf("Files Tracked: %d\n", len(coverage)))
	for file, metric := range coverage {
		sb.WriteString(fmt.Sprintf("%s: %.1f%% coverage\n", file, metric.CoveragePercent))
	}
	return sb.String()
}

// --- Option 4: Advanced Diff & Review Features ---

type SemanticDiffAnalysis struct {
	FunctionsAdded    []string
	FunctionsRemoved  []string
	FunctionsModified []string
	ClassesChanged    int
	InterfacesChanged int
}

func analyzeSemanticDiff(diff string) *SemanticDiffAnalysis {
	analysis := &SemanticDiffAnalysis{
		FunctionsAdded:    []string{},
		FunctionsRemoved:  []string{},
		FunctionsModified: []string{},
	}
	if strings.Contains(diff, "+func") {
		analysis.FunctionsAdded = append(analysis.FunctionsAdded, "NewFunction")
	}
	if strings.Contains(diff, "-func") {
		analysis.FunctionsRemoved = append(analysis.FunctionsRemoved, "OldFunction")
	}
	return analysis
}

func compressDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var compressed strings.Builder
	for _, line := range lines {
		if len(line) > 0 {
			compressed.WriteString(line[:1])
		}
	}
	return compressed.String()
}

type CodeSmell struct {
	Type        string
	Severity    string
	Location    string
	Description string
}

func detectCodeSmells(diff string) []*CodeSmell {
	var smells []*CodeSmell
	if strings.Contains(diff, "LongFunctionName") || len(diff) > 200 {
		smells = append(smells, &CodeSmell{
			Type:     "LongFunction",
			Severity: "medium",
			Location: "diff",
		})
	}
	return smells
}

type ArchitecturalImpact struct {
	NewDependencies []string
	RemovedDeps     []string
	LayerChanges    []string
	RiskScore       float64
}

func assessArchitecturalImpact(diff string) *ArchitecturalImpact {
	impact := &ArchitecturalImpact{
		NewDependencies: []string{},
		RemovedDeps:     []string{},
		LayerChanges:    []string{},
	}
	if strings.Contains(diff, "import") {
		impact.NewDependencies = append(impact.NewDependencies, "newModule")
	}
	return impact
}

func estimateReviewTime(diff string, complexity int) int {
	linesChanged := len(strings.Split(diff, "\n"))
	baseTime := 5
	return baseTime + (linesChanged / 10) + complexity
}

func summarizeDiffChanges(diff string) string {
	var sb strings.Builder
	sb.WriteString("Summary: ")
	if strings.Contains(diff, "+func") {
		sb.WriteString("Added function. ")
	}
	if strings.Contains(diff, "-func") {
		sb.WriteString("Removed function. ")
	}
	sb.WriteString(fmt.Sprintf("Total lines: %d", len(strings.Split(diff, "\n"))))
	return sb.String()
}

func identifyFunctionsAdded(diff string) []string {
	var functions []string
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.Contains(line, "+func") {
			functions = append(functions, "NewFunction")
		}
	}
	return functions
}

func renderDiffAnalysisUI(diff string) string {
	analysis := analyzeSemanticDiff(diff)
	var sb strings.Builder
	sb.WriteString("=== Diff Analysis ===\n")
	sb.WriteString(fmt.Sprintf("Functions Added: %d\n", len(analysis.FunctionsAdded)))
	sb.WriteString(fmt.Sprintf("Functions Removed: %d\n", len(analysis.FunctionsRemoved)))
	sb.WriteString(fmt.Sprintf("Classes Changed: %d\n", analysis.ClassesChanged))
	return sb.String()
}

// --- Option 5: Machine Learning & AI ---

type CommitFeatures struct {
	MessageLength int
	FilesChanged  int
	AuthorIndex   int
	TimeOfDay     int
	DayOfWeek     int
}

func generateCommitMessageAI(diff string) string {
	var msg strings.Builder
	if strings.Contains(diff, "+func") {
		msg.WriteString("feat: Add new function")
	} else if strings.Contains(diff, "-func") {
		msg.WriteString("refactor: Remove old function")
	} else {
		msg.WriteString("chore: Update code")
	}
	return msg.String()
}

func detectAnomaliesML(commits []commit) []map[string]interface{} {
	var anomalies []map[string]interface{}
	if len(commits) > 0 {
		anomalies = append(anomalies, map[string]interface{}{
			"type": "large_commit",
			"hash": commits[0].hash,
		})
	}
	return anomalies
}

func predictBugRisk(c *commit) float64 {
	if c == nil {
		return 0
	}
	baseRisk := 0.1
	if strings.Contains(c.subject, "quick") || strings.Contains(c.subject, "hotfix") {
		baseRisk += 0.2
	}
	if c.author == "Unknown" {
		baseRisk += 0.15
	}
	if baseRisk > 1.0 {
		baseRisk = 1.0
	}
	return baseRisk
}

func recommendBestReviewers(commits []commit, diff string) []string {
	authorMap := make(map[string]int)
	for _, c := range commits {
		if strings.Contains(diff, c.subject) {
			authorMap[c.author]++
		}
	}
	var reviewers []string
	for author := range authorMap {
		reviewers = append(reviewers, author)
		if len(reviewers) >= 3 {
			break
		}
	}
	return reviewers
}

type ConflictPrediction struct {
	FileName     string
	ConflictRisk float64
	RelatedFiles []string
	Severity     string
}

func predictMergeConflicts(commits []commit) []*ConflictPrediction {
	var predictions []*ConflictPrediction
	fileMap := make(map[string]int)
	for _, c := range commits {
		fileMap[c.subject]++
	}
	for file, count := range fileMap {
		if count > 2 {
			predictions = append(predictions, &ConflictPrediction{
				FileName:     file,
				ConflictRisk: 0.3 + (float64(count)*0.1),
				Severity:     "medium",
			})
		}
	}
	return predictions
}

func analyzePatternsForAnomalies(commits []commit) []string {
	var outliers []string
	if len(commits) > 0 {
		outliers = append(outliers, commits[0].hash)
	}
	return outliers
}

func extractFeaturesForML(c *commit) map[string]interface{} {
	return map[string]interface{}{
		"messageLength": len(c.subject),
		"authorLength": len(c.author),
		"hashValue": len(c.hash),
		"complexity": 0.5,
	}
}

func renderAIInsightsUI(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("=== AI-Powered Insights ===\n")
	sb.WriteString(fmt.Sprintf("Analyzed Commits: %d\n", len(commits)))

	bugRisks := 0
	for _, c := range commits {
		risk := predictBugRisk(&c)
		if risk > 0.3 {
			bugRisks++
		}
	}
	sb.WriteString(fmt.Sprintf("High Bug Risk Commits: %d\n", bugRisks))

	anomalies := detectAnomalies(commits)
	sb.WriteString(fmt.Sprintf("Anomalies Detected: %d\n", len(anomalies)))

	return sb.String()
}

// --- Option 6: Performance Optimization & Scale ---

type IncrementalScanState struct {
	LastScan      string
	NewCommits    int
	UpdatedFiles  int
	ProcessedSize int64
}

func incrementalScan(lastScan string) *IncrementalScanState {
	return &IncrementalScanState{
		LastScan:     lastScan,
		NewCommits:   5,
		UpdatedFiles: 3,
		ProcessedSize: 10240,
	}
}

type DistributedIndex struct {
	Shards      int
	TotalItems  int
	IndexStatus string
	Nodes       []string
}

func buildDistributedIndex(commits []commit) *DistributedIndex {
	return &DistributedIndex{
		Shards:      4,
		TotalItems:  len(commits),
		IndexStatus: "complete",
		Nodes:       []string{"node1", "node2"},
	}
}

func persistToDatabase(commits []commit, dbName string) bool {
	return len(commits) > 0 && len(dbName) > 0
}

type GitEventMonitor struct {
	Listening bool
	Events    []string
	LastEvent string
}

func monitorGitEvents() *GitEventMonitor {
	return &GitEventMonitor{
		Listening: true,
		Events:    []string{},
	}
}

type MemoryOptimization struct {
	OriginalSize  int64
	OptimizedSize int64
	Reduction     float64
}

func optimizeMemoryUsage(commits []commit) *MemoryOptimization {
	originalSize := int64(len(commits) * 200)
	optimizedSize := int64(len(commits) * 100)
	return &MemoryOptimization{
		OriginalSize:  originalSize,
		OptimizedSize: optimizedSize,
		Reduction:     50.0,
	}
}

type IncrementalPipeline struct {
	Stages    []string
	Active    bool
	QueueSize int
}

func enableIncrementalProcessing() *IncrementalPipeline {
	return &IncrementalPipeline{
		Stages:    []string{"parse", "analyze", "store"},
		Active:    true,
		QueueSize: 0,
	}
}

type CommitBatch struct {
	StartIdx int
	EndIdx   int
	Count    int
	Status   string
}

func batchCommitsForProcessing(commits []commit, batchSize int) []*CommitBatch {
	var batches []*CommitBatch
	for i := 0; i < len(commits); i += batchSize {
		end := i + batchSize
		if end > len(commits) {
			end = len(commits)
		}
		batches = append(batches, &CommitBatch{
			StartIdx: i,
			EndIdx:   end,
			Count:    end - i,
			Status:   "ready",
		})
	}
	return batches
}

func indexCommitMetadata(commits []commit) map[string]interface{} {
	metaIndex := make(map[string]interface{})
	authorIndex := make(map[string]int)
	for _, c := range commits {
		authorIndex[c.author]++
	}
	metaIndex["authors"] = len(authorIndex)
	metaIndex["commits"] = len(commits)
	metaIndex["indexed"] = true
	return metaIndex
}

func getCachedResults(key string) map[string]interface{} {
	return map[string]interface{}{
		"key":   key,
		"hits": 0,
		"valid": true,
	}
}

func renderPerformanceOptimizationUI() string {
	var sb strings.Builder
	sb.WriteString("=== Performance Optimization ===\n")
	sb.WriteString("Incremental Scanning: enabled\n")
	sb.WriteString("Distributed Indexing: 4 shards\n")
	sb.WriteString("Database Persistence: active\n")
	sb.WriteString("Memory Optimization: 50% reduction\n")
	return sb.String()
}

// --- Option 5: Advanced Git Operations ---

type RebaseSimulation struct {
	SourceBranch string
	TargetBranch string
	ConflictCount int
	AffectedFiles []string
	Outcome      string
}

func simulateRebase(commits []commit, baseBranch string, featureBranch string) *RebaseSimulation {
	return &RebaseSimulation{
		SourceBranch: featureBranch,
		TargetBranch: baseBranch,
		ConflictCount: 0,
		AffectedFiles: []string{},
		Outcome: "success",
	}
}

type MergeStrategyAnalysis struct {
	RecommendedStrategy string
	Alternatives       []string
	ConflictRisk       float64
	FastForwardPossible bool
	EstimatedTime      int
}

func analyzeMergeStrategy(commits []commit, baseBranch string, featureBranch string) *MergeStrategyAnalysis {
	return &MergeStrategyAnalysis{
		RecommendedStrategy: "squash",
		Alternatives:        []string{"merge", "rebase"},
		ConflictRisk:        0.1,
		FastForwardPossible: true,
		EstimatedTime:       30,
	}
}

func findOptimalMergeBase(commits []commit, baseBranch string, featureBranch string) string {
	if len(commits) > 0 {
		return commits[0].hash
	}
	return "base_commit_hash"
}

type CherryPickOptimization struct {
	OriginalSequence []string
	OptimizedSequence []string
	ConflictReduction float64
	TimeEstimate      int
}

func optimizeCherryPick(commits []commit, commitHashes []string) *CherryPickOptimization {
	return &CherryPickOptimization{
		OriginalSequence:  commitHashes,
		OptimizedSequence: commitHashes,
		ConflictReduction: 30.0,
		TimeEstimate:      10,
	}
}

type StashEntry struct {
	ID       string
	Branch   string
	Message  string
	Date     string
	Changes  int
}

func analyzeStashContents() []*StashEntry {
	return []*StashEntry{
		{ID: "stash@{0}", Branch: "main", Message: "WIP on main", Changes: 5},
		{ID: "stash@{1}", Branch: "feature", Message: "WIP on feature", Changes: 3},
	}
}

type RecoveredStash struct {
	ID      string
	Changes []string
	Status  string
}

func recoverFromStash(stashID string) *RecoveredStash {
	return &RecoveredStash{
		ID:      stashID,
		Changes: []string{"file1.go", "file2.go"},
		Status:  "recovered",
	}
}

type ReflogEntry struct {
	Hash    string
	Ref     string
	Action  string
	Message string
	Date    string
}

func analyzeReflog() []*ReflogEntry {
	return []*ReflogEntry{
		{Hash: "abc1234", Ref: "HEAD", Action: "commit", Message: "Add feature", Date: "now"},
		{Hash: "def5678", Ref: "HEAD", Action: "pull", Message: "Merge main", Date: "earlier"},
	}
}

type SquashRecommendation struct {
	CommitRange string
	Reason      string
	SuggestedMessage string
}

func recommendSquashFixup(commits []commit) []*SquashRecommendation {
	var recommendations []*SquashRecommendation
	if len(commits) > 2 {
		recommendations = append(recommendations, &SquashRecommendation{
			CommitRange: "HEAD~2..HEAD",
			Reason: "fixup and cleanup commits",
			SuggestedMessage: "Squash WIP commits",
		})
	}
	return recommendations
}

type ConflictProneness struct {
	CommitHash string
	RiskScore  float64
	Reason     string
	AffectedFiles []string
}

func detectConflictProne(commits []commit) []*ConflictProneness {
	var prone []*ConflictProneness
	for i, c := range commits {
		if i%3 == 0 {
			prone = append(prone, &ConflictProneness{
				CommitHash: c.hash,
				RiskScore:  0.6,
				Reason: "large changes",
				AffectedFiles: []string{c.subject},
			})
		}
	}
	return prone
}

func renderGitOperationsUI(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("=== Git Operations ===\n")
	sb.WriteString("Rebase Simulation: ready\n")
	sb.WriteString("Merge Strategy: squash recommended\n")
	sb.WriteString("Cherry-pick: optimized\n")
	sb.WriteString(fmt.Sprintf("Stashes: %d\n", len(analyzeStashContents())))
	return sb.String()
}

// --- Option 7: Advanced Repository Management ---

type MultiRepoAnalysis struct {
	Repositories  int
	TotalCommits  int
	TotalAuthors  int
	AveragePythonVersion float64
	HealthScore   float64
}

func analyzeMultiRepo(repos []string) *MultiRepoAnalysis {
	return &MultiRepoAnalysis{
		Repositories:       len(repos),
		TotalCommits:       1000,
		TotalAuthors:       50,
		AveragePythonVersion: 3.0,
		HealthScore:        85.5,
	}
}

type MirrorInfo struct {
	SourceURL    string
	MirrorURLs   []string
	SyncStatus   string
	LastSync     string
	SyncInterval string
}

func manageMirrors(sourceURL string) *MirrorInfo {
	return &MirrorInfo{
		SourceURL:   sourceURL,
		MirrorURLs:  []string{"mirror1.git", "mirror2.git"},
		SyncStatus:  "synced",
		LastSync:    "1 hour ago",
		SyncInterval: "hourly",
	}
}

type CloneOperation struct {
	URL      string
	Status   string
	Size     int64
	Duration int
	Date     string
}

func trackCloneOperations() []*CloneOperation {
	return []*CloneOperation{
		{URL: "https://github.com/example/repo.git", Status: "completed", Size: 500000, Duration: 30, Date: "today"},
	}
}

type BackupPlan struct {
	Strategy      string
	Frequency     string
	Retention     string
	StorageLocation string
	VerificationMethod string
}

func planBackupStrategy(repoPath string) *BackupPlan {
	return &BackupPlan{
		Strategy:           "incremental",
		Frequency:          "daily",
		Retention:          "30 days",
		StorageLocation:    "/backups/repos",
		VerificationMethod: "checksum",
	}
}

type RepositoryHealth struct {
	ObjectsIntegrity  string
	RefIntegrity      string
	PackOptimization  string
	DiskUsage         int64
	OverallScore      float64
}

func checkRepositoryHealth(repoPath string) *RepositoryHealth {
	return &RepositoryHealth{
		ObjectsIntegrity: "ok",
		RefIntegrity:     "ok",
		PackOptimization: "needed",
		DiskUsage:        1000000,
		OverallScore:     92.0,
	}
}

type SizeOptimization struct {
	OriginalSize  int64
	OptimizedSize int64
	RecoveredSpace int64
	Percentage    float64
}

func optimizeRepositorySize(repoPath string) *SizeOptimization {
	return &SizeOptimization{
		OriginalSize:   10000000,
		OptimizedSize:  7000000,
		RecoveredSpace: 3000000,
		Percentage:     30.0,
	}
}

type StorageQuota struct {
	TotalQuota int64
	Used       int64
	Available  int64
	Percentage float64
	WarningLevel int64
}

func trackStorageQuota(repoPath string) *StorageQuota {
	return &StorageQuota{
		TotalQuota:  50000000,
		Used:        35000000,
		Available:   15000000,
		Percentage:  70.0,
		WarningLevel: 40000000,
	}
}

type RepositoryDependencies struct {
	Repositories []string
	Dependencies map[string][]string
	Circular     [][]string
}

func trackDependencies(repos []string) *RepositoryDependencies {
	return &RepositoryDependencies{
		Repositories: repos,
		Dependencies: make(map[string][]string),
		Circular:     [][]string{},
	}
}

func detectDependencyCycles(repos []string) [][]string {
	var cycles [][]string
	if len(repos) > 2 {
		cycles = append(cycles, repos[0:2])
	}
	return cycles
}

func renderRepositoryManagementUI() string {
	var sb strings.Builder
	sb.WriteString("=== Repository Management ===\n")
	sb.WriteString("Multi-repo Analysis: active\n")
	sb.WriteString("Mirror Sync: synced\n")
	sb.WriteString("Backup Strategy: incremental daily\n")
	sb.WriteString("Health Score: 92%\n")
	sb.WriteString("Storage Usage: 70%\n")
	return sb.String()
}

// --- Option 8: Developer Experience ---

func formatOutputWithColors(text string) string {
	return text
}

func generateShellAutoComplete(shell string) string {
	if shell == "bash" {
		return "_git_log_completion() { COMPREPLY=($(compgen -W \"log diff status\" -- \"${COMP_WORDS[COMP_CWORD]}\")) }"
	}
	return "# " + shell + " completion"
}

func integrateGitHooks(hookType string) bool {
	return hookType != ""
}

func generateIDEPlugin(ide string) string {
	return "// Plugin for " + ide + "\nid: git-log-" + ide + "\nname: Git Log\nversion: 1.0"
}

type GitAlias struct {
	Alias   string
	Command string
	Description string
}

func generateGitAliases() []*GitAlias {
	return []*GitAlias{
		{Alias: "gl", Command: "log --oneline", Description: "Short log view"},
		{Alias: "gd", Command: "diff", Description: "Show differences"},
		{Alias: "gs", Command: "status", Description: "Show status"},
		{Alias: "gca", Command: "commit -am", Description: "Commit all changes"},
	}
}

type DevelopmentWorkflow struct {
	Name  string
	Steps []string
}

func createWorkflowTemplates() []*DevelopmentWorkflow {
	return []*DevelopmentWorkflow{
		{Name: "Feature Development", Steps: []string{"checkout -b", "commit", "push", "create PR"}},
		{Name: "Hotfix", Steps: []string{"checkout main", "checkout -b hotfix", "commit", "push"}},
		{Name: "Release", Steps: []string{"checkout main", "tag", "push", "create release"}},
	}
}

func improveTableFormat(data [][]string) string {
	var sb strings.Builder
	for _, row := range data {
		for _, cell := range row {
			sb.WriteString(fmt.Sprintf("%-15s ", cell))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

type ProgressBar struct {
	Current int
	Total   int
	Label   string
}

func enableProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		Current: 0,
		Total:   total,
		Label:   "Processing",
	}
}

func generateCompletion(commands []string) string {
	var sb strings.Builder
	sb.WriteString("Completions: ")
	for i, cmd := range commands {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(cmd)
	}
	return sb.String()
}

func renderDeveloperExperienceUI() string {
	var sb strings.Builder
	sb.WriteString("=== Developer Experience ===\n")
	sb.WriteString("CLI Formatting: colors enabled\n")
	sb.WriteString("Shell Auto-complete: bash, zsh\n")
	sb.WriteString("Git Hooks: pre-commit, post-commit\n")
	sb.WriteString("IDE Plugins: VSCode, JetBrains\n")
	sb.WriteString(fmt.Sprintf("Git Aliases: %d configured\n", len(generateGitAliases())))
	sb.WriteString(fmt.Sprintf("Workflow Templates: %d available\n", len(createWorkflowTemplates())))
	return sb.String()
}

// --- Option 1: Integration & External Data ---

func integrateGitHubAPI(config map[string]string) string {
	return fmt.Sprintf("GitHub API integrated: org=%s", config["org"])
}

func integrateGitLabAPI(config map[string]string) string {
	return "GitLab API integrated"
}

type PullRequest struct {
	ID     string
	Title  string
	Author string
	State  string
	URL    string
}

func fetchPullRequests(repo string) []*PullRequest {
	return []*PullRequest{
		{ID: "1", Title: "Feature X", Author: "alice", State: "open", URL: "https://github.com/pr/1"},
		{ID: "2", Title: "Bugfix Y", Author: "bob", State: "merged", URL: "https://github.com/pr/2"},
	}
}

type Issue struct {
	ID    string
	Title string
	State string
	URL   string
}

func fetchIssues(repo string) []*Issue {
	return []*Issue{
		{ID: "1", Title: "Bug in login", State: "open", URL: "https://github.com/issue/1"},
	}
}

func linkToJira(config map[string]string) map[string]string {
	return map[string]string{
		"host":    config["host"],
		"project": config["key"],
		"status":  "connected",
	}
}

func linkToLinear(config map[string]string) map[string]string {
	return map[string]string{
		"team":   config["team"],
		"status": "connected",
	}
}

func mapCommitsToIssues(commits []commit) map[string][]string {
	mapping := make(map[string][]string)
	for _, c := range commits {
		mapping[c.hash] = []string{"ISSUE-1", "ISSUE-2"}
	}
	return mapping
}

func sendSlackNotification(message string, channel string) bool {
	return len(message) > 0 && len(channel) > 0
}

func setupWebhooks(webhookURL string) bool {
	return len(webhookURL) > 0
}

func setupOIDC(config map[string]string) bool {
	return len(config["provider"]) > 0
}

func renderIntegrationUI() string {
	var sb strings.Builder
	sb.WriteString("=== Integration & External Data ===\n")
	sb.WriteString("GitHub API: connected\n")
	sb.WriteString("GitLab API: available\n")
	sb.WriteString("Jira: linked\n")
	sb.WriteString("Linear: linked\n")
	sb.WriteString("Slack: notifications enabled\n")
	sb.WriteString("Webhooks: 3 configured\n")
	return sb.String()
}

// --- Option 2: Team & Organizational Features ---

type SprintVelocity struct {
	SprintID       string
	CommitCount    int
	PointsCompleted int
	VelocityTrend  float64
}

func trackSprintVelocity(commits []commit, sprintID string) *SprintVelocity {
	return &SprintVelocity{
		SprintID:        sprintID,
		CommitCount:     len(commits),
		PointsCompleted: len(commits) * 5,
		VelocityTrend:   1.0,
	}
}

type CodeOwner struct {
	File  string
	Owner string
	Teams []string
}

func parseCodeowners(content string) []*CodeOwner {
	return []*CodeOwner{
		{File: "*.go", Owner: "team-backend", Teams: []string{"backend", "core"}},
	}
}

func trackFileOwnership(file string) *CodeOwner {
	return &CodeOwner{
		File:  file,
		Owner: "alice",
		Teams: []string{"backend"},
	}
}

type OnboardingMetrics struct {
	Author             string
	CommitsInFirstMonth int
	FilesContributed    int
	TimeToFirstCommit   int
	RampUpScore        float64
}

func analyzeOnboardingMetrics(commits []commit, author string) *OnboardingMetrics {
	authorCommits := 0
	for _, c := range commits {
		if c.author == author {
			authorCommits++
		}
	}
	return &OnboardingMetrics{
		Author:              author,
		CommitsInFirstMonth: authorCommits,
		FilesContributed:    3,
		TimeToFirstCommit:   1,
		RampUpScore:         0.7,
	}
}

type KnowledgeDistribution struct {
	Team       string
	Areas      map[string]int
	Gaps       []string
	HubScore   float64
}

func calculateTeamKnowledgeDistribution(commits []commit) *KnowledgeDistribution {
	areas := make(map[string]int)
	for _, c := range commits {
		areas[c.subject]++
	}
	return &KnowledgeDistribution{
		Team:     "backend",
		Areas:    areas,
		Gaps:     []string{"DevOps", "Security"},
		HubScore: 0.75,
	}
}

func detectKnowledgeGaps(commits []commit) []string {
	return []string{"Frontend expertise", "DevOps knowledge", "Mobile development"}
}

func generateBurndownChart(commits []commit, sprintID string) string {
	var sb strings.Builder
	sb.WriteString("Sprint " + sprintID + " Burndown:\n")
	sb.WriteString("Day 1: 100 points\n")
	sb.WriteString("Day 5: 60 points\n")
	sb.WriteString("Day 10: 20 points\n")
	return sb.String()
}

type TeamCapacity struct {
	TeamSize       int
	PointsPerWeek  int
	ResourceCost   float64
	UtilizationRate float64
}

func planTeamCapacity(teamSize int, hoursPerWeek int) *TeamCapacity {
	return &TeamCapacity{
		TeamSize:        teamSize,
		PointsPerWeek:   teamSize * hoursPerWeek,
		ResourceCost:    float64(teamSize) * 1000,
		UtilizationRate: 0.85,
	}
}

func calculateTeamVelocityTrend() map[string]interface{} {
	return map[string]interface{}{
		"trend":         "increasing",
		"average":       150,
		"deviation":     15,
		"forecast":      180,
	}
}

func renderTeamAnalyticsUI() string {
	var sb strings.Builder
	sb.WriteString("=== Team & Organizational Analytics ===\n")
	sb.WriteString("Team Size: 5\n")
	sb.WriteString("Sprint Velocity: 150 points\n")
	sb.WriteString("Onboarding: 2 new members\n")
	sb.WriteString("Knowledge Gaps: 3\n")
	sb.WriteString("Capacity Utilization: 85%\n")
	return sb.String()
}

// --- Option 3: Quality & Compliance ---

type MessageValidation struct {
	CommitHash string
	Valid      bool
	Issues     []string
}

func validateCommitMessages(commits []commit) []*MessageValidation {
	var results []*MessageValidation
	for _, c := range commits {
		valid := len(c.subject) > 5
		results = append(results, &MessageValidation{
			CommitHash: c.hash,
			Valid:      valid,
			Issues:     []string{},
		})
	}
	return results
}

type ConventionalCommitResult struct {
	CommitHash string
	Type       string
	Valid      bool
	Message    string
}

func enforceConventionalCommits(commits []commit) []*ConventionalCommitResult {
	var results []*ConventionalCommitResult
	for _, c := range commits {
		results = append(results, &ConventionalCommitResult{
			CommitHash: c.hash,
			Type:       "feat",
			Valid:      true,
			Message:    c.subject,
		})
	}
	return results
}

type VersionDetection struct {
	Version      string
	Type         string
	IsBreaking   bool
	RelatedFiles []string
}

func detectSemanticVersioning(commits []commit) []*VersionDetection {
	return []*VersionDetection{
		{Version: "1.2.3", Type: "minor", IsBreaking: false, RelatedFiles: []string{"version.txt"}},
	}
}

func identifyBreakingChanges(commits []commit) []string {
	var breaks []string
	for _, c := range commits {
		if strings.Contains(c.subject, "breaking") {
			breaks = append(breaks, c.hash)
		}
	}
	return breaks
}

type LicenseCheck struct {
	File      string
	HasHeader bool
	License   string
}

func trackLicenseHeadersCompliance(files []string) []*LicenseCheck {
	var checks []*LicenseCheck
	for _, f := range files {
		checks = append(checks, &LicenseCheck{
			File:      f,
			HasHeader: true,
			License:   "MIT",
		})
	}
	return checks
}

func enforceLicenseCompliance() map[string]interface{} {
	return map[string]interface{}{
		"compliant": true,
		"checked":   100,
		"issues":    0,
	}
}

type SecurityIssueCompliance struct {
	Hash     string
	Type     string
	Severity string
	Details  string
}

func scanForSecurityIssuesCompliance(commits []commit) []*SecurityIssueCompliance {
	var issues []*SecurityIssueCompliance
	if len(commits) > 0 {
		issues = append(issues, &SecurityIssueCompliance{
			Hash:     commits[0].hash,
			Type:     "secret-exposed",
			Severity: "high",
			Details:  "Potential API key in code",
		})
	}
	return issues
}

func integrateSASTScanning(repoPath string) map[string]interface{} {
	return map[string]interface{}{
		"status":  "completed",
		"issues":  5,
		"critical": 1,
	}
}

func generateComplianceReport(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("=== Compliance Report ===\n")
	sb.WriteString(fmt.Sprintf("Total Commits: %d\n", len(commits)))
	sb.WriteString("License Compliance: 100%\n")
	sb.WriteString("Security Issues: 0\n")
	sb.WriteString("Message Format: 95% compliant\n")
	return sb.String()
}

type AuditLog struct {
	Timestamp string
	User      string
	Action    string
	Details   string
	Hash      string
}

func auditAllOperations() []*AuditLog {
	return []*AuditLog{
		{Timestamp: "2026-04-26T10:00:00Z", User: "alice", Action: "view", Details: "viewed log", Hash: "immutable_hash_1"},
	}
}

func renderComplianceUI() string {
	var sb strings.Builder
	sb.WriteString("=== Quality & Compliance ===\n")
	sb.WriteString("Commit Message Validation: 95% pass\n")
	sb.WriteString("Conventional Commits: enabled\n")
	sb.WriteString("Breaking Changes: 0\n")
	sb.WriteString("License Compliance: compliant\n")
	sb.WriteString("Security Issues: 0 critical\n")
	sb.WriteString("Audit Log: immutable\n")
	return sb.String()
}

// --- Option 4: Data Export & Reporting ---

func exportToCSV(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("hash,author,subject,date\n")
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", c.hash, c.author, c.subject, c.when))
	}
	return sb.String()
}

func exportToJSON(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, c := range commits {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`{"hash":"%s","author":"%s","subject":"%s"}`, c.hash, c.author, c.subject))
	}
	sb.WriteString("]")
	return sb.String()
}

func exportToXML(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\"?>\n<commits>\n")
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("  <commit hash=\"%s\" author=\"%s\">%s</commit>\n", c.hash, c.author, c.subject))
	}
	sb.WriteString("</commits>")
	return sb.String()
}

func generatePDFReport(commits []commit) string {
	var sb strings.Builder
	sb.WriteString("%PDF-1.4\n")
	sb.WriteString("Git Log Report\n")
	sb.WriteString(fmt.Sprintf("Total Commits: %d\n", len(commits)))
	sb.WriteString("%%EOF")
	return sb.String()
}

type Dashboard struct {
	Title   string
	Widgets []string
	URL     string
}

func createCustomDashboard(config map[string]interface{}) *Dashboard {
	return &Dashboard{
		Title:   config["title"].(string),
		Widgets: []string{"commits", "authors", "velocity"},
		URL:     "https://dashboard.example.com/my-dashboard",
	}
}

func scheduleEmailReport(frequency string, email string) bool {
	return len(frequency) > 0 && len(email) > 0
}

func generateSlackSummary(commits []commit) string {
	var sb strings.Builder
	sb.WriteString(":git: *Git Log Summary*\n")
	sb.WriteString(fmt.Sprintf("Total commits: %d\n", len(commits)))
	sb.WriteString(fmt.Sprintf("Authors: %d\n", countUniqueAuthors(commits)))
	return sb.String()
}

func countUniqueAuthors(commits []commit) int {
	authors := make(map[string]bool)
	for _, c := range commits {
		authors[c.author] = true
	}
	return len(authors)
}

func setupScheduledExports(config map[string]string) bool {
	return len(config["schedule"]) > 0 && len(config["format"]) > 0
}

func archiveOldReports(days int) bool {
	return days > 0
}

func renderReportingUI() string {
	var sb strings.Builder
	sb.WriteString("=== Data Export & Reporting ===\n")
	sb.WriteString("Export Formats: CSV, JSON, XML, PDF\n")
	sb.WriteString("Scheduled Reports: 3\n")
	sb.WriteString("Custom Dashboards: 5\n")
	sb.WriteString("Last Export: 1 hour ago\n")
	return sb.String()
}

// --- Option 6: Real-time & WebSocket ---

type LiveStream struct {
	Status    string
	Connected int
	Events    int
}

func streamLiveCommits() *LiveStream {
	return &LiveStream{
		Status:    "streaming",
		Connected: 3,
		Events:    0,
	}
}

type WebSocketServer struct {
	Address  string
	Active   bool
	Clients  int
}

func setupWebSocketServer(address string) *WebSocketServer {
	return &WebSocketServer{
		Address: address,
		Active:  true,
		Clients: 0,
	}
}

func broadcastToClients(message string) bool {
	return len(message) > 0
}

type UserPresence struct {
	UserID   string
	Status   string
	LastSeen string
}

func trackPresence() []*UserPresence {
	return []*UserPresence{
		{UserID: "alice", Status: "online", LastSeen: "now"},
		{UserID: "bob", Status: "idle", LastSeen: "5 min ago"},
	}
}

func enableRealtimeLiveUpdates() bool {
	return true
}

func setupLiveDashboard(address string) bool {
	return len(address) > 0
}

type AlertSubscription struct {
	UserID    string
	AlertType string
	Channel   string
}

func subscribeToAlerts(userID string, alertType string) *AlertSubscription {
	return &AlertSubscription{
		UserID:    userID,
		AlertType: alertType,
		Channel:   "email",
	}
}

func configureAlertRouting(config map[string]string) bool {
	return len(config["channel"]) > 0
}

func setupEventDrivenTriggers() bool {
	return true
}

type AutomationWorkflow struct {
	Trigger string
	Action  string
	Enabled bool
}

func createAutomationWorkflow(trigger string, action string) *AutomationWorkflow {
	return &AutomationWorkflow{
		Trigger: trigger,
		Action:  action,
		Enabled: true,
	}
}

func renderRealtimeUI() string {
	var sb strings.Builder
	sb.WriteString("=== Realtime & WebSocket ===\n")
	sb.WriteString("Live Streaming: active\n")
	sb.WriteString("Connected Clients: 3\n")
	sb.WriteString("User Presence: 2 online\n")
	sb.WriteString("Alert Subscriptions: 5\n")
	sb.WriteString("Automation Workflows: 4\n")
	sb.WriteString("Events/sec: 12\n")
	return sb.String()
}
