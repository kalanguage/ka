package types

type Overload struct {
  Params []string
  Types  []string
  Body   []Action
}

type OmmFunc struct {
  Overloads []Overload
}

func (f OmmFunc) Format() string {
  return "(function) { ... }"
}

func (f OmmFunc) Type() string {
  return "function"
}

func (f OmmFunc) TypeOf() string {
  return f.Type()
}
