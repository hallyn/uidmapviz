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

```
+-----------+--------------+------------+-----------------+---------------+------------+----------+
| CONTAINER | PARENT START | PARENT END | CONTAINER START | CONTAINER END | HOST START | HOST END |
+-----------+--------------+------------+-----------------+---------------+------------+----------+
| c1/c2     |       100000 |     165536 | 0               |         65536 |     200000 |   265536 |
| c1        |       100000 |     300000 | 0               |        200000 |     100000 |   300000 |
+-----------+--------------+------------+-----------------+---------------+------------+----------+
```

A child container must come after the parent definition.
If a child container's mapping is not contained within the
parent's available range, then uidmapviz will warn you.

## More details

To install uidmapviz in your ~/go/bin, you can:

```
go get github.com/hallyn/uidmapviz
cd go/src/github.com/hallyn/uidmapviz/
go build
```

```
ubuntu@lxd1:~$ uidmapviz
Your current default allocation is:
host uid 100000 - 165535 mapping to 0 - 65535 in container
host gid 100000 - 165535 mapping to 0 - 65535 in container
```

We'll create a little file showing the uid mappings we'd like to use:

```
cat > containers << EOF
c1 0:100000:65536
c1/c2 0:100000:65536
c1/c2/c3 0:200000:65536
EOF
```

Here the first field is a container name.  If the name contains slashes,
then it is read as grandparent/parent/child.  The second field is the
range, in the same format as /etc/subuid: the first id in the container,
the first id in the parent container, and the range of the mapping.  For
each container we are mapping 65536 uids, starting at 0 in the container,
to the range starting at 100000 in the parent.

Let's see what uidmapviz says about this:

ubuntu@lxd1:~$ uidmapviz containers
Error opening file containers: "Mapping for c1/c2 exceeds its parent's, parentids should be between 0 - 65535"

c1's uid range is too narrow, as we should have known.  When we provide
a container with a mapping for uids 0-20, then only those uids exist in
the container.  It cannot delegate uids which do not exist.  If we want
to hand 20 uids to a child container while reserving 20 for ourselves,
then we'll need 40 uids in total.

So let's give our first container 3*65k, or 196608 uids, and hand 65536-196604
to the second container.  Then we'll hand the top half of c2's uids to c3.

```
cat > containers2 << EOF
c1 0:100000:196605
c1/c2 0:65536:131072
c1/c2/c3 0:65536:65536
EOF

ubuntu@lxd1:~$ uidmapviz containers2
Looking at c1
Looking at c1/c2
Looking at c1/c2/c3
+-----------+--------------+------------+-----------------+---------------+------------+----------+
| CONTAINER | PARENT START | PARENT END | CONTAINER START | CONTAINER END | HOST START | HOST END |
+-----------+--------------+------------+-----------------+---------------+------------+----------+
| c1        |       100000 |     296605 | 0               |        196605 |     100000 |   296605 |
| c1/c2     |        65536 |     196608 | 0               |        131072 |     165536 |   296608 |
| c1/c2/c3  |        65536 |     131072 | 0               |         65536 |     231072 |   296608 |
+-----------+--------------+------------+-----------------+---------------+------------+----------+
```

(Usually I use wider and easier to read allocations, i.e.
b:100000:200000)
Perfect.  Now we need to make sure that the root user on the host is allowed to
delegate the the range required by c1.  We do this using /etc/subuid and
/etc/subgid:

```
root:100000:196605
```

Restart lxd so the deamon will re-read its idmap.

We'll now create container c1.  In order for it to be able to run
nested containers, we'll set the 'security.nesting' flag to true:

lxc launch wily c1 --config security.nesting=true

Per-tenant idmaps are a future todo, for now the container will receive
the full default id mapping allocated to the root user.  So now we can
log into the c1 container

lxc exec c1 bash

and install and setup lxd.  This time we'll allocate the root user the
mapping we wanted for c2:

root:65536:196605

That should be it.  now you can

lxc launch wily c2

for an unprivileged container inside an unprivileged container.  Configuring
c3 is left as an exercise for the reader.
