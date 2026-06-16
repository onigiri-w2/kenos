# kenos 再設計: session ↔ ticket 紐付け (2026-06-16)

`.kenos/current-ticket` を廃止し、session ↔ ticket の対応を「session 自身が持つ」設計に変える。
実装は別 session で。

## 背景・動機

現状の `.kenos/current-ticket` はプロジェクト直下の**単一スロット**。これに2つの問題がある:

1. **並行作業で干渉する** — terminal A で `/task A`、terminal B で `/task B` を実行すると、`.current-ticket` が B に上書きされる。A の session-end hook が走ると、A の transcript が ticket-B の `transcripts` に追記される。
2. **メンタルモデルを縛る** — 「現在の ticket は1つ」という実装の都合が、Ken の頭の中に「今は1つしかやれない」という制約として漏れる。実際は ticket-A の最中に ticket-B も触りたい場面はある。

本質は **session→ticket の紐付けをどこに置くか**。現状はグローバル状態(`.current-ticket`)に置いている。これを session 自身に持たせれば、両方が解ける。

## 調査結果

Claude Code は `CLAUDE_CODE_SESSION_ID` を env var で公開している。skill 内から `bash` で取得可能。
transcript file は `~/.claude/projects/<encoded-cwd>/<session-id>.jsonl` に存在し、`encoded-cwd` は cwd の `/` を `-` に置換したもの。

つまり `/task <ticket>` 実行時に **その場で自分の transcript_path が組み立てられる**。

```bash
encoded=$(pwd | sed 's|/|-|g')
transcript="${HOME}/.claude/projects/${encoded}/${CLAUDE_CODE_SESSION_ID}.jsonl"
```

これにより、session→ticket の紐付けを **事後 lookup** ではなく **事前に正しい場所へ直接書く** 形に倒せる。

---

## 新しい設計

### 紐付けの保管場所

`.tasks/<ticket>/transcripts` に、その ticket に紐づく transcript path を append する。これだけで紐付けが成立する。

- `.kenos/current-ticket` は廃止
- `.kenos/sessions/<session-id>` のような lookup table も**作らない**(per-session ファイルすら不要)
- 処理済み管理用の別ファイル(`reflected` 等)も**作らない**。`transcripts` のフォーマットで状態を表す
- session-end hook も**不要**(`/task` 起動時点で書き終わっているため)

### `transcripts` ファイルのフォーマット

markdown チェックリスト形式。`now.md` / `roadmap.md` と揃える。

```
- [ ] /Users/hjwk14/.claude/projects/-Users-hjwk14-projects-kenos/<sid-A>.jsonl
- [x] /Users/hjwk14/.claude/projects/-Users-hjwk14-projects-kenos/<sid-B>.jsonl
```

- `- [ ]` = 未処理(reflect 待ち)
- `- [x]` = 処理済み

利点:
- ファイルが増えない
- 他の kenos ファイルとフォーマットが揃って認知負荷が低い
- 目で見て即わかる

### `/task <ticket>` の挙動

新規フロー / 再開フローの両方で、最後に以下を行う:

1. 自セッションの transcript_path を組み立てる(上述)
2. **衝突検出**: `.tasks/*/transcripts` を全走査し、同じ transcript_path が**別の** ticket に既に登録されていればエラーで止まる(=1 session 1 ticket 強制。理由: session は ticket の作業実行単位=子なので、1 session が複数 ticket を持つのは親子関係として破綻する。status が `[ ]` でも `[x]` でも同じ判定)
3. 自 ticket の `.tasks/<ticket>/transcripts` に `- [ ] <path>` の形で append(dedup: 同じ path が既にあれば追記しない)

### `kenos reflect` の挙動

- **引数なし**: `.tasks/*/transcripts` を全 ticket で走査し、`- [ ]` 行(未処理)を集めて ticket ごとに `claude -p` で振り返らせる
- **引数あり (`kenos reflect <transcript-path>`)**: `.tasks/*/transcripts` を逆引きして該当 ticket を見つけ、その transcript の行だけ処理する。見つからなければエラー
- 処理完了後、対象行を `- [ ]` → `- [x]` に書き換える(temp file + rename で atomic に rewrite)

---

## ファイルごとの変更

### `payload/skills/task/SKILL.md`

- **新規フロー step 5** を書き換え:
  - 旧: `.kenos/current-ticket` に `.tasks/<ticket-no>` を書く
  - 新: 自セッションの transcript_path を組み立て、衝突検出してから `.tasks/<ticket-no>/transcripts` に append(dedup)
- **再開フロー step 2** も同様に書き換え
- 「裏 Claude 機構用」というコメントは削除して、新しい仕組みに合わせた説明にする

### `payload/hooks/session-end.sh`

