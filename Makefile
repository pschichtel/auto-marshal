MODULE := github.com/pschichtel/auto-marshal

local:
	go build "$(MODULE)/cmd/auto-marshal"

run:
	go run "$(MODULE)/cmd/auto-marshal"