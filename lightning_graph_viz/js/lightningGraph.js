/*

    Copyright (C) 2013 Abram Connelly

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    this program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/


function lightningGraph() {
  this.font_size = 15;
  this.font_color = "rgba(128,128,128,0.5)";
  this.font_angle = 0;
  this.font_offset_h = "L";
  this.font_offset_v = "C";

  this.line_width = 20;
  this.line_color = "rgba(128,128,128,0.2)";

  //this.tag_color = "rgba(128,0,0,0.5)";
  this.tag_color = "rgba(185, 114, 229, 0.5)";
  this.body_color = "rgba(0,0,128,0.5)";

  this.path_radius = 50;

  this.min_step = -1;
  this.max_step = -1;
  this.component = {};

  this.component_pos = {};
  this.graphjoin_line = [];

  this.sequence_zoom_detail = 0.13;

  this.allele_path = {};
  this.sample_allele = {};
  this.allele_to_sample_map = {};

  this.highlight_sample_flag = false;
  this.highlight_sample_name = "hu_test";

  this.call_set = {};
  this.call_set_id = {};

  this.path = "none";
  this.shift_step = 0;
}


lightningGraph.prototype.load_sample_path = function(sample_name) {

  if (!(sample_name in this.call_set)) {
    console.log("ERROR: could not fine sample:", sample_name);
    return;
  }
  var callid = this.call_set[sample_name].id;

  var cmd = "select alleleID from AlleleCall where callSetID = " + callid;
  var z = g_db.exec(cmd);
  if (z.length==0) { console.log("ERROR: did not get any alleleIDs for callSetID", callid); return; }

  this.sample_allele[sample_name] = {};

  var vals = z[0].values;
  for (var ind=0; ind<vals.length; ind++) {
    var allele_id = vals[ind][0];

    z = g_db.exec("select name from Allele where ID = " + allele_id);
    if (z.length==0) { consoole.log("ERROR: no name for allele", allele_id); return; }

    var allele_name = z[0].values[0];


    z = g_db.exec("select pathItemIndex, sequenceID from AllelePathItem where alleleID = " + allele_id + " order by pathItemIndex");
    if (z.length==0) { consoole.log("ERROR: no results in AllelePathItem for", allele_id,allele_name); return; }

    this.sample_allele[sample_name][allele_name] = [];

    var p = z[0].values;
    for (var p_i=0; p_i<p.length; p_i++) {
      this.sample_allele[sample_name][allele_name].push( p[p_i][1] );
    }

  }

  return "ok";
}



lightningGraph.prototype.highlight_sample = function(sample_name) {
  var e = this.load_sample_path(sample_name);
  if (typeof(e) === "undefined") { return; }

  this.highlight_sample_flag = true;
  this.highlight_sample_name = sample_name;
}

lightningGraph.prototype.unhighlight_sample = function() {
  this.highlight_sample_flag = false;
}

lightningGraph.prototype.init_allele_path = function() {

  // Load all names
  //
  var z = g_db.exec("select ID, name from CallSet");
  var vals = z[0].values;
  for (var ind=0; ind<vals.length; ind++) {
    this.call_set_id[ vals[ind][0] ] = vals[ind][1];
    this.call_set[ vals[ind][1] ] = { "id" : vals[ind][0] };
  }

}

lightningGraph.prototype.init_graphjoin = function() {
  var z = g_db.exec("select ID, side1SequenceID, side1Position, side1StrandIsForward," +
                               "side2SequenceID, side2Position, side2StrandIsForward from GraphJoin");
  if ((z.length!=1) || (!("values" in z[0]))) {
    console.log("ERROR, didn't get expected response from DB query");
    return;
  }

  var vals = z[0].values;
  for (var ind=0; ind<vals.length; ind++) {
    var seq0_id = vals[ind][1];
    var seq1_id = vals[ind][4];

    if (!(seq0_id in this.component_pos)) {
      console.log("ERROR:, seq0_id:", seq0_id, " not in component_pos, skipping");
      continue;
    }

    if (!(seq1_id in this.component_pos)) {
      console.log("ERROR:, seq1_id:", seq1_id, " not in component_pos, skipping");
      continue;
    }

    // component_pos
    // 0 1  2      3     4    5    6     7      8
    // x y width height type path step seq_id seq_name

    var p0 = this.component_pos[seq0_id];
    var p1 = this.component_pos[seq1_id];

    if (p0.length < 5) {
      console.log("ERROR: seq0_id", seq0_id, "seq1_id", seq1_id, ", p0:", p0, ", p1:", p1);
      continue;
    }

    var seq0_name = p0[8];
    var seq1_name = p1[8];

    var sp0 = seq0_name.split("+");
    var dx = -200;
    if (sp0.length == 2) { dx = -200 - parseInt(sp0[1])*50; }

    var sp1 = seq1_name.split("+");
    if (sp1.length == 2) { dx = -200 - parseInt(sp1[1])*50; }

    // Trying to make graph network a little more intelligable
    //
    var t_height = (p0[1]+p0[3]/2) + (p1[1]+p1[3]/2);
    var fudge_dx = 5*Math.floor(t_height/200);
    if (fudge_dx >  5*this.line_width) { fudge_dx =  5*this.line_width; }
    if (fudge_dx < -5*this.line_width) { fudge_dx = -5*this.line_width; }
    fudge_dx = 0;


    var s0_type = p0[4];
    var s0_step = p0[6];

    var s1_type = p1[4];
    var s1_step = p1[6];

    if (s0_step == s1_step) {
      if (s0_type == "t") {
        this.graphjoin_line.push([ p0[0]+p0[2], p0[1]+p0[3]/2, p1[0], p1[1]+p1[3]/2, (p1[0]-(p0[0]+p0[2]))+dx ]);
      } else {
        this.graphjoin_line.push([ p1[0]+p1[2], p1[1]+p1[3]/2, p0[0], p0[1]+p0[3]/2, (p0[0]-(p1[0]+p1[2]))+dx ]);
      }
    }

    else if (s0_step < s1_step) {
      this.graphjoin_line.push([ p0[0]+p0[2], p0[1]+p0[3]/2, p1[0], p1[1]+p1[3]/2, (p1[0]-(p0[0]+p0[2]))+dx ]);
    }

    else if (s1_step < s0_step) {
      this.graphjoin_line.push([ p1[0]+p1[2], p1[1]+p1[3]/2, p0[0], p0[1]+p0[3]/2, (p0[0]-(p1[0]+p1[2]))+dx ]);
    }

  }

}

lightningGraph.prototype.init = function() {

  this.min_step = -1;
  this.max_step = -1;
  this.component = {};

  this.component_pos = {};
  this.graphjoin_line = [];

  this.sequence_zoom_detail = 0.13;

  this.allele_path = {};
  this.sample_allele = {};
  this.allele_to_sample_map = {};

  this.highlight_sample_flag = false;
  this.highlight_sample_name = "hu_test";

  this.call_set = {};
  this.call_set_id = {};


  var z = g_db.exec("select ID, sequenceRecordName, md5checksum, length from Sequence");

  if ((z.length!=1) || (!("values" in z[0]))) {
    console.log("ERROR, didn't get expected response from DB query");
    return;
  }

  var min_path = -1;
  var max_path = -1;
  var min_step = -1;
  var max_step = -1;

  var vals = z[0].values;
  for (var ind=0; ind<vals.length; ind++) {
    var id   = vals[ind][0];
    var name = vals[ind][1];

    var name_part = name.split(".");
    var path = parseInt(name_part[1], 16);
    var step = parseInt(name_part[2], 16);

    if (min_path<0) {
      min_path = path;
      max_path = path;
      min_step = step;
      max_step = step;
    }

    if (min_path>path) min_path = path;
    if (max_path<path) max_path = path;

    if (min_step>step) min_step = step;
    if (max_step<step) max_step = step;
  }

  this.min_step = min_step;
  this.max_step = max_step;

  var component = this.component;

  for (var ind=0; ind<vals.length; ind++) {
    var id    = vals[ind][0];
    var name  = vals[ind][1];
    var m5    = vals[ind][2];
    var len   = vals[ind][3];

    var name_part = name.split(".");
    var path = parseInt(name_part[1], 16);
    var step = parseInt(name_part[2], 16);
    var rank_part = name_part[3].split("+");

    // tag
    //
    if (name_part[3][0] == 't') {
      if (!(step in component)) component[step] = [];
      //component[step].push(["t", id, name, len]);

      // oeis A000120
      // count of set bits (0-f)
      //
      var bc = [ 0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4 ];
      var bitcount = 0;
      var bitvec = name_part[3];
      for (var i=1; i<bitvec.length; i++) {
        bitcount += bc[ parseInt(bitvec[i], 16) ];
      }
      component[step].push(["t", id, name, len, 0, bitcount-17]);
    }

    // Body
    //
    else if (name_part[3][0] == 'r') {

      var rank = parseInt(rank_part[0].slice(1), 16);
      var seedlen = rank_part[1];

      if (!(step in component)) component[step] = [];

      //component[step].push(["r", id, name, len, rank_part[1]]);
      component[step].push(["r", id, name, len, seedlen, rank]);
    }

    else {
      conosle.log("couldn't make sense of", name, "at", ind);
    }

  }

  // Sort each steps non-tag element by it's rank.
  //
  for (var step in this.component) {
    this.component[step].sort(function(a,b) {
      if (a[5] < b[5]) { return -1; }
      if (a[5] > b[5]) { return  1; }
      return 0;
    })
  }

  this.chooseCoords();

  this.graphjoin_line = [];
  this.init_graphjoin();
  this.init_allele_path();

}

lightningGraph.prototype.chooseCoords = function() {
  //var shift_step = 0;

  // for path 2c5
  //var shift_step = 972;

  // for path 247
  //var shift_step = 2746;

  var shift_step = this.shift_step;

  var fold_size = 50;
  var tag_len = 24;

  var edge_col_width = 30;

  //var box_h = this.font_size + 4*this.line_width;
  var box_h = this.font_size;
  var shift_y = box_h + this.font_size*2 + 4*this.line_width;
  var tag_shift = this.font_size*24 + this.font_size*2 + 2*this.line_width + edge_col_width*this.font_size;

  var shift_x = fold_size*this.font_size + 2*this.line_width + 2*this.font_size + tag_shift + edge_col_width*this.font_size;

  var fudge_y = 2*this.font_size + 2*this.line_width;

  //var path = "2c5";
  var path = this.path;

  for (var step=this.min_step; step<=this.max_step; step++) {
    if (!(step in this.component)) { continue; }

    x = (step - shift_step)*shift_x;
    y = 0;


    var tag_height = 0;

    // try flattening out the normal case
    //
    tag_height = ((box_h*5+fudge_y) - (box_h+fudge_y))/2;

    var body_height = 0;
    var tot_height = 0;

    var seed_tile_height = 2*this.line_width;

    for (var ind=0; ind<this.component[step].length; ind++) {

      var comp = this.component[step][ind];
      var seq_id   = comp[1];
      var seq_name = comp[2];

      // component_pos
      // 0 1  2      3     4    5    6     7      8
      // x y width height type path step seq_id seq_name


      if (this.component[step][ind][0] == "t") {
        var h = box_h + fudge_y;
        this.component_pos[comp[1]] = [ x, y+tag_height, comp[3]*this.font_size, h, "t", path, step, seq_id, seq_name ];
        //this.component_pos[comp[1]] = [ x, y, comp[3]*this.font_size, h, "t", path, step, seq_id, seq_name ];
        tag_height += shift_y;
      } else if (this.component[step][ind][0] == "r")  {


        //var h = box_h*(Math.floor(comp[3]/fold_size) + (((comp[3] % fold_size)==0)?0:1));
        var h = box_h*(Math.floor(comp[3]/fold_size) + (((comp[3] % fold_size)==0)?0:1)) + fudge_y;

        var stl = parseInt(comp[4]);
        var fudge_width = 0;
        if (stl>1) {
          fudge_width = 40;
        }


        // component_pos
        // 0 1  2      3     4    5    6     7      8
        // x y width height type path step seq_id seq_name

        if (stl==1) {
          this.component_pos[comp[1]] = [ x+tag_shift, y+body_height, stl*this.font_size*fold_size + fudge_width, h, "r", path,step, seq_id, seq_name ];
          body_height += (h+fudge_y);
        } else {
          this.component_pos[comp[1]] = [ x+tag_shift, -(y+seed_tile_height+h), stl*this.font_size*fold_size + fudge_width, h, "r",path,step,seq_id,seq_name ];
          seed_tile_height += (h+fudge_y);
        }

      } else {
      }

      tot_height += shift_y;

    }

  }


}


lightningGraph.prototype.drawText = function(text, cx, cy) {
  a = text.split("\n");

  for (var i=0; i<a.length; i++) {
    var fs = 1.4*this.font_size;
    if (i==0) fs = 0.7*this.font_size;


    g_painter.drawText(a[i],
        cx,
        cy + i*(this.font_size*1.1),
        this.font_color,
        fs,
        this.font_angle,
        this.font_offset_h);

    // Highlight in red nocall positions
    //
    if (i>0) {
      var hi_nocall = a[i].replace(/[actg]/g, ' ');
      if (hi_nocall.search("n")>=0) {

        g_painter.drawText(hi_nocall,
            cx,
            cy + i*(this.font_size*1.1),
            "rgba(255,0,0,0.5)",
            fs,
            this.font_angle,
            this.font_offset_h);
      }
    }

  }
}

lightningGraph.prototype.lineHeightLightness = function(a, b) {
  var min_v = 40;
  var max_v = 60;
  var a = (Math.abs(a-b)%257) / 257;
  var g = Math.floor((max_v - min_v)*a + min_v);
  //var g = Math.floor( 256*((Math.abs(y0-) % 257) / 64 / 6) );
  return g;
}

lightningGraph.prototype.drawJoin = function(sx, sy, ex, ey, dx) {

  var r = this.path_radius;
  //var g = Math.floor( 256*((Math.abs(ey-ex) % 257) / 64 / 6) );
  var g = this.lineHeightLightness(ey, ex);
  var lc = "rgba(" + g + "," + g + "," + g + ",0.4)";

  if (Math.abs(ey-sy) < 2*r) {
    g_painter.line(sx, sy, ex, ey, lc, this.line_width);
    return;
  }

  // Swap start and end point if end is to the left
  // of start
  //
  var lx, ly, rx, ry;
  lx = sx; ly = sy;
  rx = ex; ry = ey;

  if (sx > ex) {
    lx = ex; ly = ey;
    rx = sx; ry = sy;
  }
  else if (ex == sx) {
    if (ey > sy) {
      lx = ex; ly = ey;
      rx = sx; ry = sy;
    }
  }

  var mx = 0;
  var my = 0;

  // There are different draws depending on whether we're going up or down.
  //
  if (ly < ry) {

    if (typeof(dx) === "undefined") { mx = (rx - lx)/2 + lx; }
    else                            { mx = lx + dx; }
    my = (ry - ly)/2 + ly;

    g_painter.line(lx, ly, mx-r, ly, lc, this.line_width);
    g_painter.drawArc(mx-r, ly+r, r, -Math.PI/2.0, 0, false, this.line_width, lc);
    g_painter.line(mx, ly+r, mx, ry-r, lc, this.line_width);
    g_painter.drawArc(mx+r, ry-r, r, Math.PI/2.0, Math.PI, false, this.line_width, lc);
    g_painter.line(mx+r, ry, rx, ry, lc, this.line_width);
  } else {

    if (typeof(dx) === "undefined") { mx = (rx - lx)/2 + lx; }
    else                            { mx = lx + dx; }
    my = ry + (ly - ry)/2 ;

    g_painter.line(lx, ly, mx-r, ly, lc, this.line_width);
    g_painter.drawArc(mx-r, ly-r, r, Math.PI/2.0, 0, true, this.line_width, lc);
    g_painter.line(mx, ly-r, mx, ry+r, lc, this.line_width);
    g_painter.drawArc(mx+r, ry+r, r, -Math.PI/2.0, Math.PI, true, this.line_width, lc);
    g_painter.line(mx+r, ry, rx, ry, lc, this.line_width);
  }


}

lightningGraph.prototype._box_box_intersect = function( bb0, bb1, box_fudge )
{
  box_fudge = ( (typeof box_fudge === 'undefined') ? 0 : box_fudge );

  return !( ( bb1[0][0] >  (bb0[1][0] + box_fudge)) ||
            ( bb1[1][0] <  (bb0[0][0] - box_fudge)) ||
            (-bb1[1][1] > -(bb0[0][1] - box_fudge)) ||
            (-bb1[0][1] < -(bb0[1][1] + box_fudge)) );

}


lightningGraph.prototype.draw = function() {
  var x = 0;
  var y = 0;
  var w = 1000;
  var h = 1000;
  var cx = x + w/2;
  var cy = y + h/2;
  var fudge = 10;

  g_painter.line_join_type = "bevel";

  var w = $(window).width();
  var h = $(window).height();

  var p = [[0,0],[0,0]];

  var ul = g_painter.devToWorld(0, 0);
  var lr = g_painter.devToWorld(w, h);

  for (var id in this.component_pos) {
    var v = this.component_pos[id];

    // Skip over rectangles that are out of our view window.
    // Only checks for X axis.
    //
    if ((v[0]-this.line_width)        > lr.x) { continue; }
    if ((v[0]+v[2]+2*this.line_width) < ul.x) { continue; }

    //y clip
    if ((v[1]+v[3]+2*this.line_width) < ul.y) { continue; }
    if ((v[1]-this.line_width)        > lr.y) { continue; }


    var c = ( (v[4]=="t") ? this.tag_color : this.body_color );
    g_painter.drawRectangle(v[0], v[1], v[2], v[3], this.line_width, c);

    if (g_painter.zoom > this.sequence_zoom_detail) {

      //g_painter.drawText("test", v[0]+this.line_width, v[1] + this.line_width, c, this.font_size );

      // component_pos
      // 0 1  2      3     4    5    6     7      8
      // x y width height type path step seq_id seq_name
      var seq_name = v[8];

      if (seq_name in g_sequence) {
        this.drawText(g_sequence[seq_name], v[0]+this.line_width, v[1] + this.line_width);
      } else {
        console.log("ERROR: sequence name", seq_name, "not found in g_sequence");
      }

    }

  }

  // Highlight alleles
  //
  if (this.highlight_sample_flag) {
    var hiname = this.highlight_sample_name;

    //orangish
    //var color_allele = { "0" : "rgba(200,150,0,0.8)", "1":"rgba(150,100,50,0.8)" };

    //aquaish
    var color_allele = { "0" : "rgba(0,150,250,0.8)", "1":"rgba(100,180,230,0.8)" };

    var allele_pass = 1;
    for (var allele_name in this.sample_allele[hiname]) {
      var allele = this.sample_allele[hiname][allele_name];

      var hi_fudge = this.line_width*allele_pass;

      for (var ind=0; ind<allele.length; ind++) {

        if (!(allele[ind] in this.component_pos)) {
          console.log("ERROR: could not find ", allele[ind], "in component_pos");
          continue;
        }

        var v = this.component_pos[ allele[ind] ];

        // Skip over rectangles that are out of our view window.
        // Only checks for X axis.
        //
        if ((v[0]-this.line_width-hi_fudge)          > lr.x) { continue; }
        if ((v[0]+v[2]+2*this.line_width+2*hi_fudge) < ul.x) { continue; }

        //y clip
        if ((v[1]+v[3]+2*this.line_width+2*hi_fudge) < ul.y) { continue; }
        if ((v[1]-this.line_width-hi_fudge)          > lr.y) { continue; }

        var c = ( (v[4]=="t") ? this.tag_color : this.body_color );

        c = color_allele[(allele_pass-1)];
        g_painter.drawRectangle(v[0]-hi_fudge, v[1]-hi_fudge, v[2]+2*hi_fudge, v[3]+2*hi_fudge, this.line_width, c);

      }

      allele_pass++;

    }
  }

  for (var ind=0; ind<this.graphjoin_line.length; ind++) {
    var uv = this.graphjoin_line[ind];

    // Skip joints that are out of our view window.
    // Only checks for X axis.
    //
    if ((uv[0] > lr.x) && (uv[2] > lr.x)) { continue; }
    if ((uv[2] < ul.x) && (uv[2] < ul.x)) { continue; }

    //y clip
    //if (v[1] < ul.y) { continue; }
    //if ((v[1]-v[3]) > lr.y) { continue; }

    //this.drawJoin(uv[0], uv[1], uv[2], uv[3]);
    this.drawJoin(uv[0], uv[1], uv[2], uv[3], uv[4] );
  }

}

