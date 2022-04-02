# pdfind

tool to search some text into PDF files and directory containing PDFs.

## Usage

```
usage:
	pdfind [FLAG]... PATTERNS [FILE]...
  -color
    	colored search expression context (default true)
  -context uint
    	context displayed (how many character before and after the match is output (default 30)
  -out-style string
    	output style: "grep" or "csv" (default "grep")
```

## Limitations

- for now, it only works on textual PDF, for scanned PDF, we need to add some OCR system like tesseract.
- can not find the actual line in the PDF (the plaintext transformation breaks the line system), the info is ordered by page, and then by character position in the page.
