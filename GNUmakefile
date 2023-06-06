default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

docs: examples/* internal/provider/*
	go generate ./...

install: internal/* main.go
	go install .