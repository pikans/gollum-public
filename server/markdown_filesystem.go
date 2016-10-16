package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	markdown "github.com/shurcooL/github_flavored_markdown"
)

const before = `<html><head><title>pika wiki - public</title></head><body>`
const after = `</body></html>`

type markdownFile struct {
	*bytes.Reader

	file http.File
}

var _ http.File = &markdownFile{}

func openMarkdown(f http.File) (*markdownFile, error) {
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	body_bs := markdown.Markdown(bs)
	page_bs := []byte(before)
	page_bs = append(page_bs, body_bs...)
	page_bs = append(page_bs, after...)
	log.Printf("%s", page_bs)
	reader := bytes.NewReader(page_bs)
	return &markdownFile{reader, f}, nil
}

func (f *markdownFile) Readdir(count int) ([]os.FileInfo, error) {
	return f.file.Readdir(count)
}

type markdownFileInfo struct {
	os.FileInfo

	*markdownFile
}

func (fi *markdownFileInfo) Size() int64 {
	return fi.Reader.Size()
}

func (f *markdownFile) Stat() (os.FileInfo, error) {
	fi, err := f.file.Stat()
	if err != nil {
		return nil, err
	}
	return &markdownFileInfo{fi, f}, nil
}

func (f *markdownFile) Close() error {
	return f.file.Close()
}

type markdownFs struct {
	fs http.FileSystem
}

var _ http.FileSystem = &markdownFs{}

func (fs *markdownFs) Open(name string) (http.File, error) {
	realf, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return openMarkdown(realf)
}

func MarkdownFileSystem(fs http.FileSystem) http.FileSystem {
	return &markdownFs{fs}
}
