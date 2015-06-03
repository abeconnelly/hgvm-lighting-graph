
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

var g_fa_text = null
var g_fa_array = [];
var g_sequence = {};

var g_controller = null;
var g_model = null;
var g_db = null;



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

function sample_deselect() {
  g_model.unhighlight_sample();
  g_controller.display_text = "brca1";
}

function sample_select(sample_name) {
  g_model.highlight_sample(sample_name);
  g_controller.display_text = "brca1 / " + sample_name;
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


  var xhr = new XMLHttpRequest();
  xhr.open('GET', 'db/tilegraph.sqlite3', true);
  xhr.responseType = 'arraybuffer';

  xhr.onload = function(e) {
    console.log(">>>ok", e);

    var uInt8Array = new Uint8Array(this.response);
    g_db = new SQL.Database(uInt8Array);

    console.log(">>> database loaded....", g_db);

    console.log("loading lightningGraph...");
    g_model.init();

    var ele = document.getElementById("dropdown-sample-list");
    for (var sample_name in g_model.call_set) {
      $("#dropdown-sample-list").append('<li role="presentation">' +
          //'<a role="menuitem" tabindex="-1" href="#" onclick="g_model.highlight_sample(' + "'" + sample_name + "'" + ');"  >' +
          '<a role="menuitem" tabindex="-1" href="#" onclick="sample_select(' + "'" + sample_name + "'" + ');"  >' +
          sample_name + '</a></li>');
    }

  };
  xhr.send();

  var fa_req = new XMLHttpRequest();
  fa_req.open('GET', 'db/out_pgp174.fa', true);
  fa_req.onloadend = function() {

    g_fa_text = fa_req.responseText;

    //DEBUG
    console.log("ok!!", g_fa_text.length);


    g_fa_array = g_fa_text.split("\n\n");

    for (var i=0; i<g_fa_array.length; i++) {
      var z = g_fa_array[i].split("\n");
      var header = z[0];

      var h = header.slice(1);
      g_sequence[h] = g_fa_array[i];

    }
  }
  fa_req.send();



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


