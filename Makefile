.PHONY: clean build

clean:
	rm -f ./survivor ./loot-survivor

survivor:
	go build -o survivor ./...

build: clean survivor
