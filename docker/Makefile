REPOSITORY="https://github.com/ax-rwnd/d7024e-distfile.git"

all: build test_store

build:
	docker build -t d7024e_deploy .

test_store:
	docker run -it --rm --workdir /kademlia d7024e_deploy  sh -c "./kademliad & sleep 1s; ./dfs store kademliad.toml;"
