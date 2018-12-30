writer: get-dependencies fmt vet lint

get-dependencies:
	@echo "Downloading dependencies"
	go get
	go get -u golang.org/x/lint/golint

fmt:
	@echo "Running formatting"
	go fmt

vet:
	@echo "Running go vet"
	go vet

lint:
	@echo "Running lint"
	golint --set_exit_status ./...

test:
	go test