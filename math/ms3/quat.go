// Copyright 2014 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ms3

import (
	"unsafe"

	math "github.com/chewxy/math32"
)

const sizeofFloat = unsafe.Sizeof(float32(0))

var (
	_ = [1]byte{}[unsafe.Sizeof(Vec{})-4*sizeofFloat]  // Compile time check that Vec is 3 float32s long.
	_ = [1]byte{}[unsafe.Sizeof(Quat{})-4*sizeofFloat] // Compile time check that Quat is 4 float32s long.
)

// RotationOrder is the order in which rotations will be transformed for the
// purposes of AnglesToQuat.
type RotationOrder int

// The RotationOrder constants represent a series of rotations along the given
// axes for the use of AnglesToQuat.
const (
	XYX RotationOrder = iota
	XYZ
	XZX
	XZY
	YXY
	YXZ
	YZY
	YZX
	ZYZ
	ZYX
	ZXZ
	ZXY
)

// Quat represents a Quaternion, which is an extension of the imaginary numbers;
// there's all sorts of interesting theory behind it. In 3D graphics we mostly
// use it as a cheap way of representing rotation since quaternions are cheaper
// to multiply by, and easier to interpolate than matrices.
//
// A Quaternion has two parts: W, the so-called scalar component, and "V", the
// vector component. The vector component is considered to be the part in 3D
// space, while W (loosely interpreted) is its 4D coordinate.
//
// The imaginary V part is guaranteed to have an offset of zero in the Quat struct:
//
//	unsafe.Offsetof(q.V) // == 0
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
type Quat struct {
	// V contains I, J and K imaginary parts.
	I, J, K float32
	W       float32
}

// IJK returns I,J,K fields of q as a vector with set fields X,Y,Z, respectively.
func (q Quat) IJK() Vec { return Vec{X: q.I, Y: q.J, Z: q.K} }

// WithIJK replaces I, J and K fields of q with X,Y and Z fields of argument Vec ijk and returns the result.
func (q Quat) WithIJK(ijk Vec) Quat {
	return Quat{
		W: q.W,
		I: ijk.X,
		J: ijk.Y,
		K: ijk.Z,
	}
}

// QuatIdent returns the quaternion identity: W=1; V=(0,0,0).
//
// As with all identities, multiplying any quaternion by this will yield the same
// quaternion you started with.
func QuatIdent() Quat {
	return Quat{W: 1.}
}

// RotationQuat creates a rotation quaternion
// that rotates an angle relative an axis.
// Call Rotate method on Quat to apply rotation.
func RotationQuat(angle float32, axis Vec) Quat {
	// angle = (float32(math.Pi) * angle) / 180.0
	s, c := math.Sincos(0.5 * angle)
	return Quat{
		W: c,
		I: axis.X * s,
		J: axis.Y * s,
		K: axis.Z * s,
	}
}

// Add adds two quaternions. It's no more complicated than
// adding their W and V components.
func (q1 Quat) Add(q2 Quat) Quat {
	return Quat{
		W: q1.W + q2.W,
		I: q1.I + q2.I,
		J: q1.J + q2.J,
		K: q1.K + q2.K,
	}
}

// Sub subtracts two quaternions. It's no more complicated than
// subtracting their W and V components.
func (q1 Quat) Sub(q2 Quat) Quat {
	return Quat{
		W: q1.W - q2.W,
		I: q1.I - q2.I,
		J: q1.J - q2.J,
		K: q1.K - q2.K,
	}
}

// Mul multiplies two quaternions. This can be seen as a rotation. Note that
// Multiplication is NOT commutative, meaning q1.Mul(q2) does not necessarily
// equal q2.Mul(q1).
func (q1 Quat) Mul(q2 Quat) Quat {
	v1 := q1.IJK()
	v2 := q2.IJK()
	m := Add(Cross(v1, v2), Scale(q1.W, v2))
	return Quat{
		W: q1.W*q2.W - Dot(v1, v2),
		I: m.X + q2.W*v1.X,
		J: m.Y + q2.W*v1.Y,
		K: m.Z + q2.W*v1.Z,
	}
}

