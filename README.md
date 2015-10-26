# uidmapviz

A little program to visualize uid mappings

## Running

You can either use "go run mapviz", or build an executable called
"uidmapviz" using "go build".

You can run uidmap with no arguments to show the default
uid mapping (i.e. your first full map entry).

If you provide a filename as an argument, it should represent a
set of uid mappings for containers.  For instance

c1 0:100000:200000
c1/c2 0:100000:65536

represents two containers, one nested inside the other.  Running
uidmapviz agains tthis gives you

+-----------+--------------+------------+-----------------+---------------+------------+----------+
| CONTAINER | PARENT START | PARENT END | CONTAINER START | CONTAINER END | HOST START | HOST END |
+-----------+--------------+------------+-----------------+---------------+------------+----------+
| c1/c2     |       100000 |     165536 | 0               |         65536 |     200000 |   265536 |
| c1        |       100000 |     300000 | 0               |        200000 |     100000 |   300000 |
+-----------+--------------+------------+-----------------+---------------+------------+----------+

A child container must come after the parent definition.
If a child container's mapping is not contained within the
parent's available range, then uidmapviz will warn you.
