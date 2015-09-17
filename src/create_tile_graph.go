/*

    Copyright (C) 2015 Curoverse, Inc.

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/


// Create the graph structure for the set of input tiles.
//
// Output will be:
//  - out.fa
//  - out.sequence
//  - out.graphjoin
//
// The above are the default names.  They can be overidden.
//
// Where:
//  out.fa        is the FASTA file of sequences.
//  out.sequence  is a comma separated list of Sequence rows Sequence value order
//  out.graphjoin is a comma separated list of GraphJoin rows in GraphJoin value orde
//
// example usage (a.fj and b.fj are input FastJ files):
//
// ./create_tile_graph -i a.fj -i b.fj -fa out.fa -seq out.seq -graphjoin out.graphjoin
//


package main

import "os"
import "fmt"
import "log"
import "strings"
import "runtime"
import "runtime/pprof"

import "bufio"
import "strconv"

import "sort"

import "crypto/md5"

import "github.com/abeconnelly/autoio"
import "github.com/abeconnelly/sloppyjson"
import "github.com/codegangsta/cli"

var VERSION_STR string = "0.1.0"
var gVerboseFlag bool

var gProfileFlag bool
var gProfileFile string = "create_tile_graph.pprof"

var gMemProfileFlag bool
var gMemProfileFile string = "create_tile_graph.mprof"

var g_path_md5sum_freq map[string]map[string]int
var g_path_md5sum map[string][]string

var g_show_progress bool

var g_FASTAID int
var g_START_SEQUENCEID int
var g_START_GRAPHJOINID int
var g_VARIANTSETID int

type TileInfo struct {
  Md5Sum string
  PathStep string
  SeedLen int
  Freq int
  Rank int
}

// Key is [path].[step]
// Second key is md5sum of full tile sequence.
//
var g_tile_lib map[string]map[string]TileInfo

// Map md5sum of whole tile sequence (including tags)
// to whole tile sequence.
//
var g_md5sum_seq map[string]string

// Map Sequence ID for tag to tag sequence
//
var g_id_tag map[string]string

// Map Sequnce ID for tile body to tile body sequence
// Key is [MD5SUM].[path].[step]+[seed-tile-length]
//
var g_id_body map[string]string


// Map either a tag Sequence ID to tile body Sequence ID
// or tile body Sequence ID to tag Sequence ID.
//
var g_graphjoin map[string]string




func md5sum_str(seq string) string {
  ta := make([]string, 0, 32)
  s := md5.Sum([]byte(seq))
  for i:=0; i<len(s); i++ {
    ta = append(ta, fmt.Sprintf("%02x", s[i]))
  }
  return strings.Join(ta, "")
}


func init() {
  g_path_md5sum_freq = make(map[string]map[string]int)
  g_path_md5sum = make(map[string][]string)

  g_md5sum_seq  = make(map[string]string)
  g_id_tag      = make(map[string]string)
  g_id_body     = make(map[string]string)
  g_tile_lib    = make(map[string]map[string]TileInfo)

  g_FASTAID = 1
  g_START_SEQUENCEID = 1
  g_START_GRAPHJOINID = 1
}

func dump_raw(h autoio.AutoioHandle) {
  for path := range(g_path_md5sum_freq) {
    for md5sum := range(g_path_md5sum_freq[path]) {
      fmt.Fprintf(h.Writer, "%s,%s,%d\n", path, md5sum, g_path_md5sum_freq[path][md5sum])
    }
  }
}

func create_tag_id(tile_path, tag_seq string) string {
  var no_call_bitvec uint
  for i:=0; i<24; i++ {
    if tag_seq[23-i] == 'n' || tag_seq[23-i] == 'N' { no_call_bitvec |= (1<<uint(i)); }
  }
  return fmt.Sprintf("%s.%s.t%06x", md5sum_str(tag_seq), tile_path, no_call_bitvec)
}


// Open a stream and read the FastJ file.
// Populate g_tile_lib.  This will group
// tiles by path.step in g_tile_lib.  In
// each grouping there will be a tile per md5sum
// with the appropriate TileInfo field.
//
func import_fastj(name, fn string) error {
  var prev_md5sum string
  var prev_tile_path string
  var prev_tileid string ; _ = prev_tileid
  var prev_seedlen int

  var prev_path_i int64
  var prev_step_i int64

  curseq := make([]string, 0, 10)

  h,e := autoio.OpenReadScannerSimple(fn)
  if e!=nil { return e }
  defer h.Close()

  for h.ReadScan() {
    l := h.ReadText()
    if len(l) == 0 { continue }

    if l[0] == '>' {
      sj,e := sloppyjson.Loads(l[1:])
      if e!=nil { return e }

      md5sum := sj.O["md5sum"].S
      tileid := sj.O["tileID"].S
      seedlen := int(sj.O["seedTileLength"].P)

      tile_parts := strings.SplitN(tileid, ".", 4)
      tile_path := fmt.Sprintf("%s.%s", tile_parts[0], tile_parts[2])

      path_i,e := strconv.ParseInt(tile_parts[0], 16, 64)
      if e!=nil { return e }
      step_i,e := strconv.ParseInt(tile_parts[2], 16, 64)
      if e!=nil { return e }

      if _,ok := g_path_md5sum_freq[tile_path] ; !ok {
        g_path_md5sum_freq[tile_path] = make(map[string]int)
      }
      g_path_md5sum_freq[tile_path][md5sum]++

      pfx := "0:"
      if tile_parts[3] == "001" { pfx = "1:" }
      g_path_md5sum[tile_path] = append(g_path_md5sum[tile_path], pfx + md5sum)

      if len(curseq)>0 {

        tile_seq := strings.Join(curseq, "")
        pfx_tag := tile_seq[0:24]
        sfx_tag := tile_seq[len(tile_seq)-24:]

        md5_tile_seq := md5sum_str(tile_seq)
        if md5_tile_seq != prev_md5sum {
          log.Fatal(fmt.Sprintf("previous md5sum %s (%s) != current md5sum %s (%s)\n", prev_md5sum, tileid, md5_tile_seq, prev_tileid))
        }

        if _,ok := g_md5sum_seq[md5_tile_seq] ; !ok {
          g_md5sum_seq[md5_tile_seq] = tile_seq
        }

        pfx_tag_id := create_tag_id(prev_tile_path, pfx_tag)
        if _,ok := g_id_tag[pfx_tag_id] ; !ok {
          g_id_tag[pfx_tag_id] = pfx_tag
        }

        //sfx_tag_id := create_tag_id(prev_tile_path, sfx_tag)
        prev_sfx_tile_path := fmt.Sprintf("%03x.%04x", prev_path_i, prev_step_i+int64(prev_seedlen))
        sfx_tag_id := create_tag_id(prev_sfx_tile_path, sfx_tag)

        if _,ok := g_id_tag[sfx_tag_id] ; !ok {
          g_id_tag[sfx_tag_id] = sfx_tag
        }

        if _,ok := g_tile_lib[prev_tile_path] ; !ok {
          g_tile_lib[prev_tile_path] = make(map[string]TileInfo)
        }

        if _,ok := g_tile_lib[prev_tile_path][prev_md5sum] ; !ok {
          g_tile_lib[prev_tile_path][prev_md5sum] = TileInfo{ prev_md5sum, prev_tile_path, prev_seedlen, 1, -1 }
        } else {
          z := g_tile_lib[prev_tile_path][prev_md5sum]
          z.Freq++
          g_tile_lib[prev_tile_path][prev_md5sum] = z
        }

      } else {
      }

      curseq = curseq[0:0]

      prev_md5sum     = md5sum
      prev_tile_path  = tile_path
      prev_tileid     = tileid
      prev_seedlen    = seedlen
      prev_path_i     = path_i
      prev_step_i     = step_i

      continue

    }

    curseq = append(curseq, l)

  }

  if len(curseq)>0 {
    tile_seq := strings.Join(curseq, "")
    pfx_tag := tile_seq[0:24]
    sfx_tag := tile_seq[len(tile_seq)-24:]

    md5_tile_seq := md5sum_str(tile_seq)
    if md5_tile_seq != prev_md5sum {
      log.Fatal(fmt.Sprintf("previous md5sum %s != current md5sum %s (%s)\n", prev_md5sum, md5_tile_seq, prev_tileid))
    }

    if _,ok := g_md5sum_seq[md5_tile_seq] ; !ok {
      g_md5sum_seq[md5_tile_seq] = tile_seq
    }

    pfx_tag_id := create_tag_id(prev_tile_path, pfx_tag)
    if _,ok := g_id_tag[pfx_tag_id] ; !ok {
      g_id_tag[pfx_tag_id] = pfx_tag
    }

    //sfx_tag_id := create_tag_id(prev_tile_path, sfx_tag)
    prev_sfx_tile_path := fmt.Sprintf("%03x.%04x", prev_path_i, prev_step_i+int64(prev_seedlen))
    sfx_tag_id := create_tag_id(prev_sfx_tile_path, sfx_tag)

    if _,ok := g_id_tag[sfx_tag_id] ; !ok {
      g_id_tag[sfx_tag_id] = sfx_tag
    }

    if _,ok := g_tile_lib[prev_tile_path] ; !ok {
      g_tile_lib[prev_tile_path] = make(map[string]TileInfo)
    }

    if _,ok := g_tile_lib[prev_tile_path][prev_md5sum] ; !ok {
      g_tile_lib[prev_tile_path][prev_md5sum] = TileInfo{ prev_md5sum, prev_tile_path, prev_seedlen, 1, -1 }
    } else {
      z := g_tile_lib[prev_tile_path][prev_md5sum]
      z.Freq++
      g_tile_lib[prev_tile_path][prev_md5sum] = z
    }

  }

  _ = name
  return nil

}

func emit_fasta_sql_csv(ofp *bufio.Writer, fasta_ofn string) {
  l := fmt.Sprintf("%d,%s\n", g_FASTAID, fasta_ofn)
  ofp.Write([]byte(l))
}

func emit_fasta(ofp *bufio.Writer) {
  fold := 50

  seen := make(map[string]bool)

  for path_step := range g_tile_lib {
    for m := range g_tile_lib[path_step] {
      tile_seq := g_md5sum_seq[m]

      pfx_tag := tile_seq[0:24]
      sfx_tag := tile_seq[len(tile_seq)-24:]

      pfx_tag_id := create_tag_id(path_step, pfx_tag)

      //sfx_tag_id := create_tag_id(path_step, sfx_tag)
      seedlen := g_tile_lib[path_step][m].SeedLen
      path_step_parts := strings.Split(path_step, ".")
      path_i,e := strconv.ParseInt(path_step_parts[0], 16, 64)
      if e!=nil { log.Fatal(e) }
      step_i,e := strconv.ParseInt(path_step_parts[1], 16, 64)
      if e!=nil { log.Fatal(e) }
      sfx_path_step := fmt.Sprintf("%03x.%04x", path_i, step_i+int64(seedlen))
      sfx_tag_id := create_tag_id(sfx_path_step, sfx_tag)

      body_md5 := md5sum_str(tile_seq[24:len(tile_seq)-24])
      body_id := fmt.Sprintf("%s.%s.r%x+%0x",
        body_md5,
        path_step,
        g_tile_lib[path_step][m].Rank,
        g_tile_lib[path_step][m].SeedLen)


      if !seen[pfx_tag_id] {
        seen[pfx_tag_id] = true
        l := fmt.Sprintf(">%s\n", pfx_tag_id)
        ofp.Write([]byte(l))
        ofp.Write([]byte(pfx_tag))
        ofp.Write([]byte("\n\n"))
      }

      if !seen[sfx_tag_id] {
        seen[sfx_tag_id] = true
        l := fmt.Sprintf(">%s\n", sfx_tag_id)
        ofp.Write([]byte(l))
        ofp.Write([]byte(sfx_tag))
        ofp.Write([]byte("\n\n"))
      }

      if !seen[body_id] {
        seen[body_id] = true
        l := fmt.Sprintf(">%s\n", body_id)
        ofp.Write([]byte(l))

        body := tile_seq[24:len(tile_seq)-24]
        p:=0
        for p=0; p<(len(body)-fold); p+=fold {

          /*
          //DEBUG
          fmt.Printf(">>> p %d, len(body) %d, fold %d (%d)\n", p, len(body), fold, len(body)-fold)

          l:=fmt.Sprintf(">>>> p %d s%d:e%d (%d-%d=%d)\n", p, fold*p, fold*(p+1), len(body), fold, len(body)-fold)
          ofp.Write([]byte(l))
          ofp.Flush()
          */

          ofp.Write([]byte(body[p:p+fold]))
          ofp.Write([]byte("\n"))
        }
        ofp.Write([]byte(body[p:]))
        ofp.Write([]byte("\n\n"))
      }

    }
  }

}