// Scale every element of the quaternion by some constant factor.
func (q1 Quat) Scale(c float32) Quat {
	return Quat{
		W: q1.W * c,
		I: q1.I * c,
		J: q1.J * c,
		K: q1.K * c,
	}
}

// Conjugate returns the conjugate of a quaternion. Equivalent to
// Quat{q1.W, q1.V.Mul(-1)}.
func (q1 Quat) Conjugate() Quat {
	return Quat{
		W: q1.W,
		I: -q1.I,
		J: -q1.J,
		K: -q1.K,
	}
}

// Norm returns the euclidean length of the quaternion.
func (q1 Quat) Norm() float32 {
	return math.Sqrt(q1.Dot(q1))
}

// Normalize the quaternion, returning its versor (unit quaternion).
//
// This is the same as normalizing it as a Vec4.
func (q1 Quat) Unit() Quat {
	length := q1.Norm()

	if math.Abs(1-length) < 1e-8 {
		return q1
	}
	if length == 0 {
		return QuatIdent()
	}
	if math.IsInf(length, 0) {
		length = math.Copysign(math.MaxFloat32, length)
	}
	inv := 1. / length
	return q1.Scale(inv)
}

// Inverse of a quaternion. The inverse is equivalent
// to the conjugate divided by the square of the length.
//
// This method computes the square norm by directly adding the sum
// of the squares of all terms instead of actually squaring q1.Len(),
// both for performance and precision.
func (q1 Quat) Inverse() Quat {
	return q1.Conjugate().Scale(1 / q1.Dot(q1))
}

// Rotate a vector by the rotation this quaternion represents.
// This will result in a 3D vector. Strictly speaking, this is
// equivalent to q1.v.q* where the "."" is quaternion multiplication and v is interpreted
// as a quaternion with W 0 and V v. In code:
// q1.Mul(Quat{0,v}).Mul(q1.Conjugate()), and
// then retrieving the imaginary (vector) part.
//
// In practice, we hand-compute this in the general case and simplify
// to save a few operations.
func (q1 Quat) Rotate(v Vec) Vec {
	v1 := q1.IJK()
	cross := Cross(v1, v)
	// v + 2q_w * (q_v x v) + 2q_v x (q_v x v)
	finalTerm := Cross(Scale(2, v1), cross)
	x := Add(Scale(2*q1.W, cross), finalTerm)
	return Add(v, x)
}

// Mat4 returns the homogeneous 3D rotation matrix corresponding to the
// quaternion.
// func (q1 Quat) Mat4() Mat4 {
// 	w, x, y, z := q1.W, q1.V[0], q1.V[1], q1.V[2]
// 	return Mat4{
// 		1 - 2*y*y - 2*z*z, 2*x*y + 2*w*z, 2*x*z - 2*w*y, 0,
// 		2*x*y - 2*w*z, 1 - 2*x*x - 2*z*z, 2*y*z + 2*w*x, 0,
// 		2*x*z + 2*w*y, 2*y*z - 2*w*x, 1 - 2*x*x - 2*y*y, 0,
// 		0, 0, 0, 1,
// 	}
// }

// Dot product between two quaternions, equivalent to if this was a Vec4.
func (q1 Quat) Dot(q2 Quat) float32 {
	return q1.W*q2.W + q1.I*q2.I + q1.J*q2.J + q1.K*q2.K
}

// QuatSlerp is Spherical Linear intERPolation, a method of interpolating
// between two quaternions. This always takes the straightest path on the sphere between
// the two quaternions, and maintains constant velocity.
//
// However, it's expensive and QuatSlerp(q1,q2) is not the same as QuatSlerp(q2,q1)
func QuatSlerp(q1, q2 Quat, amount float32) Quat {
	q1, q2 = q1.Unit(), q2.Unit()
	dot := q1.Dot(q2)

	// If the inputs are too close for comfort, linearly interpolate and normalize the result.
	if dot > 0.9995 {
		return QuatNlerp(q1, q2, amount)
	}

	// This is here for precision errors, I'm perfectly aware that *technically* the dot is bound [-1,1], but since Acos will freak out if it's not (even if it's just a liiiiitle bit over due to normal error) we need to clamp it
	dot = math.Max(-1, math.Min(1, dot))

	theta := math.Acos(dot) * amount

	s, c := math.Sincos(theta)
	rel := q2.Sub(q1.Scale(dot)).Unit()

	return q1.Scale(c).Add(rel.Scale(s))
}

