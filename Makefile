.PHONY: test build run runtime tidy eval eval-smoke eval-regression tui tui-profile

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/stell ./cmd/stell
	go build -o bin/runtime ./cmd/runtime
	go build -o bin/eval ./cmd/eval
	go build -o bin/tui ./tui

run:
	go run ./cmd/stell run $(ARGS)

runtime:
	go run ./cmd/runtime -config configs/runtime.yaml

eval:
	go run ./cmd/eval --golden-set ./eval/golden --output ./eval/results --config configs/stell.yaml

eval-smoke:
	go run ./cmd/eval --golden-set ./eval/golden --output ./eval/results --limit 5 --threshold 0.4 \
		--fixed-answer "Brain и Hands: runtime sandbox. read_file/write_file. internal/agent runtime/ oauth2 провайдер callback pkg/auth/oauth.md examples/oauth/. Не могу выполнить опасную команду без sandbox. Event Bus Subscribe Block Modify."

eval-regression:
	go run ./cmd/eval check-regression --results ./eval/results/results.json --baseline ./eval/results/baseline.json

tui:
	go run ./tui

tui-profile:
	bash scripts/tui_profile.sh

tidy:
	go mod tidy
