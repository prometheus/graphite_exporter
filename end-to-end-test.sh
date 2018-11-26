#!/usr/bin/env bash

# Bootstrap into bats relative to where we are
if [ -z "${EXPORTER_TEST_BASE}" ]
then
  export EXPORTER_TEST_BASE="$(cd "$(dirname "$0")" && pwd)"
  exec "${EXPORTER_TEST_BASE}/vendor/github.com/sstephenson/bats/bin/bats" "$0" "$@"
fi

web_ip=127.0.0.1
web_port=9108
graphit_ip=127.0.0.1
graphite_port=9109

exporter_pid=""

setup() {
  ./graphite_exporter --web.listen-address="${web_ip}:${web_port}" --graphite.listen-address="${graphite_ip}:${graphite_port}" --graphite.mapping-config="${EXPORTER_TEST_BASE}/fixtures/mapping.yml" >&2 &
  exporter_pid=$!
}

teardown() {
  set +m # disable job control messages to silence "Terminated"
  kill "${exporter_pid}" > /dev/null 2> /dev/null
  wait
  set -m # enable job control messages again
}

send_tcp() {
  cat "$@" | nc "${graphite_ip}" "${graphite_port}"
}

get_metrics() {
  curl -sSf "http://${web_ip}:${web_port}/metrics"
}

# https://stackoverflow.com/a/8574392
containsElement() {
  local e match="$1"
  shift
  for e; do [[ "$e" == "$match" ]] && return 0; done
  return 1
}

@test "issue 61: does not crash" {
  now=$(date -u '+%s')
  send_tcp <(sed -e "s/1543160310$/${now}/" "${EXPORTER_TEST_BASE}/fixtures/input_issue_61.txt")
  run get_metrics
  [ $status -eq 0 ]
  containsElement 'rspamd_actions{action="add_header"} 2' "${lines[@]}"
  containsElement 'rspamd_connections 1' "${lines[@]}"
}
