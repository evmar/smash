# This script does 60fps screen recording using gstreamer 1.0.
# The Gnome3 screen recorder is full-screen only.  Other screen recorders
# are xdamage unaware and drop frames.
# All the gstreamer examples I could find online were for older
# versions of gstreamer, so this is pieced together from their docs.

# The stages of the pipeline:
# - read screen
# - grab the output at 60fps
# - use a buffer and a thread
# - convert video format
# - encode as vp8
# - stash vp8 in a webm format
# - save to a file
gst-launch-1.0 -e \
               ximagesrc starty=40 endx=300 endy=240 ! \
               video/x-raw,framerate=60/1 ! \
               queue ! \
               videoconvert ! \
               vp8enc ! \
               webmmux ! \
               filesink location=screen.webm
