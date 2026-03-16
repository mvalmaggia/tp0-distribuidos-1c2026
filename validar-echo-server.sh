#!/bin/bash

NETWORK="tp0_testing_net"
SERVER_CONTAINER="server"
CONFIG_PATH="./server/config.ini"
TEST_MSG="echo_test_message"

# Get the server IP address from Docker network
SERVER_IP=$(docker network inspect $NETWORK \
  --format '{{range .Containers}}{{if eq .Name "'"$SERVER_CONTAINER"'"}}{{.IPv4Address}}{{end}}{{end}}' | cut -d'/' -f1)

if [ -z "$SERVER_IP" ]; then
  echo "action: test_echo_server | result: fail"
  exit 1
fi

# Get the port from config.ini on host
SERVER_PORT=$(grep '^SERVER_PORT' ./server/config.ini | cut -d'=' -f2 | tr -d ' ')

if [ -z "$SERVER_PORT" ]; then
  echo "action: test_echo_server | result: fail"
  exit 1
fi

# Run a temporary busybox container to send a message via netcat
RESULT=$(docker exec $SERVER_CONTAINER sh -c "echo $TEST_MSG | nc $SERVER_IP $SERVER_PORT")

if [ "$RESULT" = "$TEST_MSG" ]; then
  echo "action: test_echo_server | result: success"
else
  echo "action: test_echo_server | result: fail"
fi