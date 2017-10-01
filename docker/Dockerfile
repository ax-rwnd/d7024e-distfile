FROM frolvlad/alpine-glibc

EXPOSE 8000 8001 8002

VOLUME /kademlia/

ADD build/bin /kademlia

COPY ./build/d7024e-distfile/src/dfs/dfs.toml /kademlia/dfs.toml
COPY ./build/d7024e-distfile/src/server/kademliad.toml /kademlia/kademliad.toml
