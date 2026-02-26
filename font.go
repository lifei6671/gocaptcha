// Package gocaptcha 提供生成验证码图片的功能
// 本文件实现了字体管理相关的功能，用于管理验证码中使用的字体
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

// DefaultFontFamily 默认字体家族
// 用于存储和管理验证码中使用的字体
var DefaultFontFamily = NewFontFamily()

// ErrNoFontsInFamily 字体家族中没有字体的错误
var ErrNoFontsInFamily = os.ErrNotExist

// SetFonts 设置默认字体家族
// 参数：
// - fonts: 字体文件路径列表
// 返回值：
// - error: 设置过程中的错误
// 影响：将指定的字体文件添加到默认字体家族中
// 示例：
// ```go
// // 设置默认字体家族
// err := gocaptcha.SetFonts("/path/to/font1.ttf", "/path/to/font2.ttf")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ```
func SetFonts(fonts ...string) error {
	for _, font := range fonts {
		if err := DefaultFontFamily.AddFont(font); err != nil {
			return err
		}
	}
	return nil
}

// SetFontPath 从目录设置默认字体家族
// 参数：
// - fontDirPath: 字体目录路径
// 返回值：
// - error: 设置过程中的错误
// 影响：将指定目录中的所有 .ttf 文件添加到默认字体家族中
// 示例：
// ```go
// // 从目录设置默认字体家族
// err := gocaptcha.SetFontPath("/path/to/fonts")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ```
func SetFontPath(fontDirPath string) error {
	return DefaultFontFamily.AddFontPath(fontDirPath)
}

// FontFamily 字体家族
// 用于管理一组字体，支持权重、缓存和 fallback 机制
// 字段：
// - fonts: 字体文件路径列表
// - fontCache: 字体缓存，存储已解析的字体
// - weights: 字体权重映射，用于加权随机选择
// - fallbackFonts:  fallback 字体列表，当首选字体加载失败时使用
// - r: 随机数生成器
// - mu: 读写锁，保护字体列表和权重
// - randMu: 随机数生成器的互斥锁
// - randOnce: 确保随机数生成器只初始化一次

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

// maxFontWeightMultiplier 最大字体权重乘数
// 影响：限制字体的最大权重，防止单个字体权重过大
const maxFontWeightMultiplier = 32

// Random 从字体家族中随机选择一个字体
// 返回值：
// - *truetype.Font: 随机选择的字体
// - error: 选择过程中的错误
// 影响：从字体家族中随机选择一个字体，使用 fallback 机制
func (f *FontFamily) Random() (*truetype.Font, error) {
	return f.RandomWithFallback()
}

// RandomWithFallback 从字体家族中加权随机选择一个字体，如果失败则使用 fallback
// 返回值：
// - *truetype.Font: 选择的字体
// - error: 选择过程中的错误
// 影响：根据字体权重随机选择字体，当首选字体加载失败时使用 fallback 字体
func (f *FontFamily) RandomWithFallback() (*truetype.Font, error) {
	// 获取字体选择的快照
	fontFiles, weights, fallbackFonts, err := f.snapshotSelection()
	if err != nil {
		return nil, err
	}

	// 确保随机数生成器已初始化
	f.randOnce.Do(func() {
		if f.r == nil {
			f.r = newSecureSeededRand()
		}
	})

	// 加权随机选择字体文件
	f.randMu.Lock()
	selected := chooseWeightedFontFile(f.r, fontFiles, weights)
	f.randMu.Unlock()

	// 构建 fallback 候选字体
	candidates := buildFallbackCandidates(selected, fallbackFonts, fontFiles)
	var firstErr error
	// 尝试加载候选字体
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

// WeightedCachedFonts 返回按权重扩展的字体，用于高效的加权随机选择
// 返回值：
// - []*truetype.Font: 按权重扩展的字体列表
// - error: 处理过程中的错误
// 影响：返回一个字体列表，其中每个字体根据其权重出现多次，便于后续的随机选择
func (f *FontFamily) WeightedCachedFonts() ([]*truetype.Font, error) {
	// 获取字体选择的快照
	fontFiles, weights, fallbackFonts, err := f.snapshotSelection()
	if err != nil {
		return nil, err
	}

	// 构建 fallback 候选字体
	candidates := buildFallbackCandidates("", fallbackFonts, fontFiles)
	loaded := make(map[string]*truetype.Font, len(candidates))
	var firstErr error
	// 加载候选字体
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

	// 计算总权重
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

	// 按权重扩展字体列表
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

	// 如果没有字体，使用 fallback 字体
	if len(out) == 0 {
		for _, fontFile := range fallbackFonts {
			if fontFace, ok := loaded[fontFile]; ok {
				out = append(out, fontFace)
			}
		}
	}
	// 如果仍然没有字体，使用所有加载的字体
	if len(out) == 0 {
		for _, fontFace := range loaded {
			out = append(out, fontFace)
		}
	}
	// 如果仍然没有字体，返回错误
	if len(out) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, ErrNoFontsInFamily
	}
	return out, nil
}

