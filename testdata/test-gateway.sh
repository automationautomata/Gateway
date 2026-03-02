
test_host() {
    echo $(date '+%H:%M:%S') query to $1, expected response: $2
    if wget -qO- $1 | grep -q "$2"; then
        echo suscess
    else 
        exit 1
    fi   
}

test_all_hosts() {    
    test_host http://multiple.path.ex/v1 test.host.multi

    test_host http://multiple.path.ex/v2 test.host.multi

    test_host http://single.path.ex test.host.single

    test_host http://somehost.ex test.host.single 
}

echo start reverse proxy test
test_all_hosts
echo test proxy - success

echo && echo start edge limiter test 

echo wait... 
sleep 1.3
test_all_hosts

echo $(date '+%H:%M:%S') query to http://somehost.ex, expected response: Too Many Requests
if wget -qO- http://somehost.ex 2>&1 | grep -q "Too Many Requests"; then
    echo suscess
else 
    exit 1
fi   
echo test edge limiter - success
