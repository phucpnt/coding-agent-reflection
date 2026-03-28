.PHONY: install run setup-claude setup-gemini setup-codex reflect verify status

COLLECTOR_URL ?= http://localhost:19321

install:
	go build -o collector ./cmd/collector/
	go build -o query ./cmd/query/
	@echo "Built: ./collector ./query"

run:
	go run ./cmd/collector/

setup-claude:
	./scripts/setup-claude.sh

setup-gemini:
	./scripts/setup-gemini.sh

setup-codex:
	./scripts/setup-codex.sh

reflect:
	@curl -s -X POST $(COLLECTOR_URL)/jobs/daily-reflection | jq .

verify:
	./scripts/verify.sh

status:
	@curl -s $(COLLECTOR_URL)/interactions | jq '[.[] | {ts: .Ts, provider: .Provider, prompt: (.UserPrompt | .[0:80])}]'

test:
	go test ./... -count=1
