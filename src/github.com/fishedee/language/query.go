package language

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"
)

//基础类函数QuerySelect
type QuerySelectMacroHandler func(data interface{}, selectFunctor interface{}) interface{}

func QuerySelectMacroRegister(data interface{}, selectFunctor interface{}, handler QuerySelectMacroHandler) {
	id := registerQueryTypeId([]string{reflect.TypeOf(data).String(), reflect.TypeOf(selectFunctor).String()})
	querySelectMacroMapper[id] = handler
}

func QuerySelectReflect(data interface{}, selectFuctor interface{}) interface{} {
	dataValue := reflect.ValueOf(data)
	dataLen := dataValue.Len()

	selectFuctorValue := reflect.ValueOf(selectFuctor)
	selectFuctorType := selectFuctorValue.Type()
	selectFuctorOuterType := selectFuctorType.Out(0)
	resultType := reflect.SliceOf(selectFuctorOuterType)
	resultValue := reflect.MakeSlice(resultType, dataLen, dataLen)
	callArgument := []reflect.Value{reflect.Value{}}

	for i := 0; i != dataLen; i++ {
		singleDataValue := dataValue.Index(i)
		callArgument[0] = singleDataValue
		singleResultValue := selectFuctorValue.Call(callArgument)[0]
		resultValue.Index(i).Set(singleResultValue)
	}
	return resultValue.Interface()
}

func QuerySelect(data interface{}, selectFunctor interface{}) interface{} {
	id := getQueryTypeId([]string{reflect.TypeOf(data).String(), reflect.TypeOf(selectFunctor).String()})
	handler, isExist := querySelectMacroMapper[id]
	if isExist {
		return handler(data, selectFunctor)
	} else {
		return QuerySelectReflect(data, selectFunctor)
	}
}

//基础类函数QueryWhere
type QueryWhereMacroHandler func(data interface{}, whereFunctor interface{}) interface{}

func QueryWhereMacroRegister(data interface{}, whereFunctor interface{}, handler QueryWhereMacroHandler) {
	id := registerQueryTypeId([]string{reflect.TypeOf(data).String(), reflect.TypeOf(whereFunctor).String()})
	queryWhereMacroMapper[id] = handler
}

func QueryWhereReflect(data interface{}, whereFuctor interface{}) interface{} {
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()
	dataLen := dataValue.Len()

	whereFuctorValue := reflect.ValueOf(whereFuctor)
	resultType := reflect.SliceOf(dataType.Elem())
	resultValue := reflect.MakeSlice(resultType, 0, 0)

	for i := 0; i != dataLen; i++ {
		singleDataValue := dataValue.Index(i)
		singleResultValue := whereFuctorValue.Call([]reflect.Value{singleDataValue})[0]
		if singleResultValue.Bool() {
			resultValue = reflect.Append(resultValue, singleDataValue)
		}
	}
	return resultValue.Interface()
}

func QueryWhere(data interface{}, whereFuctor interface{}) interface{} {
	id := getQueryTypeId([]string{reflect.TypeOf(data).String(), reflect.TypeOf(whereFuctor).String()})
	handler, isExist := queryWhereMacroMapper[id]
	if isExist {
		return handler(data, whereFuctor)
	} else {
		return QueryWhereReflect(data, whereFuctor)
	}
}

//基础类函数QuerySort
type querySortInterface struct {
	lenHandler  func() int
	lessHandler func(i int, j int) bool
	swapHandler func(i int, j int)
}

func (this *querySortInterface) Len() int {
	return this.lenHandler()
}

func (this *querySortInterface) Less(i int, j int) bool {
	return this.lessHandler(i, j)
}

func (this *querySortInterface) Swap(i int, j int) {
	this.swapHandler(i, j)
}

