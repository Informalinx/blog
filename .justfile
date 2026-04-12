_default:
    @just --list

generate-secret:
    go run cmd/console/main.go secret:generate

migrations action:
    go tool goose {{ action }}
