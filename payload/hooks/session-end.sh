#!/bin/bash
# kenos session-end hook: session終了時に transcript パスを ticket に追記するだけ。
# AI は一切起動しない(起動すると fork bomb になる)。振り返りは `kenos reflect` で手動発火。

set -u

# reflect 由来の裏 Claude session を記録に混ぜない(ループ再発ガード)
if [ -n "${KENOS_REFLECTING:-}" ]; then
  exit 0
fi

input=$(cat)
transcript=$(echo "$input" | jq -r '.transcript_path // empty')

if [ -z "$transcript" ]; then
  exit 0
fi

ticket_dir=$(cat .kenos/current-ticket 2>/dev/null || echo "")
if [ -z "$ticket_dir" ]; then
  exit 0
fi

if [ ! -d "$ticket_dir" ]; then
  exit 0
fi

transcripts_file="$ticket_dir/transcripts"

# dedup: 既に同じパスが末尾にあれば追記しない
if [ -f "$transcripts_file" ]; then
  last_line=$(tail -n 1 "$transcripts_file" 2>/dev/null || echo "")
  if [ "$last_line" = "$transcript" ]; then
    exit 0
  fi
fi

echo "$transcript" >> "$transcripts_file"
exit 0
