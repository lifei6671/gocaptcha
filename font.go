package gocaptcha

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"sync"

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
	fonts         []string
	fontCache     *sync.Map
	weights       map[string]int
	fallbackFonts []string
	r             *rand.Rand
	mu            sync.RWMutex
	randMu        sync.Mutex
	randOnce      sync.Once
}

const maxFontWeightMultiplier = 32

// Random returns a random font from the family
func (f *FontFamily) Random() (*truetype.Font, error) {
	return f.RandomWithFallback()
}

// RandomWithFallback returns a weighted-random font and falls back on configured chain if needed.
func (f *FontFamily) RandomWithFallback() (*truetype.Font, error) {
	fontFiles, weights, fallbackFonts, err := f.snapshotSelection()
	if err != nil {
		return nil, err
	}

	f.randOnce.Do(func() {
		if f.r == nil {
			f.r = newSecureSeededRand()
		}
	})

	f.randMu.Lock()
	selected := chooseWeightedFontFile(f.r, fontFiles, weights)
	f.randMu.Unlock()

	candidates := buildFallbackCandidates(selected, fallbackFonts, fontFiles)
	var firstErr error
	for _, fontFile := range candidates {
		fontFace, loadErr := f.loadCachedFont(fontFile)
		if loadErr == nil {
			return fontFace, nil
		}
		if firstErr == nil {
			firstErr = loadErr
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return nil, ErrNoFontsInFamily
}

// WeightedCachedFonts returns fonts expanded by weight for efficient weighted random picking.
func (f *FontFamily) WeightedCachedFonts() ([]*truetype.Font, error) {
	fontFiles, weights, fallbackFonts, err := f.snapshotSelection()
	if err != nil {
		return nil, err
	}

	candidates := buildFallbackCandidates("", fallbackFonts, fontFiles)
	loaded := make(map[string]*truetype.Font, len(candidates))
	var firstErr error
	for _, fontFile := range candidates {
		fontFace, loadErr := f.loadCachedFont(fontFile)
		if loadErr != nil {
			if firstErr == nil {
				firstErr = loadErr
			}
			continue
		}
		loaded[fontFile] = fontFace
	}

	if len(loaded) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, ErrNoFontsInFamily
	}

	totalWeight := 0
	for _, fontFile := range fontFiles {
		if _, ok := loaded[fontFile]; !ok {
			continue
		}
		weight := weights[fontFile]
		if weight <= 0 {
			weight = 1
		}
		if weight > maxFontWeightMultiplier {
			weight = maxFontWeightMultiplier
		}
		totalWeight += weight
	}

	out := make([]*truetype.Font, 0, maxInt(1, totalWeight))
	for _, fontFile := range fontFiles {
		fontFace, ok := loaded[fontFile]
		if !ok {
			continue
		}
		weight := weights[fontFile]
		if weight <= 0 {
			weight = 1
		}
		if weight > maxFontWeightMultiplier {
			weight = maxFontWeightMultiplier
		}
		for i := 0; i < weight; i++ {
			out = append(out, fontFace)
		}
	}

	if len(out) == 0 {
		for _, fontFile := range fallbackFonts {
			if fontFace, ok := loaded[fontFile]; ok {
				out = append(out, fontFace)
			}
		}
	}
	if len(out) == 0 {
		for _, fontFace := range loaded {
			out = append(out, fontFace)
		}
	}
	if len(out) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, ErrNoFontsInFamily
	}
	return out, nil
}

// CachedFonts returns a snapshot of parsed fonts from the family.
func (f *FontFamily) CachedFonts() ([]*truetype.Font, error) {
	f.mu.RLock()
	if len(f.fonts) == 0 {
		f.mu.RUnlock()
		return nil, ErrNoFontsInFamily
	}
	fontFiles := append([]string(nil), f.fonts...)
	f.mu.RUnlock()

	fonts := make([]*truetype.Font, 0, len(fontFiles))
	for _, fontFile := range fontFiles {
		if v, ok := f.fontCache.Load(fontFile); ok {
			fonts = append(fonts, v.(*truetype.Font))
			continue
		}
		font, err := f.parseFont(fontFile)
		if err != nil {
			return nil, err
		}
		f.fontCache.Store(fontFile, font)
		fonts = append(fonts, font)
	}
	return fonts, nil
}

