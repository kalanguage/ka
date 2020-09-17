package interpreter

import (
	. "github.com/tusklang/tusk/lang/types"
)

func number__divide__number(val1, val2 TuskType, instance *Instance, stacktrace []string, line uint64, file string) *TuskType {
	num1, num2 := val1.(TuskNumber), val2.(TuskNumber)
	ensurePrec(&num1, &num2, (*instance).Params)

	//maybe in a future version switch to the algorithm python uses
	//knuth division
	//https://skanthak.homepage.t-online.de/division.html

	//num2 is the divisor
	//num1 is the dividend

	if isEqual(num2, zero) { //if it is n/0, throw an error
		TuskPanic("Divide by zero error", line, file, stacktrace)
	}
	if isEqual(num1, zero) { //if it is 0/n return 0
		var ztype TuskType = zero
		return &ztype
	}

	decPlaces := len(*num1.Integer) + len(*num2.Decimal)
	num1n := append(*num1.Decimal, *num1.Integer...)
	num2n := zero
	tmp := append(*num2.Decimal, *num2.Integer...)
	num2n.Integer = &tmp

	a := zero
	a.Integer = &num1n

	for i := len(num1n); uint64(i) < (*instance).Params.Prec; i++ {
		num1n = append([]int64{0}, num1n...)
	}

	curVal := zero    //current value under the "house" of the division
	var final []int64 //final value

	num2Abs := abs(num2n, stacktrace, (*instance).Params).(TuskNumber)

	a = zero
	a.Integer = &num1n

	for i := len(num1n) - 1; i >= 0; i-- {
		v := num1n[i]

		tmpCV := append([]int64{v}, *curVal.Integer...)
		curVal.Integer = &tmpCV
		curValAbs := abs(curVal, stacktrace, (*instance).Params).(TuskNumber)

		if isLess(curValAbs, num2Abs) {
			final = append([]int64{0}, final...)
			continue
		}

		var curQuotient TuskNumber = zero
		var added TuskNumber = zero

		for addedTemp := added; func() bool {
			addedTemp = (*number__plus__number(addedTemp, num2Abs, instance, stacktrace, line, file)).(TuskNumber)
			return isLessOrEqual(addedTemp, curValAbs)
		}(); added = addedTemp {
			curQuotient = (*number__plus__number(curQuotient, one, instance, stacktrace, line, file)).(TuskNumber) //increment the current quotient
		}

		apn2 := (*number__plus__number(added, num2Abs, instance, stacktrace, line, file)).(TuskNumber)

		if isEqual(apn2, curValAbs) {
			added = apn2
			curQuotient = (*number__plus__number(curQuotient, one, instance, stacktrace, line, file)).(TuskNumber)
		}

		if isLess(num1, zero) {
			curQuotient = (*number__times__number(curQuotient, neg_one, instance, stacktrace, line, file)).(TuskNumber)
		}

		//remove leading zeros from the curQuotient
		for len(*curQuotient.Integer) != 1 && (*curQuotient.Integer)[len(*curQuotient.Integer)-1] == 0 {
			*curQuotient.Integer = (*curQuotient.Integer)[:len(*curQuotient.Integer)-1]
		}

		curVal = (*number__minus__number(curValAbs, added, instance, stacktrace, line, file)).(TuskNumber)
		final = append(*curQuotient.Integer, final...)
	}

	if isLess(num2, zero) { //if num2 is negative, multiply the final by -1
		finalAct := zero
		finalAct.Integer = &final
		finalAct = (*number__times__number(finalAct, neg_one, instance, stacktrace, line, file)).(TuskNumber)
		final = *finalAct.Integer
	}

	ret := zero
	tmpInt := final[len(final)-decPlaces:]
	tmpDec := final[:len(final)-decPlaces]
	ret.Integer, ret.Decimal = &tmpInt, &tmpDec

	var retType TuskType = ret

	return &retType
}
