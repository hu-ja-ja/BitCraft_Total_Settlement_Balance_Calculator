package tui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"bitcrafttsbc/internal/api"
	"bitcrafttsbc/internal/usecase"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenModeSelect screen = iota
	screenSearchInput
	screenCandidateSelect
	screenEmpireRankFilter
	screenProcessing
	screenResult
	screenError
)

type mode int

const (
	modeTown mode = iota
	modeEmpire
)

type searchCompletedMsg struct {
	candidates []api.Candidate
}

type searchFailedMsg struct {
	err error
}

type tickMsg struct{}

type calcJob struct {
	mu sync.RWMutex

	done    int
	total   int
	current string

	finished bool
	err      error

	townResult   *usecase.TownResult
	empireResult *usecase.EmpireResult
}

func (j *calcJob) setProgress(p usecase.Progress) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.done = p.Done
	j.total = p.Total
	j.current = p.MemberName
}

func (j *calcJob) setTownResult(result usecase.TownResult) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.townResult = &result
	j.finished = true
}

func (j *calcJob) setEmpireResult(result usecase.EmpireResult) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.empireResult = &result
	j.finished = true
}

func (j *calcJob) setError(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.err = err
	j.finished = true
}

type jobSnapshot struct {
	done         int
	total        int
	current      string
	finished     bool
	err          error
	townResult   *usecase.TownResult
	empireResult *usecase.EmpireResult
}

func (j *calcJob) snapshot() jobSnapshot {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return jobSnapshot{
		done:         j.done,
		total:        j.total,
		current:      j.current,
		finished:     j.finished,
		err:          j.err,
		townResult:   j.townResult,
		empireResult: j.empireResult,
	}
}

type Model struct {
	apiClient  *api.Client
	calculator *usecase.Calculator

	width  int
	height int

	screen screen
	mode   mode

	modeCursor int
	choices    []api.Candidate
	choiceIdx  int
	rank9Cursor int
	includeRank9 bool
	pendingEmpireCandidate *api.Candidate

	input textinput.Model
	spin  spinner.Model

	isLoading bool
	job       *calcJob

	progressDone  int
	progressTotal int
	progressName  string

	resultTown   *usecase.TownResult
	resultEmpire *usecase.EmpireResult
	resultOffset int

	errMessage string
}

