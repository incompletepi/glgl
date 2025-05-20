package ms2

import (
	"testing"

	math "github.com/chewxy/math32"
)

var testoffsets = []Vec{{-1, -2}, {-2, 1}, {2, -1}, {}, {1, 0}, {0, 1}, {1, 1}}

func TestPolygon_circle_smoothing(t *testing.T) {
	var poly PolygonBuilder
	for _, offset := range testoffsets {
		for _, radius := range []float32{1e-6, 0.1, 1, 5, 100} {
			r := radius
			smooth := radius
			for _, facets := range []int{2, 3, 4, 6, 7} {
				poly.Reset()

				poly.AddXY(offset.X+r, offset.Y+0)
				poly.AddXY(offset.X+r, offset.Y+r).Smooth(smooth, facets)
				poly.AddXY(offset.X+0, offset.Y+r)
				poly.AddXY(offset.X-r, offset.Y+r).Smooth(smooth, facets)
				poly.AddXY(offset.X-r, offset.Y+0)
				poly.AddXY(offset.X-r, offset.Y-r).Smooth(smooth, facets)
				poly.AddXY(offset.X+0, offset.Y-r)
				poly.AddXY(offset.X+r, offset.Y-r).Smooth(smooth, facets)

				verts, err := poly.AppendVecs(nil)
				if err != nil {
					t.Fatal(err, offset, r)
				}

				const tol = 1e-4
				wantVerts := 4 + (facets-1)*4
				if len(verts) != wantVerts {
					t.Errorf("want %d vertices, got %d", wantVerts, len(verts))
				}
				for _, v := range verts {
					vr := Sub(v, offset)
					gotR := Norm(vr)
					if math.IsNaN(gotR) {
						t.Error("NaN", v)
					}
					diff := gotR - r
					if math.Abs(diff) > tol || math.IsNaN(diff) {
						t.Error(v)
					}
				}
			}
		}
	}
}

func TestPolygon_circle_arcing(t *testing.T) {
	var poly PolygonBuilder
	for _, offset := range testoffsets {
		for _, radius := range []float32{0.1, 1, 5, 100} {
			r := radius
			for _, facets := range []int{2, 3, 4, 6, 7} {
				poly.Reset()
				poly.AddXY(offset.X+r, offset.Y).Arc(r, facets)
				poly.AddXY(offset.X-r, offset.Y).Arc(r, facets)

				verts, err := poly.AppendVecs(nil)
				if err != nil {
					t.Fatal(err, offset, radius)
				}

				const tol = 1e-4
				wantVerts := 2 + (facets-1)*2
				if len(verts) != wantVerts {
					t.Errorf("want %d vertices, got %d", wantVerts, len(verts))
				}
				for _, v := range verts {
					vr := Sub(v, offset)
					gotR := Norm(vr)
					if math.Abs(gotR-r) > tol {
						t.Error(offset, v)
					}
				}
			}
		}
	}
}

func TestSmooth_radiusLimitBug(t *testing.T) {
	var poly PolygonBuilder
	poly.AddXY(0.8, 0.6)
	poly.AddXY(0.4, 0.6).Smooth(0.2, 5)
	poly.AddXY(0.3, 0)
	vecs, err := poly.AppendVecs(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(vecs)
}

func TestArc_radiusLimitBug(t *testing.T) {
	var poly PolygonBuilder
	poly.AddXY(1, 0)
	poly.AddXY(0, 38).Arc(380, 2)
	vecs, err := poly.AppendVecs(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(vecs)
}

func TestPolygon_IsClockwise(t *testing.T) {
	var tests = []struct {
		verts  []Vec
		wantCW bool
	}{
		{ // Counterclockwise triangle.
			verts:  []Vec{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 0}},
			wantCW: false,
		},
		{ // Clockwise triangle.
			verts:  []Vec{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}},
			wantCW: true,
		},
	}
	var poly PolygonBuilder
	for _, test := range tests {
		poly.Reset()
		for _, v := range test.verts {
			poly.Add(v)
		}
		gotCW := poly.IsClockwise()
		if test.wantCW != gotCW {
			t.Errorf("want CW=%v got CW=%v", test.wantCW, gotCW)
		}
	}
}

func TestArc_invalidArc(t *testing.T) {
	var cases = []struct {
		start, end, center Vec
	}{
		// 180 degree:
		// {start: Vec{X: 165.36844, Y: 125.63215}, end: Vec{X: 165.59346, Y: 125.85769}, center: Vec{X: 165.48108, Y: 125.74479}},
		//
		{start: Vec{X: 168.19885, Y: 129.9802}, end: Vec{X: 167.97331, Y: 129.75517}, center: Vec{X: 168.08597, Y: 129.86781}},
		{start: Vec{X: 135.1107, Y: 116.67478}, end: Vec{X: 135.1107, Y: 116.67478}, center: Vec{X: 135.10947, Y: 116.673546}},
		{start: Vec{X: -1.05, Y: 149.07}, end: Vec{X: -1.12, Y: 148.7}, center: Vec{X: -1.0500132, Y: 149}},
	}
	var poly PolygonBuilder
	for _, test := range cases {
		start, end, center := test.start, test.end, test.center
		radius1 := Norm(Sub(start, center))
		radius2 := Norm(Sub(end, center))
		if math.Abs(radius1-radius2)/radius1 > 0.001 {
			t.Log("start/end not equidistant from center:", radius1, radius2)
		}
		poly.Reset()
		poly.Add(start)
		poly.Add(end).Arc(radius1, 3)
		vecs, err := poly.AppendVecs(nil)
		if err == nil {
			t.Error("expected invalid input")
			for _, v := range vecs {
				if v != v {
					t.Error("effectively got NaNs")
				}
				gotRadius := Norm(Sub(v, center))
				diffcenter := math.Abs(gotRadius - radius1)
				if diffcenter/radius1 > 0.01 {
					t.Errorf("bad radius got=%f want=%f", gotRadius, radius1)
				}
			}
		}
	}
}