func QuerySortInternal(length int, lessHandler func(i, j int) int, swapHandler func(i, j int)) {
	sort.Stable(&querySortInterface{
		lenHandler: func() int {
			return length
		},
		lessHandler: func(i int, j int) bool {
			return lessHandler(i, j) < 0
		},
		swapHandler: swapHandler,
	})

}

type QuerySortMacroHandler func(data interface{}, sortType string) interface{}

func QuerySortMacroRegister(data interface{}, sortType string, handler QuerySortMacroHandler) {
	id := registerQueryTypeId([]string{reflect.TypeOf(data).String(), sortType})
	querySortMacroMapper[id] = handler
}

func QuerySortReflect(data interface{}, sortType string) interface{} {
	//拷贝一份
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()
	dataElemType := dataType.Elem()
	dataValueLen := dataValue.Len()

	dataResult := reflect.MakeSlice(dataType, dataValueLen, dataValueLen)
	reflect.Copy(dataResult, dataValue)

	//排序
	targetCompares := getQueryExtractAndCompares(dataElemType, sortType)
	targetCompare := combineQueryCompare(targetCompares)
	result := dataResult.Interface()
	swapper := reflect.Swapper(result)

	QuerySortInternal(dataValueLen, func(i int, j int) int {
		left := dataResult.Index(i)
		right := dataResult.Index(j)
		return targetCompare(left, right)
	}, swapper)

	return result
}

func QuerySort(data interface{}, sortType string) interface{} {
	id := getQueryTypeId([]string{reflect.TypeOf(data).String(), sortType})
	handler, isExist := querySortMacroMapper[id]
	if isExist {
		return handler(data, sortType)
	} else {
		return QuerySortReflect(data, sortType)
	}
}

func QueryJoin(leftData interface{}, rightData interface{}, joinPlace string, joinType string, joinFuctor interface{}) interface{} {
	//解析配置
	leftJoinType, rightJoinType := analyseJoin(joinType)

	leftDataValue := reflect.ValueOf(leftData)
	leftDataType := leftDataValue.Type()
	leftDataElemType := leftDataType.Elem()
	leftDataValueLen := leftDataValue.Len()
	leftDataJoinType, leftDataJoinExtract := getQueryExtract(leftDataElemType, leftJoinType)

	rightData = QuerySort(rightData, rightJoinType+" asc")
	rightDataValue := reflect.ValueOf(rightData)
	rightDataType := rightDataValue.Type()
	rightDataElemType := rightDataType.Elem()
	rightDataValueLen := rightDataValue.Len()
	_, rightDataJoinExtract := getQueryExtract(rightDataElemType, rightJoinType)

	joinFuctorValue := reflect.ValueOf(joinFuctor)
	joinFuctorType := joinFuctorValue.Type()
	joinCompare := getQueryCompare(leftDataJoinType)
	resultValue := reflect.MakeSlice(reflect.SliceOf(joinFuctorType.Out(0)), 0, 0)

	rightHaveJoin := make([]bool, rightDataValueLen, rightDataValueLen)
	joinPlace = strings.Trim(strings.ToLower(joinPlace), " ")
	if ArrayIn([]string{"left", "right", "inner", "outer"}, joinPlace) == -1 {
		panic("invalid joinPlace [" + joinPlace + "] ")
	}

	//开始join
	for i := 0; i != leftDataValueLen; i++ {
		//二分查找右边对应的键
		singleLeftData := leftDataValue.Index(i)
		singleLeftDataJoin := leftDataJoinExtract(singleLeftData)
		j := sort.Search(rightDataValueLen, func(j int) bool {
			return joinCompare(rightDataJoinExtract(rightDataValue.Index(j)), singleLeftDataJoin) >= 0
		})
		//合并双边满足条件
		haveFound := false
		for ; j < rightDataValueLen; j++ {
			singleRightData := rightDataValue.Index(j)
			singleRightDataJoin := rightDataJoinExtract(singleRightData)
			if joinCompare(singleLeftDataJoin, singleRightDataJoin) != 0 {
				break
			}
			singleResult := joinFuctorValue.Call([]reflect.Value{singleLeftData, singleRightData})[0]
			resultValue = reflect.Append(resultValue, singleResult)
			haveFound = true
			rightHaveJoin[j] = true
		}
		//合并不满足的条件
		if !haveFound && (joinPlace == "left" || joinPlace == "outer") {
			singleRightData := reflect.New(rightDataElemType).Elem()
			singleResult := joinFuctorValue.Call([]reflect.Value{singleLeftData, singleRightData})[0]
			resultValue = reflect.Append(resultValue, singleResult)
		}
	}
	//处理剩余的右侧元素
	if joinPlace == "right" || joinPlace == "outer" {
		singleLeftData := reflect.New(leftDataElemType).Elem()
		rightHaveJoinLen := len(rightHaveJoin)
		for j := 0; j != rightHaveJoinLen; j++ {
			if rightHaveJoin[j] {
				continue
			}
			singleRightData := rightDataValue.Index(j)
			singleResult := joinFuctorValue.Call([]reflect.Value{singleLeftData, singleRightData})[0]
			resultValue = reflect.Append(resultValue, singleResult)
		}
	}
	return resultValue.Interface()
}

