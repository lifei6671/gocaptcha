package gocaptcha

import (
	"embed"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//go:embed fonts
var _fs embed.FS
var ioFS fs.FS = _fs

func FileSystem(fs embed.FS) error {
	err := readFontFromFS(fs)
	if err != nil {
		return err
	}
	ioFS = fs
	return nil
}

//获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func ReadFonts(dirPth string, suffix string) (err error) {
	fontFamily = fontFamily[:0]

	dir, err := ioutil.ReadDir(dirPth)

	if err != nil {
		return err
	}
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			fontFamily = append(fontFamily, filepath.Join(dirPth, fi.Name()))
		}
	}
	ioFS = os.DirFS(dirPth)
	return nil
}

//获取所及字体.
func RandFontFamily() (*truetype.Font, error) {
	fontFile := fontFamily[r.Intn(len(fontFamily))]
	fontBytes, err := ReadFile(fontFile)
	if err != nil {
		log.Printf("读取文件失败 -> %s - %+v\n", fontFile, err)
		return nil, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Printf("解析字体内容试标 -> %s - %+v\n", fontFile, err)
		return nil, err
	}
	return f, nil
}

//添加一个字体路径到字体库.
func AddFontFamily(fontPath ...string) {
	fontFamily = append(fontFamily, fontPath...)
}

//ReadFile 读取文件，优先从设置的文件系统中读取，失败后从内置文件系统读取，失败后从磁盘读取.
func ReadFile(name string) ([]byte, error) {
	if ioFS != nil {
		b, err := fs.ReadFile(ioFS, name)
		if err == nil {
			return b, nil
		}
	}
	return os.ReadFile(name)
}

func readFontFromFS(fs embed.FS) error {
	files, err := fs.ReadDir("fonts")
	if err != nil {
		log.Printf("解析字体文件失败 -> %+v", err)
		return err
	} else {
		fontFamily = fontFamily[:0]
		for _, fi := range files {
			if fi.IsDir() { // 忽略目录
				continue
			}
			if strings.HasSuffix(strings.ToLower(fi.Name()), ".ttf") { //匹配文件
				fontFamily = append(fontFamily, filepath.Join("fonts", fi.Name()))
			}
		}
	}
	return nil
}
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = readFontFromFS(_fs)
}
