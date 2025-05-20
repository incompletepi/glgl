//go:build !tinygo && cgo

package glgl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"golang.org/x/exp/constraints"
)

// Version returns the running OpenGL version as a string.
func Version() string { return gl.GoStr(gl.GetString(gl.VERSION)) }

const (
	Int8    Type = gl.BYTE
	Uint8   Type = gl.UNSIGNED_BYTE
	Int16   Type = gl.SHORT
	Uint16  Type = gl.UNSIGNED_SHORT
	Int32   Type = gl.INT
	Uint32  Type = gl.UNSIGNED_INT
	Float32 Type = gl.FLOAT
)

var (
	ErrStringNotNullTerminated = errors.New("string not null terminated")
)

// MaxComputeInvocations returns the maximum total number of invocations (threads)
// in a single compute work group. This value represents the upper limit for the product
// of the local work group sizes in the X, Y, and Z dimensions, i.e.,
// local_size_x * local_size_y * local_size_z.
//
// This value is used to ensure that the total number of threads within a work group
// does not exceed what the hardware and OpenGL implementation can handle.
//
// The OpenGL context must be current when calling this function.
func MaxComputeInvocations() int {
	var p runtime.Pinner
	var invoc int32
	p.Pin(&invoc)
	defer p.Unpin()
	gl.GetIntegerv(gl.MAX_COMPUTE_WORK_GROUP_INVOCATIONS, &invoc)
	return int(invoc)
}

// MaxComputeWorkGroupCount returns the maximum number of work groups that can be
// dispatched in each dimension (X, Y, Z) when invoking a compute shader using
// [Program.RunCompute] (glDispatchCompute).
//
// The OpenGL context must be current when calling this function.
func MaxComputeWorkGroupCount() (Wcx, Wcy, Wcz int) {
	var wcx, wcy, wcz int32
	var p runtime.Pinner
	p.Pin(&wcx)
	p.Pin(&wcy)
	p.Pin(&wcz)
	defer p.Unpin()
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 0, &wcx)
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 1, &wcy)
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 2, &wcz)
	return int(wcx), int(wcy), int(wcz)
}

// MaxComputeWorkGroupSize returns the maximum size of a work group that can be
// used in each dimension (X, Y, Z) within a compute shader. This corresponds to
// the limits for the local work group sizes specified in the shader using the
// layout qualifiers, such as layout(local_size_x = X, local_size_y = Y, local_size_z = Z) in;
//
// The OpenGL context must be current when calling this function.
func MaxComputeWorkGroupSize() (Wsx, Wsy, Wsz int) {
	var wsx, wsy, wsz int32
	var p runtime.Pinner
	p.Pin(&wsx)
	p.Pin(&wsy)
	p.Pin(&wsz)
	defer p.Unpin()
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_SIZE, 0, &wsx)
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_SIZE, 1, &wsy)
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_SIZE, 2, &wsz)
	return int(wsx), int(wsy), int(wsz)
}

// EnableDebugOutput writes debug output to log via glDebugMessageCallback.
// If log is nil then the default slog package logger is used.
func EnableDebugOutput(log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		attrs := []slog.Attr{
			slog.Uint64("source", uint64(source)),
			slog.Uint64("gltype", uint64(gltype)),
			slog.Uint64("severity", uint64(severity)),
			slog.Uint64("length", uint64(length)),
		}
		var level slog.Level
		switch gltype {
		case gl.DEBUG_TYPE_ERROR:
			level = slog.LevelError
		case gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:
			level = slog.LevelWarn
		// case gl.DEBUG_TYPE_OTHER:
		// 	level = slog.LevelDebug
		default:
			level = slog.LevelInfo
		}
		log.LogAttrs(context.Background(), level, message, attrs...)
	}, nil)
}

