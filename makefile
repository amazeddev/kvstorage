BINARY_NAME=kvstore
 
build:
	go build -o ~/.dac/${BINARY_NAME} main.go
 
run:
	go build -o bin/${BINARY_NAME} main.go
	bin/${BINARY_NAME}

install:
	go install
 
clean:
	go clean
	rm ${BINARY_NAME}