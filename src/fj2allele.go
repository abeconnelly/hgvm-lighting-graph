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

// Create CSV for Allele and AllelePathItem
//
// Output will be:
//  - out.allele
//  - out.allelepath
//  - out.callset
//
// The above are the default names.  They can be overidden.
//
// Where:
//  out.allele is a comma separated list of Allele rows
//  out.allelepath is a comma separated list of AllelePathItem rows
//  out.callset is a comma separated list of CallSet rows
//
// example usage (a.fj and b.fj are input FastJ files):
//
// ./fj2allele -i a.fj -sequence in.seq -allele out.allele -allele-path out.allelepath -callset out.callset
//
// See:
//  https://github.com/ga4gh/server/blob/graph/tests/data/graphs/graphSQL_v023.sql
//  https://github.com/ga4gh/server/blob/graph/tests/data/graphs/graphData_v023.sql
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

import "crypto/md5"

import "github.com/abeconnelly/autoio"
import "github.com/abeconnelly/sloppyjson"
import "github.com/codegangsta/cli"

var VERSION_STR string = "0.1.0"
var gVerboseFlag bool

var gProfileFlag bool
var gProfileFile string = "fj2allele.pprof"

var gMemProfileFlag bool
var gMemProfileFile string = "fj2allele.mprof"

// Map of md5sum of sequence to it's corresponding SequenceID, read
// in from the Sequence file.
//
var g_md5_seqid_map map[string]int64

// Each map entry is the allele of the input sequence (presumably two
// entries).  The value is an array of SequenceIDs.
//
var g_allele_sequenceid_path map[string][]int64

// Map CallSet.name to sampleID
//
var g_callsetname_to_sampleid map[string]string
var g_callsetname_to_id map[string]int64


type CallSet struct {
  Id int
  Name string
  SampleId string
}

type Allele struct {
  Id int
  VariantSetId int
  Name string
  CurPathIdx int
}

type AllelePathItem struct {
  AlleleId int
  PathItemIndex int
  SequenceId int
  Start int
  Length int
  StrandIsForward string
}

type AlleleCall struct {
  AlleleId int
  CallSetId int
  Ploidy int
}

var g_ALLELE_ID int
var g_START_CALLSET_ID int

// named sample as key
//
var g_callset map[string]CallSet

// named sample colon allele as key
// e.g. hu826751:0
//
var g_allele map[string]Allele

// named sample colon allele as key
// e.g. hu826751:1
//
var g_allele_path_item map[string][]AllelePathItem

// named sample colon allele as key
// e.g. hu826751:0
//
var g_allele_call map[string]AlleleCall

func md5sum_str(seq string) string {
  ta := make([]string, 0, 32)
  s := md5.Sum([]byte(seq))
  for i:=0; i<len(s); i++ {
    ta = append(ta, fmt.Sprintf("%02x", s[i]))
  }
  return strings.Join(ta, "")
}


