package main

import (
	"bytes"
	"image"
	_ "image/png"
	"math"
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

func TestGenerateMapImageBorders(t *testing.T) {
	buf, err := generateMapImage([]string{"Brazil"})
	if err != nil {
		t.Fatalf("generateMapImage failed: %v", err)
	}
	if buf == nil || buf.Len() == 0 {
		t.Fatal("expected non-empty PNG bytes")
	}
	assertValidPNG(t, buf, 1024, 512)
}

func TestMercatorY(t *testing.T) {
	tests := []struct {
		lat   float64
		wantY float64
		eps   float64
	}{
		{0, 0.0, 1e-10},
		{45, 0.8813736, 1e-6},
		{60, 1.3169579, 1e-6},
		{-45, -0.8813736, 1e-6},
	}
	for _, tt := range tests {
		got := mercatorY(tt.lat)
		if math.Abs(got-tt.wantY) > tt.eps {
			t.Errorf("mercatorY(%v) = %v, want %v", tt.lat, got, tt.wantY)
		}
	}
	// Clamping: latitudes beyond ±85 should return the same value as ±85
	if mercatorY(90) != mercatorY(85) {
		t.Error("mercatorY(90) should equal mercatorY(85) due to clamping")
	}
	if mercatorY(-90) != mercatorY(-85) {
		t.Error("mercatorY(-90) should equal mercatorY(-85) due to clamping")
	}
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
