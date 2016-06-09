package py

import (
	"os"
	"testing"

	"github.com/DataDog/datadog-agent/aggregator"
	"github.com/sbinet/go-python"
)

// Setup the test module
func TestMain(m *testing.M) {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}
	// Set the PYTHONPATH
	path := python.PySys_GetObject("path")
	python.PyList_Append(path, python.PyString_FromString("."))

	// Initialize acquires the GIL but we don't need it, release it
	state := python.PyEval_SaveThread()

	// for now, only Python needs it, build and pass it on the fly
	aggregator.InitApi(aggregator.NewUnbufferedAggregator())

	ret := m.Run()

	python.PyEval_RestoreThread(state)
	python.Finalize()

	os.Exit(ret)
}

// cut down some boilerplate
func assertNil(t *testing.T, sclass *python.PyObject) {
	if sclass != nil {
		t.Fatalf("Expected nil, found: %v", sclass)
	}
}

func TestFindSubclassOf(t *testing.T) {
	// Lock the GIL and release it at the end of the run
	_gstate := python.PyGILState_Ensure()
	defer func() {
		python.PyGILState_Release(_gstate)
	}()

	fooModule := python.PyImport_ImportModuleNoBlock("tests.foo")
	fooClass := fooModule.GetAttrString("Foo")
	barModule := python.PyImport_ImportModuleNoBlock("tests.bar")
	barClass := barModule.GetAttrString("Bar")

	// invalid input
	sclass := findSubclassOf(nil, nil)
	assertNil(t, sclass)

	// pass something that's not a Type
	sclass = findSubclassOf(python.PyTuple_New(0), fooModule)
	assertNil(t, sclass)
	sclass = findSubclassOf(fooClass, python.PyTuple_New(0))
	assertNil(t, sclass)

	// Foo in foo module, only Foo itself found
	sclass = findSubclassOf(fooClass, fooModule)
	assertNil(t, sclass)

	// Bar in foo module, no class found
	sclass = findSubclassOf(barClass, fooModule)
	assertNil(t, sclass)

	// Foo in bar module, get Bar
	sclass = findSubclassOf(fooClass, barModule)
	if sclass == nil || sclass.RichCompareBool(barClass, python.Py_EQ) < 1 {
		t.Fatalf("Expected Bar, found: %v", sclass)
	}
}

func TestGetModuleName(t *testing.T) {
	name := getModuleName("foo.bar.baz")
	if name != "baz" {
		t.Fatalf("Expected baz, found: %s", name)
	}

	name = getModuleName("baz")
	if name != "baz" {
		t.Fatalf("Expected baz, found: %s", name)
	}

	name = getModuleName("")
	if name != "" {
		t.Fatalf("Expected empty string, found: %s", name)
	}
}
