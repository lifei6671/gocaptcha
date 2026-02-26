package gocaptcha

import (
	"testing"
)

func TestFontFamily_Random(t *testing.T) {
	type args struct {
	}
	fontFamily := NewFontFamily()
	tests := []struct {
		name     string
		t        *FontFamily
		args     args
		fn       func(*FontFamily) error
		wantErr  bool
		wantFont bool
	}{
		{
			name: "test1",
			t:    fontFamily,
			args: args{},
			fn: func(family *FontFamily) error {
				return nil
			},
			wantErr:  true,
			wantFont: false,
		},
		{
			name: "test2",
			t:    fontFamily,
			args: args{},
			fn: func(family *FontFamily) error {
				family.fonts = append(family.fonts, "./testdata/Hiragino Sans GB.ttc")
				return nil
			},
			wantErr:  true,
			wantFont: false,
		},
		{
			name: "test3",
			t:    fontFamily,
			args: args{},
			fn: func(family *FontFamily) error {
				family.fonts = []string{"./fonts/3Dumb.ttf"}
				return nil
			},
			wantErr:  false,
			wantFont: true,
		},
		{
			name: "test4",
			t:    fontFamily,
			args: args{},
			fn: func(family *FontFamily) error {
				return family.AddFont("./fonts/3Dumb.ttf")
			},
			wantErr:  false,
			wantFont: true,
		},
		{
			name: "test5",
			t:    fontFamily,
			args: args{},
			fn: func(family *FontFamily) error {
				return family.AddFont("./fonts/Comismsh.ttf")
			},
			wantErr:  false,
			wantFont: true,
		},
		{
			name: "test6",
			t:    fontFamily,
			args: args{},
			fn: func(family *FontFamily) error {
				return family.AddFont("./testdata/Hiragino Sans GB.ttc")
			},
			wantErr:  false,
			wantFont: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.fn(tt.t)

			if font, err := tt.t.Random(); (font == nil) == tt.wantFont || (err != nil) != tt.wantErr {
				t.Errorf("FontFamily.Random() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetFonts(t *testing.T) {
	type args struct {
		fonts []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				fonts: []string{"./fonts/3Dumb.ttf"},
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				fonts: []string{"./testdata/Hiragino Sans GB.ttc"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetFonts(tt.args.fonts...); (err != nil) != tt.wantErr {
				t.Errorf("SetFonts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFontFamily_AddFontPath(t *testing.T) {
	type args struct {
		dirPath string
	}
	fontFamily := NewFontFamily()
	tests := []struct {
		name    string
		t       *FontFamily
		args    args
		wantErr bool
	}{
		{
			name:    "test1",
			t:       fontFamily,
			args:    args{dirPath: "./fonts"},
			wantErr: false,
		},
		{
			name:    "test2",
			t:       fontFamily,
			args:    args{dirPath: "./testdata/not_exist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.AddFontPath(tt.args.dirPath); (err != nil) != tt.wantErr {
				t.Errorf("FontFamily.AddFontPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFontFamily_SetFontWeight(t *testing.T) {
	family := NewFontFamily()
	if err := family.AddFont("./fonts/3Dumb.ttf"); err != nil {
		t.Fatal(err)
	}

	if err := family.SetFontWeight("./fonts/3Dumb.ttf", 5); err != nil {
		t.Fatalf("SetFontWeight() error = %v", err)
	}
	if err := family.SetFontWeight("./fonts/not-exist.ttf", 2); err == nil {
		t.Fatal("SetFontWeight() expected error for missing font")
	}
	if err := family.SetFontWeight("./fonts/3Dumb.ttf", 0); err == nil {
		t.Fatal("SetFontWeight() expected error for non-positive weight")
	}
}

func TestFontFamily_RandomWithFallback(t *testing.T) {
	family := NewFontFamily()
	family.fonts = []string{"./testdata/not-exist.ttf", "./fonts/3Dumb.ttf"}
	family.weights["./testdata/not-exist.ttf"] = 100
	family.weights["./fonts/3Dumb.ttf"] = 1
	if err := family.SetFallbackFonts("./fonts/3Dumb.ttf"); err != nil {
		t.Fatalf("SetFallbackFonts() error = %v", err)
	}

	fontFace, err := family.RandomWithFallback()
	if err != nil {
		t.Fatalf("RandomWithFallback() error = %v", err)
	}
	if fontFace == nil {
		t.Fatal("RandomWithFallback() returned nil font")
	}
}
