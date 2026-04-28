# 管理対象ツール一覧

このプロジェクトで Ken と一緒に磨いているツールのメタ情報。
各ツールの実物(本体ファイル)はここには置かない。改善議論をする時、Ken が会話に最新版を貼る。

---

## task.md

- **Usage**: `/task <チケットURL>` で起動する Claude Code のスラッシュコマンド。チケットを起点に作業場所(`.tasks/<ticket-no>/`)を作り、ticket本文を取得し、log.md を雛形付きで生成し、「現在地を埋めるフェーズ」に入る
- **どこで動くか**: Claude Code(対象リポジトリの `.claude/commands/task.md`)
- **生成物**:
  - `log.md`: ticket ごとに `.tasks/<ticket-no>/log.md` として生成される記録ファイル。メタ情報、Ticket本文、AI要約、現在地、時系列メモ、癖の観察を1ファイルに凝集する
- **関連ツール**: なし

## task-sync.md

- **Usage**: `/task-sync` で起動する Claude Code のスラッシュコマンド。`.tasks/` 配下の log.md を走査し、チケット管理ツール側で完了しているのに log.md が完了になっていないticketを検出して一括で `完了` に書き換える。ツールに紐づかないticketは Ken に手動判定を促す
- **どこで動くか**: Claude Code(対象リポジトリの `.claude/commands/task-sync.md`)
- **生成物**: なし(既存の log.md のステータス欄を書き換える)
- **関連ツール**: なし(task.md とは独立)
