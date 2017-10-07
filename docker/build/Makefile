PROJECTURL=https://github.com/ax-rwnd/d7024e-distfile.git
PROJECTFOLDER=./d7024e-distfile
BINARIES=$(PWD)/bin
LOCAL=$(PWD)/local

all: clean clone docker

clean:
	rm -rf $(PROJECTFOLDER)
	rm -rf $(BINARIES)
	rm -rf $(LOCAL)

clone:
	git clone $(PROJECTURL)

pull:
	cd $(PROJECTFOLDER) && git pull
	
docker:
	mkdir -p $(BINARIES)
	docker build -t d7024e_build .
	docker run --rm d7024e_build cat /go/dfs > $(BINARIES)/dfs
	docker run --rm d7024e_build cat /go/kademliad > $(BINARIES)/kademliad
	chmod +x $(BINARIES)/kademliad
	chmod +x $(BINARIES)/dfs

.PHONY: local
local:
	mkdir -p $(LOCAL)
	cp $(BINARIES)/kademliad $(LOCAL)/
	cp $(BINARIES)/dfs $(LOCAL)/
	#cp $(PROJECTFOLDER)/src/server/kademliad.toml $(LOCAL)/
	#cp $(PROJECTFOLDER)/src/dfs/dfs.toml $(LOCAL)/
	chmod +x $(BINARIES)/kademliad
	chmod +x $(BINARIES)/dfs
