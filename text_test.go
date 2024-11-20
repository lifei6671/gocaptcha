package gocaptcha

import (
	"image"
	"image/draw"
	"math/rand"
	"testing"
)

func Test_textDrawer_DrawString(t *testing.T) {
	type args struct {
		canvas draw.Image
		text   string
	}
	tests := []struct {
		name    string
		t       *textDrawer
		args    args
		wantErr bool
	}{
		{
			name: "Successful DrawString",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "Hello, World!",
			},
			wantErr: false,
		},
		{
			name: "DrawString with empty text",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "",
			},
			wantErr: true,
		},
		{
			name: "DrawString with nil canvas",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: nil,
				text:   "Hello, World!",
			},
			wantErr: true,
		},
	}
	err := DefaultFontFamily.AddFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawString(tt.args.canvas, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("textDrawer.DrawString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_twistTextDrawer_DrawString(t *testing.T) {
	type args struct {
		canvas draw.Image
		text   string
	}
	tests := []struct {
		name    string
		t       *twistTextDrawer
		args    args
		wantErr bool
	}{
		{
			name: "Successful DrawString",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "Hello, World!",
			},
			wantErr: false,
		},
		{
			name: "DrawString with empty text",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "",
			},
			wantErr: true,
		},
		{
			name: "DrawString with nil canvas",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: nil,
				text:   "Hello, World!",
			},
			wantErr: true,
		},
	}
	err := DefaultFontFamily.AddFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawString(tt.args.canvas, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("twistTextDrawer.DrawString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