func init() {
  g_md5_seqid_map = make(map[string]int64)
  g_allele_sequenceid_path = make(map[string][]int64)
  g_callsetname_to_sampleid = make(map[string]string)
  g_callsetname_to_id = make(map[string]int64)

  g_callset           = make(map[string]CallSet)
  g_allele            = make(map[string]Allele)
  g_allele_path_item  = make(map[string][]AllelePathItem)
  g_allele_call       = make(map[string]AlleleCall)
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
  var prev_tile_path string ; _ = prev_tile_path
  var prev_tileid string ; _ = prev_tileid
  var prev_seedlen int ; _ = prev_seedlen

  var prev_path_i int64 ; _ = prev_path_i
  var prev_step_i int64 ; _ = prev_step_i

  var prev_tile_allele int64 ; _ = prev_tile_allele
  var prev_allele_name_id string

  curseq := make([]string, 0, 10)

  h,e := autoio.OpenReadScannerSimple(fn)
  if e!=nil { return e }

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

      tile_allele,e := strconv.ParseInt(tile_parts[3], 16, 64)
      if e!=nil { return e }

      allele_name_id := fmt.Sprintf("%s:%d", name, tile_allele)

      // Initialize everything if we haven't seen it before
      //
      if _,ok := g_allele[allele_name_id] ; !ok {

        callset_id := g_callset[name].Id
        ploidy := 1
        variant_set_id := -1

        g_allele[allele_name_id] = Allele{ g_ALLELE_ID, variant_set_id, allele_name_id, 0 }
        g_allele_path_item[allele_name_id] = make([]AllelePathItem, 0, 1024)
        g_allele_call[allele_name_id] = AlleleCall{ g_ALLELE_ID, callset_id, ploidy  }
        g_ALLELE_ID++
      }

      if len(curseq)>0 {

        tile_seq := strings.Join(curseq, "")
        pfx_tag := tile_seq[0:24]
        sfx_tag := tile_seq[len(tile_seq)-24:]
        body_seq := tile_seq[24:len(tile_seq)-24]

        md5_tile_seq := md5sum_str(tile_seq)
        if md5_tile_seq != prev_md5sum {
          log.Fatal(fmt.Sprintf("previous md5sum %s (%s) != current md5sum %s (%s)\n", prev_md5sum, tileid, md5_tile_seq, prev_tileid))
        }

        var ok bool
        var seqid int64

        //allele_name := fmt.Sprintf("%d", prev_tile_allele)
        allele_id := g_allele[prev_allele_name_id].Id
        allele_path := g_allele_path_item[prev_allele_name_id]
        cur_idx := len(allele_path)


        pfx_md5 := md5sum_str(pfx_tag)
        if seqid,ok = g_md5_seqid_map[pfx_md5] ; !ok {
          log.Fatal(fmt.Sprintf("ERROR: could not find tag '%s' (%s) in Sequence map", pfx_tag, pfx_md5))
        }

        //fmt.Printf("t,%s,%d,%d\n", pfx_md5, seqid, prev_tile_allele)

        // Only add the prefix tag if it's the first one in the AllelePathItem
        //
        //if _,ok := g_allele_sequenceid_path[allele_name] ; !ok {
        //  g_allele_sequenceid_path[allele_name] = append(g_allele_sequenceid_path[allele_name], seqid)
        //}

        if cur_idx==0 {
          allele_path = append(allele_path, AllelePathItem{allele_id, cur_idx, int(seqid), 0, -1, "'TRUE'"})
          cur_idx++
        }

        body_md5 := md5sum_str(body_seq)
        if seqid,ok = g_md5_seqid_map[body_md5] ; !ok {
          log.Fatal(fmt.Sprintf("ERROR: could not find body (%s) in Sequence map", body_md5))
        }

        //fmt.Printf("b,%s,%d,%d\n", body_md5, seqid, prev_tile_allele)
        //g_allele_sequenceid_path[allele_name] = append(g_allele_sequenceid_path[allele_name], seqid)
        allele_path = append(allele_path, AllelePathItem{allele_id, cur_idx, int(seqid), 0, -1, "'TRUE'"})
        cur_idx++

        sfx_md5 := md5sum_str(sfx_tag)
        if seqid,ok = g_md5_seqid_map[sfx_md5] ; !ok {
          log.Fatal(fmt.Sprintf("ERROR: could not find tag '%s' (%s) in Sequence map", sfx_tag, sfx_md5))
        }

        //fmt.Printf("t,%s,%d,%d\n", sfx_md5, seqid, prev_tile_allele)
        //g_allele_sequenceid_path[allele_name] = append(g_allele_sequenceid_path[allele_name], seqid)
        allele_path = append(allele_path, AllelePathItem{allele_id, cur_idx, int(seqid), 0, -1, "'TRUE'"})
        cur_idx++

        g_allele_path_item[prev_allele_name_id] = allele_path

      } else {
      }

      curseq = curseq[0:0]

      prev_md5sum     = md5sum
      prev_tile_path  = tile_path
      prev_tileid     = tileid
      prev_seedlen    = seedlen
      prev_path_i     = path_i
      prev_step_i     = step_i
      prev_tile_allele = tile_allele
      prev_allele_name_id = allele_name_id

      continue

    }

    curseq = append(curseq, l)

  }

  if len(curseq)>0 {
    tile_seq := strings.Join(curseq, "")
    pfx_tag := tile_seq[0:24]
    sfx_tag := tile_seq[len(tile_seq)-24:]
    body_seq := tile_seq[24:len(tile_seq)-24]

    md5_tile_seq := md5sum_str(tile_seq)
    if md5_tile_seq != prev_md5sum {
      log.Fatal(fmt.Sprintf("previous md5sum %s != current md5sum %s (%s)\n", prev_md5sum, md5_tile_seq, prev_tileid))
    }

    var ok bool
    var seqid int64
    //allele_name := fmt.Sprintf("%d", prev_tile_allele)

    pfx_md5 := md5sum_str(pfx_tag)
    if seqid,ok = g_md5_seqid_map[pfx_md5] ; !ok {
      log.Fatal(fmt.Sprintf("ERROR: could not find tag '%s' (%s) in Sequence map", pfx_tag, pfx_md5))
    }


    // Only add the prefix tag if it's the first one in the AllelePathItem
    //
    allele_id := g_allele[prev_allele_name_id].Id
    allele_path := g_allele_path_item[prev_allele_name_id]
    cur_idx := len(allele_path)

    if cur_idx==0 {
      allele_path = append(allele_path, AllelePathItem{allele_id, cur_idx, int(seqid), 0, -1, "'TRUE'"} )
      cur_idx++
    }


    body_md5 := md5sum_str(body_seq)
    if seqid,ok = g_md5_seqid_map[body_md5] ; !ok {
      log.Fatal(fmt.Sprintf("ERROR: could not find body (%s) in Sequence map", body_md5))
    }

    //fmt.Printf("b,%s,%d,%d\n", body_md5, seqid, prev_tile_allele)
    //g_allele_sequenceid_path[allele_name] = append(g_allele_sequenceid_path[allele_name], seqid)
    allele_path = append(allele_path, AllelePathItem{allele_id, cur_idx, int(seqid), 0, -1, "'TRUE'"} )
    cur_idx++



    sfx_md5 := md5sum_str(sfx_tag)
    if seqid,ok = g_md5_seqid_map[sfx_md5] ; !ok {
      log.Fatal(fmt.Sprintf("ERROR: could not find tag '%s' (%s) in Sequence map", sfx_tag, sfx_md5))
    }

    //fmt.Printf("t,%s,%d,%d\n", sfx_md5, seqid, prev_tile_allele)
    //g_allele_sequenceid_path[allele_name] = append(g_allele_sequenceid_path[allele_name], seqid)
    allele_path = append(allele_path, AllelePathItem{allele_id, cur_idx, int(seqid), 0, -1, "'TRUE'"} )
    cur_idx++

    g_allele_path_item[prev_allele_name_id] = allele_path

  }

  _ = name
  return nil

}

