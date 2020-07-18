package compiler

import "fmt"
import "os"

import "lang/interpreter"
import . "lang/types"

type _CompileErr struct {
  Msg   string
  FName string
  Line  uint64
}

func (e _CompileErr) Print() {
  fmt.Println("Error while compiling", e.FName, "on line", e.Line)
  fmt.Println(e.Msg)
  os.Exit(1)
}

type CompileErr interface {
  Print()
}

func makeCompilerErr(msg, fname string, line uint64) CompileErr {
  return _CompileErr{
    Msg: msg,
    FName: fname,
    Line: line,
  }
}

//export Compile
func Compile(file, filename string) ([]Action, map[string][]Action, CompileErr) {

  var e CompileErr

  lex, e := lexer(file, filename)

  if e != nil {
    return []Action{}, nil, e
  }

  groups := makeGroups(lex)
  operations := makeOperations(groups)
  actions, e := actionizer(operations)

  if e != nil {
    return []Action{}, nil, e
  }

  //a bunch of validations and initializers
  e = has_non_global_prototypes(actions, true)
  if e != nil {
    return []Action{}, nil, e
  }
  put_proto_types(actions)
  e = validate_types(actions)
  if e != nil {
    return []Action{}, nil, e
  }
  /////////////////////////////////////////

  vars, e := getvars(actions)
  if e != nil {
    return nil, nil, e
  }

  //make each var have only it's name
  var varnames = make(map[string]string)

  for k := range vars {

    if vars[k][0].Type != "function" {
      changevarnames(vars[k], varnames) //ensure none of the globals use the globals from below
    }
    varnames[k] = k
  }

  //also account for the gofuncs
  for k := range interpreter.GoFuncs {
    varnames["$" + k] = "$" + k
  }


  for k := range vars {
    e = changevarnames(vars[k], varnames)

    if e != nil {
      return []Action{}, nil, e
    }
  }

  return actions, vars, nil
}
