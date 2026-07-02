---
name: task
description: チケットURLから作業場所を準備し、タスクを開始する
disable-model-invocation: true
---

# /task: Ticket着手の初期化

チケットのURLを受け取り、作業場所を準備する。

## 入力

- `$ARGUMENTS`: チケットの URL **または** ticket番号 (例: `PROJ-123`)

## 分岐: 新規 or 再開

`$ARGUMENTS` を見て分岐する:

- **ticket番号だけ** (例: `PROJ-123`) かつ `.kenos/tickets/<ticket-no>/overview.md` が存在する → **再開フロー**へ
- **URL** → **新規フロー**へ
- **ticket番号だけ** だが `.kenos/tickets/<ticket-no>/overview.md` が存在しない → 「そのticketの作業場所はまだありません。URLを貼ってください」と返す

---

## ファイル構造

`.kenos/tickets/<ticket-no>/` の中身は以下:

```
.kenos/tickets/<ticket>/
├── overview.md     # 設計図 (背景・ゴール・スコープ・現在地・メタ情報)
├── roadmap.md      # 方向性。フェーズの大きな区切り
├── now.md          # 今のフェーズで動くタスク (チェックリスト)
├── issues.md       # 問い・リスク・気になっていることの inbox (flat bullet)
├── followup.md     # スコープ外で見つけたもの
├── log.md          # 時系列メモ (裏 Claude が更新、表 Claude は読まない)
├── habits.md       # 癖の観察 (裏 Claude が記録、表 Claude は読まない)
└── findings/      # 調査・設計知識 (ad hoc命名、ticket内閉じ)
```

| ファイル | 更新方法 | 主な読み手 |
|---|---|---|
| `overview.md` | 漸近的に更新(現状把握が進むたび) | session 再開時のAI/人間、振り返り |
| `roadmap.md` | フェーズが進んだら更新。細かい TODO は書かない | session 再開時のAI/人間 |
| `now.md` | チェック消化、順序が動的に変わる。フェーズが進んだら書き直す | session毎に表 Claude が更新 |
| `issues.md` | 問いが出たら flat bullet で追加、解消したら消す | 表 Claude / 人間 |
| `followup.md` | 見つけたら追加、ticket完了後に引き渡し | ticket後の人間 |
| `log.md` | append-only | 振り返り時のみ(表 Claude は触らない) |
| `habits.md` | append-only | 振り返り時のみ(表 Claude は触らない) |
| `findings/*` | 自然な単位で追加 | 調査結果を再利用する時 |

**issues と now の関係**: issues は「移送」しない。そのフェーズで解消すべき問いがあれば now にタスクが現れる(手で書く)。now のタスク完了で関連 issue が解消されたら issues.md から消す。ファイル上で紐付けない。

---

## session の紐付け

`/task` の最後に、自セッションを ticket に紐付ける。session は ticket の作業実行単位なので、
1 session が複数 ticket に紐付くのは親子関係として破綻するため禁止する(衝突したらエラーで止める)。

新規/再開どちらのフローでも、`<ticket-no>` を埋めて以下を実行する:

```bash
if [ -z "${CLAUDE_CODE_SESSION_ID:-}" ]; then
  echo "ERROR: CLAUDE_CODE_SESSION_ID が未設定です" >&2
  exit 1
fi
encoded=$(pwd | sed 's|/|-|g')
transcript="${HOME}/.claude/projects/${encoded}/${CLAUDE_CODE_SESSION_ID}.jsonl"
target=".kenos/tickets/<ticket-no>/transcripts"

# 衝突検出: 同じ transcript が別 ticket に登録されていればエラー(状態 [ ] / [x] 問わず)
for f in .kenos/tickets/*/transcripts; do
  [ -f "$f" ] || continue
  if grep -qF "$transcript" "$f" && [ "$f" != "$target" ]; then
    echo "ERROR: この session は $f に既に紐付いている。新しい terminal で /task を実行してください" >&2
    exit 1
  fi
done

# append (dedup)
mkdir -p "$(dirname "$target")"
grep -qF "$transcript" "$target" 2>/dev/null || echo "- [ ] $transcript" >> "$target"
```

衝突エラーが返ったら、Ken にエラーメッセージを伝えてフローを止める。

---

## 再開フロー

1. `.kenos/tickets/<ticket-no>/overview.md`, `roadmap.md`, `now.md`, `issues.md` の4つを読む(**log.md は読まない**)
2. 「session の紐付け」を実行する
3. 以下を表示する:
   - メタ情報(期限、ステータス)
   - 現在地(わかっていること / わかっていないこと)
   - 次にやること(now の未チェック先頭)
   - 未解消の問い(issues の bullet 一覧)
4. 「ここから再開します。何から手をつけますか?」と聞いて始める

---

## 新規フロー

1. URLからticket番号を抽出する (例: `PROJ-123`)
2. `.kenos/tickets/<ticket-no>/` ディレクトリを作成する (既存ならskip)
3. チケット管理ツールの MCP を使って ticket 本文を取得する
4. 以下8つを雛形で生成する:

