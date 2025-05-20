package main

import (
	_ "embed"
	"fmt"
	_ "image/png"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"log/slog"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/incompletepi/glgl/v4.6-core/glgl"
)

// Very basic index buffer example.

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

//go:embed uniformtriangle.glsl
var shader string

// Square with indices:
// 3----2
// |    |
// 0----1
var positions = []float32{
	-0.5, -0.5, // 0
	0.5, -0.5, // 1
	0.5, 0.5, // 2
	-0.5, 0.5, //3
}

// We'll solve this one using index buffers.
var indices = []uint32{
	0, 1, 2, // Lower right triangle.
	0, 2, 3, // Upper left triangle.
}

func main() {
	window, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:  "Hello triangle",
		Width:  800,
		Height: 800,
	})
	if err != nil {
		slog.Error("glfw or gl init failed", "err", err.Error())
		return
	}
	defer terminate()
	fmt.Println("OpenGL version", glgl.Version())

	// Parse and compile source code.
	source, err := glgl.ParseCombined(strings.NewReader(shader))
	if err != nil {
		slog.Error("parse program failed", "err", err.Error())
		return
	}
	prog, err := glgl.CompileProgram(source)
	if err != nil {
		slog.Error("compile program failed", "err", err.Error())
		return
	}
	prog.Bind()
	defer prog.Delete()

	err = prog.BindFrag("outputColor\x00")
	if err != nil {
		slog.Error("program bind frag fail", "err", err.Error())
		return
	}
	// Configure the Vertex Array Object.
	vao := glgl.NewVAO()

	// Create the Position Buffer Object.
	vbo, err := glgl.NewVertexBuffer(glgl.StaticDraw, positions)
	if err != nil {
		slog.Error("creating positions vertex buffer", "err", err.Error())
		return
	}
	err = vao.AddAttribute(vbo, glgl.AttribLayout{
		Program: prog,
		Type:    glgl.Float32,
		Name:    "vert\x00",
		Packing: 2,
		Stride:  2 * 4, // 2 floats, each 4 bytes wide.
	})
	if err != nil {
		slog.Error("adding attribute vert", "err", err.Error())
		return
	}

	// Create Index Buffer Object.
	_, err = glgl.NewIndexBuffer(indices)
	if err != nil {
		slog.Error("creating index buffer", "err", err.Error())
		return
	}

	// Set uniform variable `u_color` in source code.
	colorLoc, err := prog.UniformLocation("u_color\x00")
	if err != nil {
		slog.Error("searching for color uniform", "err", err.Error())
		return
	}
	err = prog.SetUniformf(colorLoc, 0.2, 0.3, 0.8, 1)
	if err != nil {
		slog.Error("creating index buffer", "err", err.Error())
		return
	}
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

		prog.SetUniformf(colorLoc, float32(time.Now().UnixMilli()%1000)/1000, .5, .3, 1)
		// Maintenance
		glfw.SwapInterval(1) // Can prevent epilepsy for high frequency
		window.SwapBuffers()
		glfw.PollEvents()
		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			window.SetShouldClose(true)
		}
	}
}
