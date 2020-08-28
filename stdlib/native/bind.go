package native

import (
	. "github.com/omm-lang/omm/lang/types"
)

type OmmGoFunc struct {
	Function func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType
}

func (ogf OmmGoFunc) Format() string {
	return "{ native gofunc }"
}

func (ogf OmmGoFunc) Type() string {
	return "gofunc"
}

func (ogf OmmGoFunc) TypeOf() string {
	return ogf.Type()
}

func (_ OmmGoFunc) Deallocate() {}

//the simple native functions are stored in lang/interpreter/nativefuncs.go

//GetStd returns all of the more complex natives (they are not all functions)
func GetStd() (map[string]*OmmType, map[string]func(val1, val2 OmmType, instance *Instance, stacktrace []string, line uint64, file string, stacksize uint) *OmmType) {

	//return 1 is the native vars, and return 2 is the native operations (operations that the native vars make)

	var native = make(map[string]OmmType)
	var operations = make(map[string]func(val1, val2 OmmType, instance *Instance, stacktrace []string, line uint64, file string, stacksize uint) *OmmType)

	native["url.request"] = OmmGoFunc{
		Function: urlrequest,
	}
	operations["http-response :: string"] = http_resp_field

	nativeptrs := make(map[string]*OmmType) //make everything into a pointer to an OmmType
	for k, v := range native {
		nativeptrs[k] = &v
	}

	return nativeptrs, operations
}
