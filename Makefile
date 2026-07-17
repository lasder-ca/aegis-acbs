.PHONY: test race vet build reproduce research compare release clean

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

build:
	mkdir -p bin
	go build -trimpath -o bin/aegis ./cmd/aegis

reproduce:
	scripts/reproduce.sh

research: reproduce
	bin/aegis benchmark --graph artifacts/hatfield-uk.aegis --queries 100 --repeats 5 --batch 8 --research --seed 1010 --output artifacts/research.json --html artifacts/research.html

compare: reproduce
	scripts/compare-tags.sh artifacts/hatfield-uk.aegis artifacts/tag-comparison

release:
	scripts/build-release.sh

clean:
	rm -rf bin dist .data artifacts/*.aegis
