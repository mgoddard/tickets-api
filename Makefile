.PHONY: default
default: build ;

build: clean
	cd pgx-api/ && go mod init pgx-api && go mod tidy && go build . && cd -
	cd gopg-api/ && go mod init gopg-api && go mod tidy && go build . && cd -

clean:
	cd pgx-api/ && go clean && rm -f go.mod go.sum && cd -
	cd gopg-api/ && go clean && rm -f go.mod go.sum && cd -

