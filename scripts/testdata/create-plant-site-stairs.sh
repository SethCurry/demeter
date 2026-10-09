#!/bin/bash

for x in $(seq 10); do
    for y in $(seq 10); do
        for z in $(seq 10); do
            if [ "$y" == "$z" ]; then
                sqlite3 ./demeter.sqlite "INSERT INTO plant_site (flow_id, x, y, z) VALUES (1, $x, $y, $z)"
            fi
        done
    done
done
