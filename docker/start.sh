#!/bin/bash
echo "WARNING: hacky script, make sure that kademnet is defined and corresponds to the environment variable"
echo 'WARNING: 2<=$1<=254, or the script will probably break'
make build

for i in `seq 2 $1`;
do
    docker run -d --rm --workdir /kademlia --net kademnet -e "KADIP=172.18.0.$i" --ip 172.18.0.$i d7024e_deploy sh -c "./kademliad" 
    sleep 1s
done
