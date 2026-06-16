package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func reflectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reflect [transcript-path]",
		Short: "裏 Claude による振り返りを手動発火する(記録された transcript をまとめ読み)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var transcript string
			if len(args) == 1 {
				transcript = args[0]
			}
			return runReflect(transcript)
		},
	}
}

type transcriptEntry struct {
	processed bool
	path      string
}

type reflectTarget struct {
	ticketDir string
	paths     []string
}

func runReflect(transcript string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var targets []reflectTarget

	if transcript != "" {
		ticketDir, err := findTicketByTranscript(wd, transcript)
		if err != nil {
			return err
		}
		targets = append(targets, reflectTarget{ticketDir: ticketDir, paths: []string{transcript}})
	} else {
		tickets, err := listTickets(wd)
		if err != nil {
			return err
		}
		for _, t := range tickets {
			entries, err := parseTranscripts(filepath.Join(t, "transcripts"))
			if err != nil {
				return err
			}
			var unprocessed []string
			for _, e := range entries {
				if !e.processed {
					unprocessed = append(unprocessed, e.path)
				}
			}
			if len(unprocessed) > 0 {
				targets = append(targets, reflectTarget{ticketDir: t, paths: unprocessed})
			}
		}
	}

	if len(targets) == 0 {
		fmt.Println(styleSkip.Render("未処理の transcript はありません"))
		return nil
	}

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude が見つかりません")
	}

	for _, target := range targets {
		var existing []string
		for _, p := range target.paths {
			if _, err := os.Stat(p); err == nil {
				existing = append(existing, p)
			}
		}
		if len(existing) == 0 {
			fmt.Println(styleSkip.Render(fmt.Sprintf("- %s: 読める transcript がありません", target.ticketDir)))
			continue
		}

		var list strings.Builder
		for _, t := range existing {
			fmt.Fprintf(&list, "- %s\n", t)
		}

		prompt := fmt.Sprintf(`Read these session transcripts (古い順):
%s
Append brief, dated timeline entries to: %s/log.md
- 各 session について、何をやったか、何が決まったか、引っかかったポイントを1-3行で
- Be concise. 議事録ではなく、後で振り返るためのメモ

If any of Ken's habits appeared (good or bad), append to: %s/habits.md
- 場面、出た癖、AIの指摘、Kenの反応、学び
- 良い動きも同じフォーマットで記録する
- 該当する場面がなければ書かない

Stay concise.`, list.String(), target.ticketDir, target.ticketDir)

		fmt.Println(styleInfo.Render(fmt.Sprintf("ticket: %s (%d 件)", target.ticketDir, len(existing))))
		fmt.Println(styleInfo.Render("裏 Claude を起動して振り返りを書き込みます..."))

		c := exec.Command(claudePath, "-p", prompt)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("claude 実行に失敗: %w", err)
		}

		if err := markProcessed(target.ticketDir, existing); err != nil {
			return err
		}
	}

	fmt.Println(styleSuccess.Render("✓ done"))
	return nil
}

// parseTranscripts は <ticketDir>/transcripts を読んで全エントリを返す。
// 旧形式の bare path 行(マーカーなし)は未処理として扱う(初回 reflect 時に処理される)。
func parseTranscripts(path string) ([]transcriptEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []transcriptEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "- [x] "):
			entries = append(entries, transcriptEntry{processed: true, path: strings.TrimPrefix(line, "- [x] ")})
		case strings.HasPrefix(line, "- [ ] "):
			entries = append(entries, transcriptEntry{processed: false, path: strings.TrimPrefix(line, "- [ ] ")})
		default:
			entries = append(entries, transcriptEntry{processed: false, path: line})
		}
	}
	return entries, scanner.Err()
}

// writeTranscripts は entries を canonical 形式で atomic に書き直す。
func writeTranscripts(path string, entries []transcriptEntry) error {
	var b strings.Builder
	for _, e := range entries {
		if e.processed {
			fmt.Fprintf(&b, "- [x] %s\n", e.path)
		} else {
			fmt.Fprintf(&b, "- [ ] %s\n", e.path)
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".transcripts.tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

// listTickets は .kenos/tickets/ 配下の ticket directory を全列挙する。
func listTickets(wd string) ([]string, error) {
	ticketsDir := filepath.Join(wd, ".kenos", "tickets")
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tickets []string
	for _, e := range entries {
		if e.IsDir() {
			tickets = append(tickets, filepath.Join(ticketsDir, e.Name()))
		}
	}
	return tickets, nil
}

// findTicketByTranscript は .kenos/tickets/*/transcripts を逆引きして
// 指定 transcript path を含む ticket directory を返す。
func findTicketByTranscript(wd, target string) (string, error) {
	tickets, err := listTickets(wd)
	if err != nil {
		return "", err
	}
	for _, t := range tickets {
		entries, err := parseTranscripts(filepath.Join(t, "transcripts"))
		if err != nil {
			return "", err
		}
		for _, e := range entries {
			if e.path == target {
				return t, nil
			}
		}
	}
	return "", fmt.Errorf("transcript %s に紐づく ticket が見つかりません", target)
}

// markProcessed は ticket の transcripts を読み、processed の path に該当する行を [x] に書き換える。
func markProcessed(ticketDir string, processed []string) error {
	transcriptsPath := filepath.Join(ticketDir, "transcripts")
	entries, err := parseTranscripts(transcriptsPath)
	if err != nil {
		return err
	}
	set := make(map[string]bool, len(processed))
	for _, p := range processed {
		set[p] = true
	}
	for i := range entries {
		if set[entries[i].path] {
			entries[i].processed = true
		}
	}
	return writeTranscripts(transcriptsPath, entries)
}
