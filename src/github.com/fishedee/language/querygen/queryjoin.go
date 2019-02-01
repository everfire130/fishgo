package main

import (
	. "github.com/fishedee/language"
	"go/constant"
	"go/types"
	"html/template"
	"strings"
)

func analyseJoin(line string, joinType string) (string, string) {
	joinTypeArray := strings.Split(joinType, "=")
	if len(joinTypeArray) != 2 {
		Throw(1, "%v:join type should be two argument with equal operator", line)
	}
	leftJoinType := strings.Trim(joinTypeArray[0], " ")
	rightJoinType := strings.Trim(joinTypeArray[1], " ")
	return leftJoinType, rightJoinType
}

func QueryJoinGen(request queryGenRequest) *queryGenResponse {
	args := request.args
	line := request.pkg.FileSet().Position(request.expr.Pos()).String()

	//解析第一个参数
	firstArgSlice := getSliceType(line, args[0].Type)
	firstArgSliceNamed := getNamedType(line, firstArgSlice.Elem())
	firstArgSliceStruct := getStructType(line, firstArgSliceNamed.Underlying())

	//解析第二个参数
	secondArgSlice := getSliceType(line, args[1].Type)
	secondArgSliceNamed := getNamedType(line, secondArgSlice.Elem())
	secondArgSliceStruct := getStructType(line, secondArgSliceNamed.Underlying())

	//解析第三个参数
	thirdArgJoinPlace := getContantStringValue(line, args[2].Value)
	joinPlace := strings.Trim(strings.ToLower(thirdArgJoinPlace), " ")
	if joinPlace != "left" && joinPlace != "right" &&
		joinPlace != "inner" && joinPlace != "outer" {
		Throw(1, "%v:invalid join place %v", line, joinPlace)
	}

	//解析第四个参数
	forthArgJoinType := getContantStringValue(line, args[3].Value)
	leftJoinColumn, rightJoinColumn := analyseJoin(line, forthArgJoinType)
	leftFieldType := getFieldType(line, firstArgSliceStruct, leftJoinColumn)
	rightFieldType := getFieldType(line, secondArgSliceStruct, rightJoinColumn)
	if leftFieldType.String() != rightFieldType.String() {
		Throw(1, "%v:left join type should be equal to right join type %v!=%v", line, leftFieldType.String(), rightFieldType.String())
	}

	//解析第五个参数
	fifthArgFunc := getFunctionType(line, args[4].Type)
	fifthArgFuncArgument := getArgumentType(line, fifthArgFunc)
	fifthArgFuncReturn := getReturnType(line, fifthArgFunc)
	if len(fifthArgFuncArgument) != 2 {
		Throw(1, "%v:should be two argument", line)
	}
	if len(fifthArgFuncReturn) != 1 {
		Throw(1, "%v:should be one return", line)
	}
	if fifthArgFuncArgument[0].String() != firstArgSliceNamed.String() {
		Throw(1, "%v:joinFuctor first argument should be equal with first argument %v!=%v", line, fifthArgFuncArgument[0], firstArgSliceNamed)
	}
	if fifthArgFuncArgument[1].String() != secondArgSliceNamed.String() {
		Throw(1, "%v:joinFuctor second argument should be equal with second argument %v!=%v", line, fifthArgFuncArgument[1], secondArgSliceNamed)
	}

	//生成函数
	signature := getFunctionSignature(line, args, []bool{false, false, true, true, false})
	if hasQueryJoinGenerate[signature] == true {
		return nil
	}
	hasQueryJoinGenerate[signature] = true
	importPackage := map[string]bool{}
	setImportPackage(line, firstArgSliceNamed, importPackage)
	setImportPackage(line, secondArgSliceNamed, importPackage)
	setImportPackage(line, fifthArgFuncReturn[0], importPackage)
	argumentDefine := getFunctionArgumentCode(line, args, []bool{false, false, true, true, false})
	funcBody := excuteTemplate(queryJoinFuncTmpl, map[string]string{
		"signature":               signature,
		"firstArgElemType":        getTypeDeclareCode(line, firstArgSliceNamed),
		"secondArgElemType":       getTypeDeclareCode(line, secondArgSliceNamed),
		"fifthArgType":            getTypeDeclareCode(line, fifthArgFunc),
		"fifthArgReturnType":      getTypeDeclareCode(line, fifthArgFuncReturn[0]),
		"firstArgElemTypeDefine":  getTypeDefineCode(line, firstArgSliceNamed),
		"secondArgElemTypeDefine": getTypeDefineCode(line, secondArgSliceNamed),
		"joinPlace":               joinPlace,
		"secondArgSortCode":       getLessCompareCode(line, "newRightData[i]", rightJoinColumn, "newRightData[j]", rightJoinColumn, true, rightFieldType),
		"fifthArgSortCode":        getLessCompareCode(line, "leftDataIn[i]", leftJoinColumn, "newRightData[j]", rightJoinColumn, true, rightFieldType),
	})
	initBody := excuteTemplate(queryJoinInitTmpl, map[string]string{
		"signature":      signature,
		"argumentDefine": argumentDefine,
	})
	return &queryGenResponse{
		importPackage: importPackage,
		funcName:      "queryJoin_" + signature,
		funcBody:      funcBody,
		initBody:      initBody,
	}
}