// Map the character name of the Sequence to it's id
//
var g_sequence_id map[string]int

func emit_sequences(ofp *bufio.Writer) {
  //seq_id  := 1
  //fa_id   := 1

  seq_id  := g_START_SEQUENCEID
  fa_id   := g_FASTAID

  g_sequence_id = make(map[string]int)

  for id_tag := range g_id_tag {
    m5 := md5sum_str(g_id_tag[id_tag])

    l := fmt.Sprintf("%d,%d,%s,%s,%d\n", seq_id, fa_id, id_tag, m5, 24)

    if _,ok := g_sequence_id[id_tag] ; !ok {
      ofp.Write( []byte(l) )
      g_sequence_id[id_tag] = seq_id
      seq_id++
    }
  }

  for path_step := range g_tile_lib {
    for m := range g_tile_lib[path_step] {
      tile_seq := g_md5sum_seq[m]

      if len(tile_seq)<48 {
        log.Fatal(fmt.Sprintf(">>>> path_step:%s, md5sum:%s ???? %s\n", path_step, m, tile_seq))
      }

      body_md5 := md5sum_str(tile_seq[24:len(tile_seq)-24])

      body_id := fmt.Sprintf("%s.%s.r%x+%0x",
        body_md5,
        path_step,
        g_tile_lib[path_step][m].Rank,
        g_tile_lib[path_step][m].SeedLen)
      l := fmt.Sprintf("%d,%d,%s,%s,%d\n", seq_id, fa_id, body_id, body_md5, len(tile_seq)-48)

      if _,ok := g_sequence_id[body_id] ; !ok {
        ofp.Write([]byte(l))
        g_sequence_id[body_id]=seq_id
        seq_id++
      }

    }
  }

}