// Parse Sequence CSV file.
// Assumes no header.
// Assumes Sequence element order is:
//
//  0    1           2                3        4
// ID,fastaID,sequenceRecordName,md5checksum,length
//
func import_sequence(fn string) error {
  h,e := autoio.OpenReadScannerSimple(fn)
  if e!=nil { return e }
  defer h.Close()

  line_no:=-1

  for h.ReadScan() {
    line_no++
    l := h.ReadText()
    if len(l)==0 { continue }

    line_parts := strings.Split(l, ",")

    id,e := strconv.ParseInt(line_parts[0], 10, 64)
    if e!=nil { return fmt.Errorf("ERROR: parsing ID in Sequence file (line %d): %s", line_no, line_parts[0]) }

    fastaid := line_parts[1] ; _ = fastaid
    seqname := line_parts[2] ; _ = seqname
    m5      := line_parts[3] ; _ = m5

    seqlen,e := strconv.ParseInt(line_parts[4], 10, 64) ; _ = seqlen
    if e!=nil { return fmt.Errorf("ERROR: parsing seqlen in Sequence file (line %d): %s", line_no, line_parts[4]) }

    g_md5_seqid_map[m5] = id
  }

  return nil

}

var g_allele_name_id_map map[string]int

func emit_allele_call(ofp *bufio.Writer) {
  for k := range g_allele_call {
    s := fmt.Sprintf("%d,%d,%d\n",
      g_allele_call[k].AlleleId,
      g_allele_call[k].CallSetId,
      g_allele_call[k].Ploidy)
    ofp.Write([]byte(s))
  }
}

