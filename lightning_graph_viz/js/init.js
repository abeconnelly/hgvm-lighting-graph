
var drawing = true,
  running = true,
  mouseDown = false,
  visible = true;

var msec = 0;
var frame = 0;
var lastTime = new Date();

var g_show_fps = false;
var g_painter = null;
var g_canvas = "canvas";
var g_context = null;

//var g_fa_text = null
//var g_fa_array = [];
var g_sequence = {};

var g_controller = null;
var g_model = null;
var g_db = null;
var g_db_cache = {};



function canvasFocus() {
  var c = document.getElementById("canvas");
  c.focus();
}

function draw() {

}

function loop() {

  frame = frame + 1;
  if ( frame >= 30 ) {
    var d = new Date();
    msec = (d.getTime() - lastTime ) / frame;
    lastTime = d;
    frame = 0;
  }

  if (g_show_fps)
  {
    var container = document.querySelector( 'section' );
    container.innerHTML = "FPS: " + (1000.0/msec);
  }

  var saved_dirty = g_painter.dirty_flag;

  g_controller.redraw();
  requestAnimationFrame( loop, 1 );
}

var g_loci_name = "BRCA1";

function sample_deselect() {
  g_model.unhighlight_sample();
  g_controller.display_text = g_loci_name;
}

function sample_select(sample_name) {
  g_model.highlight_sample(sample_name);
  g_controller.display_text = g_loci_name + " / " + sample_name;
}

function loci_select(loci_name) {
  g_loci_name = loci_name;
}

$(document).ready( function() {

  //Load database

  // I don't know what this is
  var cvs = document.getElementById('canvas');
  cvs.onselectstart = function() { return false; }

  g_painter = new bleepsixRender( "canvas" );
  g_painter.setGrid ( 1 );

  g_controller = new bleepsixController();
  g_controller.init("canvas");

  g_model = new lightningGraph();
  g_controller.registerDrawCallback(function() { g_model.draw(); });

  requestAnimationFrame( loop, 1 );

  // fit to size of window
  //
  var w = $(window).width();
  var h = $(window).height();

  var canv = document.getElementById('canvas');
  canv.width = w;
  canv.height = h;
  g_painter.setWidthHeight( w, h );
  g_controller.resize(w,h);

  $(window).resize( function(ee) {
    var w = $(window).width();
    var h = $(window).height();
    var canv = document.getElementById('canvas');
    canv.width = w;
    canv.height = h;
    g_painter.setWidthHeight( w, h );
    g_controller.resize(w,h);
  });

  // Have dropdown button get value of selection.
  //
  $("#dropdown-loci-list li a").click(function() {
    var loci_name = $(this).text();

    g_db = g_db_cache[loci_name];

    if (loci_name == "BRCA1") {
      g_model.shift_step = 2746;
      g_model.path = "247";
    } else if (loci_name == "BRCA2") {
      g_model.shift_step = 972;
      g_model.path = "2c5";
    }

    g_model.init();

    $("#dropdown-sample-list").empty();
    $("#dropdown-sample-list").append('<li role="presentation"><a role="menuitem" tabindex="-1" href="#" onclick="sample_deselect();"; >None</a></li>');
    for (var sample_name in g_model.call_set) {
      $("#dropdown-sample-list").append('<li role="presentation">' +
          '<a role="menuitem" tabindex="-1" href="#" onclick="sample_select(' + "'" + sample_name + "'" + ');"  >' +
          sample_name + '</a></li>');
    }

    $("#sample_main_text").text("None");
    $("#sample_main_text").val("None");
    sample_deselect();

    // This needs to happen after the dropdown list has been created
    //
    $("#dropdown-sample-list li a").click(function() {
      $("#sample_main_text").text( $(this).text() );
      $("#sample_main_text").val( $(this).text() );
    });


    $("#loci_main_text").text(loci_name);
    $("#loci_main_text").val(loci_name);

  });


  // Grab BRCA1 database
  //
  var xhr = new XMLHttpRequest();
  xhr.open('GET', 'db/tilegraph_247.sqlite3', true);
  xhr.responseType = 'arraybuffer';
  xhr.onload = function(e) {

    var uInt8Array = new Uint8Array(this.response);
    g_db_cache["BRCA1"] = new SQL.Database(uInt8Array);

    var ele = document.getElementById("dropdown-sample-list");
    for (var sample_name in g_model.call_set) {
      $("#dropdown-sample-list").append('<li role="presentation">' +
          '<a role="menuitem" tabindex="-1" href="#" onclick="sample_select(' + "'" + sample_name + "'" + ');"  >' +
          sample_name + '</a></li>');
    }

    // Have dropdown button get value of selection.
    // This needs to happen after the dropdown list has been created
    //
    $("#dropdown-sample-list li a").click(function() {
      $("#sample_main_text").text( $(this).text() );
      $("#sample_main_text").val( $(this).text() );
    });

  };
  xhr.send();

  // Grab BRCA2 database
  //
  var xhr2 = new XMLHttpRequest();
  xhr2.open('GET', 'db/tilegraph_2c5.sqlite3', true);
  xhr2.responseType = 'arraybuffer';
  xhr2.onload = function(e) {

    var uInt8Array = new Uint8Array(this.response);
    g_db = new SQL.Database(uInt8Array);
    g_db_cache["BRCA2"] = g_db;

    g_model.shift_step = 972;
    g_model.path = "247";
    g_model.init();

    var ele = document.getElementById("dropdown-sample-list");
    for (var sample_name in g_model.call_set) {
      $("#dropdown-sample-list").append('<li role="presentation">' +
          '<a role="menuitem" tabindex="-1" href="#" onclick="sample_select(' + "'" + sample_name + "'" + ');"  >' +
          sample_name + '</a></li>');
    }

    // This needs to happen after the dropdown list has been created
    //
    $("#dropdown-sample-list li a").click(function() {
      $("#sample_main_text").text( $(this).text() );
      $("#sample_main_text").val( $(this).text() );
    });

  };
  xhr2.send();

  var fa_req = new XMLHttpRequest();
  fa_req.open('GET', 'db/pgp174_247.fa', true);
  fa_req.onloadend = function() {
    var fa_text = fa_req.responseText;
    var fa_array = fa_text.split("\n\n");
    for (var i=0; i<fa_array.length; i++) {
      var z = fa_array[i].split("\n");
      var header = z[0];
      var h = header.slice(1);
      g_sequence[h] = fa_array[i];
    }
  }
  fa_req.send();

  var fa_req2 = new XMLHttpRequest();
  fa_req2.open('GET', 'db/pgp174_2c5.fa', true);
  fa_req2.onloadend = function() {
    var fa_text = fa_req.responseText;
    var fa_array = fa_text.split("\n\n");
    for (var i=0; i<fa_array.length; i++) {
      var z = fa_array[i].split("\n");
      var header = z[0];
      var h = header.slice(1);
      g_sequence[h] = fa_array[i];
    }
  }
  fa_req2.send();



});


// more magic
//
if ( !window.requestAnimationFrame ) {
  window.requestAnimationFrame = ( function() {
  return window.webkitRequestAnimationFrame ||
         window.mozRequestAnimationFrame ||
         window.oRequestAnimationFrame ||
         window.msRequestAnimationFrame ||
         function( callback, element ) {
            window.setTimeout( callback, 1000 );
         };
       } )();
}


