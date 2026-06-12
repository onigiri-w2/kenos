# バグ修正計画: 裏 Claude fork bomb (2026-05-27)

## 症状
claude プロセスが無限増殖し、CPU と token を食い潰した。

## 原因
SessionEnd / SessionStart フックの中で `claude -p`(AI)をバックグラウンド起動していた。
headless の `claude -p` 自身も SessionStart/SessionEnd フックを発火するため、
裏 Claude が起動するたびに次の裏 Claude を生む連鎖(fork bomb)になった。

## 方針
**「記録(自動・AI なし)」と「振り返り(手動・AI あり)」を分離する。**
危険なのは AI 発火の自動化。ただの記録(transcript パスの追記)は AI を呼ばないので安全。

---

## 変更内容

### 1. 記録フックを「追記だけ」に書き換え (payload/hooks/session-end.sh)
- stdin の JSON から `transcript_path` を取得。
- `.kenos/current-ticket` が指す ticket ディレクトリの `transcripts` ファイルに1行追記。
- **dedup**: 既に同じパスが末尾にあれば追記しない。
- **claude は一切起動しない。**
- ループ再発ガード: `KENOS_REFLECTING` が設定されていれば即 exit(reflect 由来の session を記録しない)。

### 2. SessionStart フックを廃止 (payload/hooks/session-start.sh 削除)
- 取りこぼし回収のために裏 Claude を起動していた → 不要。
- SessionStart には transcript_path が来ず、projects dir scan で誤爆もあった。
- 記録は SessionEnd の追記だけで足りる。

### 3. settings.json から SessionStart を除去 (payload/settings.json)
- SessionEnd(記録フック)のみ残す。

### 4. reflect.go を「記録リストをまとめ読み」に書き換え (cmd/kenos/reflect.go)
- `findLatestUnprocessedTranscript`(projects dir scan)を廃止。
- 代わりに ticket の `transcripts` ファイルを正とし、記録された全 transcript を集める。
- `last-processed` で既処理分を除外し、未処理分だけをまとめて1つの `claude -p` に渡す。
- プロンプトは現状どおり log.md(時系列)+ habits.md(癖)へ append。
- `claude -p` 起動時に `KENOS_REFLECTING=1` を環境変数で渡し、自己 session が記録に混ざるのを防ぐ。
- 完了後、処理した transcript 群の末尾を `last-processed` に記録。

### 5. embed.go 調整 (payload/embed.go)
- `hooks/session-start.sh` の embed を外し、`hooks/session-end.sh` のみにする。

### 6. init.go 整合確認 (cmd/kenos/init.go)
- hooks インストール処理はそのまま流用可。embed 対象が session-end.sh だけになるので動作確認のみ。
- settings マージは SessionEnd だけになる。

---

## 安全性の検証ポイント
- フックは echo 追記のみ → AI 起動なし → 構造的にループ不能。
- `kenos reflect` の `claude -p` が SessionEnd を発火しても、`KENOS_REFLECTING` ガードで記録スキップ。
  仮にガードが効かなくても、追記が1行増えるだけで spawn は起きない(loop しない)。

## 動作確認
- `go build ./...` が通る。
- `kenos init` で session-end.sh と settings.json(SessionEnd のみ)が入る。
- ダミー ticket を作り、SessionEnd 相当の入力を流して `transcripts` に1行入ること、二重に入らないこと。
- `kenos reflect` で transcripts をまとめ読みして log.md / habits.md が書かれること。
