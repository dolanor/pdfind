package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/ledongthuc/pdf"
)

type config struct {
	color       bool
	contextSize uint
	outStyle    string
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Printf("usage:\n\tpdfind [FLAG]... PATTERNS [FILE]...\n")
	flag.PrintDefaults()
}

var cfg config

func run(args []string) error {
	colorcfg := flag.Bool("color", true, "colored search expression context")
	contextSize := flag.Uint("context", 30, "context displayed (how many character before and after the match is output")
	outStyle := flag.String("out-style", "grep", `output style: "grep" or "csv"`)

	flag.Usage = printUsage
	flag.Parse()
	cfg = config{
		color:       *colorcfg,
		contextSize: *contextSize,
		outStyle:    *outStyle,
	}

	if flag.NArg() < 1 {
		flag.Usage()
		return errors.New("not enough arguments passed to CLI")
	}

	searchExpr := flag.Arg(0)
	var files []string
	if flag.NArg() == 1 {
		// we scan current dir by default
		files = []string{"."}
	} else {

		files = flag.Args()[1:]
	}

	fmt.Printf(`"%s","%s","%s","%s"\n`, "filename", "page", "colum", "context")
	for _, f := range files {
		s := PDFSearcher{filename: f, searchExpr: searchExpr}
		err := s.search()
		if err != nil {
			// we don't stop the whole program for 1 file
			log.Printf(`could not search in file "%s"`, f)
		}
	}
	return nil
}

type PDFSearcher struct {
	filename string
	reader   *pdf.Reader

	searchExpr string
}

func (s *PDFSearcher) search() error {
	err := s.searchFileName()
	if err != nil {
		return err
	}

	return nil
}

func (s PDFSearcher) searchDir(f *os.File) error {
	err := filepath.WalkDir(f.Name(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("walkdir:", err)
			return fs.SkipDir
		}

		lName := strings.ToLower(d.Name())
		if !strings.HasSuffix(lName, ".pdf") {
			return nil
		}

		s := PDFSearcher{filename: path, searchExpr: s.searchExpr}
		err = s.searchFileName()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s PDFSearcher) searchFileName() error {
	f, err := os.Open(s.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}

	if st.IsDir() {
		err = s.searchDir(f)
		return err
	}

	err = s.searchFile(f, st.Size())
	if err != nil {
		return err
	}

	return nil
}

func (s PDFSearcher) searchFile(f *os.File, size int64) error {
	var err error
	s.reader, err = pdf.NewReader(f, size)
	if err != nil {
		return err
	}

	for i := 1; i <= s.reader.NumPage(); i++ {
		err := s.searchInPage(s.searchExpr, i)
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
		switch cfg.outStyle {
		case "csv":
			fmt.Printf(`"%s","%d","%d","%s"
`, s.filename, page, i, res)
		case "grep":
			fallthrough
		default:
			fmt.Printf("%s:%d:%d: %s\n", s.filename, page, i, res)
		}
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

	ctxBegin := i - int(cfg.contextSize)
	ctxEnd := i + len(searchExpr) + int(cfg.contextSize)
	if ctxBegin < 0 {
		ctxBegin = 0
	}
	if ctxEnd >= len(text) {
		ctxEnd = len(text) - 1
	}

	searchText := text[i : i+len(searchExpr)]
	if cfg.color {
		searchText = color.RedString(searchText)
	}
	t := fmt.Sprintf("%s%s%s", text[ctxBegin:i], searchText, text[i+len(searchExpr):ctxEnd])
	return i, t
}
