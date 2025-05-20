package main

import (
	_ "embed"
	"fmt"
	_ "image/png"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/incompletepi/glgl/v4.6-core/glgl"
)

//go:embed triangle.glsl
var shader string

var triangleVertices = []float32{
	-0.5, -0.5,
	0.0, 0.5,
	0.5, -0.5,
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	window, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:  "Hello triangle",
		Width:  800,
		Height: 800,
	})
	if err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer terminate()

	fmt.Println("OpenGL version", glgl.Version())

	// Separate vertex and fragment shaders from source code.
	source, err := glgl.ParseCombined(strings.NewReader(shader))
	if err != nil {
		panic(err)
	}
	//
	prog, err := glgl.CompileProgram(source)
	if err != nil {
		panic(err)
	}
	prog.Bind()
	defer prog.Delete()
	// Bind output of fragment shader.
	prog.BindFrag("outputColor\x00")

	// Create the vertex array object to store data layout.
	vao := glgl.NewVAO()

	// float32 is 4 bytes wide.
	vbo, err := glgl.NewVertexBuffer(glgl.StaticDraw, triangleVertices)
	if err != nil {
		panic(err)
	}
	const attrSize = 4
	err = vao.AddAttribute(vbo, glgl.AttribLayout{
		Program: prog,
		Type:    glgl.Float32,
		Name:    "vert\x00",   // The name of the variable in the triangle.glsl shader source code.
		Packing: 2,            // Is a 2D point, 2 floats.
		Stride:  2 * attrSize, // To reach next vertex must traverse two float32, each 4 bytes.
	})
	if err != nil {
		panic(err)
	}
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		// NOTE: If nothing is visible maybe add a gl.BindVertexArray(vao) call in here and file a bug!
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			window.SetShouldClose(true)
		}
	}
}
