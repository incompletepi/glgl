# glgl
OpenGL bindings for Go that are in the goldilocks level of abstraction zone.

_WIP_.

## Highlight #1
Excellent documentation. Well suited to learn OpenGL without the pain.
More consistent and precise naming, i.e. "Packing" is much more precise
than OpenGL's Size parameter for AttributePointer functions since Size does not
refer to a size in bytes but rather the amount of packed types at the attribute location

```go
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
	// When working with a vec3 attribute in the shader source code
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
```

## Migration to `geometry` package
The `math` package is being moved to [`soypat/geometry`](https://github.com/soypat/geometry). To migrate replace the following import paths in your project:

- `"github.com/incompletepi/glgl/math/ms1"` -> `"github.com/soypat/geometry/ms1"`
- `"github.com/incompletepi/glgl/math/ms2"` -> `"github.com/soypat/geometry/ms2"`
- `"github.com/incompletepi/glgl/math/ms3"` -> `"github.com/soypat/geometry/ms3"`
- `"github.com/incompletepi/glgl/math/md1"` -> `"github.com/soypat/geometry/md1"`
- `"github.com/incompletepi/glgl/math/md2"` -> `"github.com/soypat/geometry/md2"`
- `"github.com/incompletepi/glgl/math/md3"` -> `"github.com/soypat/geometry/md3"`

Or more simply, run a single replace directive in all of your project:
- `"github.com/incompletepi/glgl/math/` -> `"github.com/soypat/geometry/`
                                                                                                    