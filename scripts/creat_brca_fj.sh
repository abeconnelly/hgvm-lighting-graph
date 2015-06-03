#!/bin/bash

tlib="t/2c5.tagset"
for s in `echo ../brca1/GI262359905_rc.seq  ../brca1/GI528476558.seq`
do
  ofn=`basename $s .seq`
  ./tileset2fj -i $s -t $tlib > out-data/$ofn.fj
  gzip out-data/$ofn.fj
done

tlib="t/247.tagset"
for s in `echo ../brca2/GI388428999.seq  ../brca2/GI528476586.seq`
do
  ofn=`basename $s .seq`
  ./tileset2fj -i $s -t $tlib > out-data/$ofn.fj
  gzip out-data/$ofn.fj
done