var g_graphjoin_id_list []int

func emit_graphjoin(ofp *bufio.Writer) {

  g_graphjoin_id_list = make([]int,0,1024)

  //gj_id := 1
  gj_id := g_START_GRAPHJOINID

  seen_hash := make(map[string]bool)

  for path_step := range g_tile_lib {
    for m := range g_tile_lib[path_step] {
      tile_seq := g_md5sum_seq[m]

      pfx_tag := tile_seq[0:24]
      sfx_tag := tile_seq[len(tile_seq)-24:]

      pfx_tag_id := create_tag_id(path_step, pfx_tag)

      //sfx_tag_id := create_tag_id(path_step, sfx_tag)
      seedlen := int64(g_tile_lib[path_step][m].SeedLen)
      path_step_parts := strings.Split(path_step, ".")
      path_i,e := strconv.ParseInt(path_step_parts[0], 16, 64)
      if e!=nil { log.Fatal(e) }
      step_i,e := strconv.ParseInt(path_step_parts[1], 16, 64)
      if e!=nil { log.Fatal(e) }
      sfx_path_step := fmt.Sprintf("%03x.%04x", path_i, step_i+seedlen)
      sfx_tag_id := create_tag_id(sfx_path_step, sfx_tag)


      body_md5 := md5sum_str(tile_seq[24:len(tile_seq)-24])
      body_id := fmt.Sprintf("%s.%s.r%x+%0x",
        body_md5,
        path_step,
        g_tile_lib[path_step][m].Rank,
        g_tile_lib[path_step][m].SeedLen)


      pfx_seq_id := g_sequence_id[pfx_tag_id]
      sfx_seq_id := g_sequence_id[sfx_tag_id]
      body_seq_id := g_sequence_id[body_id]

      key := fmt.Sprintf("%x:%x", pfx_seq_id, body_seq_id)
      if _,seen := seen_hash[key]; !seen {
        if pfx_tag_id < body_id {
          //l:=fmt.Sprintf("%d,%d,%d,'FALSE',%d,%d,'TRUE'\n", gj_id, pfx_seq_id, 23, body_seq_id, len(tile_seq)-49)
          l:=fmt.Sprintf("%d,%d,%d,'FALSE',%d,%d,'TRUE'\n", gj_id, pfx_seq_id, 23, body_seq_id, 0)
          ofp.Write([]byte(l))
        } else {
          //l:=fmt.Sprintf("%d,%d,%d,'TRUE',%d,%d,'FALSE'\n", gj_id, body_seq_id, len(tile_seq)-49, pfx_seq_id, 23)
          l:=fmt.Sprintf("%d,%d,%d,'TRUE',%d,%d,'FALSE'\n", gj_id, body_seq_id, 0, pfx_seq_id, 23)
          ofp.Write([]byte(l))
        }

        g_graphjoin_id_list = append(g_graphjoin_id_list, gj_id)

        gj_id++
        seen_hash[key]=true
      }

      key = fmt.Sprintf("%x:%x", sfx_seq_id, body_seq_id)
      if _,seen := seen_hash[key]; !seen {
        if body_id < sfx_tag_id {
          //l:=fmt.Sprintf("%d,%d,%d,'TRUE',%d,%d,'FALSE'\n", gj_id, body_seq_id, len(tile_seq)-49, sfx_seq_id, 23)
          l:=fmt.Sprintf("%d,%d,%d,'FALSE',%d,%d,'TRUE'\n", gj_id, body_seq_id, len(tile_seq)-49, sfx_seq_id, 0)
          ofp.Write([]byte(l))
        } else {
          //l:=fmt.Sprintf("%d,%d,%d,'FALSE',%d,%d,'TRUE'\n", gj_id, sfx_seq_id, 23, body_seq_id, len(tile_seq)-49)
          l:=fmt.Sprintf("%d,%d,%d,'TRUE',%d,%d,'FALSE'\n", gj_id, sfx_seq_id, 0, body_seq_id, len(tile_seq)-49)
          ofp.Write([]byte(l))
        }

        g_graphjoin_id_list = append(g_graphjoin_id_list, gj_id)

        gj_id++
        seen_hash[key]=true
      }

    }
  }

}

