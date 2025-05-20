package ms3

import (
	"fmt"
	"math"
	"testing"
)

func TestRotation(t *testing.T) {
	const tol = 1e-7
	v := Vec{X: 1}
	y90 := RotatingMat4(math.Pi/2, Vec{Y: 1})
	got := y90.MulPosition(v)
	want := Vec{Z: -1}
	if !EqualElem(got, want, tol) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestSVD(t *testing.T) {
	const tol = 1e-6
	a := mat3(-0.558253, -0.0461681, -0.505735, -0.411397, 0.0365854, 0.199707, 0.285389, -0.313789, 0.200189)
	// Using approximate sqrt:
	// Accurate sqrt values:
	uwant := mat3(-0.849310, -0.354882, -0.390809, -0.278376, 0.930100, -0.239626, 0.448530, -0.094725, -0.888734)
	swant := mat3(0.860883, -0.000000, 0.000000, 0.000000, 0.413613, -0.000000, -0.000000, 0.000000, -0.296320)
	vwant := mat3(0.832469, -0.511493, -0.213002, -0.129771, 0.193746, -0.972431, 0.538660, 0.837160, 0.094911)
	u, s, v := a.SVD()

	if !EqualMat3(u, uwant, tol) {
		t.Error("U mismatch")
		fmt.Println("U got:")
		printMat(u)
		fmt.Println("U want:")
		printMat(uwant)
	}
	if !EqualMat3(s, swant, tol) {
		t.Error("S mismatch")
		fmt.Println("S got:")
		printMat(s)
		fmt.Println("S want:")
		printMat(swant)
	}
	if !EqualMat3(v, vwant, tol) {
		t.Error("V mismatch")
		fmt.Println("V got:")
		printMat(v)
		fmt.Println("V want:")
		printMat(vwant)
	}
}

func printMat(a Mat3) {
	fmt.Printf("%f %f %f \n", a.x00, a.x01, a.x02)
	fmt.Printf("%f %f %f \n", a.x10, a.x11, a.x12)
	fmt.Printf("%f %f %f \n", a.x20, a.x21, a.x22)
}