func New() Model {
	input := textinput.New()
	input.Placeholder = "検索文字列を入力"
	input.CharLimit = 64
	input.Focus()
	input.Width = 40

	spin := spinner.New()
	spin.Spinner = spinner.Dot

	apiClient := api.New("BitCraft_TSBC_Go (discord: hu_ja_ja_)")

	return Model{
		apiClient:  apiClient,
		calculator: usecase.New(apiClient),
		screen:     screenModeSelect,
		mode:       modeTown,
		includeRank9: true,
		input:      input,
		spin:       spin,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case searchCompletedMsg:
		m.isLoading = false
		m.choices = msg.candidates
		m.choiceIdx = 0
		if len(msg.candidates) == 0 {
			m.screen = screenError
			m.errMessage = "候補が見つかりませんでした。Escで戻って再検索できます。"
			return m, nil
		}
		m.screen = screenCandidateSelect
		return m, nil
	case searchFailedMsg:
		m.isLoading = false
		m.screen = screenError
		m.errMessage = fmt.Sprintf("検索中にエラー: %v", msg.err)
		return m, nil
	case tickMsg:
		if m.screen != screenProcessing || m.job == nil {
			return m, nil
		}
		s := m.job.snapshot()
		m.progressDone = s.done
		m.progressTotal = s.total
		m.progressName = s.current
		if !s.finished {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, tea.Batch(cmd, pollTick())
		}
		if s.err != nil {
			m.screen = screenError
			m.errMessage = fmt.Sprintf("集計中にエラー: %v", s.err)
			m.job = nil
			return m, nil
		}
		m.resultTown = s.townResult
		m.resultEmpire = s.empireResult
		m.resultOffset = 0
		m.screen = screenResult
		m.job = nil
		return m, nil
	}

	if m.screen == screenSearchInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	if m.screen == screenProcessing {
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" || key == "q" {
		return m, tea.Quit
	}

	switch m.screen {
	case screenModeSelect:
		switch key {
		case "up", "k":
			if m.modeCursor > 0 {
				m.modeCursor--
			}
		case "down", "j":
			if m.modeCursor < 1 {
				m.modeCursor++
			}
		case "enter":
			if m.modeCursor == 0 {
				m.mode = modeTown
			} else {
				m.mode = modeEmpire
			}
			m.screen = screenSearchInput
			m.input.SetValue("")
			m.input.Focus()
		}
	case screenSearchInput:
		switch key {
		case "esc":
			m.screen = screenModeSelect
			m.input.Blur()
			return m, nil
		case "enter":
			query := strings.TrimSpace(m.input.Value())
			if query == "" {
				m.screen = screenError
				m.errMessage = "検索文字列が空です。Escで戻って入力してください。"
				return m, nil
			}
			m.screen = screenProcessing
			m.isLoading = true
			m.progressDone = 0
			m.progressTotal = 0
			m.progressName = ""
			m.resultTown = nil
			m.resultEmpire = nil
			return m, tea.Batch(m.spin.Tick, searchCandidatesCmd(m.apiClient, m.mode, query))
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	case screenCandidateSelect:
		switch key {
		case "up", "k":
			if m.choiceIdx > 0 {
				m.choiceIdx--
			}
		case "down", "j":
			if m.choiceIdx < len(m.choices)-1 {
				m.choiceIdx++
			}
		case "esc":
			m.screen = screenSearchInput
			m.input.Focus()
		case "enter":
			selected := m.choices[m.choiceIdx]
			if m.mode == modeEmpire {
				m.pendingEmpireCandidate = &selected
				m.rank9Cursor = 0
				m.screen = screenEmpireRankFilter
				return m, nil
			}
			m.startCalcJob(selected)
			m.screen = screenProcessing
			m.progressDone = 0
			m.progressTotal = 0
			m.progressName = ""
			return m, tea.Batch(m.spin.Tick, pollTick())
		}
	case screenEmpireRankFilter:
		switch key {
		case "up", "k":
			if m.rank9Cursor > 0 {
				m.rank9Cursor--
			}
		case "down", "j":
			if m.rank9Cursor < 1 {
				m.rank9Cursor++
			}
		case "esc":
			m.screen = screenCandidateSelect
		case "enter":
			m.includeRank9 = m.rank9Cursor == 0
			if m.pendingEmpireCandidate == nil {
				m.screen = screenCandidateSelect
				return m, nil
			}
			selected := *m.pendingEmpireCandidate
			m.startCalcJob(selected)
			m.screen = screenProcessing
			m.progressDone = 0
			m.progressTotal = 0
			m.progressName = ""
			m.pendingEmpireCandidate = nil
			return m, tea.Batch(m.spin.Tick, pollTick())
		}
	case screenResult:
		switch key {
		case "up", "k":
			m.resultOffset--
		case "down", "j":
			m.resultOffset++
		case "pgup", "b":
			m.resultOffset -= m.resultPageSize()
		case "pgdown", "f":
			m.resultOffset += m.resultPageSize()
		case "home", "g":
			m.resultOffset = 0
		case "end", "G":
			m.resultOffset = 1<<31 - 1
		case "r":
			m.screen = screenSearchInput
			m.input.SetValue("")
			m.input.Focus()
		case "m":
			m.screen = screenModeSelect
		}
		m.clampResultOffset()
	case screenError:
		switch key {
		case "esc":
			if m.isLoading {
				m.screen = screenSearchInput
			} else {
				m.screen = screenSearchInput
			}
			m.input.Focus()
		case "m":
			m.screen = screenModeSelect
		}
	}

	return m, nil
}

func (m *Model) startCalcJob(selected api.Candidate) {
	job := &calcJob{}
	m.job = job

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	if m.mode == modeTown {
		go func() {
			result, err := m.calculator.ComputeTownBalance(ctx, selected.EntityID, job.setProgress)
			if err != nil {
				job.setError(err)
				return
			}
			if result.TownName == "" {
				result.TownName = selected.Name
			}
			job.setTownResult(result)
		}()
		return
	}

	go func() {
		result, err := m.calculator.ComputeEmpireHexite(ctx, selected.EntityID, m.includeRank9, job.setProgress)
		if err != nil {
			job.setError(err)
			return
		}
		if result.EmpireName == "" {
			result.EmpireName = selected.Name
		}
		job.setEmpireResult(result)
	}()
}

func pollTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func searchCandidatesCmd(client *api.Client, selectedMode mode, query string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if selectedMode == modeTown {
			candidates, err := client.SearchTowns(ctx, query)
			if err != nil {
				return searchFailedMsg{err: err}
			}
			return searchCompletedMsg{candidates: candidates}
		}

		candidates, err := client.SearchEmpires(ctx, query)
		if err != nil {
			return searchFailedMsg{err: err}
		}
		return searchCompletedMsg{candidates: candidates}
	}
}

