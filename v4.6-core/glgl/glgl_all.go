package glgl

import (
	"errors"
	"log/slog"
)

type WindowConfig struct {
	Title        string
	NotResizable bool
	Version      [2]int

	OpenGLProfile int // Use [ProfileCore], [ProfileCompat], [ProfileAny].
	ForwardCompat bool
	Width, Height int
	HideWindow    bool // Set glfw.Visible to false
	DebugLog      *slog.Logger
}

type Program struct {
	rid uint32
}

func CompileProgram(ss ShaderSource) (prog Program, err error) {
	if ss.Compute != "" && (ss.Fragment != "" || ss.Vertex != "") {
		return Program{}, errors.New("cannot compile compute and frag/vertex together")
	}
	if ss.Compute == "" && ss.Fragment == "" && ss.Vertex == "" {
		if ss.Include != "" {
			return Program{}, errors.New("only found `#shader include` part of program")
		}
		return Program{}, errors.New("empty program")
	}

	prog, err = compileSources(ss)
	return prog, err
}

type Type uint32

// VertexArray ties data layout with vertex buffer(s).
// Is aware of data layout via VertexAttribPointer* calls.
// Vertex array parameters are client state, that is to say the GPU is unaware of it.
// Loosely speaking, a vertex array
type VertexArray struct {
	rid uint32
}

// AttribLayout is a low level configuration struct
// for adding vertex buffers attribute layouts to a vertex array object.
type AttribLayout struct {
	// The OpenGL program identifier.
	Program Program
	// Type is a OpenGL enum representing the underlying type. Valid types include
	// gl.FLOAT, gl.UNSIGNED_INT, gl.UNSIGNED_BYTE, gl.BYTE etc.
	Type Type
	// Name is the identifier of the attribute in the
	// vertex shader source code finished with a null terminator.
	Name string
	// Packing is a value between 1 and 4 and represents how many
	// of the type are present at the attribute location.
	//
	// Example:
	// When w orking with a vec3 attribute in the shader source code
	// with a gl.Float type, then the Packing is 3 since there are
	// 3 floats packed at each attribute location.
	Packing int
	// Stride is the distance in bytes between attributes in the buffer.
	Stride int
	// Offset is the starting offset with which to start
	// traversing the vertex buffer.
	Offset int
	// specifies whether fixed-point data values should be normalized (when true)
	// or converted directly as fixed-point values (when false) when they are accessed.
	// Usually left as false?
	Normalize bool
}

// BufferUsage is a required hint given to the GPU that provide a general description of
// how exactly the user will be using the buffer object so as to better optimize performance.
//
// There are two independent parts to the usage pattern:
// how the user will be reading/writing from/to the buffer,
// and how often the user will be changing it relative to the use of the data.
//
//   - DRAW: The user will be writing data to the buffer, but the user will not read it.
//   - READ: The user will not be writing data, but the user will be reading it back.
//   - COPY: The user will be neither writing nor reading the data.
//
// There are three hints for how frequently the user will be changing the buffer's data.
//
//   - STATIC: The user will set the data once.
//   - DYNAMIC: The user will set the data occasionally.
//   - STREAM: The user will be changing the data after every use. Or almost every use.
//
// DRAW is useful for, as the name suggests, drawing. The user is uploading data, but only the GL is reading it.
//
// READ is used when a buffer object is used as the destination for OpenGL commands.
// This could be rendering to a Buffer Texture, using arbitrary writes to buffer textures,
// doing a pixel transfer into a buffer object, using Transform Feedback, or any other OpenGL operation that writes to buffer objects.
//
// COPY is used when a buffer object is used to pass data from one place in OpenGL to another.
type BufferUsage uint32

// VertexBuffer contains bytes, no information on the layout or type.
// Buffer objects are said to be "server state", compared to vertex array parameters as "client state".
type VertexBuffer struct {
	// Renderer ID. If using OpenGL is the id set on buffer creation.
	rid uint32
}

type AccessUsage uint32

type IndexBuffer struct {
	// Renderer ID. If using OpenGL is the id set on buffer creation.
	rid uint32
}

type TextureType uint32

