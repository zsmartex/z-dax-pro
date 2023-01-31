.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go clean -testcache
	go test ./... -race -covermode=atomic -coverprofile=coverage.out

.PHONY: sync
sync:
	go run cmd/main.go sync --config.file=config.yaml

.PHONY: observability
observability:
	go run cmd/main.go observability --config.file=config.yaml


.PHONY: genenv
genenv:
	go run cmd/main.go genenv --config.file=config.yaml

.PHONY: destroy
destroy:
	cd config
	terraform destroy -var="hcloud_token=${HETZNER_TOKEN}"
	cd ../
	rm -rf config