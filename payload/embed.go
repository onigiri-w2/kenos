package payload

import "embed"

//go:embed skills/task/SKILL.md skills/task-sync/SKILL.md
var Skills embed.FS

//go:embed hooks/session-end.sh
var Hooks embed.FS

//go:embed settings.json
var Settings embed.FS
