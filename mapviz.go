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
		"host start",
		"host end",
		"container start",
		"container end"})
	table.AppendBulk(data)
	table.Render()

	return
}

type container struct {
	name   string
	mapset *shared.IdmapSet
}

func ParseFile(fName string) ([]container, error) {
	set := []container{}
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
		c := container{name: s[0], mapset: &m}
		set = append(set, c)
	}

	return set, nil
}

func Process(containers []container) ([][]string, error) {
	return [][]string{}, fmt.Errorf("Process function not implemented")
}
