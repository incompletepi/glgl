//go:build !tinygo && cgo

package glgl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
)

// RunCompute runs a the program's compute shader with defined work sizes and waits for it to finish.
func (p Program) RunCompute(workSizeX, workSizeY, workSizeZ int) error {
	gl.DispatchCompute(uint32(workSizeX), uint32(workSizeY), uint32(workSizeZ))
	err := Err()
	if err != nil {
		return err
	}
	// Wait for compute to finish.
	gl.MemoryBarrier(gl.ALL_BARRIER_BITS)
	return Err()
}

func (p Program) BindFrag(name string) error {
	if !strings.HasSuffix(name, "\x00") {
		return ErrStringNotNullTerminated
	}
	gl.BindFragDataLocation(p.rid, 0, gl.Str(name))
	return nil
}

func (p Program) ID() uint32 {
	return p.rid
}

func (p Program) Bind()   { gl.UseProgram(p.rid) }
func (p Program) Unbind() { gl.UseProgram(0) }

// Delete deletes p. Make sure program is binded before deletion.
func (p Program) Delete() {
	if p.rid == 0 {
		// A program ID of zero will be silently ignored by the GL.
		panic("got program id of zero. Did you correctly create the program?")
	}
	p.Unbind()
	gl.DeleteProgram(p.rid)
}

func (p Program) AttribLocation(name string) (uint32, error) {
	if !strings.HasSuffix(name, "\x00") {
		return 0, ErrStringNotNullTerminated
	}
	loc := gl.GetAttribLocation(p.rid, gl.Str(name))
	if loc < 0 {
		return uint32(loc), errors.New("unable to find attribute in program- did you use the identifier so it was not stripped from program?")
	}
	return 0, nil
}

func (p Program) UniformLocation(name string) (int32, error) {
	if !strings.HasSuffix(name, "\x00") {
		return -2, ErrStringNotNullTerminated
	}
	loc := gl.GetUniformLocation(p.rid, gl.Str(name))
	if loc < 0 {
		return loc, errors.New("unable to find uniform in program- did you use the identifier so it was not stripped from program?")
	}
	return loc, nil
}

func (p Program) SetUniformf(loc int32, floats ...float32) error {
	switch len(floats) {
	case 1:
		gl.Uniform1f(loc, floats[0])
	case 2:
		gl.Uniform2f(loc, floats[0], floats[1])
	case 3:
		gl.Uniform3f(loc, floats[0], floats[1], floats[2])
	case 4:
		gl.Uniform4f(loc, floats[0], floats[1], floats[2], floats[3])
	default:
		return errors.New("bad number of floats to SetUniformsf")
	}
	return Err()
}

func (p Program) SetUniformi(loc int32, ints ...int32) error {
	switch len(ints) {
	case 1:
		gl.Uniform1i(loc, ints[0])
	case 2:
		gl.Uniform2i(loc, ints[0], ints[1])
	case 3:
		gl.Uniform3i(loc, ints[0], ints[1], ints[2])
	case 4:
		gl.Uniform4i(loc, ints[0], ints[1], ints[2], ints[3])
	default:
		return errors.New("bad number of ints to SetUniformsi")
	}
	return Err()
}

func (p Program) SetUniformui(loc int32, ints ...uint32) error {
	switch len(ints) {
	case 1:
		gl.Uniform1ui(loc, ints[0])
	case 2:
		gl.Uniform2ui(loc, ints[0], ints[1])
	case 3:
		gl.Uniform3ui(loc, ints[0], ints[1], ints[2])
	case 4:
		gl.Uniform4ui(loc, ints[0], ints[1], ints[2], ints[3])
	default:
		return errors.New("bad number of uints to SetUniformsui")
	}
	return Err()
}

