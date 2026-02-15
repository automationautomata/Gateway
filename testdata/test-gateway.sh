set -e

check_hosts() {
    wget -qO- http://multyple.path.ex/v1 | grep -q "test.host.multy"
    wget -qO- http://multyple.path.ex/v2 | grep -q "test.host.multy"
    
    wget -qO- http://single.path.ex/something | grep -q "test.host.single"

    wget -qO- http://somehost.ex/ | grep -q "test.host.single"
}

# test proxy 
check_hosts

echo test proxy - success

# test edge limiter 
sleep 1.3
check_hosts && $(wget -qO- http://somehost.ex/ 2>&1 | grep -q "Too Many Requests")

echo test edge limiter - success
