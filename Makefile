build:
	go build

test1: build
	./bin/maelstrom/maelstrom test -w echo --bin flyio-challenges --time-limit 1

test2: build
	./bin/maelstrom/maelstrom test -w unique-ids --bin flyio-challenges --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

test3a: build
	./bin/maelstrom/maelstrom test -w broadcast --bin flyio-challenges --node-count 1 --time-limit 20 --rate 10


test3b: build
	./bin/maelstrom/maelstrom test -w broadcast --bin flyio-challenges --node-count 5 --time-limit 20 --rate 10

test3c: build
	./bin/maelstrom/maelstrom test -w broadcast --bin flyio-challenges --node-count 5 --time-limit 20 --rate 10 --nemesis partition

test3d: build
	./bin/maelstrom/maelstrom test -w broadcast --bin  flyio-challenges --node-count 25 --time-limit 20 --rate 100 --latency 100

test3e: build
	./bin/maelstrom/maelstrom test -w broadcast --bin  flyio-challenges --node-count 25 --time-limit 20 --rate 100 --latency 100
debug:
	./bin/maelstrom/maelstrom serve
