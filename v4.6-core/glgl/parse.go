package glgl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// Vertex and Fragment are null terminated strings with source code.
type ShaderSource struct {
	// Vertex and Fragment are null terminated strings with source code.
	Vertex   string
	Fragment string
	Compute  string
	Include  string

	// CompileFlags controls how program is compiled. See [CompileFlags].
	CompileFlags CompileFlags
}

// ParseCombinedBasic parses a file with vertex and fragment #shader pragmas inspired
// by [The Cherno]'s take on shader file segmenting. This method of writing
// shaders lets one keep vertex and fragment shader source code in the same file:
//
//	// Anything above first #shader pragma is ignored.
//	#shader vertex
//	void main() {
//	    gl_Position = vec4(1.0,0.0,0.0, 1.0);
//	}
//
//	#shader fragment
//	void main() {
//	    gl_Frag = gl_Position/2;
//	}
//
// `compute` and `includeashead` are also valid #shader pragmas.
// ParseCombined performs no calls to the GL.
//
// [The Cherno]: https://www.youtube.com/watch?v=2pv0Fbo-7ms&list=PLlrATfBNZ98foTJPJ_Ev03o2oq3-GGOS2&index=9&t=724s&ab_channel=TheCherno
func ParseCombined(r io.Reader) (ss ShaderSource, err error) {
	const (
		shaderNone = iota
		shaderVertex
		shaderFragment
		shaderCompute
		shaderHeader
		shaderNum
	)
	nothing := bytes.NewBuffer(nil)
	vertexBuf := bytes.NewBuffer(nil)
	fragBuf := bytes.NewBuffer(nil)
	computeBuf := bytes.NewBuffer(nil)
	includeBuf := bytes.NewBuffer(nil)
	buffers := [shaderNum]*bytes.Buffer{
		shaderNone:     nothing,
		shaderVertex:   vertexBuf,
		shaderFragment: fragBuf,
		shaderCompute:  computeBuf,
		shaderHeader:   includeBuf,
	}
	scanner := bufio.NewScanner(r)
	currentShader := shaderNone
	for scanner.Scan() {
		line := scanner.Bytes()
		if currentShader != shaderNone && !bytes.HasPrefix(bytes.TrimSpace(line), []byte("#shader ")) {
			buffers[currentShader].Write(line)
			buffers[currentShader].WriteByte('\n')
			continue
		}
		got := bytes.Fields(line)
		if len(got) != 2 {
			continue
		}
		switch string(got[1]) {
		case "includeashead":
			currentShader = shaderHeader
		case "vertex":
			currentShader = shaderVertex
		case "fragment", "pixel":
			currentShader = shaderFragment
		case "compute":
			currentShader = shaderCompute
		default:
			return ShaderSource{}, errors.New("unexpected #shader pragma value:" + string(got[1]))
		}
	}
	isrc := includeBuf.Bytes()
	var vsrc, fsrc, csrc []byte
	if vertexBuf.Len() > 0 {
		vsrc = append(vsrc, isrc...)
		vertexBuf.WriteByte(0)
		vsrc = append(vsrc, vertexBuf.Bytes()...)
	}
	if computeBuf.Len() > 0 {
		csrc = append(csrc, isrc...)
		computeBuf.WriteByte(0)
		csrc = append(csrc, computeBuf.Bytes()...)
	}
	if fragBuf.Len() > 0 {
		fsrc = append(fsrc, isrc...)
		fragBuf.WriteByte(0)
		fsrc = append(fsrc, fragBuf.Bytes()...)
	}
	return ShaderSource{
		Vertex:   string(vsrc),
		Fragment: string(fsrc),
		Compute:  string(csrc),
		Include:  string(isrc),
	}, scanner.Err()
}

// CompileFlags controls how the shader is compiled. Currently serves to control error handling vs performance tradeoffs.
type CompileFlags uint64

// Base compiler flags.
const (
	// CompileFlagValidateProgram flags that a validation of the program should be done. This can be an expensive operation
	// and should only be reserved for debug and test builds to avoid wasting cycles.
	CompileFlagValidateProgram CompileFlags = 1 << iota // validate program
	// Omits error handling during shader compile. Setting may cause hard to debug errors.
	CompileFlagNoCompileCheck // no compile check
	// Omits error handling during shader linking. Setting may cause hard to debug errors.
	CompileFlagNoLinkCheck // no link check
)

// Composed compiler flags.
const (
	// CompileFlagsLax disables all error handling down to the bare minimum, performing less error handling for the benefit of performance.
	// Note that the returned shader is not guaranteed to be valid if this is set. It is strongly suggested the user call [Err] at some point after compiling if using this setting.
	// Set when preferring performance over program correctness. Is polar opposite of [CompileFlagsStrict].
	CompileFlagsLax = CompileFlagNoCompileCheck | CompileFlagNoLinkCheck
	// CompileFlagsStrict enables all stricter error handling options and checking throughout the compile step. Is polar opposite of [CompileFlagsLax].
	CompileFlagsStrict = CompileFlagValidateProgram
)

func (cf CompileFlags) checkCompile() bool    { return cf&CompileFlagNoCompileCheck == 0 }
func (cf CompileFlags) checkLink() bool       { return cf&CompileFlagNoLinkCheck == 0 }
func (cf CompileFlags) validateProgram() bool { return cf&CompileFlagValidateProgram != 0 }