### `overview.md`

~~~markdown
# <ticket-no>

## メタ情報

- 期限:
- ステータス: 未着手

## Ticket本文

(取得したticket本文をそのまま貼る)

## 現在地

### わかっていること

-

### わかっていないこと

-

## ゴール

(タスクが完了した状態を1-2行で。最初は空でよい)

## スコープ

### やる

-

### やらない

-

## 問題・課題

(設計上の問題や検討すべき課題。出てきたら書く)
~~~

### `roadmap.md`

~~~markdown
# <ticket-no> roadmap

ゴールまでのざっくりした段階。チェックで進捗が見える。
細かい TODO は書かない(now.md に書く)。

- [ ]
~~~

### `now.md`

~~~markdown
# <ticket-no> now

今の段階で動くタスク。完了したらチェック。
順序が動的に変わってもよい。

- [ ]
~~~

**運用**: 1ファイルで固定。roadmap の段階が進んだら中身を新しく書き直す。過去の段階で何をやったかは log.md か git で見る。

### `issues.md`

~~~markdown
# <ticket-no> issues

問い・リスク・気になってること・疑問の inbox。
flat bullet で気軽に書く。解消したら消す。

-
~~~

### `followup.md`

~~~markdown
# <ticket-no> followup

このticketのスコープ外で見つけたもの。ticket完了後に引き渡す。

-
~~~

### `log.md`

~~~markdown
# <ticket-no> log

時系列メモ。`kenos reflect` で裏 Claude が append する。表 Claude は触らない。
~~~

### `habits.md`

~~~markdown
# <ticket-no> habits

癖の観察。良い動きも記録する。裏 Claude が transcript から拾って記録する。
~~~

### `findings/`

空ディレクトリ。調査メモを ad hoc な命名で追加していく。

5. 「session の紐付け」を実行する
6. 作成したパス一覧を表示する
7. 「現在地を埋めるフェーズ」に入る(下記参照)

---

## ステータスの書き方

`overview.md` の「ステータス」欄は以下の3値のいずれかを使う:

- `未着手`
- `進行中`
- `完了`

補足が必要な状態(レビュー待ち、ブロック中、等)は括弧書きで添える:

- `進行中 (レビュー待ち)`
- `進行中 (他チーム待ち)`

## 現在地を埋めるフェーズの進め方

**原則: AIが下書きを出し、Kenが添削する。Kenに白紙から語らせない。**

1. AIがticket本文(必要なら関連コードも)を読み、「わかっていること」の下書きを書く
2. 「わかっていないこと」はAIが埋めない。Kenへの具体的な質問の形で置く(例:「このAPIの呼び出し元はどこ?」)
3. Kenが下書きを読んで突っ込む。Kenの「それ何?」に具体的に答えられなかった項目は「わかっていること」から「わかっていないこと」へ移す
4. 質問にKenが答えたら「わかっていること」へ。答えられなければそのまま「わかっていないこと」に残す
5. 「わかっていないこと」のうち、調べればわかるものは now.md のタスクにする

## 癖の観察

**表 Claude は癖の指摘を会話に出さない。** リアルタイム矯正はやめる。癖は裏 Claude が `habits.md` に記録し、Ken が振り返りで見る。

記録用の観察の観点(会話で問わない):

- TODOを全部やろうとしていないか(削れるものを疑う動きがあったか)
- 1人で決めていい判断か疑ったか(ticket書き手・依頼者・関係者の意図に関わる判断を、確認なしで進めていないか)
- 調査や仕組み作りが本題より膨らんでいないか(止めずに記録だけする。必要な調査を止められるのはだるい)
- 良い動き(自分で立ち止まった、前提を疑った、認識の曖昧さに気づいた)も同じく記録対象

## この後のKenさんとの振る舞い方

Kenは以下の癖を自己申告している:

- タスクの前提を揃えたり疑ったりするのが苦手
- ゴールをすぐ確定させてしまう
- 影響範囲への意識が抜ける
- 知識の拡充に逃げる(仕組み作りで満足して本題が進まない)
- 計画が苦手
- もらったTodoを全部やろうとしてしまう(QCDの観点で疑うテンプレがない)
- 人に聞くのがうまくできない(心の中で違和感を感じても無視しがち)

そのため:

- Kenが現在地を埋める前にゴールや計画に飛ぼうとしたら、一度止めて現在地の確認に戻す
- 癖そのものは指摘しない。記録は裏 Claude が `habits.md` に書き、Kenが振り返りで見る

## AIの振る舞い原則

- 下書きや叩き台はAIが先に出してよい。ただし採否の判断はKenに残す。
- Kenが曖昧に言ったことを「つまり〜ということですね?」と勝手にシャープ化しない。曖昧さは曖昧さのまま扱う。
- 癖の指摘は表に出さない。裏 Claude が `habits.md` に記録し、振り返りで見る。
