ans=$(hostname);
while true; do
    echo -e "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\n$ans\n" | nc -l -p 80;
done