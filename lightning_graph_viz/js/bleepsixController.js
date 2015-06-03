/*

    Copyright (C) 2013 Abram Connelly

    This file is part of bleepsix v2.

    bleepsix is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    bleepsix is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with bleepsix.  If not, see <http://www.gnu.org/licenses/>.

    Parts based on bleepsix v1. BSD license
    by Rama Hoetzlein and Abram Connelly.

*/

function bleepsixController( viewMode ) {

  this.viewMode = ( (typeof viewMode == 'undefined') ? false : viewMode );

  this.canvas = null;
  this.context = null;

  this.focus_flag = false;
  this.focus = null;

  //this.palette = new guiPalette();

  this.mouse_left_down = false;
  this.mouse_center_down = false;
  this.mouse_right_down = false;

  this.mouse_start_x = 0;
  this.mouse_start_y = 0;
  this.mouse_cur_x = 0;
  this.mouse_cur_y = 0;

  this.width = 1200;
  this.height = 700;

  this.capState = "unknown";

  this.display_text_flag = true;
  this.display_text = "demo";

  var d = new Date();
  var curt = d.getTime();

  this.tool = new toolNav();

  this.action_text_flag = true;
  this.action_text = "init";
  this.action_text_fade  = { sustainDur : 1500, dropoffDur : 500, T : 0, lastT: curt };
}

//--------------------------

bleepsixController.prototype.registerDrawCallback = function (draw_cb)
{
  this.draw_cb = draw_cb;
}


bleepsixController.prototype.fadeMessage = function ( msg )
{
  var d = new Date();
  var curt = d.getTime();

  this.action_text = msg;
  this.action_text_fade.T = 0;
  this.action_text_fade.lastT = curt;
}

bleepsixController.prototype.redraw = function ()
{

  var action_text_touched = false;
  var action_text_val = 0.0;

  var at_s = 0.4;

  if (this.action_text_flag)
  {
    if ( this.action_text_fade.T < this.action_text_fade.sustainDur )
    {
      action_text_val = at_s ;
      action_text_touched = true;
    }
    else if (this.action_text_fade.T < ( this.action_text_fade.sustainDur + this.action_text_fade.dropoffDur) )
    {
      var t = this.action_text_fade.sustainDur + this.action_text_fade.dropoffDur - this.action_text_fade.T;
      action_text_val = at_s * t / this.action_text_fade.dropoffDur;
      action_text_touched = true;
    }
    else
      action_text_touched = false;

    if (action_text_touched)
    {
      g_painter.dirty_flag = true;
      var d = new Date();
      var curt = d.getTime();
      var dt = curt - this.action_text_fade.lastT;
      this.action_text_fade.T += dt;
      this.action_text_fade.lastT = curt;
    }

  }


  if ( g_painter.dirty_flag )
  {
    g_painter.startDraw();
    g_painter.drawGrid();
    this.tool.drawOverlay();

    if (this.draw_cb)
      this.draw_cb();
    g_painter.endDraw();

    g_painter.context.setTransform ( 1, 0, 0, 1, 0, 0 );
    if (this.display_text_flag)
    {
      var _height = this.height-20;
      g_painter.drawText(this.display_text, 10, _height, "rgba(0,0,0,0.4)", 15);
    }

    if (action_text_touched)
    {
      var _height = this.height - 50;
      g_painter.drawText(this.action_text, 10, _height, "rgba(0,0,0," + action_text_val + ")", 15);
    }

    g_painter.dirty_flag = false;
    g_painter.context.setTransform ( 1, 0, 0, 1, 0, 0 );
  }

}

bleepsixController.prototype.canvas_coords_from_global = function( x, y )
{
  var rect = this.canvas.getBoundingClientRect();
  var rl = rect.left;
  var rt = rect.top;

  var scrollx = window.scrollX;
  var scrolly = window.scrollY;

  return [ x - rl - scrollx, y - rt - scrolly ];
}

bleepsixController.prototype.mouseEnter = function( x, y )
{
}

bleepsixController.prototype.mouseLeave = function( x, y )
{
}

bleepsixController.prototype.resize = function( w, h, ev )
{
  g_painter.dirty_flag = true;
  this.width = w;
  this.height = h;
}


bleepsixController.prototype.keyDown = function( keycode, ch, ev )
{
  if ( ch == 'G' ) {

  }
  else if (ch == 'O')
  {

  }
  else if (ch == 'P')
  {

  }
  else if (ch == 'I')
  {

  }
  else if (ch == 'U')
  {

  }
  else if (ch == 'Y')
  {

  }

  var r = true;

  if (!this.focus_flag)
  {
    if (typeof this.tool.keyDown !== 'undefined' )
    {
      r = this.tool.keyDown( keycode, ch, ev );
    }
  } else {
    r = this.focus.keyDown( keycode, ch, ev );
  }

  return r;
}

