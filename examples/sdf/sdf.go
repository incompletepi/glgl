// This program takes input array and a add scalar
// and writes the result of adding them to a new output array.
// The value of this example is that input and output arrays
// may have different memory layouts. One may contain vector data
// and the other may contain scalars.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"runtime"
	"strconv"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/incompletepi/glgl/v4.6-core/glgl"
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func makeScene() SDFShaderer {
	// Make SDF shader program.
	s1, _ := NewSphere(0.5)
	s2, _ := NewSphere(1)
	s1 = Translate(s1, Vec{X: 2})
	obj := Union(s1, s2)
	return obj
}

func main() {
	// Initialize the GL.
	_, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:   "compute",
		Version: [2]int{4, 6},
		Width:   1,
		Height:  1,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer terminate()

	var source bytes.Buffer
	_, err = writeProgram(&source, makeScene())
	if err != nil {
		panic(err)
	}

	// return
	ss, err := glgl.ParseCombined(&source)
	if err != nil {
		log.Println("parsing:", err)
		return
	}
	prog, err := glgl.CompileProgram(ss)
	if err != nil {
		log.Println("creating program:", err)
		return
	}
	prog.Bind()
	const div = 4
	const min, max = -1, 1

	inputArray := make([][3]float32, div*div*div)
	for i := 0; i < div; i++ {
		off1 := i * div * div
		x := float32(i)*(max-min)/div + min
		for j := 0; j < div; j++ {
			off2 := off1 + j*div
			y := float32(j)*(max-min)/div + min
			for k := 0; k < div; k++ {
				z := float32(k)*(max-min)/div + min
				inputArray[off2+k] = [3]float32{x, y, z}
			}
		}
	}
	inputCfg := glgl.TextureImgConfig{
		Type:           glgl.Texture2D,
		Width:          len(inputArray),
		Height:         1,
		Access:         glgl.ReadOnly,
		Format:         gl.RGB,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		Xtype:          gl.FLOAT,
		InternalFormat: gl.RGBA32F,
		ImageUnit:      0,
	}
	_, err = glgl.NewTextureFromImage(inputCfg, inputArray)
	if err != nil {
		log.Println("creating input texture:", err)
		return
	}

	outputArray := make([]float32, len(inputArray))
	// Define OUTPUT texture.
	outputCfg := glgl.TextureImgConfig{
		Type:           glgl.Texture2D,
		Width:          len(outputArray),
		Height:         1,
		Access:         glgl.WriteOnly,
		Format:         gl.RED,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		Xtype:          gl.FLOAT,
		InternalFormat: gl.R32F,
		ImageUnit:      1,
	}
	outputTex, err := glgl.NewTextureFromImage(outputCfg, outputArray)
	if err != nil {
		log.Println("creating output texture", err)
		return
	}

	// Dispatch and wait for compute to finish.
	err = prog.RunCompute(len(inputArray), 1, 1)
	if err != nil {
		log.Println("running compute shader", err)
		return
	}
	err = glgl.GetImage(outputArray, outputTex, outputCfg)
	if err != nil {
		log.Println("acquiring results from GPU", err)
		return
	}
	fmt.Println("SDF table position to distance:")
	for i := range inputArray {
		pos := inputArray[i]
		fmt.Printf("x:%.2g\ty:%.2g\tz:%.2g\t-> %.3g\n", pos[0], pos[1], pos[2], outputArray[i])
	}
	// fmt.Println(source.String()) // Print generated shader source code.
}

type Vec struct {
	X, Y, Z float32
}

type SDF interface {
	Evaluate(positions []Vec, distances []float32) (int, error)
	Bounds() (min, max Vec)
}

type Sphere struct {
	R float32
}

const fltPrec = 8
const fltFmtByte = 'g'

type SDFShader struct {
	Name []byte
	Body []byte
}

func (s *Sphere) ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error { return nil }