// QuatLerp is a *L*inear Int*erp*olation between two Quaternions, cheap and simple.
//
// Not excessively useful, but uses can be found.
func QuatLerp(q1, q2 Quat, amount float32) Quat {
	return q1.Add(q2.Sub(q1).Scale(amount))
}

// QuatNlerp is a *Normalized* *L*inear Int*erp*olation between two Quaternions. Cheaper than Slerp
// and usually just as good. This is literally Lerp with Normalize() called on it.
//
// Unlike Slerp, constant velocity isn't maintained, but it's much faster and
// Nlerp(q1,q2) and Nlerp(q2,q1) return the same path. You should probably
// use this more often unless you're suffering from choppiness due to the
// non-constant velocity problem.
func QuatNlerp(q1, q2 Quat, amount float32) Quat {
	return QuatLerp(q1, q2, amount).Unit()
}

// AnglesToQuat performs a rotation in the specified order. If the order is not
// a valid RotationOrder, this function will panic
//
// The rotation "order" is more of an axis descriptor. For instance XZX would
// tell the function to interpret angle1 as a rotation about the X axis, angle2 about
// the Z axis, and angle3 about the X axis again.
//
// Based off the code for the Matlab function "angle2quat", though this implementation
// only supports 3 single angles as opposed to multiple angles.
func AnglesToQuat(angle1, angle2, angle3 float32, order RotationOrder) Quat {
	var s [3]float32
	var c [3]float32

	s[0], c[0] = math.Sincos(angle1 / 2)
	s[1], c[1] = math.Sincos(angle2 / 2)
	s[2], c[2] = math.Sincos(angle3 / 2)

	var ret Quat
	switch order {
	default:
		panic("Unsupported rotation order")
	case ZYX:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] + s[0]*s[1]*s[2]),
			I: float32(c[0]*c[1]*s[2] - s[0]*s[1]*c[2]),
			J: float32(c[0]*s[1]*c[2] + s[0]*c[1]*s[2]),
			K: float32(s[0]*c[1]*c[2] - c[0]*s[1]*s[2]),
		}

	case ZYZ:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*c[1]*s[2]),
			I: float32(c[0]*s[1]*s[2] - s[0]*s[1]*c[2]),
			J: float32(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			K: float32(s[0]*c[1]*c[2] + c[0]*c[1]*s[2]),
		}
	case ZXY:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*s[1]*s[2]),
			I: float32(c[0]*s[1]*c[2] - s[0]*c[1]*s[2]),
			J: float32(c[0]*c[1]*s[2] + s[0]*s[1]*c[2]),
			K: float32(c[0]*s[1]*s[2] + s[0]*c[1]*c[2]),
		}

	case ZXZ:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*c[1]*s[2]),
			I: float32(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			J: float32(s[0]*s[1]*c[2] - c[0]*s[1]*s[2]),
			K: float32(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
		}

	case YXZ:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] + s[0]*s[1]*s[2]),
			I: float32(c[0]*s[1]*c[2] + s[0]*c[1]*s[2]),
			J: float32(s[0]*c[1]*c[2] - c[0]*s[1]*s[2]),
			K: float32(c[0]*c[1]*s[2] - s[0]*s[1]*c[2]),
		}

	case YXY:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*c[1]*s[2]),
			I: float32(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			J: float32(s[0]*c[1]*c[2] + c[0]*c[1]*s[2]),
			K: float32(c[0]*s[1]*s[2] - s[0]*s[1]*c[2]),
		}

	case YZX:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*s[1]*s[2]),
			I: float32(c[0]*c[1]*s[2] + s[0]*s[1]*c[2]),
			J: float32(c[0]*s[1]*s[2] + s[0]*c[1]*c[2]),
			K: float32(c[0]*s[1]*c[2] - s[0]*c[1]*s[2]),
		}

	case YZY:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*c[1]*s[2]),
			I: float32(s[0]*s[1]*c[2] - c[0]*s[1]*s[2]),
			J: float32(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
			K: float32(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
		}

	case XYZ:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*s[1]*s[2]),
			I: float32(c[0]*s[1]*s[2] + s[0]*c[1]*c[2]),
			J: float32(c[0]*s[1]*c[2] - s[0]*c[1]*s[2]),
			K: float32(c[0]*c[1]*s[2] + s[0]*s[1]*c[2]),
		}

	case XYX:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*c[1]*s[2]),
			I: float32(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
			J: float32(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			K: float32(s[0]*s[1]*c[2] - c[0]*s[1]*s[2]),
		}

	case XZY:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] + s[0]*s[1]*s[2]),
			I: float32(s[0]*c[1]*c[2] - c[0]*s[1]*s[2]),
			J: float32(c[0]*c[1]*s[2] - s[0]*s[1]*c[2]),
			K: float32(c[0]*s[1]*c[2] + s[0]*c[1]*s[2]),
		}

	case XZX:
		ret = Quat{
			W: float32(c[0]*c[1]*c[2] - s[0]*c[1]*s[2]),
			I: float32(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
			J: float32(c[0]*s[1]*s[2] - s[0]*s[1]*c[2]),
			K: float32(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
		}

	}
	return ret
}

