package gitlog

import (
	"testing"
)
func TestCalculateBisectProgress_ComputesSteps(t *testing.T) {
	state := bisectState{
		good:    []string{"a", "b"},
		bad:     []string{"e", "f"},
		current: "c",
	}
	steps := calculateBisectProgress(state)
	if steps == 0 {
		t.Error("should calculate non-zero steps")
	}
}

func TestBisectFindCulprit_IdentifiesCommit(t *testing.T) {
	commits := []commit{
		{hash: "a1a", shortHash: "a1a1"},
		{hash: "b2b", shortHash: "b2b2"},
		{hash: "c3c", shortHash: "c3c3"},
		{hash: "d4d", shortHash: "d4d4"},
		{hash: "e5e", shortHash: "e5e5"},
	}
	culprit := bisectFindCulprit(commits, []string{"a1a"}, []string{"e5e"})
	if culprit == "" {
		t.Error("should find culprit commit")
	}
}

func TestBisectMarkGood_AddsToGoodList(t *testing.T) {
	m := model{
		bisectState: bisectState{
			active:  true,
			current: "bbb",
			good:    []string{},
			bad:     []string{},
		},
	}
	m = bisectMarkGood(m)
	if len(m.bisectState.good) != 1 {
		t.Errorf("expected 1 good commit, got %d", len(m.bisectState.good))
	}
}

func TestBisectMarkBad_AddsToBadList(t *testing.T) {
	m := model{
		bisectState: bisectState{
			active:  true,
			current: "aaa",
			good:    []string{},
			bad:     []string{},
		},
	}
	m = bisectMarkBad(m)
	if len(m.bisectState.bad) != 1 {
		t.Errorf("expected 1 bad commit, got %d", len(m.bisectState.bad))
	}
}

func TestInitiateBisect_CreatesState(t *testing.T) {
	m := model{
		commits: []commit{
			{hash: "aaa", shortHash: "aaa1111"},
			{hash: "bbb", shortHash: "bbb2222"},
			{hash: "ccc", shortHash: "ccc3333"},
		},
		cursor: 1,
	}
	m = initiateBisect(m)
	if !m.bisectState.active {
		t.Error("bisect should be active")
	}
	if m.bisectState.current == "" {
		t.Error("should have current commit")
	}
}

func TestCalculateProductivity_Empty(t *testing.T) {
	metrics := calculateProductivity([]commit{})
	if len(metrics) != 0 {
		t.Errorf("empty should have no metrics, got %d", len(metrics))
	}
}

func TestCalculateProductivity_LineChanges(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "first"},
		{hash: "bbb", subject: "second"},
	}
	// Would need diff data to calculate actual metrics
	metrics := calculateProductivity(commits)
	if metrics == nil {
		t.Error("should return map")
	}
}

func TestTimeBasedStats_Empty(t *testing.T) {
	stats := calculateTimeStats([]commit{})
	if len(stats) != 0 {
		t.Errorf("empty commits should have no stats, got %d", len(stats))
	}
}

func TestTimeBasedStats_ByDay(t *testing.T) {
	commits := []commit{
		{when: "1 day ago", subject: "today"},
		{when: "1 day ago", subject: "also today"},
		{when: "7 days ago", subject: "week ago"},
	}
	stats := calculateTimeStats(commits)
	if len(stats) == 0 {
		t.Error("should calculate time stats")
	}
}

func TestDetectLanguage_Go(t *testing.T) {
	AssertEqual(t, "go", detectLanguage("main.go"), "should detect go")
}

func TestDetectLanguage_Python(t *testing.T) {
	AssertEqual(t, "python", detectLanguage("script.py"), "should detect python")
}

func TestDetectLanguage_Unknown(t *testing.T) {
	if lang := detectLanguage("file.unknown"); lang == "" {
		t.Error("should return some language for unknown")
	}
}

func TestDetectLanguage_NoExtension(t *testing.T) {
	if lang := detectLanguage("Makefile"); lang != "makefile" {
		t.Errorf("expected 'makefile', got %q", lang)
	}
}

func TestDetectAnomalies_FindsUnusual(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Small fix"},
		{hash: "bbb", subject: "Massive refactor with 10000 lines"},
		{hash: "ccc", subject: "Another fix"},
	}
	anomalies := detectAnomalies(commits)
	if len(anomalies) == 0 {
		t.Error("should detect anomalies")
	}
}

func TestAnalyzeCodeOwnership_ComputesOwners(t *testing.T) {
	commits := []commit{
		{author: "Alice", subject: "Fix: main.go"},
		{author: "Alice", subject: "Fix: main.go"},
		{author: "Bob", subject: "Add: utils.go"},
		{author: "Bob", subject: "Add: utils.go"},
		{author: "Bob", subject: "Add: utils.go"},
	}
	ownership := analyzeCodeOwnership(commits)
	if len(ownership) == 0 {
		t.Error("should compute code ownership")
	}
	aliceData := ownership["Alice"]
	if aliceData.author != "Alice" {
		t.Errorf("author: got %q", aliceData.author)
	}
}

func TestAnalyzeMerges_DetectsFastForwards(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Merge branch feature"},
		{hash: "bbb", subject: "Regular commit"},
	}
	analysis := analyzeMerges(commits)
	if len(analysis) == 0 {
		t.Error("should analyze merges")
	}
}

func TestAnalyzeCommitCoupling_FindsCoChanges(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Update main.go and utils.go"},
		{hash: "bbb", subject: "Fix main.go and utils.go"},
		{hash: "ccc", subject: "Add config.go"},
	}
	couplings := analyzeCommitCoupling(commits)
	if len(couplings) == 0 {
		t.Error("should find coupled files")
	}
}

func TestAnalyzeCommitComplexity_ScoresComplexity(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Simple fix"},
		{hash: "bbb", subject: "Refactor with multiple changes"},
	}
	m := model{
		commits:       commits,
		showComplexity: true,
	}
	m = analyzeComplexity(m)
	if len(m.commitMetrics) == 0 {
		t.Error("should analyze commits")
	}
}

func TestCalculateCommitSize_ComputesLines(t *testing.T) {
	metrics := calculateCommitMetrics("abc1234", 500, 50)
	if metrics.hash != "abc1234" {
		t.Errorf("hash: got %q", metrics.hash)
	}
	if metrics.linesChanged != 500 {
		t.Errorf("lines: expected 500, got %d", metrics.linesChanged)
	}
}

func TestCalculateCommitComplexity_EstimatesScore(t *testing.T) {
	metrics := commitMetrics{
		hash:         "abc",
		linesChanged: 300,
		filesChanged: 15,
	}
	score := calculateComplexityScore(metrics)
	if score < 0 || score > 100 {
		t.Errorf("invalid complexity score: %d", score)
	}
}

