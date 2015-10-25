package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/lxc/lxd/shared"
	"github.com/olekukonko/tablewriter"
)

func help(status int) {
	me := path.Base(os.Args[0])
	fmt.Printf("Usage:\n%s\n", me)
	fmt.Printf("    show default id mapping\n")
	fmt.Printf("%s cfile\n", me)
	fmt.Printf("    show id mappings for containers in cfile\n")
	os.Exit(status)
}

func isHelp(s string) bool {
	switch s {
	case "-h":
		return true
	case "--help":
		return true
	case "help":
		return true
	default:
		return false
	}
}

func showDefaultMap() {
	set, err := shared.DefaultIdmapSet()
	if err != nil {
		fmt.Printf("Error reading default mapset: %q\n", err)
		help(1)
	}
	fmt.Printf("Your current default allocation is:\n\n")
	for _, m := range set.Idmap {
		t := "uid"
		if !m.Isuid {
			t = "gid"
		}
		hmin := m.Hostid
		hmax := m.Hostid + m.Maprange - 1
		cmin := m.Nsid
		cmax := m.Nsid + m.Maprange - 1
		fmt.Printf("host %s %d - %d mapping to %d - %d in container\n",
			t, hmin, hmax, cmin, cmax)
	}
}

func main() {
	if len(os.Args) > 2 {
		fmt.Printf("Too many arguments")
		help(1)
	}
	if len(os.Args) == 2 && isHelp(os.Args[1]) {
		help(0)
	}

	if len(os.Args) == 1 {
		showDefaultMap()
		return
	}

	// to do - parse file.  Let's just use bogus input for now
	containers, err := ParseFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error opening file %s: %q\n", os.Args[1], err)
		os.Exit(1)
	}

	data, err := Process(containers)
	if err != nil {
		fmt.Printf("Error processing file %s: %q\n", os.Args[1], err)
		os.Exit(1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Container",
		"parent start",
		"parent end",
		"container start",
		"container end",
		"host start",
		"host end"})
	table.AppendBulk(data)
	table.Render()

	return
}

type container struct {
	idmap   *shared.IdmapSet
	// for nested containers, the true host min/max
	hostmin int
	hostmax int
}

type containers map[string]container

func ParseFile(fName string) (containers, error) {
	set := containers{}
	file, err := os.Open(fName)
	if err != nil {
		return set, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// c1/c2 0:100000:65536
		s := strings.Fields(scanner.Text())
		if len(s) > 2 {
			return set, fmt.Errorf("Too many fields")
		}
		mapstr := fmt.Sprintf("b:%s", s[1])
		m, err := shared.IdmapSet{}.Append(mapstr)
		if err != nil {
			return set, err
		}
		c := container{idmap: &m, hostmin: -1, hostmax: -1}
		set[s[0]] = c
	}

	return set, nil
}

func Process(containers containers) ([][]string, error) {
	result := [][]string{}

	for name, c := range containers {
		// note - we only do cases where uid+gid are the same, so just
		// take the first idmap
		idmap := c.idmap
		r := idmap.Idmap[0].Maprange
		pstart := fmt.Sprintf("%d", idmap.Idmap[0].Hostid)
		pend   := fmt.Sprintf("%d", idmap.Idmap[0].Hostid + r)
		cstart := fmt.Sprintf("%d", idmap.Idmap[0].Nsid)
		cend   := fmt.Sprintf("%d", idmap.Idmap[0].Nsid + r)
		v1, v2, err := verifyRange(name, idmap.Idmap[0], containers)
		if err != nil {
			return result, err
		}
		c.hostmin = v1
		c.hostmax = v2
		hstart := fmt.Sprintf("%d", v1)
		hend   := fmt.Sprintf("%d", v2)
		newstr := []string{name, pstart, pend, cstart, cend, hstart, hend}
		result = append(result, newstr)
	}

	return result, nil
}

func verifyRange(name string, idmap shared.IdmapEntry, c containers) (int, int, error) {
	return idmap.Hostid, idmap.Hostid + idmap.Maprange, nil
}
