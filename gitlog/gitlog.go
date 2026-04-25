package gitlog

import (
	"fmt"
	"os/exec"
	"strings"

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
	divStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	focusIndStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	selectedBg    = lipgloss.Color("#2A2A2A")
)

type commitsLoadedMsg []commit
type diffLoadedMsg []diffLine

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
		if len(m.commits) > 0 {
			m.loading = true
			return m, fetchDiff(m.repoPath, m.commits[0].hash)
		}
		return m, nil

	case diffLoadedMsg:
		m.loading = false
		m.diffLines = []diffLine(msg)
		m.diffOffset = 0
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			return switchPanel(m), nil
		}

		panelH := diffPanelHeight(m)

		if m.focus == panelList {
			switch km.String() {
			case "j", "down":
				m = moveCursorDown(m)
				if len(m.commits) > 0 {
					m.loading = true
					return m, fetchDiff(m.repoPath, m.commits[m.cursor].hash)
				}
			case "k", "up":
				m = moveCursorUp(m)
				if len(m.commits) > 0 {
					m.loading = true
					return m, fetchDiff(m.repoPath, m.commits[m.cursor].hash)
				}
			case "g":
				if m.cursor != 0 {
					m.cursor = 0
					m.diffOffset = 0
					m.loading = true
					return m, fetchDiff(m.repoPath, m.commits[0].hash)
				}
			case "G":
				last := len(m.commits) - 1
				if last >= 0 && m.cursor != last {
					m.cursor = last
					m.diffOffset = 0
					m.loading = true
					return m, fetchDiff(m.repoPath, m.commits[last].hash)
				}
			case "l":
				return switchPanel(m), nil
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
				return m, nil
			case "G":
				return scrollDiffDown(m, len(m.diffLines)), nil
			case "h":
				return switchPanel(m), nil
			}
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

	var sb strings.Builder

	// Title bar
	sb.WriteString("\n ")
	sb.WriteString(titleStyle.Render("git log"))
	sb.WriteString("  ")
	sb.WriteString(msgStyle.Render(m.repoPath))
	sb.WriteString("\n\n")

	// Panel headers
	var listHdr string
	if m.focus == panelList {
		listHdr = focusIndStyle.Render("▌") + " " + titleStyle.Render("Commits")
	} else {
		listHdr = "  " + msgStyle.Render("Commits")
	}

	var diffHdrContent string
	switch {
	case m.loading:
		diffHdrContent = msgStyle.Render("loading…")
	case len(m.commits) == 0:
		diffHdrContent = msgStyle.Render("no commits")
	default:
		c := m.commits[m.cursor]
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

	// Visible commit window (keep cursor centered)
	start := m.cursor - panelH/2
	if start < 0 {
		start = 0
	}
	if start+panelH > len(m.commits) {
		start = len(m.commits) - panelH
		if start < 0 {
			start = 0
		}
	}

	for i := 0; i < panelH; i++ {
		sb.WriteString(renderCommitRow(m, start+i, listW))
		sb.WriteString(divStyle.Render("│"))
		sb.WriteString(renderDiffRow(m, m.diffOffset+i, diffW))
		sb.WriteString("\n")
	}

	// Footer
	sb.WriteString("\n ")
	if m.focus == panelList {
		sb.WriteString(msgStyle.Render("j/k  navigate   l/Tab  diff   g/G  top/bottom   q  quit"))
	} else {
		sb.WriteString(msgStyle.Render("j/k  scroll   d/u  half-page   g/G  top/bottom   h/Tab  commits   q  quit"))
	}
	sb.WriteString("\n\n")

	content := sb.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderCommitRow renders one row of the commit list panel.
func renderCommitRow(m model, idx int, w int) string {
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

	if idx < 0 || idx >= len(m.commits) {
		return lipgloss.NewStyle().Width(w).Render("")
	}

	c := m.commits[idx]
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