//基础类函数 QueryGroup
func QueryGroupInternal(length int, lessHandler func(i int, j int) int, swapHandler func(i int, j int), groupHandler func(i int, j int)) {
	QuerySortInternal(length, lessHandler, swapHandler)
	for i := 0; i != length; {
		j := i
		for i++; i != length; i++ {
			if lessHandler(j, i) != 0 {
				break
			}
		}
		groupHandler(j, i)
	}

}

type QueryGroupMacroHandler func(data interface{}, groupType string, groupFunctor interface{}) interface{}

func QueryGroupMacroRegister(data interface{}, groupType string, groupFunctor interface{}, handler QueryGroupMacroHandler) {
	id := registerQueryTypeId([]string{reflect.TypeOf(data).String(), groupType, reflect.TypeOf(groupFunctor).String()})
	queryGroupMacroMapper[id] = handler
}

func QueryGroupReflect(data interface{}, groupType string, groupFunctor interface{}) interface{} {
	//拷贝一份
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()
	dataElemType := dataType.Elem()
	dataValueLen := dataValue.Len()

	dataResult := reflect.MakeSlice(dataType, dataValueLen, dataValueLen)
	reflect.Copy(dataResult, dataValue)

	groupFuctorValue := reflect.ValueOf(groupFunctor)
	groupFuctorType := groupFuctorValue.Type()

	//计算最终数据
	var resultValue reflect.Value
	resultType := groupFuctorType.Out(0)
	if resultType.Kind() == reflect.Slice {
		resultValue = reflect.MakeSlice(resultType, 0, 0)
	} else {
		resultValue = reflect.MakeSlice(reflect.SliceOf(resultType), 0, 0)
	}

	//分组操作
	targetCompares := getQueryExtractAndCompares(dataElemType, groupType)
	targetCompare := combineQueryCompare(targetCompares)
	result := dataResult.Interface()
	swapper := reflect.Swapper(result)
	QueryGroupInternal(dataValueLen, func(i int, j int) int {
		left := dataResult.Index(i)
		right := dataResult.Index(j)
		return targetCompare(left, right)
	}, swapper, func(i int, j int) {
		singleResult := groupFuctorValue.Call([]reflect.Value{dataResult.Slice(i, j)})[0]
		if singleResult.Kind() == reflect.Slice {
			resultValue = reflect.AppendSlice(resultValue, singleResult)
		} else {
			resultValue = reflect.Append(resultValue, singleResult)
		}
	})
	return resultValue.Interface()
}

func QueryGroup(data interface{}, groupType string, groupFunctor interface{}) interface{} {
	id := getQueryTypeId([]string{reflect.TypeOf(data).String(), groupType, reflect.TypeOf(groupFunctor).String()})
	handler, isExist := queryGroupMacroMapper[id]
	if isExist {
		return handler(data, groupType, groupFunctor)
	} else {
		return QueryGroupReflect(data, groupType, groupFunctor)
	}
}

