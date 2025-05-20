package ms3

import (
	math "github.com/chewxy/math32"
	"github.com/incompletepi/glgl/math/ms1"
)

// Mat4 is a 4x4 matrix.
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
type Mat4 struct {
	x00, x01, x02, x03 float32
	x10, x11, x12, x13 float32
	x20, x21, x22, x23 float32
	x30, x31, x32, x33 float32
}

// NewMat4 instantiates a new 4x4 Mat4 matrix from the first 16 values in row major order.
// If v is shorter than 16 NewMat4 panics.
func NewMat4(v []float32) (m Mat4) {
	_ = v[15]
	m.x00, m.x01, m.x02, m.x03 = v[0], v[1], v[2], v[3]
	m.x10, m.x11, m.x12, m.x13 = v[4], v[5], v[6], v[7]
	m.x20, m.x21, m.x22, m.x23 = v[8], v[9], v[10], v[11]
	m.x30, m.x31, m.x32, m.x33 = v[12], v[13], v[14], v[15]
	return m
}

// IdentityMat4 returns the identity 4x4 matrix.
func IdentityMat4() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}
}

func nanMat4() Mat4 {
	return Mat4{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN()}
}

// TranslatingMat4 returns a 4x4 translation matrix.
func TranslatingMat4(v Vec) Mat4 {
	return Mat4{
		1, 0, 0, v.X,
		0, 1, 0, v.Y,
		0, 0, 1, v.Z,
		0, 0, 0, 1}
}

// ScalingMat4 returns a 4x4 scaling matrix.
// Scaling does not preserve distance. See: ScaleUniform3D()
func ScalingMat4(v Vec) Mat4 {
	return Mat4{
		v.X, 0, 0, 0,
		0, v.Y, 0, 0,
		0, 0, v.Z, 0,
		0, 0, 0, 1}
}

// RotatingMat4 returns an orthographic 4x4 rotation matrix (right hand rule).
func RotatingMat4(angleRadians float32, axis Vec) Mat4 {
	axis = Unit(axis)
	s, c := math.Sincos(angleRadians)
	m := 1 - c
	return Mat4{
		m*axis.X*axis.X + c, m*axis.X*axis.Y - axis.Z*s, m*axis.Z*axis.X + axis.Y*s, 0,
		m*axis.X*axis.Y + axis.Z*s, m*axis.Y*axis.Y + c, m*axis.Y*axis.Z - axis.X*s, 0,
		m*axis.Z*axis.X - axis.Y*s, m*axis.Y*axis.Z + axis.X*s, m*axis.Z*axis.Z + c, 0,
		0, 0, 0, 1,
	}
}

// MulMat4 multiplies two 4x4 matrices and returns the result.
func MulMat4(a, b Mat4) Mat4 {
	m := Mat4{}
	m.x00 = a.x00*b.x00 + a.x01*b.x10 + a.x02*b.x20 + a.x03*b.x30
	m.x10 = a.x10*b.x00 + a.x11*b.x10 + a.x12*b.x20 + a.x13*b.x30
	m.x20 = a.x20*b.x00 + a.x21*b.x10 + a.x22*b.x20 + a.x23*b.x30
	m.x30 = a.x30*b.x00 + a.x31*b.x10 + a.x32*b.x20 + a.x33*b.x30
	m.x01 = a.x00*b.x01 + a.x01*b.x11 + a.x02*b.x21 + a.x03*b.x31
	m.x11 = a.x10*b.x01 + a.x11*b.x11 + a.x12*b.x21 + a.x13*b.x31
	m.x21 = a.x20*b.x01 + a.x21*b.x11 + a.x22*b.x21 + a.x23*b.x31
	m.x31 = a.x30*b.x01 + a.x31*b.x11 + a.x32*b.x21 + a.x33*b.x31
	m.x02 = a.x00*b.x02 + a.x01*b.x12 + a.x02*b.x22 + a.x03*b.x32
	m.x12 = a.x10*b.x02 + a.x11*b.x12 + a.x12*b.x22 + a.x13*b.x32
	m.x22 = a.x20*b.x02 + a.x21*b.x12 + a.x22*b.x22 + a.x23*b.x32
	m.x32 = a.x30*b.x02 + a.x31*b.x12 + a.x32*b.x22 + a.x33*b.x32
	m.x03 = a.x00*b.x03 + a.x01*b.x13 + a.x02*b.x23 + a.x03*b.x33
	m.x13 = a.x10*b.x03 + a.x11*b.x13 + a.x12*b.x23 + a.x13*b.x33
	m.x23 = a.x20*b.x03 + a.x21*b.x13 + a.x22*b.x23 + a.x23*b.x33
	m.x33 = a.x30*b.x03 + a.x31*b.x13 + a.x32*b.x23 + a.x33*b.x33
	return m
}

