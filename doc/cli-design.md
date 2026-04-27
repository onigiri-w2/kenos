# kenos CLI 設計メモ

## 背景

- kenos は「AIと一緒に働くための仕組み」を置くrepo
- skills と companion scripts を、複数プロジェクトで使いたい
- volta モデル: ツール本体は global install、データはプロジェクトローカル
- 将来: プロジェクト横断のデータ集約、クロスPCも視野。今は1PC構成

## 決まったこと

| 項目 | 決定 | 理由 |
|------|------|------|
| 言語 | Go | シングルバイナリ、サブコマンド構造が楽 |
| skills 配布 | Go embed でバイナリに同梱 | CLI と skills が常に同じバージョン。将来分離可能 |
| kenos skill | embed 対象外。kenos repo の `.claude/skills/kenos/` にだけ置く | メタ会話用。各プロジェクトには不要 |
| registry | 持たない | truth は各 pjt のローカル状態。横断が要る時にキャッシュ検討 |
| deploy | GitHub Releases にビルド済みバイナリ。GoReleaser で CI ビルド | Go 不要。curl で取れる。dotfiles と相性がいい |
| install 先 | `~/.kenos/bin` | ユーザー領域、sudo 不要。volta 等と同じパターン |
| update | `kenos update` が GitHub Releases から最新バイナリを自己更新 | npm upgrade 相当 |
| 衝突時 | 上書きせず確認する | pjt 固有スキルを壊さない |
| .gitignore | init では触らない | |
| GitHub repo | public | curl で認証不要 |

## コマンド

### `kenos init`

今いるプロジェクトに kenos を配線する。

1. `kenos.json` を作成（マーカー兼設定ファイル。初期値は `{}`）
2. `.claude/skills/` にバイナリ内の skills を展開・コピー
3. 同名スキルが既にある → 上書きせず確認
4. `.tasks/` を作成（なければ）

冪等。再実行で skills が最新に更新される（衝突チェックあり）。
`kenos.json` が既にあれば init 済みと判断できる。

### `kenos update`

1. GitHub Releases から最新バイナリを取得
2. `~/.kenos/bin/kenos` を差し替える
3. skills もバイナリに同梱されているので一緒に更新される
4. 各プロジェクトの skills 反映は `kenos init` を再実行

### `kenos version`

バイナリに埋め込まれたバージョンを表示する。GoReleaser がビルド時に注入。

### `kenos task-pick`（後で作る）

1. 今いるプロジェクトの `.tasks/` を fzf で一覧表示
2. 選んだ ticket 番号で `claude` を起動し `/task <ticket-no>` を実行

## 注意: 凝集度

init は今のところ skills コピーも `.tasks/` 作成もまとめてやる。本来の凝集の単位:

- **task 系**: skills/task, skills/task-sync, `.tasks/` ディレクトリ, task-pick コマンド — 1つの塊

task 以外の仕組みが増えた時に init の分割を検討する。今は自覚した上でまとめておく。

## deploy

```zsh
# dotfiles の features/claude/install.zsh から
curl -fsSL https://github.com/ken/kenos/releases/latest/download/kenos_darwin_arm64 -o ~/.kenos/bin/kenos
chmod +x ~/.kenos/bin/kenos

# dotfiles の features/claude/shell.zsh で
export PATH="$HOME/.kenos/bin:$PATH"
```