func analyseJoin(joinType string) (string, string) {
	joinTypeArray := strings.Split(joinType, "=")
	leftJoinType := strings.Trim(joinTypeArray[0], " ")
	rightJoinType := strings.Trim(joinTypeArray[1], " ")
	return leftJoinType, rightJoinType
}

func analyseSort(sortType string) (result1 []string, result2 []bool) {
	sortTypeArray := strings.Split(sortType, ",")
	for _, singleSortTypeArray := range sortTypeArray {
		singleSortTypeArrayTemp := strings.Split(singleSortTypeArray, " ")
		singleSortTypeArray := []string{}
		for _, singleSort := range singleSortTypeArrayTemp {
			singleSort = strings.Trim(singleSort, " ")
			if singleSort == "" {
				continue
			}
			singleSortTypeArray = append(singleSortTypeArray, singleSort)
		}
		var singleSortName string
		var singleSortType bool
		if len(singleSortTypeArray) >= 2 {
			singleSortName = singleSortTypeArray[0]
			singleSortType = (strings.ToLower(strings.Trim(singleSortTypeArray[1], " ")) == "asc")
		} else {
			singleSortName = singleSortTypeArray[0]
			singleSortType = true
		}
		result1 = append(result1, singleSortName)
		result2 = append(result2, singleSortType)
	}
	return result1, result2
}

func getQueryExtractAndCompares(dataType reflect.Type, sortTypeStr string) []queryCompare {
	sortName, sortType := analyseSort(sortTypeStr)
	targetCompare := []queryCompare{}
	for index, singleSortName := range sortName {
		singleSortType := sortType[index]
		singleCompare := getQueryExtractAndCompare(dataType, singleSortName)
		if !singleSortType {
			//逆序
			singleTempCompare := singleCompare
			singleCompare = func(left reflect.Value, right reflect.Value) int {
				return singleTempCompare(right, left)
			}
		}
		targetCompare = append(targetCompare, singleCompare)
	}
	return targetCompare
}

func getQueryCompare(fieldType reflect.Type) queryCompare {
	typeKind := GetTypeKind(fieldType)
	if typeKind == TypeKind.BOOL {
		return func(left reflect.Value, right reflect.Value) int {
			leftBool := left.Bool()
			rightBool := right.Bool()
			if leftBool == rightBool {
				return 0
			} else if leftBool == false {
				return -1
			} else {
				return 1
			}
		}
	} else if typeKind == TypeKind.INT {
		return func(left reflect.Value, right reflect.Value) int {
			leftInt := left.Int()
			rightInt := right.Int()
			if leftInt < rightInt {
				return -1
			} else if leftInt > rightInt {
				return 1
			} else {
				return 0
			}
		}
	} else if typeKind == TypeKind.UINT {
		return func(left reflect.Value, right reflect.Value) int {
			leftUint := left.Uint()
			rightUint := right.Uint()
			if leftUint < rightUint {
				return -1
			} else if leftUint > rightUint {
				return 1
			} else {
				return 0
			}
		}
	} else if typeKind == TypeKind.FLOAT {
		return func(left reflect.Value, right reflect.Value) int {
			leftFloat := left.Float()
			rightFloat := right.Float()
			if leftFloat < rightFloat {
				return -1
			} else if leftFloat > rightFloat {
				return 1
			} else {
				return 0
			}
		}
	} else if typeKind == TypeKind.STRING {
		return func(left reflect.Value, right reflect.Value) int {
			leftString := left.String()
			rightString := right.String()
			if leftString < rightString {
				return -1
			} else if leftString > rightString {
				return 1
			} else {
				return 0
			}
		}
	} else if typeKind == TypeKind.STRUCT && fieldType == reflect.TypeOf(time.Time{}) {
		return func(left reflect.Value, right reflect.Value) int {
			leftTime := left.Interface().(time.Time)
			rightTime := right.Interface().(time.Time)
			if leftTime.Before(rightTime) {
				return -1
			} else if leftTime.After(rightTime) {
				return 1
			} else {
				return 0
			}
		}
	} else {
		panic(fieldType.Name() + " can not compare")
	}
}

