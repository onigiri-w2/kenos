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
kenos init          # 今いるプロジェクトに skills / .kenos/tickets/ を配線
kenos task pick     # .kenos/tickets/ からチケットを選んで claude で再開
kenos reflect       # 裏 Claude による振り返りを手動発火
kenos update        # CLI を最新版に自己更新
kenos version       # バージョン表示
```

### kenos init

プロジェクト直下に以下を作る:

- `.claude/skills/` — CLIに同梱された skill ファイル群
- `.kenos/tickets/` — チケットごとの作業ログ置き場

既にあるファイルが更新されている場合は上書きするか確認される。

### kenos task pick

カレントディレクトリから親を遡り、最初に見つけた `.kenos/tickets/` 内のチケットを fzf で一覧表示する。
チケットを選ぶと `claude` が起動し、`/task <ticket>` で作業を再開する。

### kenos reflect

裏 Claude を手動で起動し、transcript から `log.md` / `habits.md` を更新する。

- 引数なし: `.kenos/tickets/*/transcripts` の `- [ ]` 行(未処理)を ticket ごとに集め、それぞれに対して裏 Claude を発火する
- `kenos reflect <path>`: 指定の transcript の所属 ticket を逆引きし、その transcript だけ処理する

処理が完了した行は `- [ ]` → `- [x]` に書き換えられる。

session と ticket の紐付けは `/task` 実行時に `.kenos/tickets/<ticket>/transcripts` に `- [ ] <transcript-path>` の形で append される。1 session が複数 ticket に紐付くのは禁止(衝突したらエラー)。

### kenos update

GitHub Releases から最新バイナリを取得して、今動いているバイナリを置き換える。
skills は CLI に同梱されているので、update すれば一緒に最新になる。
各プロジェクトに反映するには `kenos init` を再実行する。

## チケット作業のファイル構造

`/task <URL>` で初期化すると以下が生成される:

```
.kenos/tickets/<ticket>/
├── overview.md     # 設計図(背景・ゴール・スコープ・現在地・メタ情報)
├── roadmap.md      # 方向性。フェーズの大きな区切り
├── now.md          # 今のフェーズで動くタスク(チェックリスト)
├── issues.md       # 問い・リスク・気になってることの inbox
├── followup.md     # スコープ外で見つけたもの
├── log.md          # 時系列メモ(裏 Claude が更新)
├── habits.md       # 癖の観察(裏 Claude が記録)
├── transcripts     # session ↔ ticket の紐付け(チェックリスト形式)
└── findings/      # 調査・設計知識
```

表 Claude(対話している側)は `overview` / `roadmap` / `now` / `issues` を読み書きし、
裏 Claude(`kenos reflect` で起動)が `log` / `habits` を append する。

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