func emit_graphjoin_variantset(ofp *bufio.Writer) {
  for i:=0; i<len(g_graphjoin_id_list); i++ {
    l := fmt.Sprintf("%d,%d\n", g_graphjoin_id_list[i], g_VARIANTSETID)
    ofp.Write([]byte(l))
  }
}

var path_step_order []string

type LexOrder []string
func (s LexOrder) Len() int { return len(s) }
func (s LexOrder) Swap(i,j int) { s[i],s[j] = s[j],s[i] }
func (s LexOrder) Less(i,j int) bool { return s[i] < s[j] }

type TileFreqOrder []TileInfo
func (t TileFreqOrder) Len() int { return len(t) }
func (t TileFreqOrder) Swap(i,j int) { t[i],t[j] = t[j],t[i] }
func (t TileFreqOrder) Less(i,j int) bool {
  if t[i].Freq < t[j].Freq { return false }
  if t[i].Freq > t[j].Freq { return true }
  return t[i].Md5Sum < t[j].Md5Sum
}

func rank_tile_lib() {
  path_step_order = make([]string, 0, len(g_tile_lib))
  for path_step := range g_tile_lib {
    path_step_order = append(path_step_order, path_step)

    freq_order := make([]TileInfo, 0, len(g_tile_lib[path_step]))
    for m5 := range g_tile_lib[path_step] {
      freq_order = append(freq_order, g_tile_lib[path_step][m5])
    }
    sort.Sort(TileFreqOrder(freq_order))


    for i:=0; i<len(freq_order); i++ {
      z := g_tile_lib[ freq_order[i].PathStep ][ freq_order[i].Md5Sum ]
      z.Rank = i
      g_tile_lib[ freq_order[i].PathStep ][ freq_order[i].Md5Sum ] = z
    }

  }

  sort.Sort(LexOrder(path_step_order))

}

