package types

import (
	"errors"
)

type TuskProto struct {
	ProtoName  string
	Static     map[string]*TuskType
	Instance   map[string]*TuskType
	AccessList map[string][]string
}

func getfield(full map[string]*TuskType, field string, access map[string][]string, file string) (*TuskType, error) {
	if field[0] == '_' {
		return nil, errors.New("Cannot access private member: " + field)
	}

	//check for access (protected)

	if access[field] == nil || file == "" { //if it does not name any access, automatically make it public
		goto allowed
	}

	for _, v := range access[field] {
		if file == v {
			goto allowed
		}
	}

	return nil, errors.New("File cannot acces field \"" + field + "\"")

allowed:
	fieldv := full[field]

	if fieldv == nil {
		return nil, errors.New("Prototype does not contain the field \"" + field + "\"")
	}

	return fieldv, nil
}

func (p TuskProto) Get(field string, file string) (*TuskType, error) {
	return getfield(p.Static, field, p.AccessList, file)
}

func (p TuskProto) Format() string {
	return "{" + p.ProtoName + "}"
}

func (p TuskProto) Type() string {
	return "prototype"
}

func (p TuskProto) TypeOf() string {
	return p.ProtoName + " prototype"
}

func (p TuskProto) Deallocate() {}

func (p TuskProto) Clone() *TuskType {
	return nil
}

//Range ranges over a prototype
func (p TuskProto) Range(fn func(val1, val2 *TuskType) (Returner, *TuskError)) (*Returner, *TuskError) {
	return nil, nil
}
