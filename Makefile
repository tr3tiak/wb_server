all: run
	
run:
	go run .

build:
	go build -o main .

clean:
	rm main

kafka:
	docker compose up

.PHONY: all run build clean kafka