func _main( c *cli.Context ) {

  g_START_SEQUENCEID = c.Int("start-sequence-id")
  g_START_GRAPHJOINID = c.Int("start-graphjoin-id")
  g_FASTAID = c.Int("fasta-id")
  g_VARIANTSETID = c.Int("variantset-id")

  if c.Bool( "pprof" ) {
    gProfileFlag = true
    gProfileFile = c.String("pprof-file")
  }

  if c.Bool( "mprof" ) {
    gMemProfileFlag = true
    gMemProfileFile = c.String("mprof-file")
  }

  g_show_progress = c.Bool("progress")
  gVerboseFlag = c.Bool("Verbose")

  if c.Int("max-procs") > 0 {
    runtime.GOMAXPROCS( c.Int("max-procs") )
  }

  if gProfileFlag {
    prof_f,err := os.Create( gProfileFile )
    if err != nil {
      fmt.Fprintf( os.Stderr, "Could not open profile file %s: %v\n", gProfileFile, err )
      os.Exit(2)
    }

    pprof.StartCPUProfile( prof_f )
    defer pprof.StopCPUProfile()
  }


  fasta_ofn     := c.String("fasta")
  sequence_ofn  := c.String("sequence")
  graphjoin_ofn := c.String("graphjoin")
  fasta_csv_ofn := c.String("fasta-csv")
  graphjoin_variantset_ofn := c.String("graphjoin-variantset")

  fasta_out,err := autoio.CreateWriter( fasta_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err) ; os.Exit(1) }
  defer func() { fasta_out.Flush() ; fasta_out.Close() }()

  fasta_csv_out,err := autoio.CreateWriter( fasta_csv_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err) ; os.Exit(1) }
  defer func() { fasta_csv_out.Flush() ; fasta_csv_out.Close() }()

  seq_out,err := autoio.CreateWriter( sequence_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err) ; os.Exit(1) }
  defer func() { seq_out.Flush() ; seq_out.Close() }()

  gj_out,err := autoio.CreateWriter( graphjoin_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err) ; os.Exit(1) }
  defer func() { gj_out.Flush() ; gj_out.Close() }()

  vs_gj_out,err := autoio.CreateWriter( graphjoin_variantset_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err) ; os.Exit(1) }
  defer func() { vs_gj_out.Flush() ; vs_gj_out.Close() }()


  // Process input FastJ files

  ifns := c.StringSlice("input")
  for i:=0; i<len(ifns); i++ {

    if g_show_progress {
      fmt.Fprintf(os.Stderr, ">>> %s\n", ifns[i])
    }

    name := ifns[i]
    ifn := ifns[i]
    if strings.Contains(ifns[i], ",") {
      z := strings.SplitN(ifns[i], ",", 2)
      name = z[0]
      ifn = z[1]
    }

    e := import_fastj(name, ifn)
    if e!=nil { log.Fatal(e) }
  }


  // Once the library has been created, rank
  // the resulting tiles.
  //
  rank_tile_lib()


  // Output a big FASTA file with all of our
  // sequence information.
  //
  emit_fasta(fasta_out.Writer)

  // And our FASTA.csv file
  //
  emit_fasta_sql_csv(fasta_csv_out.Writer, fasta_ofn)

  // Now output the SQL Sequences.
  //
  emit_sequences(seq_out.Writer)


  // Now emit the SQL GraphJoin
  //
  emit_graphjoin(gj_out.Writer)

  emit_graphjoin_variantset(vs_gj_out.Writer)

}

