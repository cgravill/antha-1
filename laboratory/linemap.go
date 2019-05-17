package laboratory

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

type lineMapManager struct {
	elementMaps map[string]*elementMap
}

func NewLineMapManager() *lineMapManager {
	return &lineMapManager{
		elementMaps: make(map[string]*elementMap),
	}
}

type elementMap struct {
	anthaElementPath string
	elementTypeName  string
	lineMap          map[int]int
}

// Neither the goElementPath nor the anElementPath need to be full
// paths, but they should be in filepath format, and they will be
// tested as suffixes against the frames in the stack.
func (lmm *lineMapManager) RegisterLineMap(elementTypeName, goElementPath, anElementPath string, lineMap map[int]int) {
	em := &elementMap{
		anthaElementPath: anElementPath,
		elementTypeName:  elementTypeName,
		lineMap:          lineMap,
	}
	lmm.elementMaps[goElementPath] = em
}

// ElementStackTrace creates a stack trace, detecting whether or not
// the panic occured within an element. Essentially, the normal
// debug.Stack() is returned, but modified whenever the stack trace
// passes through an element. We use the registered line maps to
// create a stack trace which refers back to the original elements,
// with the correct line numbers.
func (lmm *lineMapManager) ElementStackTrace() string {
	// This is a magic number :( It limits us to dealing with stack
	// traces that are 1000 frames deep. It is not expected this will
	// be a problem in practice!
	cs := make([]uintptr, 1000)

	// When a panic occurs, if a defer-with-recover has been
	// registered, the stack itself does not unwind. Instead, the
	// recover is invoked in a sub-frame:
	//
	// - Frame of defer-with-recover func
	// - Frame of runtime.gopanic
	// - (optional) One or more frames of panic detail, eg
	//     runtime.panicdivide, runtime.panicmem, runtime.sigpanic
	// - Frame of function that panicked
	// - ... rest of call stack ...
	//
	// If the panic is caught with a recover and then re-thrown, then
	// that adds more frames to the top of the call stack.
	//
	// In the defer, if we call ElementStackTrace then that adds
	// another frame, and then we further have to call runtime.Callers
	// in order to generate the stack frame. At this point, the stack
	// will look like this:
	//
	// - Frame of runtime.Callers
	// - Frame of ElementStackTrace
	// - Frame of defer-with-recover func  \___ maybe repeated
	// - Frame of runtime.gopanic          /
	// - (optional) One or more frames of panic detail, eg
	//     runtime.panicdivide, runtime.panicmem, runtime.sigpanic
	// - Frame of function that panicked
	// - ... rest of call stack ...
	//
	num := runtime.Callers(0, cs)
	stack := string(debug.Stack())
	if num < 0 {
		return stack
	}
	// Now, the Go runtime API doesn't allow us to construct the stack
	// trace in the same way as debug.Stack(). So rather than trying,
	// we simply walk through the frames, find any that refer to
	// elements, and try to locate and modify the corresponding lines.
	stackSplit := strings.Split(stack, "\n")
	result := make([]string, 0, len(stackSplit))

	frames := runtime.CallersFrames(cs[:num])
	frame, more := frames.Next()
	for {
		var elem *elementMap
		for suffix, em := range lmm.elementMaps {
			if strings.HasSuffix(frame.File, suffix) {
				elem = em
				break
			}
		}
		if elem != nil {
			for len(stackSplit) > 0 {
				line := stackSplit[0]
				stackSplit = stackSplit[1:]
				result = append(result, line)
				if strings.HasPrefix(line, fmt.Sprintf("\t%s:%d", frame.File, frame.Line)) { // this line is the 2nd of every frame!
					// the current line and previous line should be overwritten:
					result = result[:len(result)-2]
					lineStr := "(unknown line)"
					if line, foundLine := elem.lineMap[frame.Line]; foundLine {
						lineStr = fmt.Sprint(line)
					}
					result = append(result,
						frame.Function,
						fmt.Sprintf("\t[ElementType %s] %s:%s", elem.elementTypeName, elem.anthaElementPath, lineStr),
						fmt.Sprintf("\t[Go] %s:%d", frame.File, frame.Line))
					break
				}
			}
		}

		if more {
			frame, more = frames.Next()
		} else {
			result = append(result, stackSplit...)
			return strings.Join(result, "\n")
		}
	}

	return stack
}
