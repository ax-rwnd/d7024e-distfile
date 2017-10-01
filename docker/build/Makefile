PROJECTURL=https://github.com/ax-rwnd/d7024e-distfile.git
PROJECTFOLDER=./d7024e-distfile
BINARIES=$(PWD)/bin

all: clean clone docker

clean:
	rm -rf $(PROJECTFOLDER)
	rm -rf $(BINARIES)

clone:
	git clone $(PROJECTURL)

pull:
	cd $(PROJECTFOLDER) && git pull
	
docker:
	mkdir $(BINARIES)
	docker build -t d7024e_build .
	docker run --rm d7024e_build cat /go/dfs > $(BINARIES)/dfs
	docker run --rm d7024e_build cat /go/kademliad > $(BINARIES)/kademliad
	chmod +x $(BINARIES)/kademliad
	chmod +x $(BINARIES)/dfs
