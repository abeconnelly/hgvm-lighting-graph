We want a fasta sequence per tag and per tile sequence body.

We'll choose the FASTA ID for the tag sequences to be:
  [MD5SUM].[path].[step].t[no-call-bitmask]

Where [no-call-bitmask] is the ascii hexadecimal mask of the no-calls (should be 6 characters), 0-padded so it's fixed width.
path and step should also be 0-padded so they're fixed width.  [step] is taken to be step of the tile it is the left tag for.
This should mean tags start at 1 for paths (for this case, we're starting in the middle so it's not a big issue).

We'll choose the FASTA ID for the tile sequence body to be:
  [MD5SUM].[path].[step].r[tile-variant]+[seed-tile-length]

Where [tile-variant] is the tile variant of the tile.  The [md5sum] is taken to be the whole tiles md5sum and not
the tile body sequence (the md5sum in the FASTA ID will most likely be different than the MD5SUM as represented in
the `Sequence` SQL row entry).  The +[seed-tile-length] is optional and includes a hexadecimal digit representing
the seed tile lenght of the tile.

This means that for both for the tags and tile sequence bodies, there could be duplicate md5sums.  They should have
different FASTA IDs reflecting the different positons they occupy.


The rough outline is that it's hard to do independent steps.  We need to construct the sequences but we also
need to construct the connections between them.  Rather than trying to do extra work after an independent step
of creating the library or the actual sequences, we should use information that we already have at library and
sequence creation to construct the sequence (graph) topology.

Roughly on a tile-path by tile-path basis):

  - Create the tile library:
    * Collect all path.step tile sequences, order them by frequency for your population.
      Remember to store seed tile length.

  - For each path.step.variant in your tile library:
    * take prefix tag and the suffix tag, add these to your tag pool.
    * take the tile body and add these to the tile body pool.
    * store the connection to the tile body pool entry and the prefux and suffix tag pool entries.

  - Generate the sequence file

  - Generate the topology


----

There are some issues with the way we represent tiles.

Since we treat 'no-calls' as part of the md5sum of the sequence and we
rank tiles in terms of frequency, the bodies are derived, inheriting the
rank of the parent tile.  This means we can have bodies with the same
implied sequence, the same md5sum but different ranks.  This shows up
as two body tiles with identical md5sums, identical sequences but different
sequences.

It also means that paths through the graph might pick one or the other.

There is also a bug in how we show paths.  Since allele path items are
calculated only with the md5sum of the body to lookup, it sometimes get's
it wrong and takes the sequence with incorrect rank.  We need to store
the whole tile md5sum along with the subsequence md5sum and somehow
lookup that way in order to figure out proper paths.

**or** do it at the 'generate the graph tile library' level since
we presumably have that information there to begin with...

We could also just conflate the sequences but then we start loosing information.

My opinion is that the better way to solve this is to 'fill in' the no-calls
with ref (or consensus, or something) and then overlay 'no-call' information
when doing queries on an individual basis.
