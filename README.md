# mdtocgen

Markdown Table of Content Generator

```
go run main.go [-dir=dirPath] [-out=outFile] [-t=Title] [-asc[=true|false]]

Usage:
  -asc
    	Order the TOC in ascending order, if false, it will be in descending order (default true)
  -dir string
    	Directory to read the file (default ".")
  -out string
    	Output file
  -t dir
    	Title of output file, default is the dir
```