// CachedFonts 返回字体家族中已解析字体的快照
// 返回值：
// - []*truetype.Font: 已解析的字体列表
// - error: 处理过程中的错误
// 影响：返回字体家族中所有字体的解析结果，未解析的字体将被解析并缓存
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
		// 尝试从缓存中加载
		if v, ok := f.fontCache.Load(fontFile); ok {
			fonts = append(fonts, v.(*truetype.Font))
			continue
		}
		// 解析字体
		font, err := f.parseFont(fontFile)
		if err != nil {
			return nil, err
		}
		// 缓存字体
		f.fontCache.Store(fontFile, font)
		fonts = append(fonts, font)
	}
	return fonts, nil
}

// snapshotSelection 获取字体选择的快照
// 返回值：
// - []string: 字体文件路径列表
// - map[string]int: 字体权重映射
// - []string: fallback 字体列表
// - error: 处理过程中的错误
// 影响：获取字体家族的当前状态快照，避免并发修改问题
func (f *FontFamily) snapshotSelection() ([]string, map[string]int, []string, error) {
	f.mu.RLock()
	if len(f.fonts) == 0 {
		f.mu.RUnlock()
		return nil, nil, nil, ErrNoFontsInFamily
	}
	// 复制字体文件路径列表
	fontFiles := append([]string(nil), f.fonts...)
	// 复制字体权重映射
	weights := make(map[string]int, len(f.weights))
	for path, weight := range f.weights {
		weights[path] = weight
	}
	// 复制 fallback 字体列表
	fallbackFonts := append([]string(nil), f.fallbackFonts...)
	f.mu.RUnlock()
	return fontFiles, weights, fallbackFonts, nil
}

// loadCachedFont 加载缓存的字体
// 参数：
// - fontFile: 字体文件路径
// 返回值：
// - *truetype.Font: 解析后的字体
// - error: 加载过程中的错误
// 影响：从缓存中加载字体，如果缓存中没有则解析并缓存
func (f *FontFamily) loadCachedFont(fontFile string) (*truetype.Font, error) {
	// 尝试从缓存中加载
	if v, ok := f.fontCache.Load(fontFile); ok {
		return v.(*truetype.Font), nil
	}
	// 解析字体
	fontFace, err := f.parseFont(fontFile)
	if err != nil {
		return nil, err
	}
	// 缓存字体
	f.fontCache.Store(fontFile, fontFace)
	return fontFace, nil
}