func emit_allele(ofp *bufio.Writer) {

  for allele_key := range g_allele {
    s := fmt.Sprintf("%d,%d,%s\n", g_allele[allele_key].Id, g_allele[allele_key].VariantSetId, g_allele[allele_key].Name)
    ofp.Write([]byte(s))
  }

  return

  g_allele_name_id_map = make(map[string]int)

  g_allele_name_id_map["0"] = 0
  g_allele_name_id_map["1"] = 1

  for allele_name := range g_allele_name_id_map {
    variant_id := -1
    allele_id := g_allele_name_id_map[allele_name]
    s := fmt.Sprintf("%d,%d,%s\n", allele_id, variant_id, allele_name)
    ofp.Write([]byte(s))
  }

}

func emit_allele_path_item(ofp *bufio.Writer) {

  for k := range g_allele_path_item {

    for i:=0; i<len(g_allele_path_item[k]); i++ {
      s := fmt.Sprintf("%d,%d,%d,%d,%d,%s\n",
        g_allele_path_item[k][i].AlleleId,
        g_allele_path_item[k][i].PathItemIndex,
        g_allele_path_item[k][i].SequenceId,
        g_allele_path_item[k][i].Start,
        g_allele_path_item[k][i].Length,
        g_allele_path_item[k][i].StrandIsForward)
      ofp.Write([]byte(s))
    }
  }

  return

  for allele_name := range g_allele_sequenceid_path {

    allele_id := g_allele_name_id_map[allele_name]

    for i:=0; i<len(g_allele_sequenceid_path[allele_name]); i++ {
      seqid := g_allele_sequenceid_path[allele_name][i]
      s := fmt.Sprintf("%d,%d,%d,%s,%s,%s\n", allele_id, i, seqid, "NULL", "NULL", "'TRUE'")
      ofp.Write([]byte(s))
    }

  }
}


func emit_callset(ofp *bufio.Writer) {
  for cs_id := range g_callset {
    s:=fmt.Sprintf("%d,%s,%s\n", g_callset[cs_id].Id, g_callset[cs_id].Name, g_callset[cs_id].SampleId)
    ofp.Write([]byte(s))
  }
}


