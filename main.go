package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

func usage() string {
	return fmt.Sprintf("usage:\n\tpdfind SEARCH FILE [FILE...]")
}

func run(args []string) error {
	if len(args) < 2 {
		print(usage())
		return errors.New("not enough arguments passed to CLI")
	}

	searchExpr := args[0]

	files := args[1:]

	for _, f := range files {
		err := searchInFile(searchExpr, f)
		if err != nil {
			// we don't stop the whole program for 1 file
			log.Printf(`could not search in file "%s"`, f)
		}
	}
	return nil
}

func searchInFile(searchExpr string, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}

	r, err := pdf.NewReader(f, st.Size())
	if err != nil {
		return err
	}

	for i := 1; i <= r.NumPage(); i++ {
		err := searchInPage(searchExpr, r, i)
		if err != nil {
			log.Printf(`could not search in file "%s", page %d`, filename, i)
		}
	}

	return nil
}

func searchInPage(searchExpr string, r *pdf.Reader, page int) error {
	p := r.Page(page)
	if p.V.IsNull() {
		// nothing to see here
		return nil
	}

	b, err := r.GetPlainText()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.ReadFrom(b)

	v := strings.ToLower(buf.String())
	s := strings.ToLower(searchExpr)
	contains := strings.Contains(v, s)
	if contains {
		log.Printf("found in %s, page %d", "LOL", page)
		return nil
	}
	return nil
}