// CompileBasic compiles two OpenGL vertex and fragment shaders
// and returns a program with the current OpenGL context.
// It returns an error if compilation, linking or validation fails.
func compileSources(ss ShaderSource) (program Program, err error) {
	// Note: glDeleteShader only flags a shader for deletion.
	// They are not deleted until they are detached from the program.
	// Beware: multiple calls to glDeleteShader on the same shader will cause an error on GL's side.
	program.rid = gl.CreateProgram()
	if program.rid == 0 {
		return Program{}, errors.New("silently got invalid program ID. Are you calling from the main thread? Remember to call runtime.LockOSThread() from your main thread")
	}

	// Some inspiration taken from github.com/TheCherno/Hazel/src/Platform/OpenGL/OpenGLShader.cpp
	// Hazel detaches and deletes shaders immediately after creating program.
	// Apparently no need to persist them after the fact.
	var shaders []uint32
	var linked bool
	defer func() {
		for _, sid := range shaders {
			if linked {
				gl.DetachShader(program.rid, sid)
			}
			gl.DeleteShader(sid)
		}
	}()
	flags := ss.CompileFlags
	if len(ss.Vertex) > 0 {
		vid, err := compile(gl.VERTEX_SHADER, flags, ss.Vertex)
		if err != nil {
			return Program{}, fmt.Errorf("vertex shader compile: %w", err)
		}
		gl.AttachShader(program.rid, vid)
		shaders = append(shaders, vid) // for cleanup
	}
	if len(ss.Fragment) > 0 {
		fid, err := compile(gl.FRAGMENT_SHADER, flags, ss.Fragment)
		if err != nil {
			return Program{}, fmt.Errorf("fragment shader compile: %w", err)
		}
		gl.AttachShader(program.rid, fid)
		shaders = append(shaders, fid) // for cleanup
	}
	if len(ss.Compute) > 0 {
		cid, err := compile(gl.COMPUTE_SHADER, flags, ss.Compute)
		if err != nil {
			return Program{}, fmt.Errorf("compute shader compile: %w", err)
		}
		gl.AttachShader(program.rid, cid)
		shaders = append(shaders, cid) // for cleanup
	}

	gl.LinkProgram(program.rid)
	if flags.checkLink() {
		err = ivLogErr(program.rid, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog)
		if err != nil {
			return Program{}, fmt.Errorf("link failed: %w", err)
		}
	}

	// Link flag set so shaders detached on return, see: https://www.youtube.com/watch?v=71BLZwRGUJE&list=PLlrATfBNZ98foTJPJ_Ev03o2oq3-GGOS2&index=7&ab_channel=TheCherno
	linked = true
	if flags.validateProgram() {
		gl.ValidateProgram(program.rid)
		err = ivLogErr(program.rid, gl.VALIDATE_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog)
		if err != nil {
			return Program{}, fmt.Errorf("validation failed: %w", err)
		}
	}
	return program, nil
}

func compile(shaderType uint32, flags CompileFlags, sourceCodes ...string) (uint32, error) {
	var sourceLengths []int32
	for i := range sourceCodes {
		if !strings.HasSuffix(sourceCodes[i], "\x00") {
			return 0, errors.New("source missing null terminator")
		}
		sourceLengths = append(sourceLengths, int32(len(sourceCodes[i])))
	}

	id := gl.CreateShader(shaderType)
	if id == 0 {
		if err := Err(); err != nil {
			return 0, fmt.Errorf("got invalid shader ID: %w", err)
		}
		return 0, fmt.Errorf("silently got invalid shader id 0")
	}
	csources, free := gl.Strs(sourceCodes...)
	gl.ShaderSource(id, int32(len(sourceCodes)), csources, &sourceLengths[0])
	free()

	gl.CompileShader(id)
	// We now check the errors during compile, if there were any.
	if flags.checkCompile() {
		err := ivLogErr(id, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog)
		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

// ivLogErr calls ivLog expecting an error in return parameter. It also calls [Err] on getting a non-empty log message and joins both.
func ivLogErr(id, plName uint32, getIV func(program uint32, pname uint32, params *int32), getInfo func(program uint32, bufSize int32, length *int32, infoLog *uint8)) error {
	log := ivLog(id, plName, getIV, getInfo)
	if len(log) > 0 {
		err := errors.New(log)
		if err2 := Err(); err2 != nil {
			return errors.Join(err, err2)
		}
		return err
	}
	return nil
}

// ivLog is a helper function for extracting log data
// from a Shader compilation step or program linking.
//
//	log := ivLog(id, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog)
//	if len(log) > 0 {
//		return 0, errors.New(log)
//	}
func ivLog(id, plName uint32, getIV func(program uint32, pname uint32, params *int32), getInfo func(program uint32, bufSize int32, length *int32, infoLog *uint8)) string {
	var iv int32
	getIV(id, plName, &iv)
	if iv == gl.FALSE {
		var logLength int32
		getIV(id, gl.INFO_LOG_LENGTH, &logLength)
		if logLength == 0 {
			// panic(fmt.Sprintf("unexpected false iv for plName=0x%#X with no log", plName))
			fmt.Print("TODO: fix ivLog in shaders.go")
			return ""
		}
		log := make([]byte, logLength)
		getInfo(id, logLength, &logLength, &log[0])
		return string(log[:len(log)-1]) // we exclude the last null character.
	}
	return ""
}