// TextureImgConfig builds an image based texture.
// Below are common formats:
// - Base internal. i.e: gl.RED, gl.RG, gl.RGBA, gl.DEPTH_COMPONENT
// - Sized internal: gl.R8, gl.R16, gl.RGB4, gl.R32F, gl.RGBA32F.
type TextureImgConfig struct {
	// Specifies the target texture. Must be one of:
	//  GL_TEXTURE_2D, GL_PROXY_TEXTURE_2D, GL_TEXTURE_1D_ARRAY, GL_PROXY_TEXTURE_1D_ARRAY, GL_TEXTURE_RECTANGLE, GL_PROXY_TEXTURE_RECTANGLE, GL_TEXTURE_CUBE_MAP_POSITIVE_X, GL_TEXTURE_CUBE_MAP_NEGATIVE_X, GL_TEXTURE_CUBE_MAP_POSITIVE_Y, GL_TEXTURE_CUBE_MAP_NEGATIVE_Y, GL_TEXTURE_CUBE_MAP_POSITIVE_Z, GL_TEXTURE_CUBE_MAP_NEGATIVE_Z, or GL_PROXY_TEXTURE_CUBE_MAP.
	Type   TextureType
	Width  int
	Height int
	Border int32
	// Specifies the number of color components in the texture.
	// Can use base, sized or compressed internal formats: See [TextureImgConfig] for more.
	// If not set uses Format.
	InternalFormat int32
	// Specifies format of the pixel data. Accepts:
	//  GL_RED, GL_RG, GL_RGB, GL_BGR, GL_RGBA, GL_BGRA, GL_RED_INTEGER, GL_RG_INTEGER, GL_RGB_INTEGER, GL_BGR_INTEGER, GL_RGBA_INTEGER, GL_BGRA_INTEGER, GL_STENCIL_INDEX, GL_DEPTH_COMPONENT, GL_DEPTH_STENCIL.
	Format uint32

	// Specifies the data type of the pixel data. Accepts
	//   GL_UNSIGNED_BYTE, GL_BYTE, GL_UNSIGNED_SHORT, GL_SHORT, GL_UNSIGNED_INT, GL_INT, GL_HALF_FLOAT, GL_FLOAT, GL_UNSIGNED_BYTE_3_3_2, GL_UNSIGNED_BYTE_2_3_3_REV, GL_UNSIGNED_SHORT_5_6_5, GL_UNSIGNED_SHORT_5_6_5_REV, GL_UNSIGNED_SHORT_4_4_4_4, GL_UNSIGNED_SHORT_4_4_4_4_REV, GL_UNSIGNED_SHORT_5_5_5_1, GL_UNSIGNED_SHORT_1_5_5_5_REV, GL_UNSIGNED_INT_8_8_8_8, GL_UNSIGNED_INT_8_8_8_8_REV, GL_UNSIGNED_INT_10_10_10_2, and GL_UNSIGNED_INT_2_10_10_10_REV.
	Xtype uint32
	// Magnification filtering. gl.NEAREST or gl.LINEAR.
	MagFilter int32
	// Minification filtering. gl.NEAREST or gl.LINEAR.
	MinFilter int32
	// Textures coordinates usually range from (0,0) to (1,1). Wrap indicates
	// how OpenGL is to repeat the texture outside this range.
	// gl.REPEAT, gl.MIRRORED_REPEAT, gl.CLAMP_TO_EDGE, gl.CLAMP_TO_BORDER.
	Wrap int32

	// Specifies a token indicating the type of access that will be performed on the image.
	Access AccessUsage
	// Optional parameters below

	Layered bool
	Layer   int32
	// Specifies the level-of-detail number. Level 0 is the base image level. If target is GL_TEXTURE_RECTANGLE or GL_PROXY_TEXTURE_RECTANGLE, level must be 0.
	Level int32
	// Specifies the unit on which to bind the image onto the texture.
	// This is the binding point for image2D uniforms.
	ImageUnit uint32

	// TextureUnit is the texture unit onto which the texture is loaded (glActiveTexture).
	// TextureUnit starts at 0 and goes all the way up to MaxTextureSlots().
	TextureUnit int
}

// ShaderStorageBuffer is a generic buffer object. Commonly referred to as SSBO.
type ShaderStorageBuffer struct {
	id    uint32
	usage AccessUsage
	sz    int
}

type ShaderStorageBufferConfig struct {
	Usage AccessUsage
	// Base is the binding point for the buffer. To access the buffer within a shader ithe layout parameter `binding`
	// specifies the base of the buffer:
	// 	layout(std430, binding = 0) buffer ImageBuffer {
	// 		uint pixelData[];
	// 	};
	Base uint32
	// MemSize specifies the size in bytes of memory the ShaderStorageBuffer should take up.
	// When creating buffers with [NewShaderStorageBuffer] MemSize will decide the minimum size the buffer have,
	// if the data slice is nil.
	MemSize uint32
}