var (
	queryJoinFuncTmpl    *template.Template
	queryJoinInitTmpl    *template.Template
	hasQueryJoinGenerate map[string]bool
)

func init() {
	var err error
	queryJoinFuncTmpl, err = template.New("name").Parse(`
	func queryJoin_{{ .signature }}(leftData interface{},rightData interface{},joinPlace string,joinType string,joinFunctor interface{})interface{}{
		leftDataIn := leftData.([]{{ .firstArgElemType }})
		rightDataIn := rightData.([]{{ .secondArgElemType }})
		joinFunctorIn := joinFunctor.({{ .fifthArgType }})
		newRightData := make([]{{ .secondArgElemType }},len(rightDataIn),len(rightDataIn))
		copy(newRightData,rightDataIn)
		newData2 := make([]{{ .fifthArgReturnType }},0,len(leftDataIn))

		emptyLeftData := {{ .firstArgElemTypeDefine }}
		emptyRightData := {{ .secondArgElemTypeDefine }}
		language.QueryJoinInternal(
			"{{ .joinPlace }}",
			len(leftDataIn),
			len(rightDataIn),
			func(i int,j int)int{
				{{ .secondArgSortCode }}
				return 0
			},
			func(i int, j int){
				newRightData[j],newRightData[i] = newRightData[i],newRightData[j]
			},
			func(i int,j int)int{
				{{ .fifthArgSortCode }}
				return 0
			},
			func(i int,j int){
				left := emptyLeftData
				if i != -1{
					left = leftDataIn[i]
				}
				right := emptyRightData
				if j != -1{
					right = newRightData[j]
				}
				single := joinFunctorIn(left,right)
				newData2 = append(newData2,single)
			},
		)
		return newData2
	}
	`)
	if err != nil {
		panic(err)
	}
	queryJoinInitTmpl, err = template.New("name").Parse(`
		language.QueryJoinMacroRegister({{.argumentDefine}},queryJoin_{{.signature}})
	`)
	if err != nil {
		panic(err)
	}
	registerQueryGen("github.com/fishedee/language.QueryJoin", QueryJoinGen)
	registerQueryGen("github.com/fishedee/language.QueryLeftJoin", func(request queryGenRequest) *queryGenResponse {
		thridParty := types.TypeAndValue{
			Type:  nil,
			Value: constant.MakeString("left"),
		}
		newArgs := []types.TypeAndValue{}
		newArgs = append(newArgs, request.args[0:2]...)
		newArgs = append(newArgs, thridParty)
		newArgs = append(newArgs, request.args[2:]...)
		request.args = newArgs
		return QueryJoinGen(request)
	})
	registerQueryGen("github.com/fishedee/language.QueryRightJoin", func(request queryGenRequest) *queryGenResponse {
		thridParty := types.TypeAndValue{
			Type:  nil,
			Value: constant.MakeString("right"),
		}
		newArgs := []types.TypeAndValue{}
		newArgs = append(newArgs, request.args[0:2]...)
		newArgs = append(newArgs, thridParty)
		newArgs = append(newArgs, request.args[2:]...)
		request.args = newArgs
		return QueryJoinGen(request)
	})
	registerQueryGen("github.com/fishedee/language.QueryInnerJoin", func(request queryGenRequest) *queryGenResponse {
		thridParty := types.TypeAndValue{
			Type:  nil,
			Value: constant.MakeString("inner"),
		}
		newArgs := []types.TypeAndValue{}
		newArgs = append(newArgs, request.args[0:2]...)
		newArgs = append(newArgs, thridParty)
		newArgs = append(newArgs, request.args[2:]...)
		request.args = newArgs
		return QueryJoinGen(request)
	})
	registerQueryGen("github.com/fishedee/language.QueryOuterJoin", func(request queryGenRequest) *queryGenResponse {
		thridParty := types.TypeAndValue{
			Type:  nil,
			Value: constant.MakeString("outer"),
		}
		newArgs := []types.TypeAndValue{}
		newArgs = append(newArgs, request.args[0:2]...)
		newArgs = append(newArgs, thridParty)
		newArgs = append(newArgs, request.args[2:]...)
		request.args = newArgs
		return QueryJoinGen(request)
	})
	hasQueryJoinGenerate = map[string]bool{}
}