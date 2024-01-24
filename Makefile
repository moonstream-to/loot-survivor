.PHONY: clean build

clean:
	rm -f ./loot-survivor

loot-survivor:
	go build .

build: clean loot-survivor
