package main

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"
)

func init() {
	geoJSONPath = "countries.geo.json"
}

func TestGenerateMapImageEmpty(t *testing.T) {
	buf, err := generateMapImage([]string{})
	if err != nil {
		t.Fatalf("generateMapImage with empty list failed: %v", err)
	}
	if buf == nil || buf.Len() == 0 {
		t.Fatal("expected non-empty PNG bytes")
	}
	assertValidPNG(t, buf, 1024, 512)
}

func TestGenerateMapImageVisited(t *testing.T) {
	buf, err := generateMapImage([]string{"Germany", "France"})
	if err != nil {
		t.Fatalf("generateMapImage with visited countries failed: %v", err)
	}
	if buf == nil || buf.Len() == 0 {
		t.Fatal("expected non-empty PNG bytes")
	}
	assertValidPNG(t, buf, 1024, 512)
}

func assertValidPNG(t *testing.T, buf *bytes.Buffer, wantW, wantH int) {
	t.Helper()
	img, format, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("output is not a valid image: %v", err)
	}
	if format != "png" {
		t.Errorf("expected png format, got %s", format)
	}
	bounds := img.Bounds()
	if bounds.Dx() != wantW || bounds.Dy() != wantH {
		t.Errorf("expected %dx%d image, got %dx%d", wantW, wantH, bounds.Dx(), bounds.Dy())
	}
}
