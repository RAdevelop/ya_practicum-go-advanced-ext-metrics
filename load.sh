#!/usr/bin/env bash
set -euo pipefail

# ── настройки ─────────────────────────────────────────────────────────────
AGENTS=50
DURATION=60
SERVER_URL="http://localhost:8080"
PPROF_URL="http://localhost:6060/debug/pprof/heap"
PROFILE_OUT="${1:-profiles/base.pprof}"

AGENT_BIN="./cmd/agent/agent"
SERVER_BIN="./cmd/server/server"
SERVER_LOG="./cmd/server.log"
AGENT_LOG_DIR="./cmd/agent-logs"
AGENT_PIDS=()
SERVER_PID=""
SERVER_EXIT_CODE=""

# ── освобождение портов ───────────────────────────────────────────────────
free_ports() {
    for port in 8080 6060; do
        local pids
        pids=$(lsof -ti ":$port" 2>/dev/null || true)
        if [[ -n "$pids" ]]; then
            echo "  порт $port занят, убиваю: $pids"
            kill -TERM $pids 2>/dev/null || true
            for _ in {1..20}; do
                sleep 0.1
                pids=$(lsof -ti ":$port" 2>/dev/null || true)
                [[ -z "$pids" ]] && break
            done
            if [[ -n "$pids" ]]; then
                kill -KILL $pids 2>/dev/null || true
            fi
        fi
    done
}

# ── очистка при выходе ────────────────────────────────────────────────────
cleanup() {
    echo "→ остановка агентов..."
    for pid in "${AGENT_PIDS[@]}"; do
        kill -TERM "$pid" 2>/dev/null || true
    done

    # ждём до 5 секунд, потом SIGKILL
    for i in {1..50}; do
        alive=0
        for pid in "${AGENT_PIDS[@]}"; do
            kill -0 "$pid" 2>/dev/null && alive=$((alive + 1))
        done
        [[ $alive -eq 0 ]] && break
        sleep 0.1
    done

    for pid in "${AGENT_PIDS[@]}"; do
        kill -KILL "$pid" 2>/dev/null || true
    done

    if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
        echo "→ остановка сервера..."
        kill -TERM "$SERVER_PID" 2>/dev/null || true
        for i in {1..50}; do
            kill -0 "$SERVER_PID" 2>/dev/null || break
            sleep 0.1
        done

        if kill -0 "$SERVER_PID" 2>/dev/null; then
        kill -KILL "$SERVER_PID" 2>/dev/null || true
    fi
    fi

    if [[ -n "$SERVER_PID" ]]; then
        set +e
        wait "$SERVER_PID" 2>/dev/null
        SERVER_EXIT_CODE=$?
        set -e

        if [[ "$SERVER_EXIT_CODE" -ne 0 && "$SERVER_EXIT_CODE" -ne 143 && "$SERVER_EXIT_CODE" -ne 130 ]]; then
            echo "✗ сервер завершился с кодом $SERVER_EXIT_CODE" >&2
            if [[ -f "$SERVER_LOG" ]]; then
                echo "───── последние 100 строк лога сервера ─────" >&2
                tail -100 "$SERVER_LOG" >&2
                echo "─────────────────────────────────────────────" >&2
            fi
        fi
    fi

    # страховка: освободить порты
    echo "→ освобождение портов..."
    free_ports
}
trap cleanup EXIT

# ── освобождение портов ДО сборки ─────────────────────────────────────────
echo "→ освобождение портов 8080, 6060..."
free_ports

# ── сборка ────────────────────────────────────────────────────────────────
echo "→ сборка бинарников..."
go build -o "$SERVER_BIN" ./cmd/server
go build -o "$AGENT_BIN"  ./cmd/agent

mkdir -p "$(dirname "$PROFILE_OUT")"
mkdir -p "$AGENT_LOG_DIR"
rm -f "$SERVER_LOG"
rm -f "$AGENT_LOG_DIR"/*.log

# ── запуск сервера ────────────────────────────────────────────────────────
echo "→ запуск сервера (лог: $SERVER_LOG)..."
"$SERVER_BIN" > "$SERVER_LOG" 2>&1 &
SERVER_PID=$!

# ждём, пока pprof-порт откроется
for i in {1..30}; do
    if curl -sf "http://localhost:6060/debug/pprof/" >/dev/null 2>&1; then
        echo "  сервер готов (pprof на :6060)"
        break
    fi
    sleep 0.2
done

if ! curl -sf "http://localhost:6060/debug/pprof/" >/dev/null 2>&1; then
    echo "✗ сервер не поднялся" >&2
    echo "───── лог сервера ─────" >&2
    tail -50 "$SERVER_LOG" >&2 || true
    exit 1
fi

# ── запуск агентов ────────────────────────────────────────────────────────
echo "→ запуск $AGENTS агентов (логи: $AGENT_LOG_DIR/agent-*.log)..."
for i in $(seq 1 "$AGENTS"); do
    "$AGENT_BIN" -r 1 -p 1 -l 5 > "$AGENT_LOG_DIR/agent-$i.log" 2>&1 &
    AGENT_PIDS+=("$!")
done
echo "  запущено ${#AGENT_PIDS[@]} агентов"

# ── ждём DURATION секунд ──────────────────────────────────────────────────
echo "→ нагрузка $DURATION секунд..."
sleep "$DURATION"

# ── проверка, что сервер жив ──────────────────────────────────────────────
if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "✗ сервер упал во время нагрузки" >&2
    echo "───── последние 200 строк лога сервера ─────" >&2
    tail -200 "$SERVER_LOG" >&2 || true
    echo "─────────────────────────────────────────────" >&2

    set +e
    wait "$SERVER_PID" 2>/dev/null
    SERVER_EXIT_CODE=$?
    set -e
    echo "код завершения сервера: $SERVER_EXIT_CODE" >&2
    exit 1
fi

# ── снимаем профиль ───────────────────────────────────────────────────────
echo "→ снятие профиля в $PROFILE_OUT ..."
curl -sf -o "$PROFILE_OUT" "$PPROF_URL"

if [[ -s "$PROFILE_OUT" ]]; then
    echo "✓ профиль сохранён: $PROFILE_OUT ($(du -h "$PROFILE_OUT" | cut -f1))"
else
    echo "✗ профиль пустой" >&2
    exit 1
fi

# ── финал ─────────────────────────────────────────────────────────────────
echo "→ завершение"
echo "  лог сервера: $SERVER_LOG"
echo "  логи агентов: $AGENT_LOG_DIR/"