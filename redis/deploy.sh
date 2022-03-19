#!/bin/bash

num_nodes=$1
folder=10
dir=$(pwd)
port=6379

file='nodes.txt'
nodes=()
i=0

rm appendonly.aof && rm nodes.conf && rm dump.rdb

while read line; do
  if [ $i == "$num_nodes" ]; then
    break
  fi
  nodes+=("$line")
  i+=1
done < $file

for n in "${nodes[@]}"; do
  cd "${dir}" || return
  mkdir -m 777 "${folder}"
  cd "${folder}" && cp ../redis-server ../redis.conf .
  rm -f appendonly.aof && rm -f nodes.conf && rm -f dump.rdb
  ssh -f "${n}" "cd ${dir}/${folder}  && ./redis-server redis.conf"
  folder=$((folder+1))
done

sleep 5
cd "${dir}" || return
cmd="./redis-cli --cluster create "
for n in "${nodes[@]}"; do
  cmd+="${n}:${port} "
done
cmd+="--cluster-replicas 1"

eval "$cmd"