// NewShaderStorageBuffer creates a new SSBO and binds it.
func NewShaderStorageBuffer[T any](data []T, cfg ShaderStorageBufferConfig) (ssbo ShaderStorageBuffer, err error) {
	var z T
	if data == nil && cfg.MemSize <= 0 {
		return ssbo, errors.New("undefined SSBO size")
	} else if data != nil && cfg.MemSize != 0 {
		return ssbo, errors.New("SSBO MemSize used only when data is nil")
	} else if unsafe.Sizeof(z)%uintptr(cfg.MemSize) != 0 {
		return ssbo, errors.New("SSBO MemSize should be multiple of data type length")
	}

	var p runtime.Pinner
	p.Pin(&ssbo.id)
	gl.GenBuffers(1, &ssbo.id)
	p.Unpin()
	ssbo.sz = int(unsafe.Sizeof(z)) * len(data)
	ssbo.usage = cfg.Usage
	ptr := unsafe.Pointer(&data[0])

	ssbo.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, ssbo.sz, ptr, uint32(cfg.Usage))
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, cfg.Base, ssbo.id)
	return ssbo, Err()
}

func (ssbo ShaderStorageBuffer) Bind() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, ssbo.id)
}

func (ssbo ShaderStorageBuffer) Delete() {
	var p runtime.Pinner
	p.Pin(&ssbo.id)
	gl.DeleteBuffers(1, &ssbo.id)
	p.Unpin()
}

// CopyFromShaderStorageBuffer copies data from a readable SSBO on the GPU to the destination buffer.
func CopyFromShaderStorageBuffer[T any](dst []T, ssbo ShaderStorageBuffer) error {
	dstSize := elemSize[T]() * len(dst)
	if ssbo.usage != ReadOnly && ssbo.usage != ReadOrWrite {
		return errors.New("attempted to read from non-readable SSBO")
	} else if ssbo.sz < dstSize {
		return errors.New("attempted to read more bytes than allocated for SSBO")
	} else if len(dst) == 0 {
		return errors.New("zero length or nil buffer")
	}
	ssbo.Bind()
	ptr := gl.MapBufferRange(gl.SHADER_STORAGE_BUFFER, 0, dstSize, gl.MAP_READ_BIT)
	if ptr == nil {
		err := Err()
		if err != nil {
			return err
		}
		return errors.New("failed to map buffer")
	}
	defer gl.UnmapBuffer(gl.SHADER_STORAGE_BUFFER)
	gpuBytes := unsafe.Slice((*byte)(ptr), dstSize)
	bufBytes := unsafe.Slice((*byte)(unsafe.Pointer(&dst[0])), dstSize)
	copy(bufBytes, gpuBytes)
	return Err()
}

// NewVAO creates a vertex array object and binds it to current context.
func NewVAO() VertexArray {
	// Configure the Vertex Array Object.
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	return VertexArray{rid: vao}
}

func (vao VertexArray) Bind()   { gl.BindVertexArray(vao.rid) }
func (vao VertexArray) Unbind() { gl.BindVertexArray(0) }

func (vao VertexArray) AddAttribute(vbo VertexBuffer, layout AttribLayout) error {
	if !strings.HasSuffix(layout.Name, "\x00") {
		return ErrStringNotNullTerminated
	}
	if layout.Type == 0 || layout.Packing < 1 || layout.Packing > 4 {
		return errors.New("invalid argument")
	}
	vbo.Bind()
	vertAttrib := gl.GetAttribLocation(layout.Program.rid, gl.Str(layout.Name))
	if vertAttrib < 0 {
		return errors.New("vertex attribute not found:" + layout.Name[:len(layout.Name)-1])
	}
	gl.EnableVertexAttribArray(uint32(vertAttrib))
	// VAO: Vertex Array Object is bound to the vertex buffer on this call.
	// What this line is saying is that `vertAttrib`` index is going to be bound
	// to the current gl.ARRAY_BUFFER (vbo).
	// It also stores size, type, normalized, stride and pointer as vertex array
	// state, in addition to the current vertex array buffer object binding. https://registry.khronos.org/OpenGL-Refpages/gl4/html/glVertexAttribPointer.xhtml
	gl.VertexAttribPointerWithOffset(uint32(vertAttrib), int32(layout.Packing), uint32(layout.Type),
		layout.Normalize, int32(layout.Stride), uintptr(layout.Offset))
	return Err()
}