func (f *FontFamily) snapshotSelection() ([]string, map[string]int, []string, error) {
	f.mu.RLock()
	if len(f.fonts) == 0 {
		f.mu.RUnlock()
		return nil, nil, nil, ErrNoFontsInFamily
	}
	fontFiles := append([]string(nil), f.fonts...)
	weights := make(map[string]int, len(f.weights))
	for path, weight := range f.weights {
		weights[path] = weight
	}
	fallbackFonts := append([]string(nil), f.fallbackFonts...)
	f.mu.RUnlock()
	return fontFiles, weights, fallbackFonts, nil
}

func (f *FontFamily) loadCachedFont(fontFile string) (*truetype.Font, error) {
	if v, ok := f.fontCache.Load(fontFile); ok {
		return v.(*truetype.Font), nil
	}
	fontFace, err := f.parseFont(fontFile)
	if err != nil {
		return nil, err
	}
	f.fontCache.Store(fontFile, fontFace)
	return fontFace, nil
}

func chooseWeightedFontFile(r *rand.Rand, fontFiles []string, weights map[string]int) string {
	if len(fontFiles) == 0 {
		return ""
	}
	totalWeight := 0
	for _, fontFile := range fontFiles {
		weight := weights[fontFile]
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
	}
	if totalWeight <= 0 {
		return fontFiles[0]
	}

	target := randIntn(r, totalWeight)
	for _, fontFile := range fontFiles {
		weight := weights[fontFile]
		if weight <= 0 {
			weight = 1
		}
		target -= weight
		if target < 0 {
			return fontFile
		}
	}
	return fontFiles[len(fontFiles)-1]
}

func buildFallbackCandidates(primary string, fallbackFonts []string, allFonts []string) []string {
	seen := make(map[string]struct{}, len(allFonts)+len(fallbackFonts)+1)
	out := make([]string, 0, len(allFonts)+len(fallbackFonts)+1)

	add := func(fontFile string) {
		if fontFile == "" {
			return
		}
		if _, ok := seen[fontFile]; ok {
			return
		}
		seen[fontFile] = struct{}{}
		out = append(out, fontFile)
	}

	add(primary)
	for _, fontFile := range fallbackFonts {
		add(fontFile)
	}
	for _, fontFile := range allFonts {
		add(fontFile)
	}
	return out
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
	f.mu.RLock()
	for _, existing := range f.fonts {
		if existing == fontFile {
			f.mu.RUnlock()
			return nil
		}
	}
	f.mu.RUnlock()

	if _, ok := f.fontCache.Load(fontFile); ok {
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, existing := range f.fonts {
			if existing == fontFile {
				return nil
			}
		}
		f.fonts = append(f.fonts, fontFile)
		if f.weights == nil {
			f.weights = make(map[string]int)
		}
		if f.weights[fontFile] <= 0 {
			f.weights[fontFile] = 1
		}
		return nil
	}

	font, err := f.parseFont(fontFile)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.fonts {
		if existing == fontFile {
			return nil
		}
	}
	f.fonts = append(f.fonts, fontFile)
	if f.weights == nil {
		f.weights = make(map[string]int)
	}
	if f.weights[fontFile] <= 0 {
		f.weights[fontFile] = 1
	}
	f.fontCache.Store(fontFile, font)
	return nil
}

// SetFontWeight configures weighted random probability for a specific font.
func (f *FontFamily) SetFontWeight(fontFile string, weight int) error {
	if weight <= 0 {
		return errors.New("font weight must be greater than 0")
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.fonts {
		if existing == fontFile {
			if f.weights == nil {
				f.weights = make(map[string]int)
			}
			f.weights[fontFile] = weight
			return nil
		}
	}
	return os.ErrNotExist
}

// SetFallbackFonts sets ordered fallback fonts used when preferred font loading fails.
func (f *FontFamily) SetFallbackFonts(fontFiles ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(fontFiles) == 0 {
		f.fallbackFonts = nil
		return nil
	}

	known := make(map[string]struct{}, len(f.fonts))
	for _, fontFile := range f.fonts {
		known[fontFile] = struct{}{}
	}

	out := make([]string, 0, len(fontFiles))
	seen := make(map[string]struct{}, len(fontFiles))
	for _, fontFile := range fontFiles {
		if _, ok := known[fontFile]; !ok {
			return os.ErrNotExist
		}
		if _, ok := seen[fontFile]; ok {
			continue
		}
		seen[fontFile] = struct{}{}
		out = append(out, fontFile)
	}
	f.fallbackFonts = out
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
		weights:   make(map[string]int),
		r:         newSecureSeededRand(),
	}
}
