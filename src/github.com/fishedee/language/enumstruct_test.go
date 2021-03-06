package language_test

import (
	. "github.com/fishedee/language"
	"testing"
)

func getEnumStructError(handler func()) (lastErr string) {
	defer func() {
		err := recover()
		if err != nil {
			lastErr = err.(string)
		}
	}()
	handler()
	return ""
}

func TestEnumStruct(t *testing.T) {
	var testCase struct {
		EnumStruct
		ENUM1 int `enum:"1,枚举1"`
		ENUM2 int `enum:"2,枚举2"`
		ENUM3 int `enum:"3,枚举3"`
	}

	InitEnumStruct(&testCase)

	//断言基本枚举值
	AssertEqual(t, testCase.ENUM1, 1)
	AssertEqual(t, testCase.ENUM2, 2)
	AssertEqual(t, testCase.ENUM3, 3)

	//断言函数
	AssertEqual(t, testCase.Names(), map[string]string{
		"1": "枚举1",
		"2": "枚举2",
		"3": "枚举3",
	})
	AssertEqual(t, testCase.Entrys(), map[int]string{
		1: "枚举1",
		2: "枚举2",
		3: "枚举3",
	})
	AssertEqual(t, testCase.Datas(), []EnumData{
		{1, "枚举1"},
		{2, "枚举2"},
		{3, "枚举3"},
	})
	AssertEqual(t, ArraySort(testCase.Keys()), []int{1, 2, 3})
	AssertEqual(t, ArraySort(testCase.Values()), []string{"枚举1", "枚举2", "枚举3"})

	//错误的enum
	var err string
	var realErr = "invalid enum struct"
	var testCaseErr1 struct {
		EnumStruct
		ENUM1 int `enum:1,枚举1`
		ENUM2 int `enum:"2,枚举2"`
		ENUM3 int `enum:"3,枚举3"`
	}
	err = getEnumStructError(func() {
		InitEnumStruct(&testCaseErr1)
	})
	AssertEqual(t, err[0:len(realErr)], realErr)

	var testCaseErr2 struct {
		EnumStruct
		ENUM1 int `enum:"枚举1"`
		ENUM2 int `enum:"2,枚举2"`
		ENUM3 int `enum:"3,枚举3"`
	}
	err = getEnumStructError(func() {
		InitEnumStruct(&testCaseErr2)
	})
	AssertEqual(t, err[0:len(realErr)], realErr)

	var testCaseErr3 struct {
		EnumStruct
		ENUM1 int `enum:"1"`
		ENUM2 int `enum:"2,枚举2"`
		ENUM3 int `enum:"3,枚举3"`
	}
	err = getEnumStructError(func() {
		InitEnumStruct(&testCaseErr3)
	})
	AssertEqual(t, err[0:len(realErr)], realErr)

	var testCaseErr4 struct {
		EnumStruct
		ENUM1 int `enum:"1z,枚举1"`
		ENUM2 int `enum:"2,枚举2"`
		ENUM3 int `enum:"3,枚举3"`
	}
	err = getEnumStructError(func() {
		InitEnumStruct(&testCaseErr4)
	})
	AssertEqual(t, err[0:len(realErr)], realErr)
}

func TestEnumStructString(t *testing.T) {
	var testCase struct {
		EnumStructString
		ENUM1 string `enum:"/content/del1,枚举1"`
		ENUM2 string `enum:"/content/del2,枚举2"`
		ENUM3 string `enum:"/content/del3,枚举3"`
	}

	InitEnumStructString(&testCase)

	//断言基本枚举值
	AssertEqual(t, testCase.ENUM1, "/content/del1")
	AssertEqual(t, testCase.ENUM2, "/content/del2")
	AssertEqual(t, testCase.ENUM3, "/content/del3")

	//断言函数
	AssertEqual(t, testCase.Names(), map[string]string{
		"/content/del1": "枚举1",
		"/content/del2": "枚举2",
		"/content/del3": "枚举3",
	})
	AssertEqual(t, testCase.Entrys(), map[string]string{
		"/content/del1": "枚举1",
		"/content/del2": "枚举2",
		"/content/del3": "枚举3",
	})
	AssertEqual(t, testCase.Datas(), []EnumDataString{
		{"/content/del1", "枚举1"},
		{"/content/del2", "枚举2"},
		{"/content/del3", "枚举3"},
	})
	AssertEqual(t, ArraySort(testCase.Keys()), []string{"/content/del1", "/content/del2", "/content/del3"})
	AssertEqual(t, ArraySort(testCase.Values()), []string{"枚举1", "枚举2", "枚举3"})

	//错误的enum
	var err string
	var realErr = "invalid enum struct"
	var testCaseErr1 struct {
		EnumStructString
		ENUM1 string `enum:/content/del1,枚举1`
		ENUM2 string `enum:"/content/del2,枚举2"`
		ENUM3 string `enum:"/content/del3,枚举3"`
	}
	err = getEnumStructError(func() {
		InitEnumStructString(&testCaseErr1)
	})
	AssertEqual(t, err[0:len(realErr)], realErr)

	var testCaseErr2 struct {
		EnumStructString
		ENUM1 string `enum:"/content/del1"`
		ENUM2 string `enum:"/content/del2,枚举2"`
		ENUM3 string `enum:"/content/del3,枚举3"`
	}
	err = getEnumStructError(func() {
		InitEnumStructString(&testCaseErr2)
	})
	AssertEqual(t, err[0:len(realErr)], realErr)

	var testCaseErr3 struct {
		EnumStructString
		ENUM1 string `enum:"枚举1"`
		ENUM2 string `enum:"/content/del2,枚举2"`
		ENUM3 string `enum:"/content/del3,枚举3"`
	}
	err = getEnumStructError(func() {
		InitEnumStructString(&testCaseErr3)
	})
	AssertEqual(t, err[0:len(realErr)], realErr)
}
