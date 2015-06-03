Lightning Graph Representation for the HGVM Pilot
===

Overview
---

The Human Genome Variant Map (HGVM) Pilot Project aims:

> ... to create a draft reference structure that represents all “common” genetic variation, providing a means to stably name and canonically identify each variant. ...

We've focused on the BRCA1 and BRCA2 regions specifically.

We're using a representation that is basically a more structured graph representation.
For this exercise we split sequences up into 'tags' and 'bodies', where tags are chosen
to be unique 24mers and the body sequences are chosen to be around 200bp long.
This follows our thinking about tiling the genome with overlapping tags.
Any variant that would fall on a tag is subsumed into a longer tile to keep the tag unaltered.


This is code and supporting material for a [Lightning tile
representation](https://github.com/curoverse/lightning/blob/master/experimental/abram/Documentation/0-Overview.md) in relation to the
[GA4GH Human  Genome Variation Map (HGVM) Pilot Project](https://github.com/ga4gh/schemas/wiki/Human-Genome-Variation-Map-%28HGVM%29-Pilot-Project).

Our focus is only the BRCA1 and BRCA2 loci.  We've included the haplotype genomic information as referenced by the HGVM pilot project
page and also included the relevant portions of 174 participants from the Harvard Personal Genome Project.

This should be considered very experimental and a work in progress.

Quickstart
---

I've tried to make this repository as self contained as possible but dependencies will always be a problem.

Assuming you have all the appropriate dependencies, the following should 'just work':

```bash
$ git clone https://github.com/abeconnelly/hgvm-lighting-graph
$ cd hgvm-lightning-graph
$ ./bootstrap
```

Assuming everything went well, the final SQLite database file should be located in `db/tilegraph.sqlite3`.


License
---

All code is AGPLv3.  All genomic data is CC0.

The Curoverse logo and name are trademarks of Curoverse, Inc.