// Buffer Usages. See BufferUsage documentation for detailed information.
const (
	StaticDraw  BufferUsage = gl.STATIC_DRAW
	StaticRead  BufferUsage = gl.STATIC_READ
	StaticCopy  BufferUsage = gl.STATIC_COPY
	DynamicDraw BufferUsage = gl.DYNAMIC_DRAW
	DynamicRead BufferUsage = gl.DYNAMIC_READ
	DynamicCopy BufferUsage = gl.DYNAMIC_COPY
	StreamDraw  BufferUsage = gl.STREAM_DRAW
	StreamRead  BufferUsage = gl.STREAM_READ
	StreamCopy  BufferUsage = gl.STREAM_COPY
)

// NewVertexBuffer creates a new vertex buffer and binds it.
func NewVertexBuffer[T any](usage BufferUsage, data []T) (VertexBuffer, error) {
	var vbo VertexBuffer
	vertexSize := unsafe.Sizeof(data[0])
	vertPtr := unsafe.Pointer(&data[0])
	gl.GenBuffers(1, &vbo.rid)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.rid)
	gl.BufferData(gl.ARRAY_BUFFER, int(vertexSize)*len(data), vertPtr, uint32(usage))
	return vbo, Err()
}

func (vbo VertexBuffer) Bind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.rid)
}
func (vbo VertexBuffer) Unbind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}
func (vbo VertexBuffer) Delete() {
	gl.DeleteBuffers(1, &vbo.rid)
}

const WriteOnly, ReadOnly, ReadOrWrite AccessUsage = gl.WRITE_ONLY, gl.READ_ONLY, gl.READ_WRITE

// MapBufferData maps vertex buffer memory on the GPU to client space in the form
// of a slice.
func MapBufferData[T any](vbo VertexBuffer, length int, access AccessUsage) ([]T, error) {
	vertexSize := unsafe.Sizeof(*new(T))
	ptr := gl.MapNamedBufferRange(vbo.rid, 0, int(vertexSize)*length, uint32(access))
	err := Err()
	if err != nil {
		return nil, err
	}
	if ptr == nil {
		panic("got nil pointer from MapNamedBufferRange")
	}

	return unsafe.Slice((*T)(ptr), length), nil
}

func GetBufferData[T any](dst []T, vbo VertexBuffer) error {
	vertexSize := unsafe.Sizeof(dst[0])
	vertPtr := unsafe.Pointer(&dst[0])
	// gl.GetBufferDat
	gl.GetBufferSubData(gl.ARRAY_BUFFER, 0, len(dst)*int(vertexSize), vertPtr)
	// gl.GetNamedBufferSubData(vbo.rid, 0, len(dst)*int(vertexSize), vertPtr)
	return Err()
}

func NewIndexBuffer(data []uint32) (IndexBuffer, error) {
	return newIndexBuffer(gl.STATIC_DRAW, data)
}

func newIndexBuffer(usage uint32, data []uint32) (IndexBuffer, error) {
	var ibo IndexBuffer
	const IndexSize = unsafe.Sizeof(data[0])
	vertPtr := unsafe.Pointer(&data[0])
	gl.GenBuffers(1, &ibo.rid)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo.rid)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(IndexSize)*len(data), vertPtr, usage)
	return ibo, Err()
}

func (vbo IndexBuffer) Bind() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, vbo.rid)
}

func (vbo IndexBuffer) Unbind() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
}

func (vbo IndexBuffer) Delete() {
	gl.DeleteBuffers(1, &vbo.rid)
}

type Texture struct {
	rid uint32
	// Usually GL_TEXTURE_2D.
	target uint32
	// Usually TEXTURE0.
	unit uint32
}

