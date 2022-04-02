package main

import (
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
	i := strings.Index(v, ss)
	// no text found
	if i == -1 {
		return nil
	}

	begin := i - 30
	end := i + 30
	if begin < 0 {
		begin = 0
	}
	if end >= len(v) {
		end = len(v) - 1
	}
	fmt.Printf("%s:p%d:%d: %s\n", s.filename, page, i, v[begin:end])
	return nil
}

type PDFSearcher struct {
	filename string
	reader   *pdf.Reader

	searchExpr string
}
