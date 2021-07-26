package font

import (
	_ "embed"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed ponged-pong.ttf
var pongFontData []byte

func Load() font.Face {
	f, err := truetype.Parse(pongFontData)
	if err != nil {
		panic(err)
	}

	return truetype.NewFace(f, &truetype.Options{
		Size:              48,
		GlyphCacheEntries: 1,
	})
}
