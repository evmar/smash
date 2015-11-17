package xlib

//go:generate stringer -type=xEventType
type xEventType int

const (
	xKeyPress         xEventType = 2
	xKeyRelease       xEventType = 3
	xButtonPress      xEventType = 4
	xButtonRelease    xEventType = 5
	xMotionNotify     xEventType = 6
	xEnterNotify      xEventType = 7
	xLeaveNotify      xEventType = 8
	xFocusIn          xEventType = 9
	xFocusOut         xEventType = 10
	xKeymapNotify     xEventType = 11
	xExpose           xEventType = 12
	xGraphicsExpose   xEventType = 13
	xNoExpose         xEventType = 14
	xVisibilityNotify xEventType = 15
	xCreateNotify     xEventType = 16
	xDestroyNotify    xEventType = 17
	xUnmapNotify      xEventType = 18
	xMapNotify        xEventType = 19
	xMapRequest       xEventType = 20
	xReparentNotify   xEventType = 21
	xConfigureNotify  xEventType = 22
	xConfigureRequest xEventType = 23
	xGravityNotify    xEventType = 24
	xResizeRequest    xEventType = 25
	xCirculateNotify  xEventType = 26
	xCirculateRequest xEventType = 27
	xPropertyNotify   xEventType = 28
	xSelectionClear   xEventType = 29
	xSelectionRequest xEventType = 30
	xSelectionNotify  xEventType = 31
	xColormapNotify   xEventType = 32
	xClientMessage    xEventType = 33
	xMappingNotify    xEventType = 34
	xGenericEvent     xEventType = 35
)
