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
kenos init          # 今いるプロジェクトに skills / hooks / .tasks/ を配線
kenos task pick     # .tasks/ からチケットを選んで claude で再開
kenos reflect       # 裏 Claude による振り返りを手動発火(取りこぼし回収)
kenos update        # CLI を最新版に自己更新
kenos version       # バージョン表示
```

### kenos init

プロジェクト直下に以下を作る:

- `kenos.json` — init 済みマーカー
- `.claude/skills/` — CLIに同梱された skill ファイル群
- `.claude/hooks/` — session 終了/開始時に裏 Claude を起動する shell スクリプト
- `.claude/settings.json` — hooks 登録(既存ファイルがあればマージ)
- `.tasks/` — チケットごとの作業ログ置き場
- `.kenos/` — 裏 Claude 機構の状態置き場(`current-ticket`, `last-processed`)

既にあるファイルが更新されている場合は上書きするか確認される。
`.claude/settings.json` は既存の中身を保ったまま hooks エントリのみ追加する。

### kenos task pick

カレントディレクトリから親を遡り、最初に見つけた `.tasks/` 内のチケットを fzf で一覧表示する。
チケットを選ぶと `claude` が起動し、`/task <ticket>` で作業を再開する。

### kenos reflect

裏 Claude を手動で起動し、transcript から `log.md` / `habits.md` を更新する。

- 引数なし: `.kenos/last-processed` 以降の最新 transcript を拾う
- `kenos reflect <path>`: 指定の transcript を処理

通常は session 終了時に hook が自動で発火する。`kenos reflect` は PC強制終了などで hook が走らなかった場合の取りこぼし回収用。

### kenos update

GitHub Releases から最新バイナリを取得して、今動いているバイナリを置き換える。
skills / hooks は CLI に同梱されているので、update すれば一緒に最新になる。
各プロジェクトに反映するには `kenos init` を再実行する。

## チケット作業のファイル構造

`/task <URL>` で初期化すると以下が生成される:

```
.tasks/<ticket>/
├── overview.md     # 設計図(背景・ゴール・スコープ・現在地・メタ情報)
├── roadmap.md      # 順序つきTODO(チェックリスト形式)
├── issues.md       # 論点・人に聞くこと
├── followup.md     # スコープ外で見つけたもの
├── log.md          # 時系列メモ(裏 Claude が更新)
├── habits.md       # 癖の観察(裏 Claude が記録)
└── research/      # 調査・設計知識
```

表 Claude(対話している側)は `overview` / `roadmap` / `issues` を読み書きし、
裏 Claude(session 終了時に hook で起動)が `log` / `habits` を append する。

## 構造

```
kenos/
├── cmd/kenos/         # CLI エントリポイント
├── payload/
│   ├── embed.go       # go:embed でバイナリに同梱
│   ├── skills/        # 各プロジェクトに配線する skill
│   ├── hooks/         # 各プロジェクトに配線する hook スクリプト
│   └── settings.json  # hooks 登録テンプレート
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