// QuatLookAt creates a rotation from an eye point to a center point.
//
// It assumes the front of the rotated object at Z- and up at Y+
func QuatLookAt(eye, center, upDir Vec) Quat {
	// http://www.opengl-tutorial.org/intermediate-tutorials/tutorial-17-quaternions/#I_need_an_equivalent_of_gluLookAt__How_do_I_orient_an_object_towards_a_point__
	// https://bitbucket.org/sinbad/ogre/src/d2ef494c4a2f5d6e2f0f17d3bfb9fd936d5423bb/OgreMain/src/OgreCamera.cpp?at=default#cl-161

	direction := Unit(Sub(center, eye))

	// Find the rotation between the front of the object (that we assume towards Z-,
	// but this depends on your model) and the desired direction
	rotDir := RotationBetweenVecsQuat(Vec{X: 0, Y: 0, Z: -1}, direction)

	// Recompute up so that it's perpendicular to the direction
	// You can skip that part if you really want to force up
	//right := direction.Cross(up)
	//up = right.Cross(direction)

	// Because of the 1rst rotation, the up is probably completely screwed up.
	// Find the rotation between the "up" of the rotated object, and the desired up
	upCur := rotDir.Rotate(Vec{X: 0, Y: 1, Z: 0})
	rotUp := RotationBetweenVecsQuat(upCur, upDir)

	rotTarget := rotUp.Mul(rotDir) // remember, in reverse order.
	return rotTarget.Inverse()     // camera rotation should be inversed!
}

// RotationBetweenVecsQuat calculates the rotation between start and dest.
func RotationBetweenVecsQuat(start, dest Vec) Quat {
	// http://www.opengl-tutorial.org/intermediate-tutorials/tutorial-17-quaternions/#I_need_an_equivalent_of_gluLookAt__How_do_I_orient_an_object_towards_a_point__
	// https://github.com/g-truc/glm/blob/0.9.5/glm/gtx/quaternion.inl#L225
	// https://bitbucket.org/sinbad/ogre/src/d2ef494c4a2f5d6e2f0f17d3bfb9fd936d5423bb/OgreMain/include/OgreVector3.h?at=default#cl-654

	start = Unit(start)
	dest = Unit(dest)
	epsilon := float32(0.001)

	cosTheta := Dot(start, dest)
	if cosTheta < -1.0+epsilon {
		// special case when vectors in opposite directions:
		// there is no "ideal" rotation axis
		// So guess one; any will do as long as it's perpendicular to start
		axis := Cross(Vec{X: 1, Y: 0, Z: 0}, start)
		if Norm2(axis) < epsilon {
			// bad luck, they were parallel, try again!
			axis = Cross(Vec{X: 0, Y: 1, Z: 0}, start)
		}

		return RotationQuat(math.Pi, Unit(axis))
	}

	axis := Cross(start, dest)
	s := math.Sqrt((1.0 + cosTheta) * 2.0)

	return Quat{
		W: s * 0.5,
		I: axis.X / s,
		J: axis.Y / s,
		K: axis.Z / s,
	}
}

