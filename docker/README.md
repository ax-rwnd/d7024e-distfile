# Dockerfiles for kademliad
This folder contains dockerfiles for building- and deploying kademlia nodes.

## How to build kademliad for docker
Change to the build directory, clean the folders, grab the current project files and build them in docker.
```
        cd build && make
```
The files will now be located in build/bin, ready to deploy.

## How to deploy kademliad
Build kademlia using the above step and then build the current image
```
        make build
```
to test your image by storing a file
```
        make test_store
```