func (m Model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	totalStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("28")).Padding(0, 1)
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))

	var body strings.Builder
	body.WriteString(headerStyle.Render("BitCraft Total Balance / Empire Hexite"))
	body.WriteString("\n\n")

	switch m.screen {
	case screenModeSelect:
		modes := []string{"町の総資産を計算", "国のエネルギー総量を計算"}
		for i, item := range modes {
			prefix := "  "
			line := item
			if i == m.modeCursor {
				prefix = "> "
				line = selectedStyle.Render(item)
			}
			body.WriteString(prefix + line + "\n")
		}
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("Enter: 選択  ↑/↓: 移動  q: 終了"))

	case screenSearchInput:
		if m.mode == modeTown {
			body.WriteString("対象: 町の総資産\n\n")
			body.WriteString("町名を入力してください\n")
		} else {
			body.WriteString("対象: 国のエネルギー総量\n\n")
			body.WriteString("帝国名を入力してください\n")
		}
		body.WriteString(m.input.View() + "\n\n")
		body.WriteString(hintStyle.Render("Enter: 検索  Esc: モード選択へ戻る  q: 終了"))

	case screenCandidateSelect:
		if m.mode == modeTown {
			body.WriteString("町候補\n\n")
		} else {
			body.WriteString("帝国候補\n\n")
		}
		for i, c := range m.choices {
			line := fmt.Sprintf("%s (ID: %s)", c.Name, c.EntityID)
			prefix := "  "
			if i == m.choiceIdx {
				prefix = "> "
				line = selectedStyle.Render(line)
			}
			body.WriteString(prefix + line + "\n")
		}
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("Enter: 実行  Esc: 再検索  ↑/↓: 移動  q: 終了"))

	case screenEmpireRankFilter:
		body.WriteString("帝国集計オプション\n\n")
		body.WriteString("rank 9（一般市民）のメンバーを含めますか？\n\n")
		options := []string{"含める", "除外する"}
		for i, option := range options {
			prefix := "  "
			line := option
			if i == m.rank9Cursor {
				prefix = "> "
				line = selectedStyle.Render(option)
			}
			body.WriteString(prefix + line + "\n")
		}
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("Enter: 決定して集計開始  Esc: 帝国候補へ戻る  ↑/↓: 移動"))

	case screenProcessing:
		if m.isLoading {
			body.WriteString(fmt.Sprintf("%s 検索中...\n\n", m.spin.View()))
			body.WriteString(hintStyle.Render("しばらくお待ちください"))
			break
		}

		body.WriteString(fmt.Sprintf("%s 集計中...\n", m.spin.View()))
		if m.progressTotal > 0 {
			percent := float64(m.progressDone) / float64(m.progressTotal) * 100
			body.WriteString(fmt.Sprintf("進捗: %d / %d (%.1f%%)\n", m.progressDone, m.progressTotal, percent))
		} else {
			body.WriteString("進捗: 初期化中\n")
		}
		if m.progressName != "" {
			body.WriteString(fmt.Sprintf("処理中メンバー: %s\n", m.progressName))
		}
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("処理完了まで待機してください  q: 終了"))

	case screenResult:
		resultContent := m.resultContent()
		resultContent, current, maxOffset := m.applyResultScroll(resultContent)
		if m.mode == modeTown && m.resultTown != nil {
			body.WriteString(selectedStyle.Render("町の総資産 結果"))
			body.WriteString("\n\n")
			body.WriteString(totalStyle.Render(fmt.Sprintf("総資産: %d", m.resultTown.Total)))
			body.WriteString("\n\n")
			body.WriteString(resultContent)
		}
		if m.mode == modeEmpire && m.resultEmpire != nil {
			body.WriteString(selectedStyle.Render("帝国ヘキサイト総量 結果"))
			body.WriteString("\n\n")
			body.WriteString(totalStyle.Render(fmt.Sprintf("総量: %d", m.resultEmpire.Total)))
			body.WriteString("\n\n")
			body.WriteString(resultContent)
		}
		body.WriteString("\n")
		body.WriteString(hintStyle.Render(fmt.Sprintf("↑/↓: スクロール  PgUp/PgDn: ページ移動  g/G: 先頭/末尾  r: 再検索  m: モード選択  q: 終了  [%d/%d]", current+1, maxOffset+1)))

	case screenError:
		body.WriteString(errorStyle.Render("エラー"))
		body.WriteString("\n\n")
		body.WriteString(m.errMessage + "\n\n")
		body.WriteString(hintStyle.Render("Esc: 検索へ戻る  m: モード選択  q: 終了"))
	}

	panel := lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))
	if m.width > 0 {
		panel = panel.MaxWidth(m.width - 4)
	}
	return panel.Render(body.String())
}

