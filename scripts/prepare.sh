#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="project-sem-1"
DB_USER="validator"
DB_PASSWORD="val1dat0r"

export PGPASSWORD=$DB_PASSWORD

echo -e "${GREEN}Installing dependencies...${NC}"
go mod tidy
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to install dependencies${NC}"
  exit 1
fi

echo -e "${GREEN}Preparing database...${NC}"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f platform/storage/migrations/20252003_create_prices_table.sql
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to prepare database${NC}"
  exit 1
fi

echo -e "${GREEN}Database prepared successfully${NC}"