func MaxTextureSlots() (textureUnits int) {
	var tu int32
	ptr := &tu
	var p runtime.Pinner
	p.Pin(ptr)
	defer p.Unpin()
	gl.GetIntegerv(gl.MAX_TEXTURE_IMAGE_UNITS, ptr)
	return int(*ptr)
}

func MaxTextureBinded() (textureBounds int) {
	var tu int32
	ptr := &tu
	var p runtime.Pinner
	p.Pin(ptr)
	defer p.Unpin()
	gl.GetIntegerv(gl.MAX_COMBINED_TEXTURE_IMAGE_UNITS, ptr)
	return int(*ptr)
}

// Bind receives a slot onto which to bind from 0 to 32.
func (t Texture) Bind(activeSlot int) {
	gl.ActiveTexture(gl.TEXTURE0 + uint32(activeSlot))
	gl.BindTexture(t.target, t.rid)
}

//	func (t Texture) Unbind() {
//		if err := Err(); err != nil {
//			panic(err)
//		}
//		gl.ActiveTexture(0)
//		gl.BindTexture(t.target, 0)
//		if err := Err(); err != nil {
//			panic(err)
//		}
//	}
func (t Texture) Delete() {
	// gl.BindTexture(t.target, 0)
	// if err := Err(); err != nil {
	// 	panic(err)
	// }
	gl.DeleteTextures(1, &t.rid)
	if err := Err(); err != nil {
		panic(err)
	}
}

const Texture2D TextureType = gl.TEXTURE_2D

func (cfg TextureImgConfig) PixelSize() int {
	var mul, sz int
	switch cfg.Format {
	case gl.RED, gl.RED_INTEGER:
		mul = 1
	case gl.RG, gl.RG_INTEGER:
		mul = 2
	case gl.RGB, gl.RGB_INTEGER:
		mul = 3
	case gl.RGBA, gl.RGBA_INTEGER:
		mul = 4
	default:
		panic("unsupported format. file an issue or PR with its addition!")
	}
	switch cfg.Xtype {
	case gl.FLOAT, gl.INT:
		sz = 4
	default:
		panic("unsupported xtype. file an issue or PR with its addition!")
	}
	return mul * sz
}
func assertImgSameSize[T any](cfg TextureImgConfig, data []T) error {
	sz := cfg.PixelSize() * cfg.Width * cfg.Height
	bufSize := len(data) * int(unsafe.Sizeof(data[0])) // If you are getting panic here please use nil as data.
	if sz != bufSize {
		return errors.New("data size not match to be allocated")
	}
	return nil
}

// NewTextureFromImage creates a new Texture from an image and binds it to the current context.
func NewTextureFromImage[T any](cfg TextureImgConfig, data []T) (Texture, error) {
	var outTexture uint32
	var ptr unsafe.Pointer = nil
	if data != nil {
		if err := assertImgSameSize(cfg, data); err != nil {
			return Texture{}, err
		}
		ptr = unsafe.Pointer(&data[0])
	}
	gl.GenTextures(1, &outTexture)
	tex := Texture{
		rid:    outTexture,
		target: uint32(cfg.Type),
		unit:   uint32(gl.TEXTURE0 + cfg.TextureUnit),
	}
	tex.Bind(cfg.TextureUnit)

	internalFormat := zdefault(cfg.InternalFormat, int32(cfg.Format))
	gl.TexImage2D(tex.target, cfg.Level, internalFormat, int32(cfg.Width), int32(cfg.Height),
		cfg.Border, cfg.Format, cfg.Xtype, ptr)
	// Use default values since OpenGL does not do sane defaults: https://medium.com/@daniel.coady/compute-shaders-in-opengl-4-3-d1c741998c03
	gl.TexParameteri(tex.target, gl.TEXTURE_MAG_FILTER, zdefault(cfg.MagFilter, gl.NEAREST))
	gl.TexParameteri(tex.target, gl.TEXTURE_MIN_FILTER, zdefault(cfg.MinFilter, gl.NEAREST))
	gl.TexParameteri(tex.target, gl.TEXTURE_WRAP_S, zdefault(cfg.Wrap, gl.REPEAT))
	gl.TexParameteri(tex.target, gl.TEXTURE_WRAP_T, zdefault(cfg.Wrap, gl.REPEAT))

	// For following call: format specifies the format that is to be used when performing
	// formatted stores into the image from shaders. format must be compatible with the
	// texture's internal format and must be one of the formats listed in the following table.
	gl.BindImageTexture(cfg.ImageUnit, outTexture, cfg.Level, cfg.Layered, cfg.Layer,
		uint32(cfg.Access), uint32(internalFormat))
	return tex, Err()
}