type queryCompare func(reflect.Value, reflect.Value) int

func combineQueryCompare(targetCompare []queryCompare) queryCompare {
	return func(left reflect.Value, right reflect.Value) int {
		for _, singleCompare := range targetCompare {
			compareResult := singleCompare(left, right)
			if compareResult < 0 {
				return -1
			} else if compareResult > 0 {
				return 1
			}
		}
		return 0
	}
}

type queryExtract func(reflect.Value) reflect.Value

func getQueryExtract(dataType reflect.Type, name string) (reflect.Type, queryExtract) {
	if name == "." {
		return dataType, func(v reflect.Value) reflect.Value {
			return v
		}
	} else {
		field, ok := getFieldByName(dataType, name)
		if !ok {
			panic(dataType.Name() + " has not name " + name)
		}
		fieldIndex := field.Index
		fieldType := field.Type
		return fieldType, func(v reflect.Value) reflect.Value {
			return v.FieldByIndex(fieldIndex)
		}
	}
}

func getQueryExtractAndCompare(dataType reflect.Type, name string) queryCompare {
	fieldType, extract := getQueryExtract(dataType, name)
	compare := getQueryCompare(fieldType)
	return func(left reflect.Value, right reflect.Value) int {
		return compare(extract(left), extract(right))
	}
}

//扩展类函数 QueryColumn
type QueryColumnMacroHandler func(data interface{}, column string) interface{}

func QueryColumnMacroRegister(data interface{}, column string, handler QueryColumnMacroHandler) {
	id := registerQueryTypeId([]string{reflect.TypeOf(data).String(), column})
	queryColumnMacroMapper[id] = handler
}

func QueryColumnReflect(data interface{}, column string) interface{} {
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type().Elem()
	dataLen := dataValue.Len()
	column = strings.Trim(column, " ")
	dataFieldType, dataFieldExtract := getQueryExtract(dataType, column)

	resultValue := reflect.MakeSlice(reflect.SliceOf(dataFieldType), dataLen, dataLen)

	for i := 0; i != dataLen; i++ {
		singleDataValue := dataValue.Index(i)
		singleResultValue := dataFieldExtract(singleDataValue)
		resultValue.Index(i).Set(singleResultValue)
	}
	return resultValue.Interface()
}

func QueryColumn(data interface{}, column string) interface{} {
	id := getQueryTypeId([]string{reflect.TypeOf(data).String(), column})
	handler, isExist := queryColumnMacroMapper[id]
	if isExist {
		return handler(data, column)
	} else {
		return QueryColumnReflect(data, column)
	}
}

//扩展类函数 QueryColumnMap
type QueryColumnMapMacroHandler func(data interface{}, column string) interface{}

func QueryColumnMapMacroRegister(data interface{}, column string, handler QueryColumnMapMacroHandler) {
	id := registerQueryTypeId([]string{reflect.TypeOf(data).String(), column})
	queryColumnMapMacroMapper[id] = handler
}

func QueryColumnMapReflect(data interface{}, column string) interface{} {
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type().Elem()
	dataLen := dataValue.Len()
	column = strings.Trim(column, " ")
	dataFieldType, dataFieldExtract := getQueryExtract(dataType, column)

	resultValue := reflect.MakeMap(reflect.MapOf(dataFieldType, dataType))
	for i := dataLen - 1; i >= 0; i-- {
		singleDataValue := dataValue.Index(i)
		singleResultValue := dataFieldExtract(singleDataValue)
		resultValue.SetMapIndex(singleResultValue, singleDataValue)
	}
	return resultValue.Interface()
}