// RotationMat3 returns a rotation 3x3 matrix.
func (q Quat) RotationMat3() Mat3 {
	qv := q.IJK()
	qs := Skew(qv)
	q01 := IdentityMat3()
	q01 = ScaleMat3(q01, q.W*q.W)

	qd := IdentityMat3()
	qd = ScaleMat3(qd, Dot(qv, qv))
	qs = ScaleMat3(qs, 2*q.W)

	m := ScaleMat3(Prod(qv, qv), 2) // m = 2*[q]*[q]ᵀ
	m = AddMat3(m, q01)             // m += q.Real*q.Real * [E]
	m = AddMat3(m, qd)              // m += dot([q],[q])*[E]
	m = AddMat3(m, qs)              // m += 2*q.Real * skew([q])
	return m
}

/*

// Mat4ToQuat converts a pure rotation matrix into a quaternion
func Mat4ToQuat(m Mat4) Quat {
	// http://www.euclideanspace.com/maths/geometry/rotations/conversions/matrixToQuaternion/index.htm

	if tr := m[0] + m[5] + m[10]; tr > 0 {
		s := 0.5 / math32.Sqrt(tr+1.0)
		return Quat{
			0.25 / s,
			Vec{
				(m[6] - m[9]) * s,
				(m[8] - m[2]) * s,
				(m[1] - m[4]) * s,
			},
		}
	}

	if (m[0] > m[5]) && (m[0] > m[10]) {
		s := 2.0 * math32.Sqrt(1.0+m[0]-m[5]-m[10])
		return Quat{
			(m[6] - m[9]) / s,
			Vec{
				0.25 * s,
				(m[4] + m[1]) / s,
				(m[8] + m[2]) / s,
			},
		}
	}

	if m[5] > m[10] {
		s := 2.0 * math32.Sqrt(1.0+m[5]-m[0]-m[10])
		return Quat{
			(m[8] - m[2]) / s,
			Vec{
				(m[4] + m[1]) / s,
				0.25 * s,
				(m[9] + m[6]) / s,
			},
		}

	}

	s := 2.0 * math32.Sqrt(1.0+m[10]-m[0]-m[5])
	return Quat{
		(m[1] - m[4]) / s,
		Vec{
			(m[8] + m[2]) / s,
			(m[9] + m[6]) / s,
			0.25 * s,
		},
	}
}

// ApproxEqual returns whether the quaternions are approximately equal, as if
// FloatEqual was called on each matching element
func (q1 Quat) ApproxEqual(q2 Quat) bool {
	return FloatEqual(q1.W, q2.W) && q1.V.ApproxEqual(q2.V)
}

// ApproxEqualThreshold returns whether the quaternions are approximately equal with a given tolerence, as if
// FloatEqualThreshold was called on each matching element with the given epsilon
func (q1 Quat) ApproxEqualThreshold(q2 Quat, epsilon float32) bool {
	return FloatEqualThreshold(q1.W, q2.W, epsilon) && q1.V.ApproxEqualThreshold(q2.V, epsilon)
}

// ApproxEqualFunc returns whether the quaternions are approximately equal using the given comparison function, as if
// the function had been called on each individual element
func (q1 Quat) ApproxEqualFunc(q2 Quat, f func(float32, float32) bool) bool {
	return f(q1.W, q2.W) && q1.V.ApproxFuncEqual(q2.V, f)
}

// OrientationEqual returns whether the quaternions represents the same orientation
//
// Different values can represent the same orientation (q == -q) because quaternions avoid singularities
// and discontinuities involved with rotation in 3 dimensions by adding extra dimensions
func (q1 Quat) OrientationEqual(q2 Quat) bool {
	return q1.OrientationEqualThreshold(q2, Epsilon)
}

// OrientationEqualThreshold returns whether the quaternions represents the same orientation with a given tolerence
func (q1 Quat) OrientationEqualThreshold(q2 Quat, epsilon float32) bool {
	return Abs(q1.Normalize().Dot(q2.Normalize())) > 1-epsilon
}

*/
