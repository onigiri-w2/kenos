package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/onigiri/kenos/payload"
)

var (
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleSkip    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "現在のプロジェクトに kenos を配線する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 1. kenos.json
	kenosJSON := filepath.Join(wd, "kenos.json")
	if _, err := os.Stat(kenosJSON); os.IsNotExist(err) {
		if err := os.WriteFile(kenosJSON, []byte("{}\n"), 0644); err != nil {
			return err
		}
		fmt.Println(styleSuccess.Render("✓ created kenos.json"))
	} else {
		fmt.Println(styleSkip.Render("- kenos.json already exists"))
	}

	// 2. .claude/skills/
	skillsDir := filepath.Join(wd, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return err
	}

	skillsFS, _ := fs.Sub(payload.Skills, "skills")
	err = fs.WalkDir(skillsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		dest := filepath.Join(skillsDir, path)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		return copyWithConfirm(skillsFS, path, dest, 0644)
	})
	if err != nil {
		return err
	}
	fmt.Println(styleInfo.Render("  skills installed to .claude/skills/"))

	// 3. .claude/hooks/
	hooksDir := filepath.Join(wd, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	hooksFS, _ := fs.Sub(payload.Hooks, "hooks")
	err = fs.WalkDir(hooksFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		dest := filepath.Join(hooksDir, path)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		return copyWithConfirm(hooksFS, path, dest, 0755)
	})
	if err != nil {
		return err
	}
	fmt.Println(styleInfo.Render("  hooks installed to .claude/hooks/"))

	// 4. .claude/settings.json (マージ)
	if err := installSettings(wd); err != nil {
		return err
	}

	// 5. .tasks/
	tasksDir := filepath.Join(wd, ".tasks")
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tasksDir, 0755); err != nil {
			return err
		}
		fmt.Println(styleSuccess.Render("✓ created .tasks/"))
	} else {
		fmt.Println(styleSkip.Render("- .tasks/ already exists"))
	}

	// 6. .kenos/ (裏 Claude 機構の状態置き場)
	kenosDir := filepath.Join(wd, ".kenos")
	if _, err := os.Stat(kenosDir); os.IsNotExist(err) {
		if err := os.MkdirAll(kenosDir, 0755); err != nil {
			return err
		}
		fmt.Println(styleSuccess.Render("✓ created .kenos/"))
	} else {
		fmt.Println(styleSkip.Render("- .kenos/ already exists"))
	}

	return nil
}

// copyWithConfirm は src 上のファイルを dest に書き出す。
// dest が既存で内容が異なる場合は上書き確認をする。
func copyWithConfirm(srcFS fs.FS, srcPath, dest string, perm os.FileMode) error {
	srcData, err := fs.ReadFile(srcFS, srcPath)
	if err != nil {
		return err
	}

	relPath, _ := filepath.Rel(filepath.Dir(filepath.Dir(dest)), dest)

	if _, statErr := os.Stat(dest); statErr == nil {
		dstData, _ := os.ReadFile(dest)
		if string(srcData) == string(dstData) {
			fmt.Println(styleSkip.Render(fmt.Sprintf("- %s (unchanged)", relPath)))
			return nil
		}

		var overwrite bool
		huh.NewConfirm().
			Title(fmt.Sprintf("%s already exists and differs. overwrite?", relPath)).
			Value(&overwrite).
			Run()

		if !overwrite {
			fmt.Println(styleSkip.Render(fmt.Sprintf("- %s skipped", relPath)))
			return nil
		}
	}

	if err := os.WriteFile(dest, srcData, perm); err != nil {
		return err
	}
	fmt.Println(styleSuccess.Render(fmt.Sprintf("✓ %s", relPath)))
	return nil
}

// installSettings は payload の settings.json を .claude/settings.json にマージする。
// hooks の SessionEnd が無ければ追加し、既にあれば触らない。
func installSettings(wd string) error {
	src, err := payload.Settings.ReadFile("settings.json")
	if err != nil {
		return err
	}

	var srcMap map[string]any
	if err := json.Unmarshal(src, &srcMap); err != nil {
		return err
	}

	dest := filepath.Join(wd, ".claude", "settings.json")

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err := os.WriteFile(dest, src, 0644); err != nil {
			return err
		}
		fmt.Println(styleSuccess.Render("✓ .claude/settings.json"))
		return nil
	}

	existing, err := os.ReadFile(dest)
	if err != nil {
		return err
	}

	var existingMap map[string]any
	if err := json.Unmarshal(existing, &existingMap); err != nil {
		return fmt.Errorf(".claude/settings.json のパースに失敗: %w", err)
	}

	added, err := mergeHooks(existingMap, srcMap)
	if err != nil {
		return err
	}
	if len(added) == 0 {
		fmt.Println(styleSkip.Render("- .claude/settings.json (hooks 既に設定済み)"))
		return nil
	}

	var confirm bool
	huh.NewConfirm().
		Title(fmt.Sprintf(".claude/settings.json に hook %v を追加しますか?", added)).
		Value(&confirm).
		Run()
	if !confirm {
		fmt.Println(styleSkip.Render("- .claude/settings.json skipped"))
		return nil
	}

	merged, err := json.MarshalIndent(existingMap, "", "  ")
	if err != nil {
		return err
	}
	merged = append(merged, '\n')

	if err := os.WriteFile(dest, merged, 0644); err != nil {
		return err
	}
	fmt.Println(styleSuccess.Render(fmt.Sprintf("✓ .claude/settings.json (merged: %v)", added)))
	return nil
}

// mergeHooks は src の hooks エントリのうち、dst にまだ無いものを dst に追加する。
// 追加した hook キーのリストを返す。
func mergeHooks(dst, src map[string]any) ([]string, error) {
	srcHooks, ok := src["hooks"].(map[string]any)
	if !ok {
		return nil, nil
	}

	dstHooks, _ := dst["hooks"].(map[string]any)
	if dstHooks == nil {
		dstHooks = map[string]any{}
		dst["hooks"] = dstHooks
	}

	var added []string
	for key, val := range srcHooks {
		if _, exists := dstHooks[key]; exists {
			continue
		}
		dstHooks[key] = val
		added = append(added, key)
	}
	return added, nil
}
