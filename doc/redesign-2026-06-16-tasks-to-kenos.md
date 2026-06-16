# v0.6.0: `.tasks/` → `.kenos/tickets/` 統合、`kenos.json` 廃止

## 背景

- `.tasks/` と `.kenos/` が並列で存在。`.kenos/current-ticket` 廃止以降 `.kenos/` は空き家
- `kenos.json` は `init` 時に作るだけでどこからも読まれていない飾り
- 名前空間が散らかっている。`.tasks/` は generic で他ツールと衝突しうる

## 決定

1. **`.tasks/` 廃止、`.kenos/tickets/<ticket-id>/` に集約**
   - 一段挟むのは、将来 `.kenos/` 配下に config / cache 等を置く余地を残すため
2. **`kenos.json` 廃止**
   - コード上で読まれていない
   - init 済み判定は不要(`kenos init` は冪等)
   - 将来 schema version 等が必要になったら `.kenos/config.json` で受ければよい

## 修正対象

### Go コード

- `cmd/kenos/init.go`
  - `kenos.json` 作成ロジックを削除
  - `.tasks/` 作成 → `.kenos/tickets/` 作成に変更
- `cmd/kenos/task.go`
  - parent 遡って探すターゲットを `.tasks` → `.kenos/tickets`
  - エラーメッセージも更新
- `cmd/kenos/reflect.go`
  - `.tasks/` → `.kenos/tickets/`
  - `.tasks/*/transcripts` の glob を `.kenos/tickets/*/transcripts` に
  - `findTicketByTranscript` も同様

### Skill / payload

- `payload/skills/` 配下で `.tasks/` を参照している箇所を全て `.kenos/tickets/` に書き換え
- 特に task skill(ticket 構造を扱う側)が影響大

### ドキュメント

- `README.md`: 「チケット作業のファイル構造」「構造」のパス記述を更新
- `CLAUDE.md`(kenos repo 側): 必要なら追記

## 移行(既存ユーザー)

- 自動移行ロジックは `kenos init` に入れない方針(現時点では)
- 既存 `.tasks/` の引っ越しは利用者(= Ken の adblue 等)が手で対応する
- 必要になったら別途 `kenos migrate` 的なサブコマンドを切る

## やらないこと

- adblue 側の `.tasks/` → `.kenos/tickets/` の実移行(Ken が後でやる)
- `adblue/.tasks/.git` の削除(これも Ken が手で。今回の改修とは独立)

## 残る論点(次セッション着手前に決めても、進めながらでもよい)

- skill 内ドキュメントの用語整合:「タスク」「チケット」「ticket」「tickets」をどこまで揃えるか
- `.kenos/tickets/` 直下の `README.md`(adblue の `.tasks/README.md` に該当するもの)を payload に含めるかどうか
