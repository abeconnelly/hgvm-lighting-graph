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


function toolNav( x, y, viewMode )
{
  this.viewMode = ( (typeof viewMode === 'undefined') ? false : viewMode );
  x = ( typeof x !== 'undefined' ? x : 0 );
  y = ( typeof y !== 'undefined' ? y : 0 );

  this.mouse_down = false;
  this.mouse_cur_x = x;
  this.mouse_cur_y = y;

  this.lock_grid_flag = true;
  this.show_cursor_flag = true;

  this.cursorSize = 6;
  this.cursorWidth = 1;

  this.mouse_drag_flag = false;

  this.mouse_world_xy = g_painter.devToWorld(x, y);
  this.snap_world_xy = {"x" : this.mouse_world_xy.x, "y": this.mouse_world_xy.y};

  this.highlightBoxFlag = false;
  this.highlightBox = { x:0, y:0, w:0, h:0 };
  this.highlightBoxWidth = 10;
  this.highlightBoxColor = "rgba(0,0,0,0.4)";
}

toolNav.prototype.update = function(x, y)
{
  this.mouse_cur_x = x;
  this.mouse_cur_y = y;

  this.mouse_world_xy = g_painter.devToWorld(x, y);
  this.snap_world_xy = { "x" : this.mouse_world_xy.x, "y" : this.mouse_world_xy.y };
}


toolNav.prototype.drawOverlay = function()
{

  if ( !this.mouse_drag_flag )
  {
    this.snap_world_xy = { "x" : this.mouse_world_xy.x, "y" : this.mouse_world_xy.y };

    var s = this.cursorSize / 2;
    g_painter.drawRectangle( this.snap_world_xy["x"] - s,
                             this.snap_world_xy["y"] - s,
                             this.cursorSize ,
                             this.cursorSize ,
                             this.cursorWidth ,
                             "rgb(128, 128, 128 )" );

  }

  if ( this.highlightBoxFlag )
  {
    g_painter.drawRectangle( this.highlightBox.x,
                             this.highlightBox.y,
                             this.highlightBox.w,
                             this.highlightBox.h,
                             this.highlightBoxWidth,
                             "rgb(128,128,128)",
                             true,
                             this.highlightBoxColor );
  }

}

toolNav.prototype.mouseDown = function( button, x, y )
{
  this.mouse_down = true;

  //if (button == 3)
  if (button == 1)
    this.mouse_drag_flag = true;
}

toolNav.prototype.doubleClick = function(button, x, y)
{
  var world_coord = g_painter.devToWorld( x, y );
}

toolNav.prototype.mouseUp = function( button, x, y )
{
  this.mouse_down = false;

  //if (button == 3)
  if (button == 1)
    this.mouse_drag_flag = false;
}

toolNav.prototype.mouseMove = function( x, y )
{
  if ( this.mouse_drag_flag )
     this.mouseDrag ( x - this.mouse_cur_x, y - this.mouse_cur_y );

  this.mouse_cur_x = x;
  this.mouse_cur_y = y;

  var world_xy = g_painter.devToWorld( this.mouse_cur_x, this.mouse_cur_y );
  this.mouse_world_xy["x"] = world_xy["x"];
  this.mouse_world_xy["y"] = world_xy["y"];

  var wx = this.mouse_world_xy.x;
  var wy = this.mouse_world_xy.y;

  g_painter.dirty_flag = true;
}

toolNav.prototype.mouseDrag = function( dx, dy )
{
  g_painter.adjustPan ( dx, dy );
  g_painter.dirty_flag = true;
}

toolNav.prototype.mouseWheel = function( delta )
{
  g_painter.adjustZoom ( this.mouse_cur_x, this.mouse_cur_y, delta );
}

toolNav.prototype.keyDown = function( keycode, ch, ev )
{

  if ( keycode == 188 ) { }
  else if ( keycode == 192 ) { }
  else if ( keycode == 190 ) { }

  if ( keycode == 219 ) // '['
  {
  }
  else if (keycode == 221 ) // ']'
  {
  }

  var x = this.mouse_cur_x;
  var y = this.mouse_cur_y;
  var wc = g_painter.devToWorld(x, y);

  var wx = wc["x"];
  var wy = wc["y"];

  if ( ch == '1' ) {  // no grid
    g_painter.setGrid ( 0 );
  } else if ( ch == '2' ) {  // point grid
    g_painter.setGrid ( 1 );
  } else if ( ch == '3' ) {  // line grid
    g_painter.setGrid ( 2 );
  } else if (ch == 'A') {
  }

  else if (ch == 'I')
  {
    console.log("info");
  }

  else if (ch == 'L') { }
  else if (ch == 'V') { }
  else if (ch == 'U') { }
  else if (ch == 'K') { }

  else if (ch == 'B') { }
  else if (ch == 'C') { }

  else if (ch == 'P') { }
  else if (ch == 'J') { }
  else if (ch == 'X') { }
  else if (ch == 'S') { }
  else if (ch == 'W') { }
  else if (ch == 'R') { }
  else if (ch == 'E')  { }
  else if (ch == 'H') { }
  else if (ch == 'Y') { }
  else if (ch == 'D') { }
  else if (ch == 'M') { }

  if (keycode == '32') return false;
  return true;
}

toolNav.prototype.keyUp = function( keycode, ch, ev ) { }


