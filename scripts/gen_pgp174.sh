#!/bin/bash

mkdir -p out-data

b38chr17="data/b38_brca1_2c5.fj.gz"
b38chr13="data/b38_brca2_247.fj.gz"

b1a="data/GI262359905_rc.fj.gz"
b1b="data/GI528476558.fj.gz"

#idir="/scratch/brca/tiles/pgp174"
idir="tiles"
opt="-i <( fjfilter -i <(zcat $b38chr17) -s 2c5.00.3cd -e 2c5.00.52c )"
opt=" $opt -i <( fjfilter -i <(zcat $b1a) -s 2c5.00.3cd -e 2c5.00.52c )"
opt=" $opt -i <( fjfilter -i <(zcat $b1b) -s 2c5.00.3cd -e 2c5.00.52c )"
for d in `ls $idir/*/2c5.fj.gz`
do
  opt="$opt -i <( fjfilter -i <(zcat $d) -s 2c5.00.3cd -e 2c5.00.52c )"
done

cmd="./src/create_tile_graph --progress $opt -fasta-csv out-data/pgp174_2c5_fasta.csv -fasta out-data/pgp174_2c5.fa -sequence out-data/pgp174_2c5.seq -graphjoin out-data/pgp174_2c5.gj"
echo ">>>> $cmd"
bash -c " $cmd "

b2a="data/GI388428999.fj.gz"
b2b="data/GI528476586.fj.gz"

opt="-i <( fjfilter -i <(zcat $b38chr13) -s 247.00.abb -e 247.00.c20 )"
opt=" $opt -i <( fjfilter -i <(zcat $b2a) -s 247.00.abb -e 247.00.c20 )"
opt=" $opt -i <( fjfilter -i <(zcat $b2b) -s 247.00.abb -e 247.00.c20 )"
for d in `ls $idir/*/247.fj.gz`
do
  opt="$opt -i <( fjfilter -i <(zcat $d) -s 247.00.abb -e 247.00.c20 )"
done

starts=" -start-sequence-id 1000000 -start-graphjoin-id 1000000 -fasta-id 2"
cmd="./src/create_tile_graph --progress $opt -fasta-csv out-data/pgp174_247_fasta.csv -fasta out-data/pgp174_247.fa -sequence out-data/pgp174_247.seq -graphjoin out-data/pgp174_247.gj $starts"
echo ">>>> $cmd"
bash -c " $cmd "

cat out-data/pgp174_247_fasta.csv out-data/pgp174_2c5_fasta.csv > out-data/pgp174_fasta.csv
cat out-data/pgp174_247.seq out-data/pgp174_2c5.seq > out-data/pgp174.seq
cat out-data/pgp174_247.fa out-data/pgp174_2c5.fa > out-data/pgp174.fa
cat out-data/pgp174_247.gj out-data/pgp174_2c5.gj > out-data/pgp174.gj