// MulPosition multiplies a r3.Vec position with a rotate/translate matrix.
func (a Mat4) MulPosition(b Vec) Vec {
	return Vec{
		X: a.x00*b.X + a.x01*b.Y + a.x02*b.Z + a.x03,
		Y: a.x10*b.X + a.x11*b.Y + a.x12*b.Z + a.x13,
		Z: a.x20*b.X + a.x21*b.Y + a.x22*b.Z + a.x23}
}

// MulBox rotates/translates a 3d bounding box and resizes for axis-alignment.
func (a Mat4) MulBox(box Box) Box {
	// Below is equivalent code:
	// box2 := box
	// verts := box.Vertices()
	// for _, v := range verts {
	// 	v = a.MulPosition(v)
	// 	box2.Max = MaxElem(box2.Max, v)
	// 	box2.Min = MinElem(box2.Min, v)
	// }
	// return box2
	r := Vec{X: a.x00, Y: a.x10, Z: a.x20}
	u := Vec{X: a.x01, Y: a.x11, Z: a.x21}
	b := Vec{X: a.x02, Y: a.x12, Z: a.x22}
	t := Vec{X: a.x03, Y: a.x13, Z: a.x23}

	xa := Scale(box.Min.X, r)
	xb := Scale(box.Max.X, r)
	ya := Scale(box.Min.Y, u)
	yb := Scale(box.Max.Y, u)
	za := Scale(box.Min.Z, b)
	zb := Scale(box.Max.Z, b)
	xa, xb = MinElem(xa, xb), MaxElem(xa, xb)
	ya, yb = MinElem(ya, yb), MaxElem(ya, yb)
	za, zb = MinElem(za, zb), MaxElem(za, zb)
	newBox := Box{
		Min: Add(xa, Add(ya, Add(za, t))),
		Max: Add(xb, Add(yb, Add(zb, t))),
	}
	return newBox
}

// Determinant returns the determinant of a 4x4 matrix.
func (a Mat4) Determinant() float32 {
	return a.x00*a.x11*a.x22*a.x33 - a.x00*a.x11*a.x23*a.x32 +
		a.x00*a.x12*a.x23*a.x31 - a.x00*a.x12*a.x21*a.x33 +
		a.x00*a.x13*a.x21*a.x32 - a.x00*a.x13*a.x22*a.x31 -
		a.x01*a.x12*a.x23*a.x30 + a.x01*a.x12*a.x20*a.x33 -
		a.x01*a.x13*a.x20*a.x32 + a.x01*a.x13*a.x22*a.x30 -
		a.x01*a.x10*a.x22*a.x33 + a.x01*a.x10*a.x23*a.x32 +
		a.x02*a.x13*a.x20*a.x31 - a.x02*a.x13*a.x21*a.x30 +
		a.x02*a.x10*a.x21*a.x33 - a.x02*a.x10*a.x23*a.x31 +
		a.x02*a.x11*a.x23*a.x30 - a.x02*a.x11*a.x20*a.x33 -
		a.x03*a.x10*a.x21*a.x32 + a.x03*a.x10*a.x22*a.x31 -
		a.x03*a.x11*a.x22*a.x30 + a.x03*a.x11*a.x20*a.x32 -
		a.x03*a.x12*a.x20*a.x31 + a.x03*a.x12*a.x21*a.x30
}

// Transpose returns the transpose of a.
func (a Mat4) Transpose() Mat4 {
	return Mat4{
		x00: a.x00, x01: a.x10, x02: a.x20, x03: a.x30,
		x10: a.x01, x11: a.x11, x12: a.x21, x13: a.x31,
		x20: a.x02, x21: a.x12, x22: a.x22, x23: a.x32,
		x30: a.x03, x31: a.x13, x32: a.x23, x33: a.x33,
	}
}

