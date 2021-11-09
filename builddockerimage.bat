export DOCKER_SCAN_SUGGEST=false
docker build -t ilosentenceserver .
docker run -d -p 8083:8083 ilosentenceserver