func (m Model) resultPageSize() int {
	if m.height > 14 {
		return m.height - 14
	}
	return 10
}

func (m *Model) applyResultScroll(content string) (string, int, int) {
	lines := strings.Split(content, "\n")
	pageSize := m.resultPageSize()
	maxOffset := 0
	if len(lines) > pageSize {
		maxOffset = len(lines) - pageSize
	}
	offset := m.resultOffset
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	start := offset
	end := start + pageSize
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n"), offset, maxOffset
}

func (m *Model) clampResultOffset() {
	if m.resultOffset < 0 {
		m.resultOffset = 0
	}

	maxOffset := m.resultMaxOffset()
	if m.resultOffset > maxOffset {
		m.resultOffset = maxOffset
	}
}

func (m Model) resultMaxOffset() int {
	lines := strings.Split(m.resultContent(), "\n")
	pageSize := m.resultPageSize()
	if len(lines) <= pageSize {
		return 0
	}
	return len(lines) - pageSize
}

func (m Model) resultContent() string {
	if m.mode == modeTown && m.resultTown != nil {
		var out strings.Builder
		out.WriteString(fmt.Sprintf("町: %s (ID: %s)\n", m.resultTown.TownName, m.resultTown.TownEntityID))
		out.WriteString(fmt.Sprintf("treasury: %d\n", m.resultTown.Treasury))
		out.WriteString(fmt.Sprintf("Wallet合計: %d\n", m.resultTown.TotalWallet))
		out.WriteString(fmt.Sprintf("Market合計: %d\n", m.resultTown.TotalMarket))
		out.WriteString("\nメンバー内訳\n")
		for _, member := range m.resultTown.Members {
			out.WriteString(fmt.Sprintf("- %s: Wallet %d / Market %d / Total %d\n", member.UserName, member.Wallet, member.Market, member.Total))
		}
		return out.String()
	}

	if m.mode == modeEmpire && m.resultEmpire != nil {
		var out strings.Builder
		citizenPolicy := "含める"
		if !m.includeRank9 {
			citizenPolicy = "除外する"
		}
		out.WriteString(fmt.Sprintf("帝国: %s (ID: %s)\n", m.resultEmpire.EmpireName, m.resultEmpire.EmpireEntityID))
		out.WriteString(fmt.Sprintf("一般市民(rank 9): %s\n", citizenPolicy))
		out.WriteString(fmt.Sprintf("メンバーhexite合計: %d\n", m.resultEmpire.TotalMemberHexite))
		out.WriteString(fmt.Sprintf("帝国の所持エネルギー: %d\n", m.resultEmpire.EmpireCurrencyTreasury))
		out.WriteString("\nメンバー内訳\n")
		for _, member := range m.resultEmpire.Members {
			out.WriteString(fmt.Sprintf("- %s: Hexite %d\n", member.UserName, member.Hexite))
		}
		return out.String()
	}

	return "結果がありません"
}