func QueryColumnMap(data interface{}, column string) interface{} {
	id := getQueryTypeId([]string{reflect.TypeOf(data).String(), column})
	handler, isExist := queryColumnMapMacroMapper[id]
	if isExist {
		return handler(data, column)
	} else {
		return QueryColumnMapReflect(data, column)
	}
}

//扩展类函数
func QueryLeftJoin(leftData interface{}, rightData interface{}, joinType string, joinFuctor interface{}) interface{} {
	return QueryJoin(leftData, rightData, "left", joinType, joinFuctor)
}

func QueryRightJoin(leftData interface{}, rightData interface{}, joinType string, joinFuctor interface{}) interface{} {
	return QueryJoin(leftData, rightData, "right", joinType, joinFuctor)
}

func QueryInnerJoin(leftData interface{}, rightData interface{}, joinType string, joinFuctor interface{}) interface{} {
	return QueryJoin(leftData, rightData, "inner", joinType, joinFuctor)
}

func QueryOuterJoin(leftData interface{}, rightData interface{}, joinType string, joinFuctor interface{}) interface{} {
	return QueryJoin(leftData, rightData, "outer", joinType, joinFuctor)
}

func QueryReduce(data interface{}, reduceFuctor interface{}, resultReduce interface{}) interface{} {
	dataValue := reflect.ValueOf(data)
	dataLen := dataValue.Len()

	reduceFuctorValue := reflect.ValueOf(reduceFuctor)
	resultReduceType := reduceFuctorValue.Type().In(0)
	resultReduceValue := reflect.New(resultReduceType)
	err := MapToArray(resultReduce, resultReduceValue.Interface(), "json")
	if err != nil {
		panic(err)
	}
	resultReduceValue = resultReduceValue.Elem()

	for i := 0; i != dataLen; i++ {
		singleDataValue := dataValue.Index(i)
		resultReduceValue = reduceFuctorValue.Call([]reflect.Value{resultReduceValue, singleDataValue})[0]
	}
	return resultReduceValue.Interface()
}

func QuerySum(data interface{}) interface{} {
	dataType := reflect.TypeOf(data).Elem()
	if dataType.Kind() == reflect.Int {
		return QueryReduce(data, func(sum int, single int) int {
			return sum + single
		}, 0)
	} else if dataType.Kind() == reflect.Float32 {
		return QueryReduce(data, func(sum float32, single float32) float32 {
			return sum + single
		}, (float32)(0.0))
	} else if dataType.Kind() == reflect.Float64 {
		return QueryReduce(data, func(sum float64, single float64) float64 {
			return sum + single
		}, 0.0)
	} else {
		panic("invalid type " + dataType.String())
	}
}

func QueryMax(data interface{}) interface{} {
	dataType := reflect.TypeOf(data).Elem()
	if dataType.Kind() == reflect.Int {
		return QueryReduce(data, func(max int, single int) int {
			if single > max {
				return single
			} else {
				return max
			}
		}, math.MinInt32)
	} else if dataType.Kind() == reflect.Float32 {
		return QueryReduce(data, func(max float32, single float32) float32 {
			if single > max {
				return single
			} else {
				return max
			}
		}, math.SmallestNonzeroFloat32)
	} else if dataType.Kind() == reflect.Float64 {
		return QueryReduce(data, func(max float64, single float64) float64 {
			if single > max {
				return single
			} else {
				return max
			}
		}, math.SmallestNonzeroFloat64)
	} else {
		panic("invalid type " + dataType.String())
	}
}