- ファイルを削除

### `payload/settings.json`

- `SessionEnd` の hook 登録を削除。空でよければ `{}` に。あるいはファイル自体を削除して init 側で扱いを変える

### `cmd/kenos/init.go`

- `payload/settings.json` のマージ処理: hooks エントリがなければ settings.json は触らない(または無視できる)ように
- `payload/hooks/` が空になっても init が壊れないこと
- 既存環境向けに `.claude/hooks/session-end.sh` や `.claude/settings.json` の SessionEnd エントリを**削除はしない**。残っていても無害(ファイルが無いだけで動かないだけ)。後始末は手動でよい

### `cmd/kenos/reflect.go`

- `readCurrentTicket` を削除
- `transcripts` パーサを追加:
  - 各行を `- [ ] <path>` / `- [x] <path>` として解釈
  - 未処理(`- [ ]`)の path を返す関数、全 path を返す関数(衝突検出用)
- `unprocessedTranscripts` を per-ticket 走査に変更:
  - 入力: ticket dir
  - 動作: `<ticketDir>/transcripts` の `- [ ]` 行を返す
- `runReflect` のエントリポイント変更:
  - 引数なし: `.tasks/*/` を列挙 → 各 ticket で未処理を集める → ticket ごとに `claude -p` 発火 → 処理後 対象行を `- [x]` に書き換え
  - 引数あり: 指定 transcript の所属 ticket を `.tasks/*/transcripts` から逆引き → その ticket だけ処理 → 同じく書き換え
- `transcripts` の rewrite は atomic(temp file + `os.Rename`)
- `.kenos/last-processed` への書き込みは削除(状態は `transcripts` の行が持つ)

### `README.md`

- 「`.kenos/` 機構の状態置き場(`current-ticket`, `last-processed`)」の記述を削除
- 「裏 Claude(session 終了時に hook で起動)」の説明を「`kenos reflect` で起動」に修正(session-end hook が廃止されるため)

---

## エッジケース

- **`/task` を実行せず作業した session**: どの ticket にも紐付かない。reflect でも拾われない。意図通り
- **同じ session で `/task A` → `/task B`**: B 側で衝突検出に当たりエラー。session が「ticket の作業実行単位=子」なので、1 session が複数 ticket を持つのは親子関係として破綻するため
- **`claude --continue` で同じ transcript が伸びる**: transcript_path は同じ。`/task` を再実行しなければ `transcripts` への追記もない(dedup で弾かれる)。reflect の差分判定は path 単位なので、伸びた部分の再処理は現状と同じく考慮外(別件)
- **既存 ticket の `.tasks/<ticket>/transcripts` に古いデータが残ってる**: 旧形式は `<path>` だけの行。新形式パーサから見ると未知の行になるので、初回 reflect 時に再処理されてしまう可能性がある。移行時は手動で旧行を `- [x] <path>` に書き換える(処理済み扱いにする)か、削除する

---

## 検証手順 (実装後)

1. terminal を2つ開き、それぞれで `/task A`、`/task B` を実行
2. `.tasks/A/transcripts` と `.tasks/B/transcripts` がそれぞれ別の transcript path を持つことを確認
3. `.kenos/current-ticket` が**作られていない**ことを確認
4. 同じ session 内で `/task A` 実行後、`/task B` を試す → 衝突エラーで止まることを確認
5. それぞれの session を終了し、`kenos reflect` を実行 → 両 ticket の `log.md` / `habits.md` が更新されることを確認
6. `kenos reflect` を再度実行 → 「未処理の transcript はありません」と表示されることを確認

---

## 実装の進め方(別 session で)

1. `payload/skills/task/SKILL.md` 更新(新規 + 再開フロー)
2. `payload/hooks/session-end.sh` 削除、`payload/settings.json` 整理
3. `cmd/kenos/init.go` 微修正(hooks/settings が空でも壊れないように)
4. `cmd/kenos/reflect.go` 書き換え(per-ticket 走査、逆引き、`reflected` ファイル)
5. README 更新
6. ローカルで `kenos init` を再実行 → 検証手順を実施
7. 動作確認後リリースタグ

## 後方互換

- 既存 `.tasks/<ticket>/transcripts` は **format 変更あり**(`<path>` → `- [ ] <path>` / `- [x] <path>`)。手動で書き換えるか、放置(初回 reflect で再処理される)。kenos 自体のチケットは数が少ないので手動でよい
- 既存環境の `.kenos/current-ticket`, `.kenos/last-processed`, `.claude/hooks/session-end.sh`, `.claude/settings.json` の SessionEnd エントリは**自動削除しない**。残っていても無害
- `kenos init` を再実行すれば payload 側の更新が反映される
