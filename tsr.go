package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"time"
)

type Archive struct {
	Date time.Time
	Part bool
	Name string
}

type ArchiveList []Archive

func (d ArchiveList) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ArchiveList) Len() int           { return len(d) }
func (d ArchiveList) Less(i, j int) bool { return d[i].Date.Before(d[j].Date) }

var (
	nrOfArchives int
	del, getTime bool
	config       string
)

func init() {
	flag.BoolVar(&getTime, "time", false, "just print rfc1123 formatted time")
	flag.IntVar(&nrOfArchives, "number", 3, "number of archives to keep")
	flag.BoolVar(&del, "delete", false, "keep \"number\" of oldest archives, delete the rest")
	flag.StringVar(&config, "configfile", "~/.tarsnaprc", "location of tarsnaprc")
}

func main() {
	flag.Parse()
	if getTime {
		fmt.Println(time.Now().Format(time.RFC1123))
		return
	}
	var prefix string
	if len(flag.Args()) != 1 {
		fmt.Println("Provide a archive prefix to list/delete")
		flag.Usage()
		return
	}
	prefix = flag.Arg(0)
	archives := getArchives(prefix)
	sort.Sort(sort.Reverse(archives))
	if del {
		deleteArchives(archives[nrOfArchives:])
		return
	}
	for _, a := range archives {
		fmt.Println(a.Name)
	}
}

func deleteArchives(archives ArchiveList) {
	if len(archives) < 1 {
		fmt.Println("no archives to delete")
		return
	}
	for _, a := range archives {
		fmt.Println("deleting:", a.Name)
		cmd := exec.Command("tarsnap", "-d", "--configfile", config, "-f", fmt.Sprintf("%s", a.Name))
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err, " ", string(out))
		}
	}
}

func getArchives(prefix string) ArchiveList {
	// run tarsnap --list-archives
	cmd := exec.Command("tarsnap", "--list-archives", "--configfile", config)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err, " ", string(out))
	}
	archiveRegex := regexp.MustCompile(prefix + `\((.*)\)`)
	archives := make(ArchiveList, 0)
	for _, s := range bytes.Split(out, []byte("\n")) {
		if matches := archiveRegex.FindStringSubmatch(string(s)); len(matches) > 1 {
			t, err := time.Parse(time.RFC1123, matches[1])
			if err != nil {
				panic(err)
			}
			b := Archive{Date: t, Name: string(s)}
			archives = append(archives, b)
		}
	}
	return archives
}
