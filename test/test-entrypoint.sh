"$(curl http://example.com/api)" && echo "host.ex" && \ 
    "$(curl http://example.com/app)" && echo "host.ex" && \ 
    "$(curl http://another.com/)" && echo "host.test.ex"
