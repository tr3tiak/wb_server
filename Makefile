all: run
	

run:
	go run .

build:
	go build -o main .

clean:
	rm main

.PHONY: all run build clean