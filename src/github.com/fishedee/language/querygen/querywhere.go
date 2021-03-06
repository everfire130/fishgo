package main

import (
	. "github.com/fishedee/language"
	"html/template"
)

func QueryWhereGen(request queryGenRequest) *queryGenResponse {
	args := request.args
	line := request.pkg.FileSet().Position(request.expr.Pos()).String()

	//解析第一个参数
	firstArgSlice := getSliceType(line, args[0].Type)
	firstArgElem := firstArgSlice.Elem()

	//解析第二个参数
	secondArgFunc := getFunctionType(line, args[1].Type)
	secondArgFuncArguments := getArgumentType(line, secondArgFunc)
	secondArgFuncResults := getReturnType(line, secondArgFunc)
	if len(secondArgFuncArguments) != 1 {
		Throw(1, "%v:selector should be single argument")
	}
	if len(secondArgFuncResults) != 1 {
		Throw(1, "%v:selector should be single return")
	}
	secondArgFuncArgument := secondArgFuncArguments[0]
	secondArgFuncResult := secondArgFuncResults[0]
	if firstArgElem.String() != secondArgFuncArgument.String() {
		Throw(1, "%v:second selector argument should be equal with first argument")
	}
	if secondArgFuncResult.String() != "bool" {
		Throw(1, "%v:second selector return should be bool type")
	}

	//生成函数
	signature := getFunctionSignature(line, args, []bool{false, false})
	if hasQueryWhereGenerate[signature] == true {
		return nil
	}
	hasQueryWhereGenerate[signature] = true
	importPackage := map[string]bool{}
	setImportPackage(line, secondArgFuncArgument, importPackage)
	argumentDefine := getFunctionArgumentCode(line, args, []bool{false, false})
	funcBody := excuteTemplate(queryWhereFuncTmpl, map[string]string{
		"signature":        signature,
		"firstArgElemType": getTypeDeclareCode(line, firstArgElem),
		"secondArgType":    getTypeDeclareCode(line, secondArgFunc),
	})
	initBody := excuteTemplate(queryWhereInitTmpl, map[string]string{
		"signature":      signature,
		"argumentDefine": argumentDefine,
	})
	return &queryGenResponse{
		importPackage: importPackage,
		funcName:      "queryWhere_" + signature,
		funcBody:      funcBody,
		initBody:      initBody,
	}
}

var (
	queryWhereFuncTmpl    *template.Template
	queryWhereInitTmpl    *template.Template
	hasQueryWhereGenerate map[string]bool
)

func init() {
	var err error
	queryWhereFuncTmpl, err = template.New("name").Parse(`
	func queryWhere_{{ .signature }}(data interface{},whereFunctor interface{})interface{}{
		dataIn := data.([]{{ .firstArgElemType }})
		whereFunctorIn := whereFunctor.({{ .secondArgType }})
		result := make([]{{ .firstArgElemType }},0,len(dataIn))

		for _,single := range dataIn{
			shouldStay := whereFunctorIn(single)
			if shouldStay == true {
				result = append(result,single)
			}
		}
		return result
	}
	`)
	if err != nil {
		panic(err)
	}
	queryWhereInitTmpl, err = template.New("name").Parse(`
		language.QueryWhereMacroRegister({{.argumentDefine}},queryWhere_{{.signature}})
	`)
	if err != nil {
		panic(err)
	}
	registerQueryGen("github.com/fishedee/language.QueryWhere", QueryWhereGen)
	hasQueryWhereGenerate = map[string]bool{}
}