func _main( c *cli.Context ) {
  sequence_ifn  := c.String("sequence")
  if len(sequence_ifn)==0 { cli.ShowAppHelp(c) }

  g_ALLELE_ID = c.Int("start-allele-id")
  g_START_CALLSET_ID = c.Int("start-callset-id")

  ifns := c.StringSlice("input")
  if len(ifns)==0 { cli.ShowAppHelp(c) }

  if c.Bool( "pprof" ) {
    gProfileFlag = true
    gProfileFile = c.String("pprof-file")
  }

  if c.Bool( "mprof" ) {
    gMemProfileFlag = true
    gMemProfileFile = c.String("mprof-file")
  }

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

  allele_ofn            := c.String("allele")
  allele_path_item_ofn  := c.String("allele-path")
  callset_ofn           := c.String("callset")
  allele_call_ofn       := c.String("allele-call")

  allele_out,err := autoio.CreateWriter( allele_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err); os.Exit(1) }
  defer func() { allele_out.Flush(); allele_out.Close() }()

  allele_path_item_out,err := autoio.CreateWriter( allele_path_item_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err); os.Exit(1) }
  defer func() { allele_path_item_out.Flush(); allele_path_item_out.Close() }()

  callset_out,err := autoio.CreateWriter( callset_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err); os.Exit(1) }
  defer func() { callset_out.Flush(); callset_out.Close() }()

  allele_call_out,err := autoio.CreateWriter( allele_call_ofn )
  if err!=nil { fmt.Fprintf(os.Stderr, "%v", err); os.Exit(1) }
  defer func() { allele_call_out.Flush(); allele_call_out.Close() }()

  show_progress_flag := c.Bool("progress")


  // Process Sequence CSV file
  //
  import_sequence(sequence_ifn)


  // First populate the callset maps
  //
  for i:=0; i<len(ifns); i++ {
    name := ifns[i]
    if strings.Contains(ifns[i], ",") {
      z := strings.SplitN(ifns[i], ",", 2)
      name = z[0]
    }
    g_callset[name] = CallSet{ g_START_CALLSET_ID + i, name, name }
  }

  // Process input FastJ files
  //
  for i:=0; i<len(ifns); i++ {


    name := ifns[i]
    ifn := ifns[i]
    if strings.Contains(ifns[i], ",") {
      z := strings.SplitN(ifns[i], ",", 2)
      name = z[0]
      ifn = z[1]
    }

    if show_progress_flag { fmt.Printf(">>>> %s %s\n", name, ifn) }

    e := import_fastj(name, ifn)
    if e!=nil { log.Fatal(e) }
  }

  // Output the CallSet
  //
  emit_callset(callset_out.Writer)

  // Output Allele information
  //
  emit_allele(allele_out.Writer)

  //
  emit_allele_call(allele_call_out.Writer)

  // Output actual path
  //
  emit_allele_path_item(allele_path_item_out.Writer)

}

func main() {

  app := cli.NewApp()
  app.Name  = "fj2allele"
  app.Usage = "Create the CSV files for Allele and AllelePathItem from a FastJ and Sequence CSV file"
  app.Version = VERSION_STR
  app.Author = "Curoverse, Inc."
  app.Email = "info@curoverse.com"
  app.Action = func( c *cli.Context ) { _main(c) }

  app.Flags = []cli.Flag{
    cli.StringSliceFlag{
      Name: "input, i",
      Value: &cli.StringSlice{},
      Usage: "INPUT",
    },

    cli.BoolFlag{
      Name: "progress",
      Usage: "show progress",
    },

    cli.StringFlag{
      Name: "sequence",
      Value: "out.sequence",
      Usage: "Sequence OUTPUT",
    },

    cli.StringFlag{
      Name: "callset",
      Value: "out.callset",
      Usage: "CallSet CSV OUTPUT",
    },

    cli.StringFlag{
      Name: "allele",
      Value: "out.allele",
      Usage: "Allele CSV OUTPUT",
    },

    cli.StringFlag{
      Name: "allele-path",
      Value: "out.allelepath",
      Usage: "AllelePathItem CSV OUTPUT",
    },

    cli.StringFlag{
      Name: "allele-call",
      Value: "out.allelecall",
      Usage: "AlleleCall CSV OUTPUT",
    },

    cli.IntFlag{
      Name: "start-allele-id",
      Value: 1,
      Usage: "Start Allele ID",
    },

    cli.IntFlag{
      Name: "start-callset-id",
      Value: 1,
      Usage: "Start CallSet ID",
    },

    cli.IntFlag{
      Name: "max-procs, N",
      Value: -1,
      Usage: "MAXPROCS",
    },

    cli.BoolFlag{
      Name: "Verbose, V",
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
