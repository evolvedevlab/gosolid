package gowebi

import (
	"fmt"
	"log"

	"github.com/dop251/goja"
)

var clg = func(args ...any) {
	log.Println(append([]any{"[JS]:"}, args...)...)
}

type Runtime struct {
	VM     *goja.Runtime
	Render goja.Callable
	Meta   goja.Callable
}

// runVM runs the loaded program in a new vm.
// meta function is optional and can be nil.
func runVM(program *goja.Program) (*Runtime, error) {
	vm := newVM()

	_, err := vm.RunProgram(program)
	if err != nil {
		return nil, fmt.Errorf("failed to run program: %+v", err)
	}

	render, ok := goja.AssertFunction(vm.Get("render"))
	if !ok {
		return nil, fmt.Errorf("failed to get the render function from vm")
	}

	meta, _ := goja.AssertFunction(vm.Get("meta"))

	return &Runtime{
		VM:     vm,
		Render: render,
		Meta:   meta,
	}, nil
}

func newVM() *goja.Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	vm.Set("setTimeout", func(goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})
	vm.Set("clearTimeout", func(goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})
	vm.Set("console", map[string]func(...any){
		"log":   clg,
		"error": clg,
		"warn":  clg,
	})

	return vm
}
