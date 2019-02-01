package main

import (
	. "github.com/fishedee/language"
	"html/template"
)

func QueryCombineGen(request queryGenRequest) *queryGenResponse {
	args := request.args
	line := request.pkg.FileSet().Position(request.expr.Pos()).String()

	//解析第一个参数
	firstArgSlice := getSliceType(line, args[0].Type)
	firstArgSliceNamed := getNamedType(line, firstArgSlice.Elem())

	//解析第二个参数
	secondArgSlice := getSliceType(line, args[1].Type)
	secondArgSliceNamed := getNamedType(line, secondArgSlice.Elem())

	//解析第三个参数
	thirdArgFunc := getFunctionType(line, args[2].Type)
	thirdArgFuncArgument := getArgumentType(line, thirdArgFunc)
	thirdArgFuncReturn := getReturnType(line, thirdArgFunc)
	if len(thirdArgFuncArgument) != 2 {
		Throw(1, "%v:should be two argument", line)
	}
	if len(thirdArgFuncReturn) != 1 {
		Throw(1, "%v:should be one return", line)
	}
	if thirdArgFuncArgument[0].String() != firstArgSliceNamed.String() {
		Throw(1, "%v:groupFunctor first argument should be equal with first argument %v!=%v", line, thirdArgFuncArgument[0], firstArgSliceNamed)
	}
	if thirdArgFuncArgument[1].String() != secondArgSliceNamed.String() {
		Throw(1, "%v:groupFunctor second argument should be equal with second argument %v!=%v", line, thirdArgFuncArgument[1], secondArgSliceNamed)
	}

	//生成函数
	signature := getFunctionSignature(line, args, []bool{false, false, false})
	if hasQueryCombineGenerate[signature] == true {
		return nil
	}
	hasQueryCombineGenerate[signature] = true
	importPackage := map[string]bool{}
	setImportPackage(line, firstArgSliceNamed, importPackage)
	setImportPackage(line, secondArgSliceNamed, importPackage)
	setImportPackage(line, thirdArgFuncReturn[0], importPackage)
	argumentDefine := getFunctionArgumentCode(line, args, []bool{false, false, false})
	funcBody := excuteTemplate(queryCombineFuncTmpl, map[string]string{
		"signature":          signature,
		"firstArgElemType":   getTypeDeclareCode(line, firstArgSliceNamed),
		"secondArgElemType":  getTypeDeclareCode(line, secondArgSliceNamed),
		"thirdArgType":       getTypeDeclareCode(line, thirdArgFunc),
		"thirdArgReturnType": getTypeDeclareCode(line, thirdArgFuncReturn[0]),
	})
	initBody := excuteTemplate(queryCombineInitTmpl, map[string]string{
		"signature":      signature,
		"argumentDefine": argumentDefine,
	})
	return &queryGenResponse{
		importPackage: importPackage,
		funcName:      "queryCombine_" + signature,
		funcBody:      funcBody,
		initBody:      initBody,
	}
}

var (
	queryCombineFuncTmpl    *template.Template
	queryCombineInitTmpl    *template.Template
	hasQueryCombineGenerate map[string]bool
)

func init() {
	var err error
	queryCombineFuncTmpl, err = template.New("name").Parse(`
	func queryCombine_{{ .signature }}(leftData interface{},rightData interface{},combineFunctor interface{})interface{}{
		leftDataIn := leftData.([]{{ .firstArgElemType }})
		rightDataIn := rightData.([]{{ .secondArgElemType }})
		combineFunctorIn := combineFunctor.({{ .thirdArgType }})
		newData := make([]{{ .thirdArgReturnType }},len(leftDataIn),len(leftDataIn))

		for i := 0 ;i != len(leftDataIn);i++{
			newData[i] = combineFunctorIn(leftDataIn[i],rightDataIn[i])
		}
		return newData
	}
	`)
	if err != nil {
		panic(err)
	}
	queryCombineInitTmpl, err = template.New("name").Parse(`
		language.QueryCombineMacroRegister({{.argumentDefine}},queryCombine_{{.signature}})
	`)
	if err != nil {
		panic(err)
	}
	registerQueryGen("github.com/fishedee/language.QueryCombine", QueryCombineGen)
	hasQueryCombineGenerate = map[string]bool{}
}