// chooseWeightedFontFile 选择加权字体文件
// 参数：
// - r: 随机数生成器
// - fontFiles: 字体文件路径列表
// - weights: 字体权重映射
// 返回值：
// - string: 选择的字体文件路径
// 影响：根据字体权重随机选择一个字体文件
func chooseWeightedFontFile(r *rand.Rand, fontFiles []string, weights map[string]int) string {
	if len(fontFiles) == 0 {
		return ""
	}
	// 计算总权重
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

	// 随机选择
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

// buildFallbackCandidates 构建 fallback 候选字体
// 参数：
// - primary: 首选字体文件路径
// - fallbackFonts: fallback 字体文件路径列表
// - allFonts: 所有字体文件路径列表
// 返回值：
// - []string: 候选字体文件路径列表
// 影响：构建一个无重复的候选字体列表，顺序为首选字体、fallback 字体、所有字体
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

// parseFont 解析字体文件
// 参数：
// - fontFile: 字体文件路径
// 返回值：
// - *truetype.Font: 解析后的字体
// - error: 解析过程中的错误
// 影响：读取并解析字体文件，返回解析后的字体对象
func (f *FontFamily) parseFont(fontFile string) (*truetype.Font, error) {
	// 读取字体文件
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	// 解析字体
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return font, nil
}

// AddFont 向字体家族添加字体
// 参数：
// - fontFile: 字体文件路径
// 返回值：
// - error: 添加过程中的错误
// 影响：将指定的字体文件添加到字体家族中，如果字体已存在则忽略
// 示例：
// ```go
// // 向字体家族添加字体
// err := fontFamily.AddFont("/path/to/font.ttf")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ```
func (f *FontFamily) AddFont(fontFile string) error {
	// 检查字体是否已存在
	f.mu.RLock()
	for _, existing := range f.fonts {
		if existing == fontFile {
			f.mu.RUnlock()
			return nil
		}
	}
	f.mu.RUnlock()

	// 如果字体已在缓存中
	if _, ok := f.fontCache.Load(fontFile); ok {
		f.mu.Lock()
		defer f.mu.Unlock()
		// 再次检查字体是否已存在
		for _, existing := range f.fonts {
			if existing == fontFile {
				return nil
			}
		}
		// 添加字体
		f.fonts = append(f.fonts, fontFile)
		if f.weights == nil {
			f.weights = make(map[string]int)
		}
		if f.weights[fontFile] <= 0 {
			f.weights[fontFile] = 1
		}
		return nil
	}

	// 解析字体
	font, err := f.parseFont(fontFile)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	// 再次检查字体是否已存在
	for _, existing := range f.fonts {
		if existing == fontFile {
			return nil
		}
	}
	// 添加字体
	f.fonts = append(f.fonts, fontFile)
	if f.weights == nil {
		f.weights = make(map[string]int)
	}
	if f.weights[fontFile] <= 0 {
		f.weights[fontFile] = 1
	}
	// 缓存字体
	f.fontCache.Store(fontFile, font)
	return nil
}

// SetFontWeight 为特定字体配置加权随机概率
// 参数：
// - fontFile: 字体文件路径
// - weight: 字体权重，必须大于0
// 返回值：
// - error: 设置过程中的错误
// 影响：设置字体的权重，权重越大，被随机选中的概率越高
// 示例：
// ```go
// // 设置字体权重
// err := fontFamily.SetFontWeight("/path/to/font.ttf", 5)
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ```
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

// SetFallbackFonts 设置有序的 fallback 字体，当首选字体加载失败时使用
// 参数：
// - fontFiles: 字体文件路径列表
// 返回值：
// - error: 设置过程中的错误
// 影响：设置 fallback 字体列表，当首选字体加载失败时按顺序尝试这些字体
// 示例：
// ```go
// // 设置 fallback 字体
// err := fontFamily.SetFallbackFonts("/path/to/fallback1.ttf", "/path/to/fallback2.ttf")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ```
func (f *FontFamily) SetFallbackFonts(fontFiles ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(fontFiles) == 0 {
		f.fallbackFonts = nil
		return nil
	}

	// 检查字体是否在字体家族中
	known := make(map[string]struct{}, len(f.fonts))
	for _, fontFile := range f.fonts {
		known[fontFile] = struct{}{}
	}

	// 构建无重复的 fallback 字体列表
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

// AddFontPath 从给定目录添加所有 .ttf 文件到字体家族
// 参数：
// - dirPath: 目录路径
// 返回值：
// - error: 添加过程中的错误
// 影响：遍历目录及其子目录，添加所有 .ttf 文件到字体家族
// 示例：
// ```go
// // 从目录添加字体
// err := fontFamily.AddFontPath("/path/to/fonts")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ```
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

// NewFontFamily 创建一个新的字体家族
// 返回值：
// - *FontFamily: 新的字体家族实例
// 示例：
// ```go
// // 创建一个新的字体家族
// fontFamily := gocaptcha.NewFontFamily()
// // 添加字体
// err := fontFamily.AddFont("/path/to/font.ttf")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ```
func NewFontFamily() *FontFamily {
	return &FontFamily{
		fontCache: &sync.Map{},
		weights:   make(map[string]int),
		r:         newSecureSeededRand(),
	}
}
