set -e

: "${SERVER_PATH:?SERVER_PATH is not set}"

for arg in "$@"; do
    PORT="$arg" python "$SERVER_PATH" &
done

wait
