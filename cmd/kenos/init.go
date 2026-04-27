package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/onigiri/kenos/src/skills"
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

	err = fs.WalkDir(skills.FS, ".", func(path string, d fs.DirEntry, err error) error {
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

		if _, statErr := os.Stat(dest); statErr == nil {
			srcData, _ := fs.ReadFile(skills.FS, path)
			dstData, _ := os.ReadFile(dest)
			if string(srcData) == string(dstData) {
				fmt.Println(styleSkip.Render(fmt.Sprintf("- %s (unchanged)", path)))
				return nil
			}

			var overwrite bool
			huh.NewConfirm().
				Title(fmt.Sprintf("%s already exists and differs. overwrite?", path)).
				Value(&overwrite).
				Run()

			if !overwrite {
				fmt.Println(styleSkip.Render(fmt.Sprintf("- %s skipped", path)))
				return nil
			}
		}

		data, err := fs.ReadFile(skills.FS, path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return err
		}
		fmt.Println(styleSuccess.Render(fmt.Sprintf("✓ %s", path)))
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Println(styleInfo.Render("  skills installed to .claude/skills/"))

	// 3. .tasks/
	tasksDir := filepath.Join(wd, ".tasks")
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tasksDir, 0755); err != nil {
			return err
		}
		fmt.Println(styleSuccess.Render("✓ created .tasks/"))
	} else {
		fmt.Println(styleSkip.Render("- .tasks/ already exists"))
	}

	return nil
}
