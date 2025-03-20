#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Building the application...${NC}"
go build -o bin/main main.go
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to build application${NC}"
  exit 1
fi

echo -e "${GREEN}Running application locally on port 8080...${NC}"
./bin/main &

sleep 5
if lsof -i:8080; then
  echo -e "${GREEN}Server started successfully${NC}"
else
  echo -e "${RED}Failed to start server${NC}"
  exit 1
fi
