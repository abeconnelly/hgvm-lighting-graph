#!/bin/bash

b38chr17="data/b38_brca1_2c5.fj.gz"
b38chr13="data/b38_brca2_247.fj.gz"

b1a="data/GI262359905_rc.fj.gz"
b1b="data/GI528476558.fj.gz"

idir="tiles"

opt="-i GRCh38_2c5,<( fjfilter -i <(zcat $b38chr17) -s 2c5.00.3cd -e 2c5.00.52c )"
opt=" $opt -i GI262359905_rc,<( fjfilter -i <(zcat $b1a) -s 2c5.00.3cd -e 2c5.00.52c )"
opt=" $opt -i GI528476558,<( fjfilter -i <(zcat $b1b) -s 2c5.00.3cd -e 2c5.00.52c )"
for d in `ls $idir/*/2c5.fj.gz`
do
  nam=`dirname $d`
  nam=`basename $nam .fj`
  opt="$opt -i ${nam}_2c5,<( fjfilter -i <(zcat $d) -s 2c5.00.3cd -e 2c5.00.52c )"
done

cmd=" ./src/fj2allele $opt -progress -sequence out-data/pgp174_2c5.seq -allele out-data/pgp174_2c5.allele -allele-path out-data/pgp174_2c5.allelepath  -allele-call out-data/pgp174_2c5.allelecall -callset out-data/pgp174_2c5.callset"
echo ">>>> $cmd"
bash -c " $cmd "

b2a="data/GI388428999.fj.gz"
b2b="data/GI528476586.fj.gz"

opt="-i GRCh38_247,<( fjfilter -i <(zcat $b38chr13) -s 247.00.abb -e 2c5.00.c20 )"
opt=" $opt -i GI388428999,<( fjfilter -i <(zcat $b2a) -s 247.00.abb -e 247.00.c20 )"
opt=" $opt -i GI528476586,<( fjfilter -i <(zcat $b2b) -s 247.00.abb -e 247.00.c20 )"
for d in `ls $idir/*/247.fj.gz`
do
  nam=`dirname $d`
  nam=`basename $nam .fj`
  opt="$opt -i ${nam}_247,<( fjfilter -i <(zcat $d) -s 247.00.abb -e 247.00.c20 )"
done

starts=" -start-allele-id 100000 -start-callset-id 1000000"

cmd=" ./src/fj2allele $opt -progress -sequence out-data/pgp174_247.seq -allele out-data/pgp174_247.allele -allele-path out-data/pgp174_247.allelepath  -allele-call out-data/pgp174_247.allelecall -callset out-data/pgp174_247.callset $starts "
echo ">>>> $cmd"
bash -c " $cmd "


cat out-data/pgp174_247.seq out-data/pgp174_2c5.seq > out-data/pgp174.seq
cat out-data/pgp174_247.allele out-data/pgp174_2c5.allele > out-data/pgp174.allele
cat out-data/pgp174_247.allelepath out-data/pgp174_2c5.allelepath > out-data/pgp174.allelepath
cat out-data/pgp174_247.allelecall out-data/pgp174_2c5.allelecall > out-data/pgp174.allelecall
cat out-data/pgp174_247.callset out-data/pgp174_2c5.callset > out-data/pgp174.callset
