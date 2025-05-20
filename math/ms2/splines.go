package ms2

// Spline3 implements uniform cubic spline logic (degree 3).
// Keep in mind the iteration over the spline points and how the points are interpreted
// depend on the type of spline being worked with.
//
// Bézier example:
//
//	const Nsamples = 64 // Number of times to sample each set of two Bézier points.
//	var spline []ms2.Vec = makeBezierSpline()
//	bz := ms2.SplineBezier()
//	var curve []ms2.Vec
//	for i := 0; i < len(spline); i += 4 {
//		p0, cp0, cp1, p1 := spline[4*i], spline[4*i+1], spline[4*i+2], spline[4*i+3]
//		for t := float32(0.0); t<1; t+=1./Nsamples {
//			xy := bz.Evaluate(t, p0, cp0, cp1, p1)
//			curve = append(curve, xy)
//		}
//	}
//	plot(curve)
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
type Spline3 struct {
	m mat4
}

// NewSpline3 returns a [Spline3] ready for use.
// See [Freya Holmér's video] on splines for more information on how a matrix represents a uniform cubic spline.
//
// [Freya Holmér's video]: https://youtu.be/jvPPXbo87ds?si=Sn08aUjSKSXeRZ6D&t=419
func NewSpline3(matrix4x4 []float32) Spline3 {
	if len(matrix4x4) < 16 {
		panic("input matrix too short (need to be 4x4, row major)")
	}
	return Spline3{m: newMat4(matrix4x4)}
}

// Mat4Array returns a row-major ordered copy of the values of the cubic spline 4x4 matrix.
func (s Spline3) Mat4Array() [16]float32 {
	return s.m.Array()
}

// Evaluate evaluates the cubic spline over 4 points with a value of t. t is usually between 0 and 1 to interpolate the spline.
func (s Spline3) Evaluate(t float32, v0, v1, v2, v3 Vec) (res Vec) {
	x := vec4{x: v0.X, y: v1.X, z: v2.X, w: v3.X}
	y := vec4{x: v0.Y, y: v1.Y, z: v2.Y, w: v3.Y}
	x = matvecmul4(s.m, x)
	y = matvecmul4(s.m, y)
	v0 = Vec{X: x.x, Y: y.x}
	v1 = Vec{X: x.y, Y: y.y}
	v2 = Vec{X: x.z, Y: y.z}
	v3 = Vec{X: x.w, Y: y.w}
	res = Add(v0, Scale(t, v1))
	res = Add(res, Scale(t*t, v2))
	res = Add(res, Scale(t*t*t, v3))
	return res
}

// BasisFuncs returns the basis functions of the cubic spline corresponding to each of 4 control points.
func (s Spline3) BasisFuncs() (bs [4]func(float32) float32) {
	arr := s.m.Transpose().Array()
	for i := range bs {
		off := i * 4
		bs[i] = func(t float32) (b float32) {
			return arr[off+0] + t*arr[off+1] + t*t*arr[off+2] + t*t*t*arr[off+3]
		}
	}
	return bs
}

// BasisFuncs returns the differentiaed basis functions of the cubic spline.
func (s Spline3) BasisFuncsDiff() (bs [4]func(float32) float32) {
	arr := s.m.Transpose().Array()
	for i := range bs {
		off := i * 4
		bs[i] = func(t float32) (b float32) {
			return arr[off+1] + 2*t*arr[off+2] + 3*t*t*arr[off+3]
		}
	}
	return bs
}

// BasisFuncsDiff2 returns the twice-differentiaed basis functions of the cubic spline.
func (s Spline3) BasisFuncsDiff2() (bs [4]func(float32) float32) {
	arr := s.m.Transpose().Array()
	for i := range bs {
		off := i * 4
		bs[i] = func(t float32) (b float32) {
			return 2*arr[off+2] + 6*t*arr[off+3]
		}
	}
	return bs
}

// BasisFuncsDiff3 returns the thrice-differentiaed basis functions of the cubic spline.
func (s Spline3) BasisFuncsDiff3() (bs [4]func(float32) float32) {
	arr := s.m.Transpose().Array()
	for i := range bs {
		off := i * 4
		bs[i] = func(t float32) (b float32) {
			return 6 * arr[off+3]
		}
	}
	return bs
}

