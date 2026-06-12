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

func runReflect(transcript string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ticketDir, err := readCurrentTicket(wd)
	if err != nil {
		return err
	}

	var transcripts []string
	if transcript != "" {
		// 明示指定: その1つだけを処理する
		transcripts = []string{transcript}
	} else {
		transcripts, err = unprocessedTranscripts(wd, ticketDir)
		if err != nil {
			return err
		}
		if len(transcripts) == 0 {
			fmt.Println(styleSkip.Render("未処理の transcript はありません"))
			return nil
		}
	}

	// 存在するものだけに絞る
	var existing []string
	for _, t := range transcripts {
		if _, err := os.Stat(t); err == nil {
			existing = append(existing, t)
		}
	}
	if len(existing) == 0 {
		fmt.Println(styleSkip.Render("読める transcript がありません"))
		return nil
	}

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude が見つかりません")
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

Stay concise.`, list.String(), ticketDir, ticketDir)

	fmt.Println(styleInfo.Render(fmt.Sprintf("ticket: %s", ticketDir)))
	fmt.Println(styleInfo.Render(fmt.Sprintf("まとめ読みする transcript: %d 件", len(existing))))
	fmt.Println(styleInfo.Render("裏 Claude を起動して振り返りを書き込みます..."))

	c := exec.Command(claudePath, "-p", prompt)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	// 自己 session が記録に混ざらないようガードを立てる
	c.Env = append(os.Environ(), "KENOS_REFLECTING=1")
	if err := c.Run(); err != nil {
		return fmt.Errorf("claude 実行に失敗: %w", err)
	}

	// 処理した末尾を last-processed に記録
	kenosDir := filepath.Join(wd, ".kenos")
	_ = os.MkdirAll(kenosDir, 0755)
	last := existing[len(existing)-1]
	_ = os.WriteFile(filepath.Join(kenosDir, "last-processed"), []byte(last), 0644)

	fmt.Println(styleSuccess.Render("✓ done"))
	return nil
}

func readCurrentTicket(wd string) (string, error) {
	path := filepath.Join(wd, ".kenos", "current-ticket")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(".kenos/current-ticket が読めません(/task で ticket を開いてください): %w", err)
	}
	dir := strings.TrimSpace(string(data))
	if dir == "" {
		return "", fmt.Errorf(".kenos/current-ticket が空です")
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(wd, dir)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("ticket ディレクトリが見つかりません: %s", dir)
	}
	return dir, nil
}

// unprocessedTranscripts は ticket の transcripts ファイルを正とし、
// last-processed で記録された分を除いた未処理の transcript を古い順で返す。
func unprocessedTranscripts(wd, ticketDir string) ([]string, error) {
	recorded, err := readLines(filepath.Join(ticketDir, "transcripts"))
	if err != nil {
		return nil, err
	}
	if len(recorded) == 0 {
		return nil, nil
	}

	lastData, _ := os.ReadFile(filepath.Join(wd, ".kenos", "last-processed"))
	last := strings.TrimSpace(string(lastData))
	if last == "" {
		return recorded, nil
	}

	// last より後ろを未処理とする。last が見つからなければ全件未処理扱い。
	for i, t := range recorded {
		if t == last {
			return recorded[i+1:], nil
		}
	}
	return recorded, nil
}

// readLines はファイルを行ごとに読み、空行を除いてスライスで返す。
// ファイルが無い場合は空スライスを返す。
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
