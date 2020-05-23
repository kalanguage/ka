package lang

import "strings"
import "strconv"
import "reflect"
import "fmt"
import "os"
import "encoding/gob"

// #cgo CFLAGS: -std=c99
// #include "bind.h"
import "C"

//export Condition
type Condition struct {
  Type            string
  Condition     []Action
  Actions       []Action
}

type SubCaller struct {
  Indexes     [][]Action
  Args        [][]Action
  IsProc          bool
}

//export Action
type Action struct {
  Type            string
  Name            string
  ExpStr        []string
  ExpAct        []Action
  Params        []string
  Args        [][]Action
  Condition     []Condition
  ID              int

  //stuff for operations

  First         []Action
  Second        []Action
  Degree        []Action

  //stuff for indexes

  Value       [][]Action
  Indexes     [][]Action
  Hash_Values     map[string][]Action

  IsMutable       bool
  Access          string
  SubCall       []SubCaller
}

var operations = []string{ "+", "-", "*", "/", "^", "%", "&", "|", "=", "!=", ">", "<", ">=", "<=", ")", "(", "~~", "~~~", ":" }

func convToAct(_val []interface{}, dir, name string) []Action {
  var val []Action

  if reflect.TypeOf(_val[0]).String() == "lang.Lex" {

    var num []Lex

    for _, v := range _val {
      num = append(num, v.(Lex))
    }

    val = Actionizer(num, true, dir, name)

  } else {

    for _, v := range _val {
      val = append(val, v.(Action))
    }

  }

  return val
}

func getLeft(index int, exp []interface{}, dir, name string) ([]Action, []interface{}) {

  var _num1 []interface{}

  //_num1 loop
  for o := index - 1; o >= 0; o-- {

    _num1 = append(_num1, exp[o])
  }

  reverseInterface(_num1)

  num1 := convToAct(_num1, dir, name)

  return num1, _num1
}

func getRight(index int, exp []interface{}, dir, name string) ([]Action, []interface{}) {
  var _num2 []interface{}

  //_num2 loop
  for o := index + 1; o < len(exp); o++ {

    _num2 = append(_num2, exp[o])
  }

  num2 := convToAct(_num2, dir, name)

  return num2, _num2
}

func calcExp(index int, exp []interface{}, dir, name string) ([]Action, []Action, []interface{}, []interface{}) {

  num1, _num1 := getLeft(index, exp, dir, name)
  num2, _num2 := getRight(index, exp, dir, name)

  return num1, num2, _num1, _num2
}

func callCalcParams(i *int, lex []Lex, len_lex int, dir, filename string) ([][]Action, [][]Action, []SubCaller, bool) {

  cbCnt := 0
  glCnt := 0
  bCnt := 0
  pCnt := 1

  indexes := [][]Lex{[]Lex{}}
  var putIndexes [][]Action

  if lex[*i].Name == "." {

    cbCnt = 0
    glCnt = 0
    bCnt = 0
    pCnt = 0

    for o := (*i) + 1; o < len_lex; o++ {
      if lex[o].Name == "{" {
        cbCnt++
      }
      if lex[o].Name == "[:" {
        glCnt++
      }
      if lex[o].Name == "[" {
        bCnt++
      }
      if lex[o].Name == "(" {
        pCnt++
      }

      if lex[o].Name == "}" {
        cbCnt--
      }
      if lex[o].Name == ":]" {
        glCnt--
      }
      if lex[o].Name == "]" {
        bCnt--
      }
      if lex[o].Name == ")" {
        pCnt--
      }

      if lex[o].Name == "." {
        indexes = append(indexes, []Lex{})
      } else {

        (*i)++

        indexes[len(indexes) - 1] = append(indexes[len(indexes) - 1], lex[o])

        if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {

          if o < len_lex - 1 && lex[o + 1].Name == "." {
            continue
          } else {
            break
          }

        }
      }
    }

    for _, v := range indexes {
      putIndexes = append(putIndexes, Actionizer(v[1:len(v) - 1], true, dir, filename))
    }

    (*i)++
  }

  var params_ [][]Action
  var subcaller []SubCaller

  var isProc = false

  if *i < len_lex && lex[*i].Name == "(" {

    params := [][]Lex{[]Lex{}}

    cbCnt = 0
    glCnt = 0
    bCnt = 0
    pCnt = 0

    for o := *i; o < len_lex; o++ {
      if lex[o].Name == "{" {
        cbCnt++;
      }
      if lex[o].Name == "}" {
        cbCnt--;
      }

      if lex[o].Name == "[:" {
        glCnt++;
      }
      if lex[o].Name == ":]" {
        glCnt--;
      }

      if lex[o].Name == "[" {
        bCnt++;
      }
      if lex[o].Name == "]" {
        bCnt--;
      }

      if lex[o].Name == "(" {
        pCnt++;
      }
      if lex[o].Name == ")" {
        pCnt--;
      }

      if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
        break
      }

      if o == *i {
        continue
      }

      if lex[o].Name == "," {
        params = append(params, []Lex{})
        continue
      }

      params[len(params) - 1] = append(params[len(params) - 1], lex[o])
    }

    for _, v := range params {

      if len(v) == 0 {
        continue
      }

      params_ = append(params_, Actionizer(v, true, dir, filename))
    }

    pCnt_ := 0
    skip_nums := 0

    for o := *i; o < len_lex; o++ {
      if lex[o].Name == "(" {
        pCnt_++
      }
      if lex[o].Name == ")" {
        pCnt_--
      }

      skip_nums++;

      if pCnt_ == 0 {
        break
      }
    }

    isProc = true

    (*i)+=skip_nums

    if *i < len_lex {

      if lex[*i].Name == "(" || lex[*i].Name == "." {

        paramsSub, indexesSub, subVal, isProcSub := callCalcParams(i, lex, len_lex, dir, filename)

        subcaller = append(subcaller, SubCaller{ indexesSub, paramsSub, isProcSub })
        subcaller = append(subcaller, subVal...)
      }
    }
  }

  return params_, putIndexes, subcaller, isProc
}

//function to actionize the callers (#~)
func callCalc(i *int, lex []Lex, len_lex int, dir, filename string) ([][]Action, [][]Action, []SubCaller, string) {

  var name = lex[(*i) + 2].Name

  (*i)+=3

  params_, putIndexes, subcaller, _ := callCalcParams(i, lex, len_lex, dir, filename)

  return params_, putIndexes, subcaller, name
}

func procCalc(i *int, lex []Lex, len_lex int, dir, name string) ([]Action, []string, string) {

  var params []string
  var procName string
  var logic []Action

  if lex[(*i) + 1].Name == "~" {
    procName = lex[*i + 2].Name

    for o := (*i) + 4; o < len_lex; o++ {
      if lex[o].Name == ")" {
        break
      }

      if lex[o].Name == "," {
        (*i)++
        continue
      }

      params = append(params, lex[o].Name)
    }
    *i+=(len(params) + 5)

    var logic_ = []Lex{}

    cbCnt := 0

    for o := *i; o < len_lex; o++ {
      if lex[o].Name == "{" {
        cbCnt++
      }

      if lex[o].Name == "}" {
        cbCnt--
      }

      logic_ = append(logic_, lex[o])

      if cbCnt == 0 {
        break
      }
    }

    (*i)+=len(logic_) - 1

    logic = Actionizer(logic_, false, dir, name)
  } else {
    params = []string{}
    procName = ""

    for o := (*i) + 2; o < len_lex; o+=2 {
      if lex[o].Name == ")" {
        break
      }

      params = append(params, lex[o].Name)
    }
    *i+=(3 + len(params))

    var logic_ = []Lex{}

    cbCnt := 0

    for o := *i; o < len_lex; o++ {
      if lex[o].Name == "{" {
        cbCnt++
      }

      if lex[o].Name == "}" {
        cbCnt--
      }

      logic_ = append(logic_, lex[o])

      if cbCnt == 0 {
        break
      }
    }

    (*i)+=len(logic_) - 1

    logic = Actionizer(logic_, false, dir, name)
  }

  return logic, params, procName
}

