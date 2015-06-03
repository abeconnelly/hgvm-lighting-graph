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

// Take in an ordered tileset represented as a CSV of:
//
//   path.step,tag_sequence
//
// along with a sequence and create a FastJ file.
//
// example usage:
//  ./tileset2fj -i 'actcat....gcat' -t mytileset.csv --build-prefix 'hg19 chr17' -o out.fj
//

package main


import "fmt"
import "os"
import "runtime"
import "runtime/pprof"

import "strings"
import _ "bufio"

import "github.com/abeconnelly/autoio"
import "github.com/codegangsta/cli"

import "sort"
import "crypto/md5"

var VERSION_STR string = "0.1.0"
var gVerboseFlag bool

var gProfileFlag bool
var gProfileFile string = "tileset2fj.pprof"

var gMemProfileFlag bool
var gMemProfileFile string = "tileset2fj.mprof"

var g_tagset map[string]string
var g_tagseq map[string]string

var g_seq string

var g_build_prefix string
var g_seq_start int

func init() {
  g_build_prefix = "unknown"
  g_tagset = make(map[string]string)
  g_tagseq = make(map[string]string)
}

func md5sum_str(seq string) string {
  ta := make([]string, 0, 32)
  s := md5.Sum([]byte(seq))
  for i:=0; i<len(s); i++ {
    ta = append(ta, fmt.Sprintf("%02x", s[i]))
  }
  return strings.Join(ta, "")
}


func load_tagset(h autoio.AutoioHandle) error {
  line_no:=0

  for h.ReadScan() {
    line_no++
    l := h.ReadText()
    if len(l)==0 { continue }
    fields := strings.Split(l, ",")
    if len(fields) != 2 { return fmt.Errorf("bad read on line %s", line_no) }

    g_tagset[fields[0]] = fields[1]
    g_tagseq[fields[1]] = fields[0]
  }

  return nil

}

func load_seq(h autoio.AutoioHandle) {
  //fold := 50

  for h.ReadScan() {
    l := h.ReadText()
    if len(l) == 0 { continue }
    g_seq = l
  }

  g_seq = strings.ToLower(g_seq)

}

func find_tag_positions() {
  //for tilepos := range tagpos_ind { fmt.Printf("%s >> %d\n", tilepos, tagpos_ind[tilepos]) }
}

func print_fold(seq string, fold int) {
  if len(seq)==0 { return }

  p:=0
  for ; p<(len(seq)-fold); p+=fold {
    fmt.Printf("%s\n", seq[p:p+fold])
  }
  fmt.Printf("%s\n", seq[p:])

}

type LexOrder []string
func (s LexOrder) Len() int { return len(s) }
func (s LexOrder) Swap(i,j int) { s[i],s[j] = s[j],s[i] }
func (s LexOrder) Less(i,j int) bool { return s[i] < s[j] }


func gen_tiling() {
  tagpos_ind := make(map[string]int)

  for tilepos := range g_tagset {
    seq := g_tagset[tilepos]
    tagpos_ind[tilepos] = strings.Index(g_seq, seq)
  }

  tagpos := make([]string, 0, len(g_tagset))
  for tilepos := range g_tagset {
    tagpos = append(tagpos, tilepos)
  }
  sort.Sort(LexOrder(tagpos))

  for i:=0; i<len(tagpos); i++ {
    if (tagpos_ind[tagpos[i]]<0) { continue }
    n:=1
    for ; (i+n)<len(tagpos); n++ {
      if tagpos_ind[tagpos[i+n]]>=0 { break }
    }

    if (i+n)==len(tagpos) { continue }

    sp := tagpos_ind[tagpos[i]]
    ep := tagpos_ind[tagpos[i+n]] + 24

    seq := g_seq[sp:ep]

    tagpos_parts := strings.Split(tagpos[i], ".")
    tileid := fmt.Sprintf("%03s.00.%04s.000", tagpos_parts[0], tagpos_parts[1])
    m5 := md5sum_str(seq)

    s := sp + g_seq_start
    e := ep + g_seq_start - 1

    json_str := fmt.Sprintf(`{ "tileID" : "%s", "md5sum":"%s", "locus":[{"build":"%s %d %d"}], "n":%d, "seedTileLength":%d, ` +
      `"startTile":false, "endTile":false, "startSeq":"%s", "endSeq":"%s", ` +
      `"startTag":"%s", "endTag":"%s", "nocallCount":%d, "notes":[] }`,
      tileid, m5, g_build_prefix, s, e, len(seq), n,
      seq[0:24], seq[len(seq)-24:], seq[0:24], seq[len(seq)-24:], 0)


    fmt.Printf("> %s\n", json_str)
    print_fold(seq, 50)
    fmt.Printf("\n")

  }

}

func _main( c *cli.Context ) {

  g_build_prefix = c.String("build-prefix")
  g_seq_start = c.Int("seq-start")

  if c.String("input") == "" {
    fmt.Fprintf( os.Stderr, "Input required, exiting\n" )
    cli.ShowAppHelp( c )
    os.Exit(1)
  }

  seq_fp,err := autoio.OpenReadScannerSimple( c.String("input") ) ; _ = seq_fp
  if err!=nil {
    fmt.Fprintf(os.Stderr, "%v", err)
    os.Exit(1)
  }
  defer seq_fp.Close()

  tileset_fp,err := autoio.OpenReadScannerSimple( c.String("tileset") ) ; _ = tileset_fp
  if err!=nil {
    fmt.Fprintf(os.Stderr, "%v", err)
    os.Exit(1)
  }
  defer tileset_fp.Close()

  load_tagset(tileset_fp)
  load_seq(seq_fp)

  find_tag_positions()
  gen_tiling()


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

}

func main() {

  app := cli.NewApp()
  app.Name  = "tileset2fj"
  app.Usage = "Convert a sequence to FastJ from a tileset"
  app.Version = VERSION_STR
  app.Author = "Curoverse, Inc."
  app.Email = "info@curoverse.com"
  app.Action = func( c *cli.Context ) { _main(c) }

  app.Flags = []cli.Flag{
    cli.StringFlag{
      Name: "input, i",
      Usage: "INPUT sequence",
    },

    cli.StringFlag{
      Name: "tileset, t",
      Usage: "TileSet as a CSV path.step,tag_sequence",
    },

    cli.StringFlag{
      Name: "build-prefix",
      Usage: "Prefix to put in build note (e.g. 'grch38 chr17')",
    },

    cli.IntFlag{
      Name: "seq-start",
      Usage: "Offset of sequence, used for build-prefix calculations. 0 reference.",
    },

    cli.StringFlag{
      Name: "output, o",
      Value: "-",
      Usage: "OUTPUT",
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