func main() {

  app := cli.NewApp()
  app.Name  = "create_tile_graph"
  app.Usage = "Create the FASTA sequences and CSV of for upload into a SQL Sequence and GraphJoin table"
  app.Version = VERSION_STR
  app.Author = "Curoverse, Inc."
  app.Email = "info@curoverse.com"
  app.Action = func( c *cli.Context ) { _main(c) }

  app.Flags = []cli.Flag{
    cli.StringSliceFlag{
      Name: "input, i",
      Value: &cli.StringSlice{},
      Usage: "INPUT (can be specified more than once for multiple input files)",
    },

    cli.StringFlag{
      Name: "fasta",
      Value: "out.fa",
      Usage: "FASTA OUTPUT",
    },

    cli.StringFlag{
      Name: "fasta-csv",
      Value: "FASTA.csv",
      Usage: "FASTA SQL CSV OUTPUT",
    },

    cli.StringFlag{
      Name: "sequence",
      Value: "out.sequence",
      Usage: "Sequence OUTPUT",
    },

    cli.StringFlag{
      Name: "graphjoin",
      Value: "out.graphjoin",
      Usage: "GraphJoin OUTPUT",
    },

    cli.StringFlag{
      Name: "graphjoin-variantset",
      Value: "out.graphjoin-variantset",
      Usage: "GraphJoin_VariantSet_Join OUTPUT",
    },

    cli.IntFlag{
      Name: "max-procs, N",
      Value: -1,
      Usage: "MAXPROCS",
    },

    cli.IntFlag{
      Name: "fasta-id",
      Value: 1,
      Usage: "ID of FASTA SQL row",
    },

    cli.IntFlag{
      Name: "start-sequence-id",
      Value: 1,
      Usage: "Start ID of Sequence SQL row",
    },

    cli.IntFlag{
      Name: "start-graphjoin-id",
      Value: 1,
      Usage: "Start ID of GraphJoin SQL row",
    },

    cli.IntFlag{
      Name: "variantset-id",
      Value: 0,
      Usage: "ID of VariantSet SQL row",
    },

    cli.BoolFlag{
      Name: "Verbose, V",
      Usage: "Verbose flag",
    },

    cli.BoolFlag{
      Name: "progress",
      Usage: "Verbose flag",
    },

    cli.BoolFlag{
      Name: "pprof",
      Usage: "Profile usage",
    },

    cli.StringFlag{
      Name: "pprof-file",
      Value: gProfileFile,
      Usage: "Profile File",
    },

    cli.BoolFlag{
      Name: "mprof",
      Usage: "Profile memory usage",
    },

    cli.StringFlag{
      Name: "mprof-file",
      Value: gMemProfileFile,
      Usage: "Profile Memory File",
    },

  }

  app.Run( os.Args )

  if gMemProfileFlag {
    fmem,err := os.Create( gMemProfileFile )
    if err!=nil { panic(fmem) }
    pprof.WriteHeapProfile(fmem)
    fmem.Close()
  }

}