// Inverse returns the inverse of a 4x4 matrix. Does not check for singularity.
func (a Mat4) Inverse() Mat4 {
	m := Mat4{}
	det := a.Determinant()
	if det == 0 {
		return nanMat4()
	}
	d := 1.0 / det
	m.x00 = (a.x12*a.x23*a.x31 - a.x13*a.x22*a.x31 + a.x13*a.x21*a.x32 - a.x11*a.x23*a.x32 - a.x12*a.x21*a.x33 + a.x11*a.x22*a.x33) * d
	m.x01 = (a.x03*a.x22*a.x31 - a.x02*a.x23*a.x31 - a.x03*a.x21*a.x32 + a.x01*a.x23*a.x32 + a.x02*a.x21*a.x33 - a.x01*a.x22*a.x33) * d
	m.x02 = (a.x02*a.x13*a.x31 - a.x03*a.x12*a.x31 + a.x03*a.x11*a.x32 - a.x01*a.x13*a.x32 - a.x02*a.x11*a.x33 + a.x01*a.x12*a.x33) * d
	m.x03 = (a.x03*a.x12*a.x21 - a.x02*a.x13*a.x21 - a.x03*a.x11*a.x22 + a.x01*a.x13*a.x22 + a.x02*a.x11*a.x23 - a.x01*a.x12*a.x23) * d
	m.x10 = (a.x13*a.x22*a.x30 - a.x12*a.x23*a.x30 - a.x13*a.x20*a.x32 + a.x10*a.x23*a.x32 + a.x12*a.x20*a.x33 - a.x10*a.x22*a.x33) * d
	m.x11 = (a.x02*a.x23*a.x30 - a.x03*a.x22*a.x30 + a.x03*a.x20*a.x32 - a.x00*a.x23*a.x32 - a.x02*a.x20*a.x33 + a.x00*a.x22*a.x33) * d
	m.x12 = (a.x03*a.x12*a.x30 - a.x02*a.x13*a.x30 - a.x03*a.x10*a.x32 + a.x00*a.x13*a.x32 + a.x02*a.x10*a.x33 - a.x00*a.x12*a.x33) * d
	m.x13 = (a.x02*a.x13*a.x20 - a.x03*a.x12*a.x20 + a.x03*a.x10*a.x22 - a.x00*a.x13*a.x22 - a.x02*a.x10*a.x23 + a.x00*a.x12*a.x23) * d
	m.x20 = (a.x11*a.x23*a.x30 - a.x13*a.x21*a.x30 + a.x13*a.x20*a.x31 - a.x10*a.x23*a.x31 - a.x11*a.x20*a.x33 + a.x10*a.x21*a.x33) * d
	m.x21 = (a.x03*a.x21*a.x30 - a.x01*a.x23*a.x30 - a.x03*a.x20*a.x31 + a.x00*a.x23*a.x31 + a.x01*a.x20*a.x33 - a.x00*a.x21*a.x33) * d
	m.x22 = (a.x01*a.x13*a.x30 - a.x03*a.x11*a.x30 + a.x03*a.x10*a.x31 - a.x00*a.x13*a.x31 - a.x01*a.x10*a.x33 + a.x00*a.x11*a.x33) * d
	m.x23 = (a.x03*a.x11*a.x20 - a.x01*a.x13*a.x20 - a.x03*a.x10*a.x21 + a.x00*a.x13*a.x21 + a.x01*a.x10*a.x23 - a.x00*a.x11*a.x23) * d
	m.x30 = (a.x12*a.x21*a.x30 - a.x11*a.x22*a.x30 - a.x12*a.x20*a.x31 + a.x10*a.x22*a.x31 + a.x11*a.x20*a.x32 - a.x10*a.x21*a.x32) * d
	m.x31 = (a.x01*a.x22*a.x30 - a.x02*a.x21*a.x30 + a.x02*a.x20*a.x31 - a.x00*a.x22*a.x31 - a.x01*a.x20*a.x32 + a.x00*a.x21*a.x32) * d
	m.x32 = (a.x02*a.x11*a.x30 - a.x01*a.x12*a.x30 - a.x02*a.x10*a.x31 + a.x00*a.x12*a.x31 + a.x01*a.x10*a.x32 - a.x00*a.x11*a.x32) * d
	m.x33 = (a.x01*a.x12*a.x20 - a.x02*a.x11*a.x20 + a.x02*a.x10*a.x21 - a.x00*a.x12*a.x21 - a.x01*a.x10*a.x22 + a.x00*a.x11*a.x22) * d
	return m
}