// SetImage2D sets an existing texture's values on the GPU.
func SetImage2D[T any](tex Texture, cfg TextureImgConfig, data []T) error {
	var ptr unsafe.Pointer = nil
	if data != nil {
		ptr = unsafe.Pointer(&data[0])
	}
	internalFormat := zdefault(cfg.InternalFormat, int32(cfg.Format))
	gl.TextureBarrier()
	gl.TexImage2D(tex.unit, cfg.Level, internalFormat,
		int32(cfg.Width), int32(cfg.Height), cfg.Border, cfg.Format, cfg.Xtype, ptr)
	return Err()
}

func GetImage[T any](dst []T, tex Texture, cfg TextureImgConfig) error {
	if len(dst) == 0 {
		return errors.New("dst cannot be nil or zero length")
	}
	if err := assertImgSameSize(cfg, dst); err != nil {
		return err
	}
	gl.TextureBarrier()
	gl.GetTexImage(tex.target, cfg.Level, cfg.Format, cfg.Xtype, unsafe.Pointer(&dst[0]))
	return Err()
}

// ClearErrors clears all of OpenGL's errors in it's log.
func ClearErrors() {
	i := 0
	for gl.GetError() != gl.NO_ERROR {
		i++
		if i > 2000 {
			panic("forever loop in clear errors. Has the context terminated?")
		}
	}
}

// Err returns a non-nil glErrors if errors are foudn in OpenGL's GetError buffer.
// After a call to Err no more errors should be returned until the next GL call.
func Err() error {
	code := gl.GetError()
	if code == gl.NO_ERROR {
		return nil
	}
	errs := glErrors{glError(code)}
	for {
		code = gl.GetError()
		if code == gl.NO_ERROR {
			return errs
		}
		errs = append(errs, glError(code))
		if len(errs) > 61 {
			lastIdx := len(errs) - 1
			return fmt.Errorf("possible forever loop in Err. Context may be terminated. errs[0]=%v, errs[%d]=%v(%d)", errs[0].String(), lastIdx, errs[lastIdx].String(), uint32(errs[lastIdx]))
		}
	}
}

// Refactor this so we can unwrap a error type local to glgl.
type glErrors []glError

func (ge glErrors) Error() (errstr string) {
	if len(ge) == 0 {
		return "no gl errors"
	}
	for i, e := range ge {
		errstr += e.String()
		if i != len(ge)-1 {
			errstr += "; "
		}
	}
	return errstr
}

type glError uint32

func (ge glError) String() (s string) {
	switch ge {
	case gl.INVALID_ENUM:
		s = "invalid enum"
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		s = "invalid framebuffer operation"
	case gl.INVALID_INDEX:
		s = "invalid index"
	case gl.INVALID_OPERATION:
		s = "invalid operation"
	case gl.INVALID_VALUE:
		s = "invalid value"
	default:
		s = "glError(" + strconv.Itoa(int(ge)) + ")"
	}
	return s
}

// zdefault is a helper function that returns the Default
// value if got is zero.
func zdefault[T constraints.Integer](got, Default T) T {
	if got == 0 {
		return Default
	}
	return got
}

func elemSize[T any]() int {
	var z T
	return int(unsafe.Sizeof(z))
}
