package main

import (
	"code.google.com/p/go.net/html"
	"errors"
	"flag"
	"fmt"
	"github.com/meoow/nodefinder"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var inplace bool

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-i] PATH FILE1 [FILE2 ...]\n",
			filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSyntax for PATH can be found at \"github.com/meoow/nodefinder\", multiple rules can be separated by \":::\"\n")
		os.Exit(0)
	}
	flag.BoolVar(&inplace, "i", false, "Change file inplace.")
	log.SetOutput(os.Stderr)
}

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	path := flag.Arg(0)
	for _, f := range flag.Args()[1:] {
		found := 0
		fh, err := os.Open(f)
		if err != nil {
			log.Print(err)
			continue
		}
		defer fh.Close()
		root, err := html.Parse(fh)
		if err != nil {
			log.Print(err)
			continue
		}
		for _, p := range strings.Split(path, ":::") {
			tags := nodefinder.NewPath(p)
			nodes := nodefinder.FindByNode(tags, root)
			if err != nil {
				log.Print(err)
				continue
			}
			for _, n := range nodes {
				if n.Parent != nil {
					n.Parent.RemoveChild(n)
					found++
				} else {
					log.Print(errors.New("Found node has no parent, can not be removed"))
				}
			}
		}
		if !inplace {
			html.Render(os.Stdout, root)
		} else if found > 0 {
			tempfile, err := ioutil.TempFile(filepath.Dir(f), "htmlcleaner")
			if err != nil {
				log.Print(err)
				continue
			}
			err = html.Render(tempfile, root)
			if err != nil {
				log.Print(err)
				continue
			}
			tempfile.Close()
			err = os.Rename(tempfile.Name(), f)
			if err != nil {
				log.Print(err)
				continue
			}
		}
	}
}
