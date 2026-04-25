package gitlog

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD93D"))
	hashStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F39C12"))
	authorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	whenStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	subjStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Bold(true)
	addStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ECC71"))
	removeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C"))
	hunkStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	metaStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	msgStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	alertStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Bold(true)
	divStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	focusIndStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	selectedBg    = lipgloss.Color("#2A2A2A")
)

type commitsLoadedMsg []commit
type diffLoadedMsg []diffLine
type flashClearMsg struct{}
type editorDoneMsg struct{}

func fetchCommits(repoPath string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("git", "-C", repoPath, "log",
			"--format=%H%x00%h%x00%an%x00%ar%x00%s", "-n", "200").Output()
		if err != nil {
			return commitsLoadedMsg(nil)
		}
		return commitsLoadedMsg(parseCommits(string(out)))
	}
}

func fetchDiff(repoPath, hash string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("git", "-C", repoPath, "show",
			"--stat", "--patch", "--color=never", hash).Output()
		if err != nil {
			return diffLoadedMsg(nil)
		}
		return diffLoadedMsg(parseDiff(string(out)))
	}
}

func flashCmd() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg {
		return flashClearMsg{}
	})
}

// copyToClipboard tries common clipboard tools in order.
func copyToClipboard(s string) error {
	for _, args := range [][]string{
		{"pbcopy"},
		{"wl-copy"},
		{"xclip", "-selection", "clipboard"},
		{"xsel", "--clipboard", "--input"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(s)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no clipboard tool found")
}

// openInEditor writes the diff to a temp file and opens $EDITOR on it.
func openInEditor(repoPath, hash string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	out, _ := exec.Command("git", "-C", repoPath, "show",
		"--color=always", "--no-pager", hash).Output()
	f, err := os.CreateTemp("", "charms-gitlog-*.diff")
	if err != nil {
		return func() tea.Msg { return editorDoneMsg{} }
	}
	_, _ = f.Write(out)
	_ = f.Close()
	name := f.Name()
	cmd := exec.Command(editor, name)
	return tea.ExecProcess(cmd, func(error) tea.Msg {
		_ = os.Remove(name)
		return editorDoneMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return fetchCommits(m.repoPath)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case commitsLoadedMsg:
		m.commits = []commit(msg)
		vc := visibleCommits(m)
		if len(vc) > 0 {
			m.loading = true
			return m, fetchDiff(m.repoPath, vc[0].hash)
		}
		return m, nil

	case diffLoadedMsg:
		m.loading = false
		m.diffLines = []diffLine(msg)
		m.diffOffset = 0
		m.fileItems = parseFileItems(m.diffLines)
		m.fileCursor = 0
		return m, nil

	case flashClearMsg:
		m.flash = ""
		return m, nil

	case editorDoneMsg:
		return m, nil
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if km.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Search mode: all keys feed the query
	if m.searching {
		switch km.Type {
		case tea.KeyEsc:
			m.searching = false
			m.query = ""
			m.cursor = 0
			vc := visibleCommits(m)
			if len(vc) > 0 {
				m.loading = true
				return m, fetchDiff(m.repoPath, vc[0].hash)
			}
		case tea.KeyEnter:
			m.searching = false
		case tea.KeyBackspace:
			runes := []rune(m.query)
			if len(runes) > 0 {
				prevFirst := firstVisibleHash(m)
				m.query = string(runes[:len(runes)-1])
				m.cursor = 0
				if h := firstVisibleHash(m); h != prevFirst && h != "" {
					m.loading = true
					return m, fetchDiff(m.repoPath, h)
				}
			}
		default:
			if len(km.Runes) > 0 {
				prevFirst := firstVisibleHash(m)
				m.query += string(km.Runes)
				m.cursor = 0
				if h := firstVisibleHash(m); h != prevFirst && h != "" {
					m.loading = true
					return m, fetchDiff(m.repoPath, h)
				}
			}
		}
		return m, nil
	}

	vc := visibleCommits(m)

	// Global bindings
	switch km.String() {
	case "q":
		return m, tea.Quit
	case "tab":
		if m.showFiles {
			m = toggleFileView(m)
		} else {
			m = switchPanel(m)
		}
		return m, nil
	case "/":
		m.searching = true
		return m, nil
	case "f":
		if len(m.fileItems) > 0 || m.showFiles {
			m = toggleFileView(m)
		}
		return m, nil
	case "y":
		if m.cursor < len(vc) {
			if err := copyToClipboard(vc[m.cursor].hash); err != nil {
				m.flash = "clipboard unavailable"
			} else {
				m.flash = "copied " + vc[m.cursor].shortHash
			}
			return m, flashCmd()
		}
		return m, nil
	case "e":
		if m.cursor < len(vc) {
			return m, openInEditor(m.repoPath, vc[m.cursor].hash)
		}
		return m, nil
	}

	panelH := diffPanelHeight(m)

	// File list navigation (overrides normal panel nav)
	if m.showFiles {
		switch km.String() {
		case "j", "down":
			if m.fileCursor < len(m.fileItems)-1 {
				m.fileCursor++
			}
		case "k", "up":
			if m.fileCursor > 0 {
				m.fileCursor--
			}
		case "enter", " ":
			if m.fileCursor < len(m.fileItems) {
				m = scrollToDiffLine(m, m.fileItems[m.fileCursor].diffIdx)
				m.showFiles = false
				m.focus = panelDiff
			}
		case "esc":
			m = toggleFileView(m)
		}
		return m, nil
	}

	if m.focus == panelList {
		switch km.String() {
		case "j", "down":
			if m.cursor < len(vc)-1 {
				m.cursor++
				m.diffOffset = 0
				m.loading = true
				return m, fetchDiff(m.repoPath, vc[m.cursor].hash)
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.diffOffset = 0
				m.loading = true
				return m, fetchDiff(m.repoPath, vc[m.cursor].hash)
			}
		case "g":
			if m.cursor != 0 && len(vc) > 0 {
				m.cursor = 0
				m.diffOffset = 0
				m.loading = true
				return m, fetchDiff(m.repoPath, vc[0].hash)
			}
		case "G":
			if last := len(vc) - 1; last >= 0 && m.cursor != last {
				m.cursor = last
				m.diffOffset = 0
				m.loading = true
				return m, fetchDiff(m.repoPath, vc[last].hash)
			}
		case "l":
			m = switchPanel(m)
		}
	} else {
		switch km.String() {
		case "j", "down":
			return scrollDiffDown(m, 1), nil
		case "k", "up":
			return scrollDiffUp(m, 1), nil
		case "d", " ":
			return scrollDiffDown(m, panelH/2), nil
		case "u":
			return scrollDiffUp(m, panelH/2), nil
		case "g":
			m.diffOffset = 0
		case "G":
			return scrollDiffDown(m, len(m.diffLines)), nil
		case "h":
			m = switchPanel(m)
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "\n  loading…\n"
	}

	listW := listPanelWidth(m.width)
	diffW := diffPanelWidth(m.width)
	panelH := diffPanelHeight(m)
	vc := visibleCommits(m)

	var sb strings.Builder

	// Title bar
	sb.WriteString("\n ")
	sb.WriteString(titleStyle.Render("git log"))
	sb.WriteString("  ")
	sb.WriteString(msgStyle.Render(m.repoPath))
	sb.WriteString("\n\n")

	// Left panel header
	var listHdr string
	switch {
	case m.showFiles:
		hdrText := fmt.Sprintf("Files (%d)", len(m.fileItems))
		if m.focus == panelList || m.showFiles {
			listHdr = focusIndStyle.Render("▌") + " " + titleStyle.Render(hdrText)
		} else {
			listHdr = "  " + msgStyle.Render(hdrText)
		}
	case m.query != "":
		hdrText := fmt.Sprintf("Commits [/%s] %d", m.query, len(vc))
		if m.focus == panelList {
			listHdr = focusIndStyle.Render("▌") + " " + titleStyle.Render(hdrText)
		} else {
			listHdr = "  " + msgStyle.Render(hdrText)
		}
	default:
		if m.focus == panelList {
			listHdr = focusIndStyle.Render("▌") + " " + titleStyle.Render("Commits")
		} else {
			listHdr = "  " + msgStyle.Render("Commits")
		}
	}

	// Right panel header
	var diffHdrContent string
	switch {
	case m.loading:
		diffHdrContent = msgStyle.Render("loading…")
	case len(vc) == 0:
		diffHdrContent = msgStyle.Render("no commits")
	default:
		c := vc[m.cursor]
		diffHdrContent = hashStyle.Render(c.shortHash) + "  " +
			msgStyle.Render(truncate(c.subject, diffW-12))
	}
	var diffHdr string
	if m.focus == panelDiff {
		diffHdr = focusIndStyle.Render("▌") + " " + diffHdrContent
	} else {
		diffHdr = "  " + diffHdrContent
	}

	sb.WriteString(lipgloss.NewStyle().Width(listW).Render(listHdr))
	sb.WriteString(divStyle.Render("│"))
	sb.WriteString(diffHdr)
	sb.WriteString("\n")

	// Visible commit window (keep cursor centred)
	commitStart := m.cursor - panelH/2
	if commitStart < 0 {
		commitStart = 0
	}
	if commitStart+panelH > len(vc) {
		commitStart = len(vc) - panelH
		if commitStart < 0 {
			commitStart = 0
		}
	}

	// Visible file window (keep fileCursor centred)
	fileStart := m.fileCursor - panelH/2
	if fileStart < 0 {
		fileStart = 0
	}
	if fileStart+panelH > len(m.fileItems) {
		fileStart = len(m.fileItems) - panelH
		if fileStart < 0 {
			fileStart = 0
		}
	}

	for i := 0; i < panelH; i++ {
		var leftLine string
		if m.showFiles {
			leftLine = renderFileRow(m, fileStart+i, listW)
		} else {
			leftLine = renderCommitRow(m, vc, commitStart+i, listW)
		}
		sb.WriteString(leftLine)
		sb.WriteString(divStyle.Render("│"))
		sb.WriteString(renderDiffRow(m, m.diffOffset+i, diffW))
		sb.WriteString("\n")
	}

	// Footer / hint
	sb.WriteString("\n ")
	switch {
	case m.flash != "":
		sb.WriteString(alertStyle.Render(m.flash))
	case m.searching:
		sb.WriteString(msgStyle.Render("[/] " + m.query + "█   Esc clear   Enter confirm"))
	case m.showFiles:
		sb.WriteString(msgStyle.Render("j/k  navigate   Enter  jump to file   f/Esc  back"))
	case m.focus == panelDiff:
		sb.WriteString(msgStyle.Render("j/k  scroll   d/u  half-page   g/G  top/bottom   h/Tab  commits   q  quit"))
	default:
		sb.WriteString(msgStyle.Render("j/k  navigate   l/Tab  diff   /  search   f  files   y  copy hash   e  editor   g/G  top/bottom   q  quit"))
	}
	sb.WriteString("\n\n")

	content := sb.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderCommitRow renders one row of the commit list panel at index idx into vc.
func renderCommitRow(m model, vc []commit, idx int, w int) string {
	const (
		cursorW = 2
		hashW   = 7
		whenW   = 8
		authW   = 9
	)
	subjectW := w - cursorW - hashW - 1 - authW - 1 - whenW - 1
	if subjectW < 4 {
		subjectW = 4
	}

	if idx < 0 || idx >= len(vc) {
		return lipgloss.NewStyle().Width(w).Render("")
	}

	c := vc[idx]
	selected := idx == m.cursor

	bg := func(st lipgloss.Style) lipgloss.Style {
		if selected && m.focus == panelList {
			return st.Background(selectedBg)
		}
		return st
	}

	var cur string
	if idx == m.cursor {
		cur = bg(cursorStyle).Width(cursorW).Render("▶")
	} else {
		cur = bg(lipgloss.NewStyle()).Width(cursorW).Render("")
	}

	hash := bg(hashStyle).Width(hashW + 1).Render(c.shortHash)
	subj := bg(subjStyle).Width(subjectW + 1).Render(truncate(c.subject, subjectW))
	auth := bg(authorStyle).Width(authW + 1).Render(truncate(firstWord(c.author), authW))
	when := bg(whenStyle).Width(whenW).Render(truncate(c.when, whenW))

	return cur + hash + subj + auth + when
}

// renderFileRow renders one row of the file list panel.
func renderFileRow(m model, idx int, w int) string {
	const cursorW = 2

	if idx < 0 || idx >= len(m.fileItems) {
		return lipgloss.NewStyle().Width(w).Render("")
	}

	fi := m.fileItems[idx]
	selected := idx == m.fileCursor

	bg := func(st lipgloss.Style) lipgloss.Style {
		if selected {
			return st.Background(selectedBg)
		}
		return st
	}

	var cur string
	if selected {
		cur = bg(cursorStyle).Width(cursorW).Render("▶")
	} else {
		cur = bg(lipgloss.NewStyle()).Width(cursorW).Render("")
	}

	path := bg(subjStyle).Width(w - cursorW).Render(truncate(fi.path, w-cursorW))
	return cur + path
}

// renderDiffRow renders one line of the diff panel.
func renderDiffRow(m model, idx int, w int) string {
	if idx < 0 || idx >= len(m.diffLines) {
		return ""
	}
	dl := m.diffLines[idx]
	text := truncate(dl.text, w)
	switch dl.kind {
	case lineAdded:
		return addStyle.Render(text)
	case lineRemoved:
		return removeStyle.Render(text)
	case lineHunk:
		return hunkStyle.Render(text)
	case lineMeta:
		return metaStyle.Render(text)
	default:
		return text
	}
}

// firstVisibleHash returns the hash of the first commit in the current filtered view.
func firstVisibleHash(m model) string {
	vc := visibleCommits(m)
	if len(vc) == 0 {
		return ""
	}
	return vc[0].hash
}

func Run() {
	out, _ := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	repoPath := strings.TrimSpace(string(out))
	if repoPath == "" {
		repoPath = "."
	}
	p := tea.NewProgram(newModel(repoPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
