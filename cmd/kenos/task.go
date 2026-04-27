package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func taskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "タスク管理",
	}
	cmd.AddCommand(taskPickCmd())
	return cmd
}

func taskPickCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pick",
		Short: "タスクを選んで claude で再開する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskPick()
		},
	}
}

type taskEntry struct {
	ticket string
	title  string
	status string
}

func findTasksDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	home, _ := os.UserHomeDir()

	for {
		candidate := filepath.Join(dir, ".tasks")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		if dir == home || dir == "/" {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf(".tasks/ が見つかりません")
}

func parseLogMD(path string) (taskEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return taskEntry{}, err
	}
	defer f.Close()

	var entry taskEntry
	var section string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			section = trimmed
			continue
		}

		if strings.HasPrefix(trimmed, "# ") && entry.ticket == "" {
			entry.ticket = strings.TrimPrefix(trimmed, "# ")
			continue
		}

		if strings.HasPrefix(trimmed, "- ステータス:") {
			entry.status = strings.TrimSpace(strings.TrimPrefix(trimmed, "- ステータス:"))
			continue
		}

		if entry.title == "" && trimmed != "" && !strings.HasPrefix(trimmed, "-") {
			if section == "## AI要約" || (section == "## Ticket本文" && entry.title == "") {
				entry.title = truncate(trimmed, 60)
			}
		}
	}

	return entry, nil
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func runTaskPick() error {
	tasksDir, err := findTasksDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return err
	}

	var lines []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		logPath := filepath.Join(tasksDir, e.Name(), "log.md")
		if _, err := os.Stat(logPath); err != nil {
			continue
		}

		entry, err := parseLogMD(logPath)
		if err != nil {
			continue
		}

		if entry.ticket == "" {
			entry.ticket = e.Name()
		}
		if entry.title == "" {
			entry.title = "(タイトルなし)"
		}
		if entry.status == "" {
			entry.status = "不明"
		}

		line := fmt.Sprintf("%-15s  %-10s  %s", entry.ticket, entry.status, entry.title)
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return fmt.Errorf("タスクが見つかりません")
	}

	fzf := exec.Command("fzf", "--reverse", "--no-sort")
	fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	fzf.Stderr = os.Stderr

	out, err := fzf.Output()
	if err != nil {
		return nil
	}

	selected := strings.TrimSpace(string(out))
	ticket := strings.Fields(selected)[0]

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude が見つかりません")
	}

	projectDir := filepath.Dir(tasksDir)
	if err := os.Chdir(projectDir); err != nil {
		return err
	}

	return syscall.Exec(claudePath, []string{"claude", fmt.Sprintf("/task %s", ticket)}, os.Environ())
}
