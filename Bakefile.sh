# shellcheck shell=bash

task.init() {
	(cd ./schools/elco/parser && poetry install)
	(cd ./schools/elco/scraper && yarn install)
}

task.run() {
	go run main.go "$@"
}

task.elco-scraper() {
	cd ./schools/elco/scraper
	yarn start
}

task.elco-parser() {
	cd ./schools/elco/parser
	poetry run python main.py
}