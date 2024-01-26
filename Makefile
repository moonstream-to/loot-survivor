.PHONY: clean build


build: loot-survivor

clean:
	rm -f ./loot-survivor

rebuild: clean build

loot-survivor:
	go build .