func (s *Sphere) AppendShader(glsl *SDFShader) error {
	r := float64(s.R)
	glsl.Name = append(glsl.Name, "sphere"...)
	glsl.Name = strconv.AppendFloat(glsl.Name, r, fltFmtByte, fltPrec, 32)
	if idx := bytes.IndexByte(glsl.Name, '.'); idx >= 0 {
		// Identifiers cannot have period in name.
		glsl.Name[idx] = 'p'
	}
	glsl.Body = append(glsl.Body, "return length(p)-"...)
	glsl.Body = strconv.AppendFloat(glsl.Body, r, fltFmtByte, fltPrec, 32)
	glsl.Body = append(glsl.Body, ';')
	return nil
}

func (s *Sphere) Evaluate(positions []Vec, distances []float32) (int, error) {
	for i, pos := range positions {
		r := norm(pos)
		distances[i] = r - s.R
	}
	return 0, nil
}

func (s *Sphere) Bounds() (min, max Vec) {
	min = Vec{X: s.R, Y: s.R, Z: s.R}
	max = Vec{X: s.R, Y: s.R, Z: s.R}
	return min, max
}

type SDFShaderer interface {
	Bounds() (min, max Vec)
	AppendShader(glsl *SDFShader) error
	ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error
}

type BinaryOpShader struct {
	s1, s2  SDFShaderer
	opname  string
	bodyFmt string
}

func Union(s1, s2 SDFShaderer) SDFShaderer {
	if s1 == nil || s2 == nil {
		panic("nil object")
	}
	return &UnionShader{
		s1: s1,
		s2: s2,
	}
}

type UnionShader struct {
	s1, s2 SDFShaderer
}

func (s *UnionShader) Bounds() (vmin, vmax Vec) {
	min1, max1 := s.s1.Bounds()
	min2, max2 := s.s2.Bounds()
	vmin = Vec{X: minf(min1.X, min2.X), Y: minf(min1.Y, min2.Y), Z: minf(min1.Z, min2.Z)}
	vmax = Vec{X: maxf(max1.X, max2.X), Y: maxf(max1.Y, max2.Y), Z: maxf(max1.Z, max2.Z)}
	return vmin, vmax
}

func (s *UnionShader) ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error {
	err := fn(flags, s.s1)
	if err != nil {
		return err
	}
	return fn(flags, s.s2)
}

func (s *UnionShader) AppendShader(glsl *SDFShader) error {
	body := glsl.Body
	glsl.Name = append(glsl.Name, "union_"...)
	id1Start := len(glsl.Name)
	err := s.s1.AppendShader(glsl)
	if err != nil {
		return err
	}
	id2Start := len(glsl.Name)
	err = s.s2.AppendShader(glsl)
	if err != nil {
		return err
	}
	glsl.Body = glsl.Body[:len(body)] // Remove union element bodies but retain longer
	glsl.Body = append(glsl.Body, "return min("...)
	glsl.Body = append(glsl.Body, glsl.Name[id1Start:id2Start]...)
	glsl.Body = append(glsl.Body, "(p),"...)
	glsl.Body = append(glsl.Body, glsl.Name[id2Start:]...)
	glsl.Body = append(glsl.Body, "(p));"...)
	return nil
}

func NewSphere(radius float32) (SDFShaderer, error) {
	return &Sphere{R: radius}, nil
}

func Translate(s SDFShaderer, to Vec) SDFShaderer {
	return &TranslateShader{
		s: s,
		p: to,
	}
}

type TranslateShader struct {
	s SDFShaderer
	p Vec
}

func (ts *TranslateShader) Bounds() (min, max Vec) {
	min, max = ts.s.Bounds()
	min = Vec{X: min.X + ts.p.X, Y: min.Y + ts.p.Y, Z: min.X + ts.p.Z}
	max = Vec{X: max.X + ts.p.X, Y: max.Y + ts.p.Y, Z: max.X + ts.p.Z}
	return min, max
}

func (s *TranslateShader) ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error {
	return fn(flags, s.s)
}

