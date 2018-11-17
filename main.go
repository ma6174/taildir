package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ma6174/taildir/lib"
)

func main() {
	match := flag.String("match", "*", "match files")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage:\n%v dir\n", os.Args[0])
		return
	}
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
	r, err := lib.NewDirReader(lib.Config{
		Dir:         flag.Arg(0),
		FilePattern: *match,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(io.Copy(os.Stdout, r))
}
