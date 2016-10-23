package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	markdown "github.com/shurcooL/github_flavored_markdown"
)

type PageData struct {
	Title   string
	Content template.HTML
}

type markdownFile struct {
	*bytes.Reader

	file http.File
}

var _ http.File = &markdownFile{}

func openMarkdown(name string, f http.File) (*markdownFile, error) {
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	body_bs := markdown.Markdown(bs)
	d := PageData{
		Title:   strings.TrimPrefix(strings.Replace(name, "-", " ", -1), "/"),
		Content: template.HTML(string(body_bs)),
	}
	buffer := new(bytes.Buffer)
	t, err := template.ParseFiles("public-wiki.html")
	if err != nil {
		return nil, err
	}
	err = t.Execute(buffer, d)
	if err != nil {
		return nil, err
	}
	return &markdownFile{bytes.NewReader(buffer.Bytes()), f}, nil
}

func (f *markdownFile) Readdir(count int) ([]os.FileInfo, error) {
	log.Panicf("Readdir() called on file")
	return nil, nil
}

func (f *markdownFile) Stat() (os.FileInfo, error) {
	fi, err := f.file.Stat()
	if err != nil {
		return nil, err
	}
	return &markdownFileInfo{thinMarkdownFileInfo{fi}, f}, nil
}

func (f *markdownFile) Close() error {
	return f.file.Close()
}

// Doesn't need to read the file; Size() panics
type thinMarkdownFileInfo struct {
	os.FileInfo
}

func (fi *thinMarkdownFileInfo) Name() string {
	return strings.TrimSuffix(fi.FileInfo.Name(), ".md")
}

func (fi *thinMarkdownFileInfo) Size() int64 {
	log.Panicf("Size() called on thinMarkdownFileInfo")
	return 0
}

type markdownFileInfo struct {
	thinMarkdownFileInfo

	*markdownFile
}

func (fi *markdownFileInfo) Size() int64 {
	return fi.Reader.Size()
}

type markdownDir struct {
	http.File
}

func (d *markdownDir) Readdir(count int) ([]os.FileInfo, error) {
	fis, err := d.File.Readdir(count)
	if err != nil {
		return nil, err
	}
	fakefis := []os.FileInfo{}
	for _, fi := range fis {
		fakefis = append(fakefis, &thinMarkdownFileInfo{fi})
	}
	return fakefis, nil
}

type markdownFs struct {
	fs http.FileSystem
}

var _ http.FileSystem = &markdownFs{}

func (fs *markdownFs) Open(name string) (http.File, error) {
	realf, err := fs.fs.Open(name)
	if os.IsNotExist(err) {
		realf, err = fs.fs.Open(name + ".md")
	}
	if err != nil {
		return nil, err
	}
	finfo, err := realf.Stat()
	if err != nil {
		return nil, err
	}
	if finfo.IsDir() {
		return realf, nil
	} else {
		return openMarkdown(name, realf)
	}
}

func MarkdownFileSystem(fs http.FileSystem) http.FileSystem {
	return &markdownFs{fs}
}
