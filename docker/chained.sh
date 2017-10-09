#!/bin/bash
echo "WARNING: hacky script, make sure that kademnet is defined and corresponds to the environment variable"
echo 'WARNING: 2<=$1<=254, or the script will probably break'
make build

for i in `seq 2 $1`;
do
    let last=$i-1
    echo "Starting $i which bootstraps $last"
    docker run -d --rm --workdir /kademlia --net kademnet -e "KADIP=172.18.0.$i" -e "BOOTIP=172.18.0.$last" --ip 172.18.0.$i d7024e_deploy sh -c "./kademliad" 
    sleep .5s
done
