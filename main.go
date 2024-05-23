package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type MDFileInfo struct {
	IsDir    bool
	Children map[string]MDFileInfo
	Title    string
	Level    int
	Path     string
}

func main() {
	var (
		wd      string
		outFile string
		title   string
		sortAsc bool
	)
	flag.StringVar(&wd, "dir", ".", "Directory to read the file")
	flag.StringVar(&outFile, "out", "", "Output file")
	flag.StringVar(&title, "t", "", "Title of output file, default is the `dir`")
	flag.BoolVar(&sortAsc, "asc", true, "Order the TOC in ascending order, if false, it will be in descending order")
	flag.Parse()

	files, err := ListMDFiles(wd)
	if err != nil {
		log.Fatal(err)
	}

	if title == "" {
		if wd == "." {
			wd, _ = os.Getwd()
		}
		files.Title = filepath.Base(wd)
	} else {
		files.Title = title
	}

	toc := CreateTocTree(files, "  ", sortAsc)

	if outFile != "" {
		err = os.WriteFile(outFile, []byte(toc), 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(toc)
	}
}

// ListMDFiles lists all the Markdown files in the given path and its subdirectories.
//
// It takes a string parameter `dirPath` which represents the directory path to search for Markdown files.
// The function returns a `MDFileInfo` struct which represents the root directory and its descendants,
// and an error if any occurred during the file walk.
//
// The `MDFileInfo` struct has the following fields:
// - `Name`: the name of the file or directory
// - `IsDir`: a boolean indicating whether the file is a directory
// - `Children`: a map of child files and directories
// - `Level`: the level of indentation for the file or directory
// - `Title`: the title of the Markdown file
// - `Path`: the full path of the file or directory
func ListMDFiles(dirPath string) (MDFileInfo, error) {
	root := MDFileInfo{
		IsDir:    true,
		Children: make(map[string]MDFileInfo),
		Level:    0,
		Title:    "",
		Path:     ".",
	}
	err := filepath.Walk(dirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// We get Markdown files only
			if !info.IsDir() && info.Name() != "README.md" && filepath.Ext(path) == ".md" {
				relPath := strings.Replace(path, dirPath, ".", 1)
				dirs := strings.Split(filepath.Dir(relPath), "/")
				p := root
				for _, d := range dirs {
					if d == "." {
						continue
					}
					if _, ok := p.Children[d]; !ok {
						p.Children[d] = MDFileInfo{
							IsDir:    true,
							Children: make(map[string]MDFileInfo),
							Level:    p.Level + 1,
							Title:    d,
							Path:     url.PathEscape(filepath.Join(p.Path, d)),
						}
					}
					p = p.Children[d]
				}
				p.Children[info.Name()] = MDFileInfo{
					IsDir: false,
					Level: p.Level + 1,
					Title: GetMDTitle(path),
					Path:  url.PathEscape(relPath),
				}
			}
			return nil
		})
	if err != nil {
		return root, err
	}
	return root, nil
}

// GetMDTitle retrieves the title of a Markdown file, the title of the file is the first H1 header.
//
// It takes a filePath string parameter, which represents the path of the Markdown file.
// The function opens the file, reads its contents line by line, and searches for an H1 header.
// If an H1 header is found, it returns the text inside the header.
// If no H1 header is found or an error occurs while opening the file, it returns an empty string.
//
// Parameters:
// - filePath: the path of the Markdown file.
//
// Return type:
// - string: the title of the Markdown file, or an empty string if no title is found or an error occurs.
func GetMDTitle(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	h1Regex := regexp.MustCompile(`^#\s+(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if h1Regex.MatchString(line) {
			return h1Regex.FindStringSubmatch(line)[1]
		}
	}

	return ""
}

// CreateTocTree generates a table of contents (TOC) tree for the given MDFileInfo.
//
// Parameters:
// - md: the MDFileInfo object representing the file or directory.
// - indent: the string used for indentation in the TOC.
// - sortAsc: a boolean indicating whether the TOC should be sorted in ascending order.
//
// Returns:
// - string: the generated TOC tree.
func CreateTocTree(md MDFileInfo, indent string, sortAsc bool) string {
	var (
		toc string
	)
	switch md.Level {
	case 0:
		toc = "# " + md.Title + "\n"
	case 1:
		if md.IsDir {
			toc = fmt.Sprintf("\n## %s\n\n", md.Title)
		} else {
			toc = fmt.Sprintf("\n## [%s](%s)\n\n", md.Title, md.Path)
		}
	default:
		if md.IsDir {
			toc = fmt.Sprintf("%s- %s\n", strings.Repeat(indent, md.Level-2), md.Title)
		} else {
			toc = fmt.Sprintf("%s- [%s](%s)\n", strings.Repeat(indent, md.Level-2), md.Title, md.Path)
		}
	}
	keys := reflect.ValueOf(md.Children).MapKeys()
	stringKeys := make([]string, len(keys))
	for i, key := range keys {
		stringKeys[i] = key.String()
	}
	if sortAsc {
		sort.Strings(stringKeys)
	} else {
		sort.Sort(sort.Reverse(sort.StringSlice(stringKeys)))
	}

	for _, key := range stringKeys {
		toc += CreateTocTree(md.Children[key], indent, sortAsc)
	}
	return toc
}