//export Actionizer
func Actionizer(lex []Lex, doExpress bool, dir, name string) []Action {
  var actions = []Action{}
  var len_lex = len(lex)

  for i := 0; i < len_lex; i++ {

    if doExpress {
      var exp []interface{}

      cbCnt := 0
      glCnt := 0
      bCnt := 0
      pCnt := 0

      for o := i; o < len_lex; o++ {
        if lex[o].Name == "{" {
          cbCnt++;
        }
        if lex[o].Name == "}" {
          cbCnt--;
        }

        if lex[o].Name == "[:" {
          glCnt++;
        }
        if lex[o].Name == ":]" {
          glCnt--;
        }

        if lex[o].Name == "[" {
          bCnt++;
        }
        if lex[o].Name == "]" {
          bCnt--;
        }

        if lex[o].Name == "(" {
          pCnt++;
        }
        if lex[o].Name == ")" {
          pCnt--;
        }

        if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
          break
        }

        exp = append(exp, lex[o])

        i++
      }

      for ;interfaceContainOperations(exp, "|") || interfaceContainOperations(exp, "&") || interfaceContainOperations(exp, "!|") || interfaceContainOperations(exp, "!&") || interfaceContainOperations(exp, "$|") || interfaceContainOperations(exp, "!$|"); {
        indexes := map[string]int{
          "|": interfaceIndexOfOperations("|", exp),
          "&": interfaceIndexOfOperations("&", exp),
          "!|": interfaceIndexOfOperations("!|", exp),
          "!&": interfaceIndexOfOperations("!&", exp),
          "$|": interfaceIndexOfOperations("$|", exp),
          "!$|": interfaceIndexOfOperations("!$|", exp),
        }

        //get max index
        var min = [2]interface{}{}

        for k, v := range indexes {
          if v != -1 {
            min = [2]interface{}{ k, v }
          }
        }

        for k, v := range indexes {
          if (v != -1 && v > min[1].(int)) || min[1].(int) == -1 {
            min = [2]interface{}{ k, v }
          }
        }

        switch min[0].(string) {
          case "|":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "or", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 71, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "&":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "and", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 72, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "!|":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "nor", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 73, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "!&":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "nand", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 74, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "$|":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "xor", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 75, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "!$|":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "xnor", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 76, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
        }
      }

      for ;interfaceContainOperations(exp, "=") || interfaceContainOperations(exp, "!=") || interfaceContainOperations(exp, ">") || interfaceContainOperations(exp, "<") || interfaceContainOperations(exp, ">=") || interfaceContainOperations(exp, "<=") || interfaceContainOperations(exp, "~~") || interfaceContainOperations(exp, "~~~"); {
        indexes := map[string]int{
          "=": interfaceIndexOfOperations("=", exp),
          "!=": interfaceIndexOfOperations("!=", exp),
          ">": interfaceIndexOfOperations(">", exp),
          "<": interfaceIndexOfOperations("<", exp),
          ">=": interfaceIndexOfOperations(">=", exp),
          "<=": interfaceIndexOfOperations("<=", exp),
          "~~": interfaceIndexOfOperations("~~", exp),
          "~~~": interfaceIndexOfOperations("~~~", exp),
        }

        //get max index
        var min = [2]interface{}{}

        for k, v := range indexes {
          if v != -1 {
            min = [2]interface{}{ k, v }
          }
        }

        for k, v := range indexes {
          if (v != -1 && v > min[1].(int)) || min[1].(int) == -1 {
            min = [2]interface{}{ k, v }
          }
        }

        switch min[0].(string) {
          case "=":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "equals", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 47, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "!=":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "notEqual", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 48, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case ">":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "greater", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 49, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "<":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "less", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 50, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case ">=":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "greaterOrEqual", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 51, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "<=":
            index := min[1].(int)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "lessOrEqual", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 52, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "~~":
            index := min[1].(int)

            var degree_ []interface{}
            doDeg := false

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            for o := index + 1; o < len(exp); o++ {
              if exp[o] == "{" {
                cbCnt++;
              }
              if exp[o] == "}" {
                cbCnt--;
              }

              if exp[o] == "[:" {
                glCnt++;
              }
              if exp[o] == ":]" {
                glCnt--;
              }

              if exp[o] == "[" {
                bCnt++;
              }
              if exp[o] == "]" {
                bCnt--;
              }

              if exp[o] == "(" {
                pCnt++;
              }
              if exp[o] == ")" {
                pCnt--;
              }

              if reflect.TypeOf(exp[o]).String() == "lang.Lex" && exp[o].(Lex).Name == ":" && cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
                doDeg = true
                break
              }

              degree_ = append(degree_, exp[o])
            }

            var degree = []Action{}
            var addDeg = 0

            if doDeg {
              degree = convToAct(degree_, dir, name)
              addDeg = len(degree_) + 1
            }

            num1, _num1 := getLeft(index, exp, dir, name)
            num2, _num2 := getRight(index + addDeg, exp, dir, name)

            var act_exp = Action{ "similar", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 54, num1, num2, degree, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + addDeg + 1:]...)

            exp = exp_
          case "~~~":
            index := min[1].(int)

            var degree_ []interface{}
            doDeg := false

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            for o := index + 1; o < len(exp); o++ {
              if exp[o] == "{" {
                cbCnt++;
              }
              if exp[o] == "}" {
                cbCnt--;
              }

              if exp[o] == "[:" {
                glCnt++;
              }
              if exp[o] == ":]" {
                glCnt--;
              }

              if exp[o] == "[" {
                bCnt++;
              }
              if exp[o] == "]" {
                bCnt--;
              }

              if exp[o] == "(" {
                pCnt++;
              }
              if exp[o] == ")" {
                pCnt--;
              }

              if reflect.TypeOf(exp[o]).String() == "lang.Lex" && exp[o].(Lex).Name == ":" && cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
                doDeg = true
                break
              }

              degree_ = append(degree_, exp[o])
            }

            var degree = []Action{}
            var addDeg = 0

            if doDeg {
              degree = convToAct(degree_, dir, name)
              addDeg = len(degree_) + 1
            }

            num1, _num1 := getLeft(index, exp, dir, name)
            num2, _num2 := getRight(index + addDeg, exp, dir, name)

            var act_exp = Action{ "strictSimilar", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 55, num1, num2, degree, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + addDeg + 1:]...)

            exp = exp_
        }
      }

      for ;interfaceContainOperations(exp, "+") || interfaceContainOperations(exp, "-"); {

        if interfaceIndexOfOperations("+", exp) > interfaceIndexOfOperations("-", exp) || interfaceIndexOfOperations("-", exp) == -1 {
          index := interfaceIndexOfOperations("+", exp)

          num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

          var act_exp = Action{ "add", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 32, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

          exp_ := append(exp[:index - len(_num1)], act_exp)
          exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

          exp = exp_
        } else {
          index := interfaceIndexOfOperations("-", exp)

          num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

          var act_exp = Action{ "subtract", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 33, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

          exp_ := append(exp[:index - len(_num1)], act_exp)
          exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

          exp = exp_
        }

      }

      for ;interfaceContainOperations(exp, "*") || interfaceContainOperations(exp, "/") || interfaceContainOperations(exp, "%"); {

        indexes := map[string]int{
          "*": interfaceIndexOfOperations("*", exp),
          "/": interfaceIndexOfOperations("/", exp),
          "%": interfaceIndexOfOperations("%", exp),
        }

        //get max index
        var min = [2]interface{}{}

        for k, v := range indexes {
          if v != -1 {
            min = [2]interface{}{ k, v }
          }
        }

        for k, v := range indexes {
          if (v != -1 && v > min[1].(int)) || min[1].(int) == -1 {
            min = [2]interface{}{ k, v }
          }
        }

        switch min[0].(string) {
          case "*":
            index := interfaceIndexOfOperations("*", exp)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "multiply", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 34, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "/":
            index := interfaceIndexOfOperations("/", exp)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "divide", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 35, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
          case "%":
            index := interfaceIndexOfOperations("%", exp)

            num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

            var act_exp = Action{ "modulo", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 37, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

            exp_ := append(exp[:index - len(_num1)], act_exp)
            exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

            exp = exp_
        }

      }

      for ;interfaceContainOperations(exp, "^"); {
        index := interfaceIndexOfOperations("^", exp)

        num1, num2, _num1, _num2 := calcExp(index, exp, dir, name)

        var act_exp = Action{ "exponentiate", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 36, num1, num2, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

        exp_ := append(exp[:index - len(_num1)], act_exp)
        exp_ = append(exp_, exp[index + len(_num2) + 1:]...)

        exp = exp_
      }

      for ;interfaceContainOperations(exp, "!"); {

        index := interfaceIndexOfOperations("!", exp)

        var num []interface{}

        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := index + 1; o < len(exp); o++ {

          if exp[o].(Lex).Name == "{" {
            cbCnt++
          }
          if exp[o].(Lex).Name == "}" {
            cbCnt--
          }

          if exp[o].(Lex).Name == "[" {
            bCnt++
          }
          if exp[o].(Lex).Name == "]" {
            bCnt--
          }

          if exp[o].(Lex).Name == "[:" {
            glCnt++
          }
          if exp[o].(Lex).Name == ":]" {
            glCnt--
          }

          if exp[o].(Lex).Name == "(" {
            pCnt++
          }
          if exp[o].(Lex).Name == ")" {
            pCnt--
          }

          if arrayContainInterface(operations, exp[o]) {
            break
          }

          num = append(num, exp[o])
        }

        numAct := convToAct(num, dir, name)

        var act_exp = Action{ "not", "operation", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 53, []Action{}, numAct, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

        exp_ := append(exp[:index], act_exp)
        exp_ = append(exp_, exp[index + len(num) + 1:]...)

        exp = exp_
      }

      var proc_indexes []int

      for ;interfaceContainWithProcIndex(exp, "(", proc_indexes); {

        index := interfaceIndexOfWithProcIndex("(", exp, proc_indexes)

        if index - 1 != -1 && (reflect.TypeOf(exp[index - 1]).String() != "lang.Lex" || ((strings.HasPrefix(exp[index - 1].(Lex).Name, "$") || exp[index - 1].(Lex).Name == "]")))  {
          proc_indexes = append(proc_indexes, index)
          continue
        }

        var pExp []Lex

        pCnt := 0

        for o := index; o < len(exp); o++ {
          if exp[o].(Lex).Name == "(" {
            pCnt++;
          }
          if exp[o].(Lex).Name == ")" {
            pCnt--;
          }

          pExp = append(pExp, exp[o].(Lex))

          if pCnt == 0 {
            break
          }
        }

        pExp = pExp[1:len(pExp) - 1]

        pExpAct := Actionizer(pExp, true, dir, name)

        scbCnt := 0
        sglCnt := 0
        sbCnt := 0
        spCnt := 0

        indexes := [][]Lex{}

        if !(index + len(pExp) + 2 >= len(exp)) {
          if exp[index + len(pExp) + 2].(Lex).Name == "." {
            for o := index + len(pExp) + 2; o < len_lex; o++ {
              if exp[o].(Lex).Name == "{" {
                scbCnt++
              }
              if exp[o].(Lex).Name == "}" {
                scbCnt--
              }

              if exp[o].(Lex).Name == "[" {
                sbCnt++
              }
              if exp[o].(Lex).Name == "]" {
                sbCnt--
              }

              if exp[o].(Lex).Name == "[:" {
                sglCnt++
              }
              if exp[o].(Lex).Name == ":]" {
                sglCnt--
              }

              if exp[o].(Lex).Name == "(" {
                spCnt++
              }
              if exp[o].(Lex).Name == ")" {
                spCnt--
              }

              if exp[o].(Lex).Name == "." {
                indexes = append(indexes, []Lex{})
              } else {

                i++

                indexes[len(indexes) - 1] = append(indexes[len(indexes) - 1], exp[o].(Lex))

                if scbCnt == 0 && sglCnt == 0 && sbCnt == 0 && spCnt == 0 {

                  if o < len(exp) - 1 && exp[o + 1].(Lex).Name == "." {
                    continue
                  } else {
                    break
                  }

                }
              }
            }

            var putIndexes [][]Action

            for _, v := range indexes {

              v = v[1:len(v) - 1]
              putIndexes = append(putIndexes, Actionizer(v, true, dir, name))
            }

            pExpAct[0].Type = "expressionIndex"
            pExpAct[0].Indexes = putIndexes
            pExpAct[0].ID = 8 //set the action ID to the epxressionIndex ID
          }
        }

        exp = append([]interface{}{ pExpAct[0] }, exp...)
      }

      if len(exp) == 0 {
        break
      }

      if reflect.TypeOf(exp[0]).String() == "lang.Lex" {

        //variale that grets convved to a []Lex
        var toa []Lex

        for _, v := range exp {
          toa = append(toa, v.(Lex))
        }

        exp[0] = Actionizer(toa, false, dir, name)[0]
      }

      actions = append(actions, exp[0].(Action))
    }

    if i >= len_lex {
      break
    }

    switch lex[i].Name {
      case "newlineN":
        actions = append(actions, Action{ "newline", "", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 0, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
      case "local":
        exp_ := []Lex{}

        //getting nb semicolons
        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i + 4; o < len_lex; o++ {

          if lex[o].Name == "{" {
            cbCnt++;
          }
          if lex[o].Name == "}" {
            cbCnt--;
          }

          if lex[o].Name == "[:" {
            glCnt++;
          }
          if lex[o].Name == ":]" {
            glCnt--;
          }

          if lex[o].Name == "[" {
            bCnt++;
          }
          if lex[o].Name == "]" {
            bCnt--;
          }

          if lex[o].Name == "(" {
            pCnt++;
          }
          if lex[o].Name == ")" {
            pCnt--;
          }

          if cbCnt != 0 || glCnt != 0 || bCnt != 0 || pCnt != 0 {
            exp_ = append(exp_, lex[o])
            continue
          }

          if lex[o].Name == "newlineS" {
            break
          }

          exp_ = append(exp_, lex[o])
        }

        exp := Actionizer(exp_, true, dir, name)

        actions = append(actions, Action{ "local", lex[i + 2].Name, []string{}, exp, []string{}, [][]Action{}, []Condition{}, 1, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=(4 + len(exp_))
      case "dynamic":
        exp_ := []Lex{}

        //getting nb semicolons
        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i + 4; o < len_lex; o++ {

          if lex[o].Name == "{" {
            cbCnt++;
          }
          if lex[o].Name == "}" {
            cbCnt--;
          }

          if lex[o].Name == "[:" {
            glCnt++;
          }
          if lex[o].Name == ":]" {
            glCnt--;
          }

          if lex[o].Name == "[" {
            bCnt++;
          }
          if lex[o].Name == "]" {
            bCnt--;
          }

          if lex[o].Name == "(" {
            pCnt++;
          }
          if lex[o].Name == ")" {
            pCnt--;
          }

          if cbCnt != 0 || glCnt != 0 || bCnt != 0 || pCnt != 0 {
            exp_ = append(exp_, lex[o])
            continue
          }

          if lex[o].Name == "newlineS" {
            break
          }

          exp_ = append(exp_, lex[o])
        }

        exp := Actionizer(exp_, true, dir, name)

        actions = append(actions, Action{ "dynamic", lex[i + 2].Name, []string{}, exp, []string{}, [][]Action{}, []Condition{}, 2, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=(4 + len(exp_))
      case "alt":

        var alter = Action{ "alt", "", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 3, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} }

        pCnt := 0

        cond_ := []Lex{}

        for o := i + 1; o < len_lex; o++ {
          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          cond_ = append(cond_, lex[o])

          if pCnt == 0 {
            break
          }
        }

        i+=len(cond_) + 1

        cond := Actionizer(cond_, true, dir, name)

        for do := true; do || lex[i].Name == "=>"; do = false {

          adder := 0

          if lex[i].Name != "=>" {
            adder = 1
          }

          cbCnt := 0

          actions_ := []Lex{}

          for o := i + adder; o < len_lex; o++ {
            if lex[o].Name == "{" {
              cbCnt++
            }
            if lex[o].Name == "}" {
              cbCnt--
            }

            actions_ = append(actions_, lex[o])

            if cbCnt == 0 {
              break
            }
          }

          i+=len(actions_)
          actions := Actionizer(actions_, true, dir, name)

          alter.Condition = append(alter.Condition, Condition{ "alt", cond, actions })
          i++

          if i >= len_lex {
            break
          }
        }

        actions = append(actions, alter)

      case "global":
        exp_ := []Lex{}

        //getting nb semicolons
        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i + 4; o < len_lex; o++ {

          if lex[o].Name == "{" {
            cbCnt++;
          }
          if lex[o].Name == "}" {
            cbCnt--;
          }

          if lex[o].Name == "[:" {
            glCnt++;
          }
          if lex[o].Name == ":]" {
            glCnt--;
          }

          if lex[o].Name == "[" {
            bCnt++;
          }
          if lex[o].Name == "]" {
            bCnt--;
          }

          if lex[o].Name == "(" {
            pCnt++;
          }
          if lex[o].Name == ")" {
            pCnt--;
          }

          if cbCnt != 0 || glCnt != 0 || bCnt != 0 || pCnt != 0 {
            exp_ = append(exp_, lex[o])
            continue
          }

          if lex[o].Name == "newlineS" {
            break
          }

          exp_ = append(exp_, lex[o])
        }

        exp := Actionizer(exp_, true, dir, name)

        actions = append(actions, Action{ "global", lex[i + 2].Name, []string{}, exp, []string{}, [][]Action{}, []Condition{}, 4, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=(4 + len(exp_))
      case "log":
        exp_ := []Lex{}

        //getting nb semicolons
        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i + 2; o < len_lex; o++ {

          if lex[o].Name == "{" {
            cbCnt++;
          }
          if lex[o].Name == "}" {
            cbCnt--;
          }

          if lex[o].Name == "[:" {
            glCnt++;
          }
          if lex[o].Name == ":]" {
            glCnt--;
          }

          if lex[o].Name == "[" {
            bCnt++;
          }
          if lex[o].Name == "]" {
            bCnt--;
          }

          if lex[o].Name == "(" {
            pCnt++;
          }
          if lex[o].Name == ")" {
            pCnt--;
          }

          if cbCnt != 0 || glCnt != 0 || bCnt != 0 || pCnt != 0 {
            exp_ = append(exp_, lex[o])
            continue
          }

          if lex[o].Name == "newlineS" {
            break
          }

          exp_ = append(exp_, lex[o])
        }

        exp := Actionizer(exp_, true, dir, name)

        actions = append(actions, Action{ "log", "", []string{}, exp, []string{}, [][]Action{}, []Condition{}, 5, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=(2 + len(exp_))
      case "print":
        exp_ := []Lex{}

        //getting nb semicolons
        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i + 2; o < len_lex; o++ {

          if lex[o].Name == "{" {
            cbCnt++;
          }
          if lex[o].Name == "}" {
            cbCnt--;
          }

          if lex[o].Name == "[:" {
            glCnt++;
          }
          if lex[o].Name == ":]" {
            glCnt--;
          }

          if lex[o].Name == "[" {
            bCnt++;
          }
          if lex[o].Name == "]" {
            bCnt--;
          }

          if lex[o].Name == "(" {
            pCnt++;
          }
          if lex[o].Name == ")" {
            pCnt--;
          }

          if cbCnt != 0 || glCnt != 0 || bCnt != 0 || pCnt != 0 {
            exp_ = append(exp_, lex[o])
            continue
          }

          if lex[o].Name == "newlineS" {
            break
          }

          exp_ = append(exp_, lex[o])
        }

        exp := Actionizer(exp_, false, dir, name)

        actions = append(actions, Action{ "print", "", []string{}, exp, []string{}, [][]Action{}, []Condition{}, 6, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=(2 + len(exp_))
      case "{":
        exp_ := []Lex{}

        //getting nb semicolons
        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i; o < len_lex; o++ {

          if lex[o].Name == "{" {
            cbCnt++;
          }
          if lex[o].Name == "}" {
            cbCnt--;
          }

          if lex[o].Name == "[:" {
            glCnt++;
          }
          if lex[o].Name == ":]" {
            glCnt--;
          }

          if lex[o].Name == "[" {
            bCnt++;
          }
          if lex[o].Name == "]" {
            bCnt--;
          }

          if lex[o].Name == "(" {
            pCnt++;
          }
          if lex[o].Name == ")" {
            pCnt--;
          }

          if cbCnt != 0 || glCnt != 0 || bCnt != 0 || pCnt != 0 {
            exp_ = append(exp_, lex[o])
            continue
          }

          exp_ = append(exp_, lex[o])

          if cbCnt == 0 {
            break
          }

          if lex[o].Name == "newlineS" {
            break
          }
        }

        exp_ = exp_[1:]
        exp_ = exp_[:len(exp_) - 1]

        exp := Actionizer(exp_, false, dir, name)

        actions = append(actions, Action{ "group", "", []string{}, exp, []string{}, [][]Action{}, []Condition{}, 9, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=(len(exp_) + 1)
      case "process":

        putFalsey := make(map[string][]Action)
        putFalsey["falsey"] = []Action{ Action{ "falsey", "", []string{ "undef" }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} } }

        logic, params, procName := procCalc(&i, lex, len_lex, dir, name)

        actions = append(actions, Action{ "process", procName, []string{}, logic, params, [][]Action{}, []Condition{}, 10, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, putFalsey, false, "private", []SubCaller{} })
      case "pargc":

        i+=2
        count := lex[i].Name
        i++

        //if it is not a number of a parameter list
        //(add parameter lists before)
        //they are just as follows
        /*

        pargc ~ (string->, number->, ...whatever datatypes) {

        }

        */
        //used to properly overload processes
        if getType(count) != "number" && count != "(" {

          //throw an error
          C.colorprint(C.CString("Error while actionizing in " + lex[i].Dir + "!\n"), C.int(12))
          fmt.Println("Expected either a numeric value or a parameter list after pargc but instead got", count, "which is of type", getType(count), "\n\nError occured on line", lex[i].Line, "\nFound near:", strings.TrimSpace(lex[i].Exp))

          //exit the process
          os.Exit(1)
        }

        var types []string

        if count == "(" {

          for ;i < len_lex; i++ {

            if lex[i].Name == "," {
              continue
            }

            if lex[i].Name == ")" {
              i++
              break
            }

            types = append(types, lex[i].Name[1:])

            if !isType(lex[i].Name[1:]) {

              //throw an error
              C.colorprint(C.CString("Error while actionizing in " + lex[i].Dir + "!\n"), C.int(12))
              fmt.Println("Expected a type value instead of", lex[i].Name[1:], "\n\nError occured on line", lex[i].Line, "\nFound near:", strings.TrimSpace(lex[i].Exp))

              //exit the process
              os.Exit(1)
            }
          }
        }

        if lex[i].Name != "{" {

          //throw an error
          C.colorprint(C.CString("Error while actionizing in " + lex[i].Dir + "!\n"), C.int(12))
          fmt.Println("Expected { instead of", lex[i].Name, "\nFound near:", strings.TrimSpace(lex[i].Exp))

          //exit the process
          os.Exit(1)
        }

        cbCnt := 0

        var exp []Lex

        for ;i < len_lex; i++ {

          if lex[i].Name == "{" {
            cbCnt++
            continue
          }
          if lex[i].Name == "}" {
            cbCnt--
            continue
          }

          exp = append(exp, lex[i])

          if cbCnt == 0 {
            break
          }
        }
        i--

        actionized := Actionizer(exp, false, dir, name)

        if count == "(" {

          actions = append(actions, Action{ "pargc_paramlist", "", []string{}, actionized, types, [][]Action{}, []Condition{}, 81, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        } else {

          actions = append(actions, Action{ "pargc_number", "", []string{ count }, actionized, []string{}, [][]Action{}, []Condition{}, 80, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        }
      case "wait":

        var exp []Lex
        pCnt := 0

        for o := i + 1; o < len_lex; o++ {
          if lex[o].Name == "(" {
            pCnt++
            continue
          }
          if lex[o].Name == ")" {
            pCnt--
            continue
          }

          if pCnt == 0 {
            break
          }

          exp = append(exp, lex[o])
        }

        actionized := Actionizer(exp, true, dir, name)

        actions = append(actions, Action{ "wait", "", []string{}, actionized, []string{}, [][]Action{}, []Condition{}, 57, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i++
      case "#":

        params_, putIndexes, subcaller, name := callCalc(&i, lex, len_lex, dir, name)

        actions = append(actions, Action{ "#", name, []string{}, []Action{}, []string{}, params_, []Condition{}, 11, []Action{}, []Action{}, []Action{}, [][]Action{}, putIndexes, make(map[string][]Action), false, "private", subcaller })
      case "@":

        params_, putIndexes, subcaller, name := callCalc(&i, lex, len_lex, dir, name)

        actions = append(actions, Action{ "@", name, []string{}, []Action{}, []string{}, params_, []Condition{}, 56, []Action{}, []Action{}, []Action{}, [][]Action{}, putIndexes, make(map[string][]Action), false, "private", subcaller })
      case "return":

        returner_ := []Lex{}

        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i + 2; o < len_lex; o++ {
          if lex[o].Name == "{" {
            cbCnt++
          }
          if lex[o].Name == "}" {
            cbCnt--
          }

          if lex[o].Name == "[:" {
            glCnt++
          }
          if lex[o].Name == ":]" {
            glCnt--
          }

          if lex[o].Name == "[" {
            bCnt++
          }
          if lex[o].Name == "]" {
            bCnt--
          }

          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
            break
          }

          returner_ = append(returner_, lex[o])
        }

        returner := Actionizer(returner_, true, dir, name)

        actions = append(actions, Action{ "return", "", []string{}, returner, []string{}, [][]Action{}, []Condition{}, 12, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=len(returner_) + 2
      case "if":

        conditions := []Condition{}

        if lex[i].Name == "if" {
          var cond_ = []Lex{}
          pCnt := 0

          for o := i + 1; o < len_lex; o++ {
            if lex[o].Name == "(" {
              pCnt++
            }
            if lex[o].Name == ")" {
              pCnt--
            }

            cond_ = append(cond_, lex[o])

            if pCnt == 0 {
              break
            }
          }

          cond := Actionizer(cond_, true, dir, name)

          cbCnt := 0
          glCnt := 0
          bCnt := 0
          pCnt = 0

          actions_ := []Lex{}

          var curlyBraceCond = lex[i + 1 + len(cond_)].Name == "{"

          for o := i + 1 + len(cond_); o < len_lex; o++ {
            if lex[o].Name == "{" {
              cbCnt++
            }
            if lex[o].Name == "}" {
              cbCnt--
            }

            if lex[o].Name == "[:" {
              glCnt++
            }
            if lex[o].Name == ":]" {
              glCnt--
            }

            if lex[o].Name == "[" {
              bCnt++
            }
            if lex[o].Name == "]" {
              bCnt--
            }

            if lex[o].Name == "(" {
              pCnt++
            }
            if lex[o].Name == ")" {
              pCnt--
            }

            actions_ = append(actions_, lex[o])

            if !curlyBraceCond {

              if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
                break
              }
            } else if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
              break
            }
          }

          acts := Actionizer(actions_, false, dir, name)

          var if_ = Condition{ "if", cond, acts }

          conditions = append(conditions, if_)

          i+=(1 + len(cond_) + len(actions_))
        }

        if !(i >= len_lex) {
          for ;lex[i].Name == "elseif"; {

            var cond_ = []Lex{}
            pCnt := 0

            for o := i + 1; o < len_lex; o++ {
              if lex[o].Name == "(" {
                pCnt++
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              cond_ = append(cond_, lex[o])

              if pCnt == 0 {
                break
              }
            }

            cond := Actionizer(cond_, true, dir, name)

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt = 0

            actions_ := []Lex{}

            var curlyBraceCond = lex[i + 1 + len(cond_)].Name == "{"

            for o := i + 1 + len(cond_); o < len_lex; o++ {
              if lex[o].Name == "{" {
                cbCnt++
              }
              if lex[o].Name == "}" {
                cbCnt--
              }

              if lex[o].Name == "[:" {
                glCnt++
              }
              if lex[o].Name == ":]" {
                glCnt--
              }

              if lex[o].Name == "[" {
                bCnt++
              }
              if lex[o].Name == "]" {
                bCnt--
              }

              if lex[o].Name == "(" {
                pCnt++
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              actions_ = append(actions_, lex[o])

              if !curlyBraceCond {

                if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
                  break
                }
              } else if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
                break
              }
            }

            acts := Actionizer(actions_, false, dir, name)

            var elseif_ = Condition{ "elseif", cond, acts }

            conditions = append(conditions, elseif_)

            i+=(1 + len(cond_) + len(actions_))

          }
        }

        if !(i >= len_lex) {
          actions_ := []Lex{}

          for ;lex[i].Name == "else"; {
            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            //allow for user to write: else ~ <do something>;
            var curlyBraceCond = lex[i + 1].Name == "{"

            for o := i + 1; o < len_lex; o++ {
              if lex[o].Name == "{" {
                cbCnt++
              }
              if lex[o].Name == "}" {
                cbCnt--
              }

              if lex[o].Name == "[:" {
                glCnt++
              }
              if lex[o].Name == ":]" {
                glCnt--
              }

              if lex[o].Name == "[" {
                bCnt++
              }
              if lex[o].Name == "]" {
                bCnt--
              }

              if lex[o].Name == "(" {
                pCnt++
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              if !curlyBraceCond {

                if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
                  break
                }
              } else if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
                break
              }

              actions_ = append(actions_, lex[o])
            }

            if !curlyBraceCond {
              actions_ = actions_[1:]
            }

            actions := Actionizer(actions_, false, dir, name)

            var else_ = Condition{ "else", []Action{}, actions }

            conditions = append(conditions, else_)

            i+=(1 + len(actions_))

            if i >= len_lex {
              break
            }
          }
        }

        actions = append(actions, Action{ "conditional", "", []string{}, []Action{}, []string{}, [][]Action{}, conditions, 13, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i--
      case "import":

        var fileDir = lex[i + 2].Name

        //remove the quotes
        fileDir = fileDir[1:len(fileDir) - 1]

        var files = []map[string]string{}

        //see if user wants to import a file from the stdlib
        if strings.HasPrefix(fileDir, "?~") {

          if strings.HasPrefix(fileDir[2:], "/") {
            files = ReadFileJS("./stdlib" + fileDir[2:])
          } else {
            files = ReadFileJS("./stdlib/" + fileDir[2:])
          }

        } else {
          files = ReadFileJS(dir + fileDir)
        }

        var lexxed = []map[string]interface{}{}

        i+=2

        if i + 1 < len_lex && lex[i + 1].Name == "newlineS" {
          i++
        }

        for _, v := range files {

          if arrayContain(imported, v["FileName"]) {
            continue
          }

          imported = append(imported, v["FileName"])

          curlex := Lexer(v["Content"], dir, strings.TrimPrefix(v["FileName"], dir) /* remove the directory part of the filename */)

          curmap := map[string]interface{}{
            "FileName": v["FileName"],
            "Content": curlex,
          }

          lexxed = append(lexxed, curmap)
        }

        var actionizedFiles [][]Action

        for _, v := range lexxed {

          if strings.HasSuffix(v["FileName"].(string), ".oat") {

            readfile, _ := os.Open(v["FileName"].(string))

            var decoded []Action

            decoder := gob.NewDecoder(readfile)
            e := decoder.Decode(&decoded)

            if e != nil {
              C.colorprint(C.CString("Error while actionizing " + dir + name + ", "), C.int(12))
              fmt.Println(v["FileName"], "was detected as an oat, but is not oat compatible.")
              os.Exit(1)
            }

            readfile.Close()

            actionizedFiles = append(actionizedFiles, decoded)

          } else {

            actionizedFiles = append(actionizedFiles, Actionizer(v["Content"].([]Lex), false, dir, name))
          }
        }

        actions = append(actions, Action{ "import", "", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 14, []Action{}, []Action{}, []Action{}, actionizedFiles, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
      case "?^": //only used in namespaces
        continue
      case "ns":

        //define calc_ns
        var calc_ns func(lex []Lex, cur_index *int) []Lex

        //assign calc_ns
        calc_ns = func(lex []Lex, cur_index *int) []Lex {

          if *cur_index >= len_lex || !strings.HasPrefix(lex[*cur_index + 1].Name, "$") {

            //throw an error
            C.colorprint(C.CString("Error while actionizing in " + lex[i].Dir + "!\n"), C.int(12))
            fmt.Println("Expected a variable name after ns::", "\n\nError occured on line", lex[i].Line, "\nFound near:", strings.TrimSpace(lex[i].Exp))

            //exit the process
            os.Exit(1)
          }

          var namespace_name = lex[*cur_index + 1].Name
          *cur_index+=2

          cbCnt := 0
          glCnt := 0
          bCnt := 0
          pCnt := 0

          var namespace_group []Lex

          for o := *cur_index; o < len_lex; o++ {
            if lex[o].Name == "{" {
              cbCnt++
            }
            if lex[o].Name == "}" {
              cbCnt--
            }

            if lex[o].Name == "[:" {
              glCnt++
            }
            if lex[o].Name == ":]" {
              glCnt--
            }

            if lex[o].Name == "[" {
              bCnt++
            }
            if lex[o].Name == "]" {
              bCnt--
            }

            if lex[o].Name == "(" {
              pCnt++
            }
            if lex[o].Name == ")" {
              pCnt--
            }

            namespace_group = append(namespace_group, lex[o])
            *cur_index++

            if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
              break
            }
          }
          *cur_index--

          //see if ?^ is activated
          var not_ns = false

          //change all variables in that lex group to start with the namespace name
          for k := 0; k < len(namespace_group); k++ {

            v := namespace_group[k]

            if v.Name == "?^" {
              not_ns = true

              if namespace_group[k + 1].Name[0] != '$' {

                //throw an error
                C.colorprint(C.CString("Error while actionizing in " + lex[i].Dir + "!\n"), C.int(12))
                fmt.Println("Unexpected ?^ before", v, ". ?^ is only allowed before a variable!", "\n\nError occured on line", lex[i].Line, "\nFound near:", strings.TrimSpace(lex[i].Exp))

                //exit the process
                os.Exit(1)

              }

              continue
            }

            //detect a nested namespace
            if v.Name == "ns" {

              cbCnt = 0
              glCnt = 0
              bCnt = 0
              pCnt = 0

              ns_key := v

              name := namespace_group[k + 1]
              var nested_namespace_group []Lex

              for j := k + 2; j < len(namespace_group); j++ {

                if namespace_group[j].Name == "{" {
                  cbCnt++
                }
                if namespace_group[j].Name == "}" {
                  cbCnt--
                }

                if namespace_group[j].Name == "[:" {
                  glCnt++
                }
                if namespace_group[j].Name == ":]" {
                  glCnt--
                }

                if namespace_group[j].Name == "[" {
                  bCnt++
                }
                if namespace_group[j].Name == "]" {
                  bCnt--
                }

                if namespace_group[j].Name == "(" {
                  pCnt++
                }
                if namespace_group[j].Name == ")" {
                  pCnt--
                }

                nested_namespace_group = append(nested_namespace_group, namespace_group[j])

                if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
                  break
                }
              }

              var nested_new = []Lex{ ns_key, name }
              nested_new = append(nested_new, nested_namespace_group...)

              zero := 0

              namespace_group = append(namespace_group, calc_ns(nested_new, &zero)...)
            }

            if strings.HasPrefix(v.Name, "$") { //if it is a variable

              if not_ns {
                not_ns = false
                continue
              }

              namespace_group[k].Name = namespace_name + "." + v.Name[1:] //make it starts with $<namespace name>
              continue
            }
          }

          return namespace_group

        }

        actionized := Actionizer(calc_ns(lex, &i), false, dir, name)

        actions = append(actions, actionized...)
      case "break":
        actions = append(actions, Action{ "break", "", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 16, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
      case "skip":
        actions = append(actions, Action{ "skip", "", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 17, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
      case "loop":

        var condition_ = []Lex{}

        pCnt := 0

        for o := i + 1; o < len_lex; o++ {
          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          condition_ = append(condition_, lex[o])

          if pCnt == 0 {
            break
          }
        }

        condition := Actionizer(condition_, true, dir, name)
        action_ := []Lex{}

        var curlyBraceCond = lex[i + 1 + len(condition_)].Name == "{"

        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt = 0

        for o := i + 1 + len(condition_); o < len_lex; o++ {
          if lex[o].Name == "{" {
            cbCnt++
          }
          if lex[o].Name == "}" {
            cbCnt--
          }

          if lex[o].Name == "[:" {
            glCnt++
          }
          if lex[o].Name == ":]" {
            glCnt--
          }

          if lex[o].Name == "[" {
            bCnt++
          }
          if lex[o].Name == "]" {
            bCnt--
          }

          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          action_ = append(action_, lex[o])

          if !curlyBraceCond {

            if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
              break
            }
          } else if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
            break
          }
        }

        action := Actionizer(action_, false, dir, name)

        actions = append(actions, Action{ "loop", "", []string{}, action, []string{}, [][]Action{}, []Condition{ { "loop", condition, action } }, 21, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        i+=len(condition_) + len(action_)
      case "[:":
        var phrase = []Lex{}

        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i; o < len_lex; o++ {
          if lex[o].Name == "{" {
            cbCnt++
          }
          if lex[o].Name == "}" {
            cbCnt--
          }

          if lex[o].Name == "[:" {
            glCnt++
          }
          if lex[o].Name == ":]" {
            glCnt--
          }

          if lex[o].Name == "[" {
            bCnt++
          }
          if lex[o].Name == "]" {
            bCnt--
          }

          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          phrase = append(phrase, lex[o])

          if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
            break
          }
        }

        i+=len(phrase)

        phrase = phrase[1:len(phrase) - 1]

        var _translated =  [][][]Lex{ [][]Lex{ []Lex{}, []Lex{} } }

        cbCnt = 0
        glCnt = 0
        bCnt = 0
        pCnt = 0

        cur := 0

        for _, v := range phrase {

          if v.Name == "{" {
            cbCnt++
          }
          if v.Name == "}" {
            cbCnt--
          }

          if v.Name == "[:" {
            glCnt++
          }
          if v.Name == ":]" {
            glCnt--
          }

          if v.Name == "[" {
            bCnt++
          }
          if v.Name == "]" {
            bCnt--
          }

          if v.Name == "(" {
            pCnt++
          }
          if v.Name == ")" {
            pCnt--
          }

          if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && v.Name == ":" {
            cur = 1
            continue
          }
          if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && v.Name == "," {
            cur = 0
            _translated = append(_translated, [][]Lex{ []Lex{}, []Lex{} })
            continue
          }

          _translated[len(_translated) - 1][cur] = append(_translated[len(_translated) - 1][cur], v)
        }

        var translated = make(map[string][]Action)

        translated["falsey"] = []Action{ Action{ "falsey", "exp_value", []string{ "undef" }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} } }

        for _, v := range _translated {

          if len(v[0]) <= 0 {
            break
          }

          var name_ = v[0][0].Name

          if strings.HasPrefix(v[0][0].Name, "'") {
            name_ = name_[1:len(name_) - 1]
          }

          if strings.HasPrefix(v[0][0].Name, "$") {
            name_ = name_[1:]
          }

          translated[name_] = Actionizer(v[1], true, dir, name)
        }

        if i >= len_lex {
          actions = append(actions, Action{ "hash", "hashed_value", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 22, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, translated, false, "private", []SubCaller{} })
          break
        }

        isMutable := false

        //checks for a runtime hash
        for ;i < len_lex && lex[i].Name == ":::"; i++ {
          isMutable = !isMutable
        }

        if i >= len_lex {
          actions = append(actions, Action{ "hash", "hashed_value", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 22, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, translated, isMutable, "private", []SubCaller{} })
          break
        }

        if lex[i].Name == "." {
          indexes := [][]Lex{ []Lex{} }

          cbCnt = 0
          glCnt = 0
          bCnt = 0
          pCnt = 0

          for o := i + 1; o < len_lex; o++ {
            if lex[o].Name == "{" {
              cbCnt++
            }
            if lex[o].Name == "[:" {
              glCnt++
            }
            if lex[o].Name == "[" {
              bCnt++
            }
            if lex[o].Name == "(" {
              pCnt++
            }

            if lex[o].Name == "}" {
              cbCnt--
            }
            if lex[o].Name == ":]" {
              glCnt--
            }
            if lex[o].Name == "]" {
              bCnt--
            }
            if lex[o].Name == ")" {
              pCnt--
            }

            if lex[o].Name == "." {
              indexes = append(indexes, []Lex{})
            } else {

              i++

              indexes[len(indexes) - 1] = append(indexes[len(indexes) - 1], lex[o])

              if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {

                if o < len_lex - 1 && lex[o + 1].Name == "." {
                  continue
                } else {
                  break
                }

              }
            }
          }

          var putIndexes [][]Action

          for _, v := range indexes {

            v = v[1:len(v) - 1]
            putIndexes = append(putIndexes, Actionizer(v, true, dir, name))
          }

          i+=3

          actions = append(actions, Action{ "hashIndex", "", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 23, []Action{}, []Action{}, []Action{}, [][]Action{}, putIndexes, translated, isMutable, "private", []SubCaller{} })
        } else {
          actions = append(actions, Action{ "hash", "hashed_value", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 22, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, translated, isMutable, "private", []SubCaller{} })
        }
      case "[":
        var phrase = []Lex{}

        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt := 0

        for o := i; o < len_lex; o++ {
          if lex[o].Name == "{" {
            cbCnt++
          }
          if lex[o].Name == "}" {
            cbCnt--
          }

          if lex[o].Name == "[:" {
            glCnt++
          }
          if lex[o].Name == ":]" {
            glCnt--
          }

          if lex[o].Name == "[" {
            bCnt++
          }
          if lex[o].Name == "]" {
            bCnt--
          }

          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          phrase = append(phrase, lex[o])

          if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
            break
          }
        }

        i+=len(phrase)

        phrase = phrase[1:len(phrase) - 1]

        var arr [][]Action

        for o := 0; o < len(phrase); o++ {

          var sub []Lex

          cbCnt := 0
          glCnt := 0
          bCnt := 0
          pCnt := 0

          for j := o; j < len(phrase); j++ {

            if phrase[j].Name == "{" {
              cbCnt++
            }
            if phrase[j].Name == "}" {
              cbCnt--
            }

            if phrase[j].Name == "[:" {
              glCnt++
            }
            if phrase[j].Name == ":]" {
              glCnt--
            }

            if phrase[j].Name == "[" {
              bCnt++
            }
            if phrase[j].Name == "]" {
              bCnt--
            }

            if phrase[j].Name == "(" {
              pCnt++
            }
            if phrase[j].Name == ")" {
              pCnt--
            }

            if phrase[j].Name == "," && cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
              break
            }
            sub = append(sub, phrase[j])
          }

          o+=len(sub)

          arr = append(arr, Actionizer(sub, true, dir, name))
        }

        hashedArr := make(map[string][]Action)

        hashedArr["falsey"] = []Action{ Action{ "falsey", "exp_value", []string{ "undef" }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} } }

        cur := "0"

        for _, v := range arr {
          hashedArr[cur] = v
          cur = C.GoString(AddC(C.CString(cur), C.CString("1"), C.CString("{}")))
        }

        if i >= len_lex {
          actions = append(actions, Action{ "array", "hashed_value", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 24, []Action{}, []Action{}, []Action{}, arr, [][]Action{}, hashedArr, false, "private", []SubCaller{} })
          break
        }

        isMutable := false

        //checks for a runtime array
        for ;i < len_lex && lex[i].Name == ":::"; i++ {
          isMutable = !isMutable
        }

        if i >= len_lex {
          actions = append(actions, Action{ "array", "hashed_value", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 24, []Action{}, []Action{}, []Action{}, arr, [][]Action{}, hashedArr, isMutable, "private", []SubCaller{} })
          break
        }

        if lex[i].Name == "." {
          indexes := [][]Lex{ []Lex{} }

          cbCnt = 0
          glCnt = 0
          bCnt = 0
          pCnt = 0

          for o := i + 1; o < len_lex; o++ {
            if lex[o].Name == "{" {
              cbCnt++
            }
            if lex[o].Name == "[:" {
              glCnt++
            }
            if lex[o].Name == "[" {
              bCnt++
            }
            if lex[o].Name == "(" {
              pCnt++
            }

            if lex[o].Name == "}" {
              cbCnt--
            }
            if lex[o].Name == ":]" {
              glCnt--
            }
            if lex[o].Name == "]" {
              bCnt--
            }
            if lex[o].Name == ")" {
              pCnt--
            }

            if lex[o].Name == "." {
              indexes = append(indexes, []Lex{})
            } else {

              i++

              indexes[len(indexes) - 1] = append(indexes[len(indexes) - 1], lex[o])

              if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {

                if o < len_lex - 1 && lex[o + 1].Name == "." {
                  continue
                } else {
                  break
                }

              }
            }
          }

          var putIndexes [][]Action

          for _, v := range indexes {

            v = v[1:len(v) - 1]
            putIndexes = append(putIndexes, Actionizer(v, true, dir, name))
          }

          i+=3

          actions = append(actions, Action{ "arrayIndex", "", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 25, []Action{}, []Action{}, []Action{}, arr, putIndexes, hashedArr, isMutable, "private", []SubCaller{} })
        } else {
          actions = append(actions, Action{ "array", "hashed_value", []string{""}, []Action{}, []string{}, [][]Action{}, []Condition{}, 24, []Action{}, []Action{}, []Action{}, arr, [][]Action{}, hashedArr, isMutable, "private", []SubCaller{} })
        }
      case "each":
        var condition_ = []Lex{}

        pCnt := 0

        for o := i + 1; o < len_lex; o++ {
          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          condition_ = append(condition_, lex[o])

          if pCnt == 0 {
            break
          }
        }

        i+=len(condition_) + 1

        condition_ = condition_[1:len(condition_) - 1]

        var _iterator []Lex

        cbCnt := 0
        glCnt := 0
        bCnt := 0
        pCnt = 0

        var stopIterIndex int

        for k, v := range condition_ {

          if v.Name == "{" {
            cbCnt++
          }
          if v.Name == "}" {
            cbCnt--
          }

          if v.Name == "[:" {
            glCnt++
          }
          if v.Name == ":]" {
            glCnt--
          }

          if v.Name == "[" {
            bCnt++
          }
          if v.Name == "]" {
            bCnt--
          }

          if v.Name == "(" {
            pCnt++
          }
          if v.Name == ")" {
            pCnt--
          }

          if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && v.Name == "newlineS" {

            //index where the iterator stopped
            stopIterIndex = k
            break
          }

          _iterator = append(_iterator, v)
          stopIterIndex = k
        }

        iterator := Actionizer(_iterator, true, dir, name)

        var var1 string
        var var2 string

        if stopIterIndex + 1 >= len(condition_) {
          var1 = "$k"
          var2 = "$v"
        } else if stopIterIndex + 3 >= len(condition_) {
          var1 = condition_[stopIterIndex + 1].Name
          var2 = "$v"
        } else {
          var1, var2 = condition_[stopIterIndex + 1].Name, condition_[stopIterIndex + 3].Name
        }

        cbCnt = 0
        glCnt = 0
        bCnt = 0
        pCnt = 0

        var exp []Lex

        var curlyBraceCond = lex[i].Name == "{"

        for o := i; o < len_lex; o++ {
          if lex[o].Name == "{" {
            cbCnt++
          }
          if lex[o].Name == "}" {
            cbCnt--
          }

          if lex[o].Name == "[:" {
            glCnt++
          }
          if lex[o].Name == ":]" {
            glCnt--
          }

          if lex[o].Name == "[" {
            bCnt++
          }
          if lex[o].Name == "]" {
            bCnt--
          }

          if lex[o].Name == "(" {
            pCnt++
          }
          if lex[o].Name == ")" {
            pCnt--
          }

          exp = append(exp, lex[o])

          if !curlyBraceCond {

            if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
              break
            }
          } else if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {
            break
          }
        }

        i+=len(exp) + 1
        actionized := Actionizer(exp, false, dir, name)
        actions = append(actions, Action{ "each", "", []string{ var1, var2 }, actionized, []string{}, [][]Action{}, []Condition{}, 59, iterator, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })

      case "kill":

        if lex[i + 1].Name == "<-" {
          actions = append(actions, Action{ "kill_thread", "", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 65, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
          i++
        } else {
          actions = append(actions, Action{ "kill", "", []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, 66, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
        }

      case "this":

        act := this_calc(&i, lex, uint(1), "this", dir, name, 70)

        i++

        isMutable := false

        //checks for a runtime value
        for ;i < len_lex && lex[i].Name == ":::"; i++ {
          isMutable = !isMutable
        }

        act.IsMutable = isMutable

        actions = append(actions, act)

      default:

        valPuts := func(lex []Lex, i int) int {

          if i >= len_lex {
            return 1
          }

          isMutable := false
          val := lex[i].Name

          //checks for a runtime value
          for ;i + 1 < len_lex && lex[i + 1].Name == ":::"; i++ {
            isMutable = !isMutable
          }

          i++

          switch getType(val) {

            case "string": {

              noQ := val[1:len(val) - 1]
              hashedString := make(map[string][]Action)

              //specify the value for the "falsey" case
              hashedString["falsey"] = []Action{ Action{ "falsey", "exp_value", []string{ "undef" }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} } }

              cur := "0"

              for _, v := range noQ {

                hashedIndex := make(map[string][]Action)
                hashedIndex["falsey"] = []Action{ Action{ "falsey", "exp_value", []string{ "undef" }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} } }

                hashedString[cur] = []Action{ Action{ "string", "exp_value", []string{ string(v) }, []Action{}, []string{}, [][]Action{}, []Condition{}, 38, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, hashedIndex, isMutable, "private", []SubCaller{} } }
                cur = C.GoString(AddC(C.CString(cur), C.CString("1"), C.CString("{}")))
              }

              actions = append(actions, Action{ "string", "exp_value", []string{ noQ }, []Action{}, []string{}, [][]Action{}, []Condition{}, 38, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, hashedString, isMutable, "private", []SubCaller{} })
            }
            case "number":

              hashed := make(map[string][]Action)

              //specify the value for the "falsey" case
              hashed["falsey"] = []Action{ Action{ "falsey", "exp_value", []string{ "undef" }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} } }

              actions = append(actions, Action{ "number", "exp_value", []string{ val }, []Action{}, []string{}, [][]Action{}, []Condition{}, 39, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, hashed, false, "private", []SubCaller{} })
            case "boolean":

              hashed := make(map[string][]Action)

              //specify the value for the "falsey" case
              hashed["falsey"] = []Action{ Action{ "boolean", "exp_value", []string{ strconv.FormatBool(val != "true") }, []Action{}, []string{}, [][]Action{}, []Condition{}, 40, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} } }

              actions = append(actions, Action{ "boolean", "exp_value", []string{ val }, []Action{}, []string{}, [][]Action{}, []Condition{}, 40, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, hashed, isMutable, "private", []SubCaller{} })
            case "falsey":

              hashed := make(map[string][]Action)

              //specify the value for the "falsey" case
              hashed["falsey"] = []Action{ Action{ "falsey", "exp_value", []string{ val }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} } }

              actions = append(actions, Action{ "falsey", "exp_value", []string{ val }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, hashed, isMutable, "private", []SubCaller{} })
            case "none":

              if strings.HasPrefix(val, "$") {

                actions = append(actions, Action{ "variable", val, []string{ val }, []Action{}, []string{}, [][]Action{}, []Condition{}, 43, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} })
              } else {

                hashedString := make(map[string][]Action)

                //specify the value for the "falsey" case
                hashedString["falsey"] = []Action{ Action{ "falsey", "exp_value", []string{ "undef" }, []Action{}, []string{}, [][]Action{}, []Condition{}, 41, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} } }

                //get it? 42?
                actions = append(actions, Action{ "none", "exp_value", []string{ val }, []Action{}, []string{}, [][]Action{}, []Condition{}, 42, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), isMutable, "private", []SubCaller{} })
              }
          }

          return 0
        }

        if i + 1 < len_lex {

          if lex[i + 1].Name == "->" {

            var val_ []Lex

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            for o := i + 2; o < len_lex; o++ {
              if lex[o].Name == "{" {
                cbCnt++
              }
              if lex[o].Name == "}" {
                cbCnt--
              }

              if lex[o].Name == "[:" {
                glCnt++
              }
              if lex[o].Name == ":]" {
                glCnt--
              }

              if lex[o].Name == "[" {
                bCnt++
              }
              if lex[o].Name == "]" {
                bCnt--
              }

              if lex[o].Name == "(" {
                pCnt++
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && arrayContainInterface(operations, lex[o].Name) {
                break
              }

              val_ = append(val_, lex[o])
            }

            val := Actionizer(val_, true, dir, name)

            getValueType := func(val string) string {
              switch getType(val) {
                case "string": fallthrough
                case "number": fallthrough
                case "boolean": fallthrough
                case "falsey":
                  return "exp_value"
                case "array": fallthrough
                case "hash":
                  return "hashed_value"
              }

              return "exp_value"
            }

            actions = append(actions, Action{ "cast", lex[i].Name, []string{ getValueType(lex[i].Name) }, val, []string{}, [][]Action{}, []Condition{}, 58, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
            i+=len(val_) + 2
            continue
          }

          if (lex[i + 1].Name == "++" || lex[i + 1].Name == "--") && strings.HasPrefix(lex[i].Name, "$") {

            id_ := []byte(lex[i + 1].Name)

            id := ""

            for o := 0; o < len(id_); o++ {
              _id := strconv.Itoa(int(id_[o]))
              id+=_id
            }

            intID, _ := strconv.Atoi(id)

            actions = append(actions, Action{ lex[i + 1].Name, lex[i].Name, []string{}, []Action{}, []string{}, [][]Action{}, []Condition{}, intID, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
            i++
            continue;
          }

          if (lex[i + 1].Name == "+=" || lex[i + 1].Name == "-=" || lex[i + 1].Name == "*=" || lex[i + 1].Name == "/=" || lex[i + 1].Name == "%=" || lex[i + 1].Name == "^=") && strings.HasPrefix(lex[i].Name, "$") {

            var by_ []Lex

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            for o := i + 2; o < len_lex; o++ {
              if lex[o].Name == "{" {
                cbCnt++
              }
              if lex[o].Name == "}" {
                cbCnt--
              }

              if lex[o].Name == "[:" {
                glCnt++
              }
              if lex[o].Name == ":]" {
                glCnt--
              }

              if lex[o].Name == "[" {
                bCnt++
              }
              if lex[o].Name == "]" {
                bCnt--
              }

              if lex[o].Name == "(" {
                pCnt++
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
                break
              }

              by_ = append(by_, lex[o])
            }

            by := Actionizer(by_, true, dir, name)

            id_ := []byte(lex[i + 1].Name)

            id := ""

            for o := 0; o < len(id_); o++ {
              _id := strconv.Itoa(int(id_[o]))
              id+=_id
            }

            intID, _ := strconv.Atoi(id)

            actions = append(actions, Action{ lex[i + 1].Name, lex[i].Name, []string{}, by, []string{}, [][]Action{}, []Condition{}, intID, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false, "private", []SubCaller{} })
            continue;
          }

          doPutIndex := false

          icbCnt := 0
          iglCnt := 0
          ibCnt := 0
          ipCnt := 0

          for o := i; o < len_lex; o++ {
            if lex[o].Name == "{" {
              icbCnt++
            }
            if lex[o].Name == "}" {
              icbCnt--
            }

            if lex[o].Name == "[:" {
              iglCnt++
            }
            if lex[o].Name == ":]" {
              iglCnt--
            }

            if lex[o].Name == "[" {
              ibCnt++
            }
            if lex[o].Name == "]" {
              ibCnt--
            }

            if lex[o].Name == "(" {
              ipCnt++
            }
            if lex[o].Name == ")" {
              ipCnt--
            }

            if icbCnt == 0 && iglCnt == 0 && ibCnt == 0 && ipCnt == 0 && lex[o].Name == "newlineS" {
              break
            }

            if icbCnt == 0 && iglCnt == 0 && ibCnt == 0 && ipCnt == 0 && lex[o].Name == ":" {
              doPutIndex = true
              break
            }
          }

          var indexes [][]Action
          varname := lex[i].Name

          if lex[i + 1].Name == "." && lex[i + 2].Name == "[" && doPutIndex {

            _indexes := [][]Lex{}

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            for o := i + 1; o < len_lex; i, o = i + 1, o + 1 {
              if lex[o].Name == "{" {
                cbCnt++
              }
              if lex[o].Name == "}" {
                cbCnt--
              }

              if lex[o].Name == "[:" {
                glCnt++
              }
              if lex[o].Name == ":]" {
                glCnt--
              }

              if lex[o].Name == "[" {
                bCnt++
              }
              if lex[o].Name == "]" {
                bCnt--
              }

              if lex[o].Name == "(" {
                pCnt++
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              if lex[o].Name == "." {
                _indexes = append(_indexes, []Lex{})
                continue
              }

              _indexes[len(_indexes) - 1] = append(_indexes[len(_indexes) - 1], lex[o])

              if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o + 1].Name == ":" {
                break
              }
            }

            for _, v := range _indexes {
              indexes = append(indexes, Actionizer(v[1:len(v) - 1], true, dir, name))
            }

            i++
          }

          if lex[i + 1].Name == ":" && (strings.HasPrefix(lex[i].Name, "$") || lex[i].Name == "]") {
            exp_ := []Lex{}

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            for o := i + 2; o < len_lex; o++ {

              if lex[o].Name == "{" {
                cbCnt++
              }
              if lex[o].Name == "}" {
                cbCnt--
              }

              if lex[o].Name == "[:" {
                glCnt++
              }
              if lex[o].Name == ":]" {
                glCnt--
              }

              if lex[o].Name == "[" {
                bCnt++
              }
              if lex[o].Name == "]" {
                bCnt--
              }

              if lex[o].Name == "(" {
                pCnt++
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && lex[o].Name == "newlineS" {
                break
              }

              exp_ = append(exp_, lex[o]);
            }

            exp := Actionizer(exp_, true, dir, name)

            actions = append(actions, Action{ "let", varname, []string{}, exp, []string{}, [][]Action{}, []Condition{}, 28, []Action{}, []Action{}, []Action{}, [][]Action{}, indexes, make(map[string][]Action), false, "private", []SubCaller{} })
            i+=(len(exp_))
            continue
          }

          if lex[i + 1].Name == "." {

            val := lex[i].Name

            cbCnt := 0
            glCnt := 0
            bCnt := 0
            pCnt := 0

            indexes := [][]Lex{ []Lex{} }

            cbCnt = 0
            glCnt = 0
            bCnt = 0
            pCnt = 0

            for o := i + 2; o < len_lex; o++ {
              if lex[o].Name == "{" {
                cbCnt++
              }
              if lex[o].Name == "[:" {
                glCnt++
              }
              if lex[o].Name == "[" {
                bCnt++
              }
              if lex[o].Name == "(" {
                pCnt++
              }

              if lex[o].Name == "}" {
                cbCnt--
              }
              if lex[o].Name == ":]" {
                glCnt--
              }
              if lex[o].Name == "]" {
                bCnt--
              }
              if lex[o].Name == ")" {
                pCnt--
              }

              if lex[o].Name == "." {
                indexes = append(indexes, []Lex{})
              } else {

                i++

                indexes[len(indexes) - 1] = append(indexes[len(indexes) - 1], lex[o])

                if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {

                  if o < len_lex - 1 && lex[o + 1].Name == "." {
                    continue
                  } else {
                    break
                  }

                }
              }
            }

            var putIndexes [][]Action

            for _, v := range indexes {

              v = v[1:len(v) - 1]
              putIndexes = append(putIndexes, Actionizer(v, true, dir, name))
            }

            i+=3

            if strings.HasPrefix(val, "$") {
              actVal := Actionizer([]Lex{ Lex{ val, "", 0, "", "", dir } }, true, dir, name)

              actions = append(actions, Action{ "variableIndex", "", []string{}, actVal, []string{}, [][]Action{}, []Condition{}, 46, []Action{}, []Action{}, []Action{}, [][]Action{}, putIndexes, make(map[string][]Action), false, "private", []SubCaller{} })
            } else {
              actVal := Actionizer([]Lex{ Lex{ val, "", 0, "", "", dir } }, true, dir, name)

              actions = append(actions, Action{ "expressionIndex", "", []string{}, actVal, []string{}, [][]Action{}, []Condition{}, 8, []Action{}, []Action{}, []Action{}, [][]Action{}, putIndexes, make(map[string][]Action), false, "private", []SubCaller{} })
            }

          }

          valPuts(lex, i)

        } else {

          valPuts(lex, i)
        }
      }
  }

  return actions
}
