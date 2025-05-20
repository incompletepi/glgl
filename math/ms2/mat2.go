package ms2

import (
	math "github.com/chewxy/math32"
)

// Mat2 is a 2x2 matrix.
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
type Mat2 struct {
	x00, x01 float32
	x10, x11 float32
}

// NewMat2 instantiates a new matrix from the first 4 floats, row major order. If v is of insufficient length NewMat2 panics.
func NewMat2(v []float32) (m Mat2) {
	_ = v[3]
	m.x00, m.x01, m.x10, m.x11 = v[0], v[1], v[2], v[3]
	return m
}

// IdentityMat2 returns the 2x2 identity matrix.
func IdentityMat2() Mat2 {
	return Mat2{
		1, 0,
		0, 1}
}

// EqualMat2 tests the equality of 2x2 matrices.
func EqualMat2(a, b Mat2, tolerance float32) bool {
	return math.Abs(a.x00-b.x00) < tolerance &&
		math.Abs(a.x01-b.x01) < tolerance &&
		math.Abs(a.x10-b.x10) < tolerance &&
		math.Abs(a.x11-b.x11) < tolerance
}

// MulMat2 multiplies two 2x2 matrices.
func MulMat2(a, b Mat2) Mat2 {
	return Mat2{
		a.x00*b.x00 + a.x01*b.x10,
		a.x00*b.x01 + a.x01*b.x11,
		a.x10*b.x00 + a.x11*b.x10,
		a.x10*b.x01 + a.x11*b.x11,
	}
}

// AddMat2 adds two 2x2 matrices together.
func AddMat2(a, b Mat2) Mat2 {
	return Mat2{
		x00: a.x00 + b.x00,
		x10: a.x10 + b.x10,
		x01: a.x01 + b.x01,
		x11: a.x11 + b.x11,
	}
}

// Prod performs vector multiplication as if they were matrices
//
//	m = v1 * v2ᵀ
func Prod(v1, v2t Vec) Mat2 {
	return Mat2{
		v1.X * v2t.X, v1.X * v2t.Y,
		v1.Y * v2t.X, v1.Y * v2t.Y,
	}
}

// MulMatVec performs matrix multiplication on v:
//
//	result = M * v
func MulMatVec(m Mat2, v Vec) (result Vec) {
	result.X = v.X*m.x00 + v.Y*m.x01
	result.Y = v.X*m.x10 + v.Y*m.x11
	return result
}

// MulMatVecTrans Performs transposed matrix multiplication on v:
//
//	result = Mᵀ * v
func MulMatVecTrans(m Mat2, v Vec) (result Vec) {
	result.X = v.X*m.x00 + v.Y*m.x10
	result.Y = v.X*m.x01 + v.Y*m.x11
	return result
}

// ScaleMat2 multiplies each 2x2 matrix component by a scalar.
func ScaleMat2(a Mat2, k float32) Mat2 {
	return Mat2{
		x00: k * a.x00,
		x10: k * a.x10,
		x01: k * a.x01,
		x11: k * a.x11,
	}
}

// Determinant returns the determinant of a 2x2 matrix.
func (a Mat2) Determinant() float32 {
	return a.x00*a.x11 - a.x10*a.x01
}

// Transpose returns the transpose of a.
func (a Mat2) Transpose() Mat2 {
	return Mat2{
		x00: a.x00, x01: a.x10,
		x10: a.x01, x11: a.x11,
	}
}

// Inverse returns the inverse of a 2x2 matrix.
func (a Mat2) Inverse() Mat2 {
	m := Mat2{}
	det := a.Determinant()
	if det == 0 {
		return Mat2{math.NaN(), math.NaN(), math.NaN(), math.NaN()}
	}
	d := 1.0 / det
	m.x00 = a.x11 * d
	m.x01 = -a.x01 * d
	m.x10 = -a.x10 * d
	m.x11 = a.x00 * d
	return m
}

// VecRow returns the ith row as a Vec. VecRow panics if i is not 0 or 1.
func (m Mat2) VecRow(i int) Vec {
	switch i {
	case 0:
		return Vec{X: m.x00, Y: m.x01}
	case 1:
		return Vec{X: m.x10, Y: m.x11}
	}
	panic("out of bounds")
}

// VecCol returns the jth column as a Vec. VecCol panics if j is not 0 or 1.
func (m Mat2) VecCol(j int) Vec {
	switch j {
	case 0:
		return Vec{X: m.x00, Y: m.x10}
	case 1:
		return Vec{X: m.x01, Y: m.x11}
	}
	panic("out of bounds")
}

// Put stores the matrix values into slice b in row major order. If b is not of length 4 or greater Put panics.
func (m Mat2) Put(b []float32) {
	_ = b[3]
	b[0], b[1], b[2], b[3] = m.x00, m.x01, m.x10, m.x11
}

// Array returns the matrix values in a static array copy in row major order.
func (m Mat2) Array() (rowmajor [4]float32) {
	m.Put(rowmajor[:])
	return rowmajor
}

// Rotate returns an orthographic 2x2 rotation matrix (right hand rule).
func RotationMat2(a float32) Mat2 {
	s, c := math.Sincos(a)
	return Mat2{
		c, -s,
		s, c,
	}
}
