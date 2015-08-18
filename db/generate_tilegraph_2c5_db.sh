#!/bin/bash

db_fn="tilegraph_2c5.sqlite3"
seq_fn="sequence_2c5.csv"
graphjoin_fn="graphjoin_2c5.csv"
fa_fn="pgp174_2c5.fa"

fasta_db_fn="fasta_2c5.csv"

variantset_fn="variantset_2c5.csv"
callset_fn="callset_2c5.csv"
variantset_callset_fn="variantset-callset-join_2c5.csv"
gj_vs_join="graphjoin-variantset-join_2c5.csv"

allelecall_fn="allelecall_2c5.csv"
allele_fn="allele_2c5.csv"
allelepath_fn="allelepath_2c5.csv"

rm -f $db_fn

echo "creating $db_fn"
cat graphSQL_v023.sql | sqlite3 $db_fn

echo "import FASTA from $fasta_db_fn"
echo -e '.separator ","\n.import '$fasta_db_fn' FASTA' | sqlite3 $db_fn

echo "import Sequence from $seq_fn"
echo -e '.separator ","\n.import '$seq_fn' Sequence' | sqlite3 $db_fn

echo "import GraphJoin from $graphjoin_fn"
echo -e '.separator ","\n.import '$graphjoin_fn' GraphJoin' | sqlite3 $db_fn

echo "import VariantSet from $variantset_fn"
echo -e '.separator ","\n.import '$variantset_fn' VariantSet' | sqlite3 $db_fn

echo "import CallSet from $callset_fn"
echo -e '.separator ","\n.import '$callset_fn' CallSet' | sqlite3 $db_fn

echo "import VariantSet_CallSet_Join from $variantset_callset_fn"
echo -e '.separator ","\n.import '$variantset_callset_fn' VariantSet_CallSet_Join' | sqlite3 $db_fn

echo "import GraphJoin_VariantSet_Join from $gj_vs_join"
echo -e '.separator ","\n.import '$gj_vs_join' GraphJoin_VariantSet_Join ' | sqlite3 $db_fn

echo "import Allele from $allele_fn"
echo -e '.separator ","\n.import '$allele_fn' Allele' | sqlite3 $db_fn

echo "import AlleleCall from $allelecall_fn"
echo -e '.separator ","\n.import '$allelecall_fn' AlleleCall' | sqlite3 $db_fn

echo "import AllelePathItem from $allelepath_fn"
echo -e '.separator ","\n.import '$allelepath_fn' AllelePathItem' | sqlite3 $db_fn
