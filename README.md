# uidmapviz

A little program to visualize uid mappings

## Running

You can simply run uidmap with no arguments to show the default
uid mapping (i.e. your first full map entry)

If you provide a filename as an argument, it should represent a
set of uid mappings for containers.  For instance

c1 0:100000:200000
c1/c2 0:100000:65536

represents two containers, one nested inside the other.
