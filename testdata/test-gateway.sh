set -e

test_host() {
    echo query to $1, expected response: $2
    wget -qO- $1 | grep -q "$2"
}

test_all_hosts() {    
    test_host http://multyple.path.ex/v1 test.host.multy

    test_host http://multyple.path.ex/v2 test.host.multy

    test_host http://single.path.ex/something test.host.single

    test_host http://somehost.ex test.host.single 
}

echo start reverse proxy test
test_all_hosts
echo test proxy - success

echo start edge limiter test 
echo wait... && sleep 1.3
test_all_hosts && $(wget -qO- http://somehost.ex/ 2>&1 | grep -q "Too Many Requests")

echo test edge limiter - success
