package gocaptcha

import (
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var DefaultFontFamily = NewFontFamily()
var ErrNoFontsInFamily = os.ErrNotExist

// SetFonts sets the default font family
func SetFonts(fonts ...string) error {
	for _, font := range fonts {
		if err := DefaultFontFamily.AddFont(font); err != nil {
			return err
		}
	}
	return nil
}

// SetFontPath sets the default font family from a directory
func SetFontPath(fontDirPath string) error {
	return DefaultFontFamily.AddFontPath(fontDirPath)
}

// FontFamily is a font family that creates a new font family
type FontFamily struct {
	fonts     []string
	fontCache *sync.Map
	r         *rand.Rand
}

// Random returns a random font from the family
func (f *FontFamily) Random() (*truetype.Font, error) {
	if len(f.fonts) == 0 {
		return nil, ErrNoFontsInFamily
	}
	fontFile := f.fonts[f.r.Intn(len(f.fonts))]
	if v, ok := f.fontCache.Load(fontFile); ok {
		return v.(*truetype.Font), nil
	}
	font, err := f.parseFont(fontFile)
	if err != nil {
		return nil, err
	}
	f.fontCache.Store(fontFile, font)
	return font, nil
}

func (f *FontFamily) parseFont(fontFile string) (*truetype.Font, error) {
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return font, nil
}

// AddFont adds a font to the family and returns an error if it fails
func (f *FontFamily) AddFont(fontFile string) error {
	if _, ok := f.fontCache.Load(fontFile); ok {
		return nil
	}
	font, err := f.parseFont(fontFile)
	if err != nil {
		return err
	}
	f.fonts = append(f.fonts, fontFile)
	f.fontCache.Store(fontFile, font)
	return nil
}

// AddFontPath adds all .ttf files from the given directory to the font family and returns an error if any
func (f *FontFamily) AddFontPath(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && filepath.Ext(path) == ".ttf" {
			return f.AddFont(path)
		}
		return nil
	})
}

// NewFontFamily creates a new font family with the given fonts
func NewFontFamily() *FontFamily {
	return &FontFamily{
		fontCache: &sync.Map{},
		r:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
