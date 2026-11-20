#!/usr/bin/env bash

set -u

# notify-send is provided by libnotify and uses GNOME's notification service.
# A missing desktop session should not turn a completed Codex run into a hook error.
if ! command -v notify-send >/dev/null 2>&1; then
  exit 0
fi

notify-send \
  --app-name="Codex" \
  --urgency=normal \
  "Работа завершена" \
  "Задача в uptimeapp завершена." \
  >/dev/null 2>&1 || true
