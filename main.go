package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type MDFileInfo struct {
	Name     string
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
	)
	flag.StringVar(&wd, "dir", ".", "Directory to read the file")
	flag.StringVar(&outFile, "out", "", "Output file")
	flag.StringVar(&title, "t", "", "Title of output file, default is the `dir`")
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

	toc := CreateTocTree(files, "  ")

	if outFile != "" {
		err = os.WriteFile(outFile, []byte(toc), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// printMDFileInfo prints the information of an MDFileInfo struct in a formatted way.
//
// It takes an MDFileInfo struct as a parameter and prints its name, level, and title.
// If the MDFileInfo struct is a directory, it recursively prints the information of its children.
// The function does not return anything and is used for debugging purposes.
func printMDFileInfo(f MDFileInfo) {
	fmt.Printf("%s %s => %s\n", strings.Repeat("\t", f.Level), f.Name, f.Title)
	if f.IsDir {
		for _, child := range f.Children {
			printMDFileInfo(child)
		}
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
		Name:     ".",
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
							Name:     d,
							IsDir:    true,
							Children: make(map[string]MDFileInfo),
							Level:    p.Level + 1,
							Title:    "",
							Path:     url.PathEscape(filepath.Join(p.Path, d)),
						}
					}
					p = p.Children[d]
				}
				p.Children[info.Name()] = MDFileInfo{
					Name:  strings.TrimRight(info.Name(), filepath.Ext(info.Name())),
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

// CreateTocTree generates a table of contents (TOC) tree for a given MDFileInfo struct.
//
// It takes two parameters:
// - md: an MDFileInfo struct representing the file or directory for which the TOC is generated.
// - indent: a string representing the indentation for each level of the TOC tree.
//
// It returns a string representing the generated TOC tree.
func CreateTocTree(md MDFileInfo, indent string) string {
	var (
		toc string
	)
	switch md.Level {
	case 0:
		toc = "# " + md.Title + "\n"
	case 1:
		if md.IsDir {
			toc = fmt.Sprintf("\n## %s\n\n", md.Name)
		} else {
			toc = fmt.Sprintf("\n## [%s](%s)\n\n", md.Title, md.Path)
		}
	default:
		if md.IsDir {
			toc = fmt.Sprintf("%s- %s\n", strings.Repeat(indent, md.Level-2), md.Name)
		} else {
			toc = fmt.Sprintf("%s- [%s](%s)\n", strings.Repeat(indent, md.Level-2), md.Title, md.Path)
		}
	}
	for _, child := range md.Children {
		toc += CreateTocTree(child, indent)
	}
	return toc
}