bleepsixController.prototype.keyPress = function( keycode, ch, ev )
{
  if (this.viewMode) { return true; }

  if (!this.focus_flag)
  {
    if (typeof this.tool.keyPress !== 'undefined' )
      this.tool.keyPress( keycode, ch, ev );
  }
  else
  {
    this.focus.keyPress( keycode, ch, ev );
  }

}


bleepsixController.prototype.mouseDown = function( button, x, y )
{

  this.focus_flag = false;
  this.focus = null;

  if (typeof this.tool.mouseDown !== 'undefined' )
    this.tool.mouseDown ( button, x, y );
}

bleepsixController.prototype.doubleClick = function( e )
{
  if (typeof this.tool.doubleClick !== 'undefined' )
    this.tool.doubleClick( e, this.mouse_cur_x, this.mouse_cur_y )
}

bleepsixController.prototype.mouseUp = function( button, x, y )
{
  if (typeof this.tool.mouseUp !== 'undefined' )
    this.tool.mouseUp ( button, x, y );
}

bleepsixController.prototype.mouseMove = function( x, y )
{
  this.mouse_cur_x = x;
  this.mouse_cur_y = y;

  if (typeof this.tool.mouseMove !== 'undefined')
    this.tool.mouseMove(x,y);
}

bleepsixController.prototype.mouseDrag = function( dx, dy )
{
  if (typeof this.tool.mouseDrag !== 'undefined' )
    this.tool.mouseDrag ( x, y );
}

bleepsixController.prototype.mouseWheel = function( delta )
{
  var x = this.mouse_cur_x;
  var y = this.mouse_cur_y;

  if (typeof this.tool.mouseWheel !== 'undefined' )
  {
    this.tool.mouseWheel ( delta );
  }

}

bleepsixController.prototype.init = function( canvas_id )
{
  this.canvas = $("#" + canvas_id)[0];
  this.context = this.canvas.getContext('2d');

  var controller = this;

  $(canvas_id).focus( function(ev) {
  });

  $(canvas_id).mouseup( function(e) {
    var xy = controller.canvas_coords_from_global( e.pageX, e.pageY );
    controller.mouseUp( e.which , xy[0], xy[1] );
  });

  $(canvas_id).mousedown( function(e) {
    var xy = controller.canvas_coords_from_global( e.pageX, e.pageY );
    controller.mouseDown( e.which, xy[0], xy[1] );
  });

  $(canvas_id).mouseover( function(e) {
  });

  $(canvas_id).mouseenter( function(e) {
    var xy = controller.canvas_coords_from_global( e.pageX, e.pageY );
    controller.mouseEnter( xy[0], xy[1] );
  });

  $(canvas_id).mouseleave( function(e) {
    var xy = controller.canvas_coords_from_global( e.pageX, e.pageY );
    controller.mouseLeave( xy[0], xy[1] );
  });

  $(canvas_id).mousemove( function(e) {
    var xy = controller.canvas_coords_from_global( e.pageX, e.pageY );
    controller.mouseMove( xy[0], xy[1] );
  });

  $(canvas_id).mousewheel( function(e, delta, detlax, deltay) {
    controller.mouseWheel( delta );
    return false;
  });

  $(canvas_id).click( function(e) {
  });

  $(canvas_id).dblclick( function(e) {
    controller.doubleClick(e);
  });

  $(canvas_id).keypress( function(e) {
  });

  $(canvas_id).keydown( function(e) {
    var key = ( e.which ? e.which : e.keyCode );
    var ch = String.fromCharCode(key);

    this.capState = $(window).capslockstate("state");

    return controller.keyDown( e.keyCode, ch, e );
  });

  $(canvas_id).keyup( function(e) {
    var key = e.which;
    var ch = String.fromCharCode(key);
  });

  $(canvas_id).resize( function(e) {
    console.log("resize");
    console.log(e);
  });

  $(canvas_id).keypress( function(e) {
    var key = e.which;
    var ch = String.fromCharCode(key);
    controller.keyPress( key, ch, e );
  });



  $(window).bind("capsOn", function(e) {
    controller.capState = "on";
  });

  $(window).bind("capsOff", function(e) {
    controller.capState = "off";
  });

  $(window).bind("capsUnknown", function(e) {
    controller.capState = "unknown";
  });

  $(window).capslockstate();

  // get rid of right click menu popup
  $(document).bind("contextmenu", function(e) { return false; });

  // put focus on the canvas
  $(canvas_id).focus();

  // do first draw
  g_painter.dirty_flag = true;

}