// matrix form of bezier curves:
//
//	                        [ a b c d ]   [ P0 ]
//	B(t) = [1  t  t²  t³] * | e f g h | * | P1 |
//	                        | i j k l |   | P2 |
//	                        [ m n o p ]   [ P3 ]
var (
	_beziermat = newMat4([]float32{
		1, 0, 0, 0,
		-3, 3, 0, 0,
		3, -6, 3, 0,
		-1, 3, -3, 1,
	})
	_hermiteMat = newMat4([]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		-3, -2, 3, -1,
		2, 1, -2, 1,
	})
	_basisMat = scalemat4(1./6, newMat4([]float32{
		1, 4, 1, 0,
		-3, 0, 3, 0,
		3, -6, 3, 0,
		-1, 3, -3, 1,
	}))
	_cardinalMat = func(s float32) mat4 {
		return newMat4([]float32{
			0, 1, 0, 0,
			-s, 0, s, 0,
			2 * s, s - 3, 3 - 2*s, -s,
			-s, 2 - s, s - 2, s,
		})
	}
	_catmullromMat      = _cardinalMat(0.5)
	_quadraticBezierMat = newMat4([]float32{
		1, 0, 0, 0,
		-2, 2, 0, 0,
		1, -2, 1, 0,
		0, 0, 0, 0,
	})
)

// SplineBezierCubic returns a Bézier cubic spline interpreter. Result splines have the following characteristics:
//   - C¹/C⁰ continuous.
//   - Interpolates some points.
//   - Manual tangents, second and third vectors are control points.
//   - Uses in shapes and vector graphics.
//
// Iterate every 3 points. Point0, ControlPoint0, ControlPoint1, Point1.
func SplineBezierCubic() Spline3 { return Spline3{m: _beziermat} }

// SplineHermite returns a Hermite cubic spline interpreter. Result splines have the following characteristics:
//   - C¹/C⁰ continuous.
//   - Interpolates all points.
//   - Explicit tangents. Second and fourth vector arguments specify velocities.
//   - Uses in animation, physics simulations and interpolation.
//
// Iterate every 2 points, Point0, Velocity0, Point1, Velocity1.
func SplineHermite() Spline3 { return Spline3{m: _hermiteMat} }

// SplineCatmullRom returns a Catmull-Rom cubic spline interpreter, a special case of Cardinal spline when scale=0.5. Result splines have the following characteristics:
//   - C¹ continuous.
//   - Interpolates all points.
//   - Automatic tangents.
//   - Used for animation and path smoothing.
func SplineCatmullRom() Spline3 { return Spline3{m: _catmullromMat} }

// SplineCardinal returns a cardinal cubic spline interpreter.
func SplineCardinal(scale float32) Spline3 { return Spline3{m: _cardinalMat(scale)} }

// SplineBasis returns a B-Spline interpreter. Result splines have the following characteristics:
//   - C² continuous.
//   - No point interpolation.
//   - Automatic tangents.
//   - Ideal for curvature-sensitive shapes and animations such as camera paths. Used in industrial design.
func SplineBasis() Spline3 { return Spline3{m: _basisMat} }

// SplineBezierQuadratic returns a quadratic spline interpreter (fourth point is inneffective).
//   - C¹ continuous.
//   - Interpolates all points.
//   - Manual tangents.
//   - Used in fonts. Cubic beziers are superior.
//
// Iterate every 2 points. Point0, ControlPoint, Point1. Keep in mind this is an innefficient implementation of a quadratic bezier. Is here for convenience.
func SplineBezierQuadratic() Spline3 { return Spline3{m: _quadraticBezierMat} }

// Spline3Sampler implements algorithms for sampling points of a cubic spline [Spline3].
type Spline3Sampler struct {
	Spline         Spline3
	v0, v1, v2, v3 Vec
	// Tolerance sets the maximum permissible error for sampling the cubic spline.
	// That is to say the resulting sampled set of line segments will approximate the curve to within Tolerance.
	Tolerance float32
}

// SetSplinePoints sets the 4 [Vec]s which define a cubic spline. They are passed to the Spline on Evaluate calls.
func (s *Spline3Sampler) SetSplinePoints(v0, v1, v2, v3 Vec) {
	s.v0, s.v1, s.v2, s.v3 = v0, v1, v2, v3
}

// Evaluate evaluates a point on the spline with points set by [Spline3Sampler.SetSplinePoints].
// It calls [Spline3.Evaluate] with t and the set points.
func (s *Spline3Sampler) Evaluate(t float32) Vec {
	return s.Spline.Evaluate(t, s.v0, s.v1, s.v2, s.v3)
}

// SampleBisect samples the cubic spline using bisection method to
// find points which discretize the curve to within [Spline3Sampler.Tol] error
// These points are then appended to dst and the result returned.
//
// It does not append points at extremes t=0 and t=1.
// maxDepth determines the max amount of times to subdivide the curve.
// The max amount of subdivisions (points appended) is given by 2**maxDepth.
func (s *Spline3Sampler) SampleBisect(dst []Vec, maxDepth int) []Vec {
	if maxDepth <= 0 {
		panic("invalid depth")
	} else if s.Tolerance < 0 {
		panic("negative tolerance")
	} else if s.Tolerance == 0 {
		panic("zero tolerance, initialize Spline3Sampler Tolerance field to a small value, i.e: 0.01")
	}
	baseRes := 1.0 / float32(uint(1)<<uint(maxDepth))
	return s.sampleBisect(dst, maxDepth, 0, s.Evaluate(0), 0, baseRes)
}