func (ts *TranslateShader) AppendShader(glsl *SDFShader) error {
	glsl.Name = append(glsl.Name, "translate"...)
	glsl.Name = strconv.AppendFloat(glsl.Name, float64(ts.p.X), fltFmtByte, fltPrec, 32)
	glsl.Name = strconv.AppendFloat(glsl.Name, float64(ts.p.Y), fltFmtByte, fltPrec, 32)
	glsl.Name = strconv.AppendFloat(glsl.Name, float64(ts.p.Z), fltFmtByte, fltPrec, 32)
	for {
		idx := bytes.IndexByte(glsl.Name, '.')
		if idx < 0 {
			break
		}
		glsl.Name[idx] = 'p'
	}
	glsl.Name = append(glsl.Name, '_')
	idStart := len(glsl.Name)
	body := glsl.Body
	err := ts.s.AppendShader(glsl)
	if err != nil {
		return err
	}
	glsl.Body = glsl.Body[:len(body)]
	glsl.Body = append(glsl.Body, "return "...)
	glsl.Body = append(glsl.Body, glsl.Name[idStart:]...)
	glsl.Body = append(glsl.Body, "(p - vec3("...)
	glsl.Body = strconv.AppendFloat(glsl.Body, float64(ts.p.X), 'f', fltPrec, 32)
	glsl.Body = append(glsl.Body, ',')
	glsl.Body = strconv.AppendFloat(glsl.Body, float64(ts.p.Y), 'f', fltPrec, 32)
	glsl.Body = append(glsl.Body, ',')
	glsl.Body = strconv.AppendFloat(glsl.Body, float64(ts.p.Z), 'f', fltPrec, 32)
	glsl.Body = append(glsl.Body, "));"...)
	return nil
}

func writeProgram(w io.Writer, obj SDFShaderer) (n int, err error) {
	var scratch SDFShader
	obj.AppendShader(&scratch)
	topname := string(scratch.Name)

	Children := []SDFShaderer{obj}
	nextChild := 0
	for len(Children[nextChild:]) > 0 {
		prev := len(Children)
		for _, obj := range Children[nextChild:] {
			fmt.Printf("%T\n", obj)
			obj.ForEachChild(0, func(flags int, s SDFShaderer) error {
				Children = append(Children, s)
				return nil
			})
		}
		nextChild = prev
	}
	const programHeader = `#shader compute
#version 430
`
	n, err = w.Write([]byte(programHeader))

	if err != nil {
		return n, err
	}

	for i := len(Children) - 1; i >= 0; i-- {
		ngot, err := writeShader(w, Children[i], &scratch)
		n += ngot
		if err != nil {
			return n, err
		}
	}
	programMain := fmt.Sprintf(`

layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
layout(rgba32f, binding = 0) uniform image2D in_tex;
// The binding argument refers to the textures Unit.
layout(r32f, binding = 1) uniform image2D out_tex;

void main() {
	// get position to read/write data from.
	ivec2 pos = ivec2( gl_GlobalInvocationID.xy );
	// Get SDF position value.
	vec3 p = imageLoad( in_tex, pos ).rgb;
	float distance = %s(p);
	// store new value in image
	imageStore( out_tex, pos, vec4( distance, 0.0, 0.0, 0.0 ) );
}
	`, topname)

	ngot, err := w.Write([]byte(programMain))
	return n + ngot, err
}

func writeShader(w io.Writer, s SDFShaderer, scratch *SDFShader) (int, error) {
	scratch.Name = scratch.Name[:0]
	scratch.Body = scratch.Body[:0]
	scratch.Name = append(scratch.Name, "float "...)
	err := s.AppendShader(scratch)
	if err != nil {
		return 0, err
	}
	scratch.Name = append(scratch.Name, "(vec3 p) {\n"...)

	scratch.Body = append(scratch.Body, "\n}\n\n"...)
	n, err := w.Write(scratch.Name)
	if err != nil {
		return n, err
	}
	n2, err := w.Write(scratch.Body)
	return n + n2, err
}

func minf(a, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

func maxf(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

// norm is equivalent to glsl `length` call.
func norm(pos Vec) float32 {
	r1 := math.Hypot(float64(pos.X), float64(pos.Y))
	r2 := math.Hypot(r1, float64(pos.Z))
	return float32(r2)
}
