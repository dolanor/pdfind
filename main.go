package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/extractor"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

func usage() {
	fmt.Printf("usage:\n\tpdfind SEARCH FILE [FILE...]")
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	fileNames := os.Args[2:]
	err := search(os.Args[1], fileNames...)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func search(searchExpr string, fileNames ...string) error {
	for _, fileName := range fileNames {
		f, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		r, err := pdf.NewPdfReader(f)
		if err != nil {
			return err
		}

		n, err := r.GetNumPages()
		if err != nil {
			return err
		}

		for i := 0; i < n; i++ {
			pageNum := i + 1

			p, err := r.GetPage(pageNum)
			if err != nil {
				return err
			}

			if pageNum == 1 {
				found, x, y, err := locateSearch(p, searchExpr)
				if err != nil {
					return err
				}

				if !found || (x == 0 && y == 0) {
					return errors.New("unable to find search term in file")
				}

				fmt.Printf("%s: found\n", fileName)
			}
		}
	}
	return nil
}

func locateSearch(page *pdf.PdfPage, searchExpr string) (found bool, x float64, y float64, err error) {

	xt, err := extractor.New(page)
	if err != nil {
		return found, x, y, err
	}

	txt, err := xt.ExtractText()
	if err != nil {
		return found, x, y, err
	}
	if strings.Contains(txt, searchExpr) {
		fmt.Printf("Tj: %s\n", txt)
		found = true
		return found, 1, 1, err
	}

	//////
	contentStr, err := page.GetAllContentStreams()
	if err != nil {
		return found, x, y, err
	}

	parser := pdfcontent.NewContentStreamParser(contentStr)
	if err != nil {
		return found, x, y, err
	}

	operations, err := parser.Parse()
	if err != nil {
		return found, x, y, err
	}

	for _, op := range *operations {
		if op.Operand == "Tm" && len(op.Params) == 6 {
			if val, ok := op.Params[4].(*pdfcore.PdfObjectFloat); ok {
				x = float64(*val)
			}
			if val, ok := op.Params[5].(*pdfcore.PdfObjectFloat); ok {
				y = float64(*val)
			}
		} else if op.Operand == "Tj" && len(op.Params) == 1 {
			val, ok := op.Params[0].(*pdfcore.PdfObjectString)
			if ok {
				str := string(*val)
				fmt.Printf("Tj: %s\n", str)
				if strings.Contains(str, searchExpr) {
					fmt.Printf("Tj: %s\n", str)
					found = true
					break
				}
			}
		}
	}
	return found, x, y, nil

}