// SampleBisectWithExtremes is same as [Spline3Sampler.SampleBisect] but adding start and end points at t=0, t=1.
func (s *Spline3Sampler) SampleBisectWithExtremes(dst []Vec, maxDepth int) []Vec {
	if maxDepth <= 0 {
		panic("invalid depth")
	} else if s.Tolerance < 0 {
		panic("negative tolerance")
	} else if s.Tolerance == 0 {
		panic("zero tolerance, initialize Spline3Sampler Tolerance field to a small value, i.e: 0.01")
	}
	baseRes := 1.0 / float32(uint(1)<<uint(maxDepth))
	xStart := s.Evaluate(0)
	dst = append(dst, xStart)
	dst = s.sampleBisect(dst, maxDepth, 0, xStart, 0, baseRes)
	dst = append(dst, s.Evaluate(1))
	return dst
}

func (s *Spline3Sampler) sampleBisect(dst []Vec, lvl, idx int, xstart Vec, tstart, baseRes float32) []Vec {
	if lvl == 0 {
		if idx != 0 {
			dst = append(dst, xstart)
		}
		return dst
	}
	// Same algorithm as octree splitting but in 1D.
	slvl := lvl - 1
	midIdx := idx + 1<<slvl
	endIdx := idx + 1<<lvl

	tend := baseRes * float32(endIdx)
	tmid := baseRes * float32(midIdx)
	xend := s.Evaluate(tend)
	xmid := s.Evaluate(tmid)
	if Collinear(xstart, xmid, xend, s.Tolerance) {
		// Check offset- curve may be undersampled.
		var k float32 = 0.45
		tmid2 := tstart + k*(tend-tstart)
		xmid2 := s.Evaluate(tmid2)
		if Collinear(xstart, xmid2, xend, s.Tolerance) {
			if idx != 0 {
				dst = append(dst, xstart)
			}
			return dst // Won't subdivide further, this section of spline is straight.
		}
	}

	dst = s.sampleBisect(dst, slvl, idx, xstart, tstart, baseRes)
	dst = s.sampleBisect(dst, slvl, midIdx, xmid, tmid, baseRes)
	return dst
}

// newMat4 instantiates a new 4x4 Mat4 matrix from the first 16 values in row major order.
// If v is shorter than 16 newMat4 panics.
func newMat4(v []float32) (m mat4) {
	_ = v[15]
	m.x00, m.x01, m.x02, m.x03 = v[0], v[1], v[2], v[3]
	m.x10, m.x11, m.x12, m.x13 = v[4], v[5], v[6], v[7]
	m.x20, m.x21, m.x22, m.x23 = v[8], v[9], v[10], v[11]
	m.x30, m.x31, m.x32, m.x33 = v[12], v[13], v[14], v[15]
	return m
}

// mat4 is a 4x4 matrix.
type mat4 struct {
	x00, x01, x02, x03 float32
	x10, x11, x12, x13 float32
	x20, x21, x22, x23 float32
	x30, x31, x32, x33 float32
}

type vec4 struct {
	x, y, z, w float32
}

func matvecmul4(m mat4, v vec4) (res vec4) {
	res.x = m.x00*v.x + m.x01*v.y + m.x02*v.z + m.x03*v.w
	res.y = m.x10*v.x + m.x11*v.y + m.x12*v.z + m.x13*v.w
	res.z = m.x20*v.x + m.x21*v.y + m.x22*v.z + m.x23*v.w
	res.w = m.x30*v.x + m.x31*v.y + m.x32*v.z + m.x33*v.w
	return res
}

func scalemat4(f float32, m mat4) mat4 {
	m.x00 *= f
	m.x01 *= f
	m.x02 *= f
	m.x03 *= f
	m.x10 *= f
	m.x11 *= f
	m.x12 *= f
	m.x13 *= f
	m.x20 *= f
	m.x21 *= f
	m.x22 *= f
	m.x23 *= f
	m.x30 *= f
	m.x31 *= f
	m.x32 *= f
	m.x33 *= f
	return m
}

// Put puts elements of the matrix in row-major order in b. If b is not of at least length 16 then Put panics.
func (m *mat4) Put(b []float32) {
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
func (m mat4) Array() (rowmajor [16]float32) {
	m.Put(rowmajor[:])
	return rowmajor
}

// Transpose returns the transpose of a.
func (a mat4) Transpose() mat4 {
	return mat4{
		x00: a.x00, x01: a.x10, x02: a.x20, x03: a.x30,
		x10: a.x01, x11: a.x11, x12: a.x21, x13: a.x31,
		x20: a.x02, x21: a.x12, x22: a.x22, x23: a.x32,
		x30: a.x03, x31: a.x13, x32: a.x23, x33: a.x33,
	}
}
