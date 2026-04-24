# kenos

AIと一緒に働くための仕組みを置いておくrepo。

## 構造

```
kenos/
├── CLAUDE.md          # このrepoを開いた時にClaude Codeが読む
├── .claude/           # kenos自身を改善するためのskill/command
│   └── skills/
│       └── kenos/     # メタ会話用skill
├── src/               # 成果物。他repoから使う
│   ├── skills/
│   └── commands/
└── docs/              # kenos自身についての記録
    ├── tools.md       # srcで管理してるツール一覧
    └── ideas.md # 仕組みアイデアの棚
```

## 2つの層

**実働層** (`src/`): タスク遂行のためのskill/command。他のrepoから参照して使う。

**メタ層** (`.claude/`): kenos自身を改善するための仕組み。`kenos/` を開いた時だけ有効。CLAUDE.md から kenos skill が起動する。

## 運用

- `kenos/` に `cd` して会話を始めると、メタ層が起動してメタ会話モードになる
- 実働層のskillを改善したい時もここで議論する
- 議論の中で出た仕組みアイデアは `docs/ideas.md` に保管する
- 実作業は別のrepoで、別セッションでやる