// Put puts elements of the matrix in row-major order in b. If b is not of at least length 16 then Put panics.
func (m *Mat4) Put(b []float32) {
	_ = b[15]
	b[0] = m.x00
	b[1] = m.x01
	b[2] = m.x02
	b[3] = m.x03

	b[4] = m.x10
	b[5] = m.x11
	b[6] = m.x12
	b[7] = m.x13

	b[8] = m.x20
	b[9] = m.x21
	b[10] = m.x22
	b[11] = m.x23

	b[12] = m.x30
	b[13] = m.x31
	b[14] = m.x32
	b[15] = m.x33
}

// Array returns the matrix values in a static array copy in row major order.
func (m Mat4) Array() (rowmajor [16]float32) {
	m.Put(rowmajor[:])
	return rowmajor
}

// RotatingBetweenVecsMat4 returns the rotation matrix that transforms "start" onto the same direction as "dest".
func RotatingBetweenVecsMat4(start, dest Vec) Mat4 {
	// is either vector == 0?
	const epsilon = 1e-12
	if EqualElem(start, Vec{}, epsilon) || EqualElem(dest, Vec{}, epsilon) {
		return IdentityMat4()
	}
	// normalize both vectors
	start = Unit(start)
	dest = Unit(dest)
	// are the vectors the same?
	if EqualElem(start, dest, epsilon) {
		return IdentityMat4()
	}

	// are the vectors opposite (180 degrees apart)?
	if EqualElem(Scale(-1, start), dest, epsilon) {
		return Mat4{
			-1, 0, 0, 0,
			0, -1, 0, 0,
			0, 0, -1, 0,
			0, 0, 0, 1,
		}
	}
	// general case
	// See:	https://math.stackexchange.com/questions/180418/calculate-rotation-matrix-to-align-vector-a-to-vector-b-in-3d
	v := Cross(start, dest)
	vx := Skew(v)
	k := 1. / (1. + Dot(start, dest))

	vx2 := MulMat3(vx, vx)
	vx2 = ScaleMat3(vx2, k)

	// Calculate sum of matrices.
	vx = AddMat3(vx, IdentityMat3())
	vx = AddMat3(vx, vx2)

	return vx.AsMat4()
}

// EqualMat4 tests the equality of 4x4 matrices.
func EqualMat4(a, b Mat4, tolerance float32) bool {
	return ms1.EqualWithinAbs(a.x00, b.x00, tolerance) &&
		ms1.EqualWithinAbs(a.x01, b.x01, tolerance) &&
		ms1.EqualWithinAbs(a.x02, b.x02, tolerance) &&
		ms1.EqualWithinAbs(a.x03, b.x03, tolerance) &&
		ms1.EqualWithinAbs(a.x10, b.x10, tolerance) &&
		ms1.EqualWithinAbs(a.x11, b.x11, tolerance) &&
		ms1.EqualWithinAbs(a.x12, b.x12, tolerance) &&
		ms1.EqualWithinAbs(a.x13, b.x13, tolerance) &&
		ms1.EqualWithinAbs(a.x20, b.x20, tolerance) &&
		ms1.EqualWithinAbs(a.x21, b.x21, tolerance) &&
		ms1.EqualWithinAbs(a.x22, b.x22, tolerance) &&
		ms1.EqualWithinAbs(a.x23, b.x23, tolerance) &&
		ms1.EqualWithinAbs(a.x30, b.x30, tolerance) &&
		ms1.EqualWithinAbs(a.x31, b.x31, tolerance) &&
		ms1.EqualWithinAbs(a.x32, b.x32, tolerance) &&
		ms1.EqualWithinAbs(a.x33, b.x33, tolerance)
}
