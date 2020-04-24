package main

import "strconv"
import "encoding/json"
import "math/big"
import "fmt"

// #cgo CFLAGS: -std=c99
import "C"

//export AddStrings
func AddStrings(num1, num2, calc_params *C.char, line C.int) *C.char {

  var cp paramCalcOpts

  _ = json.Unmarshal([]byte(C.GoString(calc_params)), &cp)

  sum := add(C.GoString(num1), C.GoString(num2), cp, int(line))

  return C.CString(sum)
}

func add(num1 string, num2 string, calc_params paramCalcOpts, line int) string {
  calc := new(big.Float)

  num1big, _ := new(big.Float).SetPrec(PREC).SetString(num1)
  num2big, _ := new(big.Float).SetPrec(PREC).SetString(num2)

  return returnInit(fmt.Sprintf("%f", calc.Add(num1big, num2big)))
}

//export AddC
func AddC(num1, num2 *C.char) *C.char {
  return C.CString(add(C.GoString(num1), C.GoString(num2), paramCalcOpts{}, -1))
}

//export Add
func Add(_num1P *C.char, _num2P *C.char, calc_paramsP *C.char, line_ C.int) *C.char {

  __num1 := C.GoString(_num1P)
  __num2 := C.GoString(_num2P)
  calc_params_str := C.GoString(calc_paramsP)

  line := int(line_)

  _ = line

  var calc_params paramCalcOpts

  _ = json.Unmarshal([]byte(calc_params_str), &calc_params)

  var _num1P_ Action
  var _num2P_ Action

  _ = json.Unmarshal([]byte(__num1), &_num1P_)
  _ = json.Unmarshal([]byte(__num2), &_num2P_)

  nums := TypeOperations{ _num1P_.Type, _num2P_.Type }

  /* TABLE OF TYPES:

    string + (* - array - none - hash) = string
    array + (* - none) = array
    none + * = falsey
    hash + (* - hash) = none
    type + (* - hash - none) = type
    num + num = num
    hash + hash = hash
    boolean + boolean = boolean
    num + boolean = num
    default = falsey
  */

  var finalRet Action

  switch nums {
    case TypeOperations{ "number", "number" }: { //detect case "num" + "num"

      numRet := returnInit(add(_num1P_.ExpStr[0], _num2P_.ExpStr[0], calc_params, line))

      finalRet = Action{ "number", "", []string{ numRet }, []Action{}, []string{}, []Action{}, []Condition{}, 39, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }
    }
    case TypeOperations{ "boolean", "boolean" }: { //detect case "boolean" + "boolean"

      val1 := _num1P_.ExpStr[0]
      val2 := _num2P_.ExpStr[0]

      boolSwitch := BoolSwitch{ val1, val2 }

      var final_ string

      switch (boolSwitch) {
        case BoolSwitch{ "true", "true" }:
          final_ = "1"
        case BoolSwitch{ "true", "false"}:
          final_ = "1"
        case BoolSwitch{ "false", "true" }:
          final_ = "1"
        case BoolSwitch{ "false", "false" }:
          final_ = "0"
        default: final_ = "0"
      }

      finalRet = Action{ "boolean", "", []string{ final_ }, []Action{}, []string{}, []Action{}, []Condition{}, 40, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }
    }
    case TypeOperations{ "number", "boolean" }: { //detect case "num" + "boolean"

      val1 := _num1P_.ExpStr[0]
      val2 := _num2P_.ExpStr[0]

      var final_ string

      if val2 == "true" {
        final_ = add(val1, "1", calc_params, line)
      } else {
        final_ = val1
      }

      final_ = returnInit(final_)

      finalRet = Action{ "number", "", []string{ final_ }, []Action{}, []string{}, []Action{}, []Condition{}, 39, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }
    }
    case TypeOperations{ "boolean", "number" }: { //detect case "num" + "boolean"

      val1 := _num1P_.ExpStr[0]
      val2 := _num2P_.ExpStr[0]

      var final_ string

      if val1 == "true" {
        final_ = add(val2, "1", calc_params, line)
      } else {
        final_ = val2
      }

      final_ = returnInit(final_)

      finalRet = Action{ "number", "", []string{ final_ }, []Action{}, []string{}, []Action{}, []Condition{}, 39, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }
    }
    default:

      if (nums.First == "string" && nums.Second != "array" && nums.Second != "none" && nums.Second != "hash") || (nums.First != "array" && nums.First != "none" && nums.First != "hash" && nums.Second == "string") { //detect case "string" + (* - "array" - "none" - "hash") = "string"
        val1 := _num1P_.ExpStr[0]
        val2 := _num2P_.ExpStr[0]

        final := val1 + val2

        finalRet = Action{ "string", "", []string{ final }, []Action{}, []string{}, []Action{}, []Condition{}, 38, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }
      } else if (nums.First == "array" && nums.Second != "none") || (nums.First != "none" && nums.Second == "array") { //detect case "array" + (* - "none") = "array"

        val1 := _num1P_
        val2 := _num2P_

        if nums.First == "array" {
          val1.Hash_Values[strconv.Itoa(len(val1.Hash_Values))] = []Action{ val2 }
        } else {
          val2, val1 = val1, val2

          nMap := make(map[string][]Action)

          for k, v := range val1.Hash_Values {
            nMap[add(k, "1", calc_params, line)] = v
          }

          nMap["0"] = []Action{ val2 }

          val1.Hash_Values = nMap
        }

        finalRet = val1

      } else if nums.First == "none" || nums.Second == "none" { //detect case "none" + * = "falsey"

        //if it is none + none, just return undefined
        finalRet = Action{ "falsey", "", []string{ "undefined" }, []Action{}, []string{}, []Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }
      } else if (nums.First == "hash" && nums.Second != "none") || (nums.First != "none" && nums.Second == "hash") { //detect case "hash" + (* - "hash") = "none"

        //get hash values of both values
        val1 := _num1P_.Hash_Values
        val2 := _num2P_.Hash_Values

        var final = make(map[string][]Action)

        //combine both hashes
        for k, v := range val1 {
          final[k] = v
        }

        for k, v := range val2 {
          final[k] = v
        }

        //return the combined hash
        finalRet = Action{ "hash", "", []string{}, []Action{}, []string{}, []Action{}, []Condition{}, 22, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, final, false }
      } else if (nums.First == "type" && nums.Second != "hash" && nums.Second != "none") || (nums.First != "hash" && nums.First != "none" && nums.Second == "type") { //detect case "type" + (* - "hash" - "none") = "type"

        val1 := _num1P_.ExpStr[0]
        val2 := _num2P_.ExpStr[0]

        final := val1 + val2

        finalRet = Action{ "string", "", []string{ final }, []Action{}, []string{}, []Action{}, []Condition{}, 38, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }

      } else {

        //if nothing was detected just return undefined
        finalRet = Action{ "falsey", "", []string{ "undefined" }, []Action{}, []string{}, []Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }
      }
  }

  reCalc(&finalRet)

  jsonNum, _ := json.Marshal(finalRet)

  return C.CString(string(jsonNum))
}
