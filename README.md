# kenos

AIと一緒に働くための仕組みを置いておくrepo。
Go製CLIでskillsを各プロジェクトに配線する。

## 構造

```
kenos/
├── cmd/kenos/         # CLIエントリポイント
├── payload/           # CLIがプロジェクトに届けるもの
│   ├── embed.go       # go:embed でバイナリに同梱する係
│   └── skills/        # 各プロジェクトに配線するskill
├── .claude/skills/    # kenos自身のメタ会話用skill
│   └── kenos/
├── doc/
│   └── ideas.md       # 仕組みアイデアの棚
├── .goreleaser.yml
└── .github/workflows/release.yml
```

## 2つの層

**payload**: CLIがプロジェクトに届けるskill/command。`kenos init` でコピーされる。

**メタ層** (`.claude/skills/kenos/`): kenos自身を改善するための仕組み。このrepoで会話した時だけ有効。

## CLI

```sh
kenos init      # 今いるプロジェクトに payload を配線
kenos version   # バージョン表示
```

## リリース

1. コードを main に commit
2. `git tag v0.x.x && git push origin main --tags`
3. GitHub Actions が GoReleaser でビルド → GitHub Releases にバイナリが上がる

ローカルに goreleaser は不要。

## 運用

- `kenos/` に `cd` して会話を始めると、メタ層が起動してメタ会話モードになる
- 実働層のskillを改善したい時もここで議論する
- 議論の中で出た仕組みアイデアは `doc/ideas.md` に保管する
- 実作業は別のrepoで、別セッションでやる
