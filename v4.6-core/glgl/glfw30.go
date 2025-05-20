//go:build glfw30 && !tinygo && cgo

package glgl

import (
	"errors"
	"fmt"

	"log/slog"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.0/glfw"
)

func InitWithCurrentWindow30(cfg WindowConfig) (*glfw.Window, func(), error) {
	if cfg.DebugLog != nil {
		glfw.SetErrorCallback(func(code glfw.ErrorCode, desc string) {
			var errstr string
			switch code {
			default:
				errstr = fmt.Sprintf("0x%#[1]X(%[1]d)", code)
			case glfw.NotInitialized:
				errstr = "not initialized"
			case glfw.NoCurrentContext:
				errstr = "no current context"
			case glfw.InvalidEnum:
				errstr = "invalid enum"
			case glfw.InvalidValue:
				errstr = "invalid value"
			case glfw.OutOfMemory:
				errstr = "out of memory"
			case glfw.ApiUnavailable:
				errstr = "api unavailable"
			case glfw.VersionUnavailable:
				errstr = "version unavailable"
			case glfw.PlatformError:
				errstr = "platform error"
			case glfw.FormatUnavailable:
				errstr = "format unavailable"
			}
			cfg.DebugLog.LogAttrs(slog.LevelError, desc, slog.String("glfwErrorCode", errstr))
		})
	}

	if b := glfw.Init(); !b {
		return nil, nil, errors.New("failed to initialized GLFW v3.0")
	}
	glfw.SetErrorCallback(func(code glfw.ErrorCode, desc string) {

	})
	glfw.WindowHint(glfw.Resizable, b2i(!cfg.NotResizable))
	if cfg.Version != [2]int{} {
		glfw.WindowHint(glfw.ContextVersionMajor, cfg.Version[0])
		glfw.WindowHint(glfw.ContextVersionMinor, cfg.Version[1])
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 6)
	}
	glfw.WindowHint(glfw.OpenglProfile, zdefault(cfg.OpenGLProfile, glfw.OpenglCoreProfile))
	glfw.WindowHint(glfw.OpenglForwardCompatible, b2i(cfg.ForwardCompat))
	window, err := glfw.CreateWindow(cfg.Width, cfg.Height, cfg.Title, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		glfw.Terminate()
		return window, nil, err
	}
	ClearErrors()
	return window, glfw.Terminate, nil
}
