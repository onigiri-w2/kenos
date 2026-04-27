---
name: task-sync
description: .tasks/配下のlog.mdとJiraのステータスを照合し、完了済みticketを一括同期する
disable-model-invocation: true
---

# /task-sync: 完了ステータスの一括同期

`.tasks/` 配下の log.md を走査し、Jira側で完了しているのに log.md が完了になっていないticketを検出して同期する。

## 手順

1. `.tasks/` 配下の各ディレクトリを走査し、各 log.md の1行目(ticket番号)と「ステータス」欄を読む
2. ticket番号を以下に分類:
   - Jira形式(例: `PJ004-982`、プロジェクトキー + ハイフン + 数字)
   - それ以外(Jiraに紐づかないticket)
3. Jira形式のticketについて、Atlassian MCP で各Jira側のステータスを取得
4. 以下の表を作って表示:

   | ticket番号 | log.md側 | Jira側 | 判定 |
   | --- | --- | --- | --- |
   | PJ004-982 | 進行中 | Done | 同期対象 |
   | PJ004-983 | 完了 | Done | 既に同期済 |
   | PJ004-984 | 進行中 (レビュー待ち) | In Review | 触らない |

5. Jira形式でないticketは「手動判定リスト」として別枠で表示:

   | ticket番号 | log.md側 |
   | --- | --- |
   | local-001 | 進行中 |

   Kenに「この中で完了しているものはありますか? 番号で指定してください」と聞く

6. 同期対象と、Kenが手動で指定した分について、該当log.mdの「ステータス」欄を `完了` に書き換える
7. 書き換えた結果をサマリ表示する

## 判定ルール

- Jira側が Done / Closed 相当 → log.md を `完了` に書き換える(log.md側のステータスが何であっても上書き)
- それ以外のズレ(Jira が In Progress、log.md が レビュー待ち、等) → 触らない

## 振る舞い原則

- Kenの確認なしに log.md を書き換えない。必ず表で提示してから書き換える
- Jira側の取得に失敗したticketがあれば、エラーとして別枠で表示する(同期対象から外す)
