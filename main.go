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
	"sync"
)

const SIMUL_GOR = 10

var inplace bool
var wg sync.WaitGroup

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
	paths := strings.Split(path, ":::")
	tagss := make([][]*nodefinder.Elem, 0, len(paths))
	for _, p := range paths {
		tagss = append(tagss, nodefinder.NewPath(p))
	}

	var simul_gor int
	if inplace {
		simul_gor = SIMUL_GOR
	} else {
		simul_gor = 1
	}
	wait_chan := make(chan int, simul_gor)

	for _, f := range flag.Args()[1:] {

		wait_chan <- 1
		wg.Add(1)

		go func(filename string) {
			defer func() {
				<-wait_chan
				wg.Done()
			}()
			found := 0
			fh, err := os.Open(filename)
			if err != nil {
				log.Print(err)
				return
			}
			defer fh.Close()
			root, err := html.Parse(fh)
			if err != nil {
				log.Print(err)
				return
			}
			for _, tags := range tagss {
				nodes := nodefinder.FindByNode(tags, root)
				if err != nil {
					log.Print(err)
					return
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
				tempfile, err := ioutil.TempFile(filepath.Dir(filename), "htmlcleaner")
				if err != nil {
					log.Print(err)
					return
				}
				err = html.Render(tempfile, root)
				if err != nil {
					log.Print(err)
					return
				}
				tempfile.Close()
				err = os.Rename(tempfile.Name(), filename)
				if err != nil {
					log.Print(err)
					return
				}
			}
		}(f)
	}
	wg.Wait()
}
