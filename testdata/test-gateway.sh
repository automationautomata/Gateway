set -e

check_hosts() {
    wget -qO- http://multypath.ex/api | grep -q "test.multy"
    wget -qO- http://multypath.ex/app | grep -q "test.multy"
    
    wget -qO- http://multypath.ex/something | grep -q "test.single"

    wget -qO- http://somehost.ex/ | grep -q "test.single"
}

# test proxy 
check_hosts

echo test proxy - success

# test edge limiter 
sleep 1 
check_hosts && $(wget -qO- http://somehost.ex/ 2>&1 | grep -q "Too Many Requests")

echo test edge limiter - success
