package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
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
		s := PDFSearcher{filename: f}
		err := s.search(searchExpr)
		if err != nil {
			// we don't stop the whole program for 1 file
			log.Printf(`could not search in file "%s"`, f)
		}
	}
	return nil
}

func (s *PDFSearcher) search(searchExpr string) error {
	f, err := os.Open(s.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}

	s.reader, err = pdf.NewReader(f, st.Size())
	if err != nil {
		return err
	}

	for i := 1; i <= s.reader.NumPage(); i++ {
		err := s.searchInPage(searchExpr, i)
		if err != nil {
			log.Printf(`could not search in file "%s", page %d: %v`, s.filename, i, err)
		}
	}

	return nil
}

func (s PDFSearcher) searchInPage(searchExpr string, page int) error {
	p := s.reader.Page(page)
	if p.V.IsNull() {
		// nothing to see here
		return nil
	}

	text, err := p.GetPlainText(nil)
	if err != nil {
		return err
	}

	v := strings.ToLower(text)
	ss := strings.ToLower(searchExpr)
	for {
		i, res := multisearch(v, ss)
		if i == -1 {
			break
		}
		fmt.Printf("%s:p%d:%d: %s\n", s.filename, page, i, res)
		v = v[i+1:]
	}
	return nil
}

func multisearch(text string, searchExpr string) (index int, context string) {
	i := strings.Index(text, searchExpr)
	// no text found
	if i == -1 {
		return -1, ""
	}

	ctxBegin := i - 30
	ctxEnd := i + 30
	if ctxBegin < 0 {
		ctxBegin = 0
	}
	if ctxEnd >= len(text) {
		ctxEnd = len(text) - 1
	}
	t := fmt.Sprintf("%s%s%s", text[ctxBegin:i], color.RedString(text[i:i+len(searchExpr)]), text[i+len(searchExpr):ctxEnd])
	return i, t
}

type PDFSearcher struct {
	filename string
	reader   *pdf.Reader

	searchExpr string
}