func QueryMin(data interface{}) interface{} {
	dataType := reflect.TypeOf(data).Elem()
	if dataType.Kind() == reflect.Int {
		return QueryReduce(data, func(min int, single int) int {
			if single < min {
				return single
			} else {
				return min
			}
		}, math.MaxInt32)
	} else if dataType.Kind() == reflect.Float32 {
		return QueryReduce(data, func(min float32, single float32) float32 {
			if single < min {
				return single
			} else {
				return min
			}
		}, math.MaxFloat32)
	} else if dataType.Kind() == reflect.Float64 {
		return QueryReduce(data, func(min float64, single float64) float64 {
			if single < min {
				return single
			} else {
				return min
			}
		}, math.MaxFloat64)
	} else {
		panic("invalid type " + dataType.String())
	}
}

func QueryReverse(data interface{}) interface{} {
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()
	dataLen := dataValue.Len()
	result := reflect.MakeSlice(dataType, dataLen, dataLen)

	for i := 0; i != dataLen; i++ {
		result.Index(dataLen - i - 1).Set(dataValue.Index(i))
	}
	return result.Interface()
}

func QueryCombine(leftData interface{}, rightData interface{}, combineFuctor interface{}) interface{} {
	leftValue := reflect.ValueOf(leftData)
	rightValue := reflect.ValueOf(rightData)
	if leftValue.Len() != rightValue.Len() {
		panic(fmt.Sprintf("len dos not equal %v != %v", leftValue.Len(), rightValue.Len()))
	}
	dataLen := leftValue.Len()
	combineFuctorValue := reflect.ValueOf(combineFuctor)
	resultType := combineFuctorValue.Type().Out(0)
	result := reflect.MakeSlice(reflect.SliceOf(resultType), dataLen, dataLen)
	for i := 0; i != dataLen; i++ {
		singleResultValue := combineFuctorValue.Call([]reflect.Value{leftValue.Index(i), rightValue.Index(i)})
		result.Index(i).Set(singleResultValue[0])
	}
	return result.Interface()
}

func QueryDistinct(data interface{}, columnNames string) interface{} {
	//提取信息
	name := Explode(columnNames, ",")
	extractInfo := []queryExtract{}
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type().Elem()
	for _, singleName := range name {
		_, extract := getQueryExtract(dataType, singleName)
		extractInfo = append(extractInfo, extract)
	}

	//整合map
	existsMap := map[interface{}]bool{}
	result := reflect.MakeSlice(dataValue.Type(), 0, 0)
	dataLen := dataValue.Len()
	for i := 0; i != dataLen; i++ {
		singleValue := dataValue.Index(i)
		newData := reflect.New(dataType).Elem()
		for _, singleExtract := range extractInfo {
			singleField := singleExtract(singleValue)
			singleExtract(newData).Set(singleField)
		}
		newDataValue := newData.Interface()
		_, isExist := existsMap[newDataValue]
		if isExist {
			continue
		}
		result = reflect.Append(result, singleValue)
		existsMap[newDataValue] = true
	}
	return result.Interface()
}

func registerQueryTypeId(data []string) int64 {
	var result int64
	for _, m := range data {
		id, isExist := queryTypeIdMapper[m]
		if isExist == false {
			id = int64(len(queryTypeIdMapper)) + 1
			queryTypeIdMapper[m] = id
		}
		result = result<<10 + id
	}
	return result
}

func getQueryTypeId(data []string) int64 {
	var result int64
	for _, m := range data {
		id, isExist := queryTypeIdMapper[m]
		if isExist == false {
			return -1
		}
		result = result<<10 + id
	}
	return result
}

var (
	queryColumnMacroMapper    = map[int64]QueryColumnMacroHandler{}
	querySelectMacroMapper    = map[int64]QuerySelectMacroHandler{}
	queryWhereMacroMapper     = map[int64]QueryWhereMacroHandler{}
	querySortMacroMapper      = map[int64]QuerySortMacroHandler{}
	queryGroupMacroMapper     = map[int64]QueryGroupMacroHandler{}
	queryColumnMapMacroMapper = map[int64]QueryColumnMapMacroHandler{}
	queryTypeIdMapper         = map[string]int64{}
)
