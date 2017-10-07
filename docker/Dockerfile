FROM frolvlad/alpine-glibc

EXPOSE 8000 8001 8002

VOLUME /kademlia/

ADD build/bin /kademlia

COPY ./dfs.toml /kademlia/dfs.toml
COPY ./kademliad.toml /kademlia/kademliad.toml
