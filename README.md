# kenos

AIと一緒に働くための仕組みを、各プロジェクトに配線するCLI。

## インストール

```sh
curl -fsSL https://github.com/onigiri-w2/kenos/releases/latest/download/kenos_darwin_arm64 \
  -o ~/.kenos/bin/kenos
chmod +x ~/.kenos/bin/kenos
```

`~/.kenos/bin` を PATH に通しておく。

## 使い方

```sh
kenos init          # 今いるプロジェクトに skills と .tasks/ を配線
kenos task pick     # .tasks/ からチケットを選んで claude で再開
kenos update        # CLI を最新版に自己更新
kenos version       # バージョン表示
```

### kenos init

プロジェクト直下に以下を作る:

- `kenos.json` — init 済みマーカー
- `.claude/skills/` — CLIに同梱された skill ファイル群
- `.tasks/` — チケットごとの作業ログ置き場

既にあるファイルが更新されている場合は上書きするか確認される。

### kenos task pick

カレントディレクトリから親を遡り、最初に見つけた `.tasks/` 内のチケットを fzf で一覧表示する。
チケットを選ぶと `claude` が起動し、`/task <ticket>` で作業を再開する。

### kenos update

GitHub Releases から最新バイナリを取得して、今動いているバイナリを置き換える。
skills は CLI に同梱されているので、update すれば skill も最新になる。

## 構造

```
kenos/
├── cmd/kenos/         # CLI エントリポイント
├── payload/
│   ├── embed.go       # go:embed でバイナリに同梱
│   └── skills/        # 各プロジェクトに配線する skill
├── .claude/skills/
│   └── kenos/         # kenos 自身のメタ会話用 skill
├── doc/
│   └── ideas.md       # 仕組みアイデアの棚
├── .goreleaser.yml
└── .github/workflows/release.yml
```

## リリース

```sh
git tag v0.x.x && git push origin main --tags
```

GitHub Actions が GoReleaser でビルドし、GitHub Releases にバイナリが上がる。ローカルに goreleaser は不要。
