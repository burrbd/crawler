test:
	go test -timeout=5s -cover -race ./...

build:
	cd app && go get && go build -o ../crawler 