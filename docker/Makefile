REPOSITORY="https://github.com/ax-rwnd/d7024e-distfile.git"

all: build test_store

.PHONY: build
build:
	docker build -t d7024e_deploy .

.PHONY: test_store
test_store:
	docker run -it --rm --workdir /kademlia d7024e_deploy  sh -c "./kademliad & sleep 5s; ./dfs store kademliad.toml;"

.PHONY: kademliad
kademliad:
	docker run -it --rm --net kademnet --workdir /kademlia d7024e_deploy  sh -c "./kademliad"
