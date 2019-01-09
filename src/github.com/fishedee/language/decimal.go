package language

import (
	"github.com/shopspring/decimal"
)

/*
* 为什么不直接用decimal包，因为decimal包原始格式无法序列化，转储只能用string，才能在数据库和前端进行无损的传递，常见的javascript就处理不了[64位长整数](https://www.zhihu.com/question/34564427)
* 详情看[这里](https://fishedee.com/2016/03/16/mysql%E7%BB%8F%E9%AA%8C%E6%B1%87%E6%80%BB/)的浮点数章节
* 本目录下的math.go的实现是错误的，将decimal包直接用float64转储会[爆炸](https://github.com/fishedee/Demo/blob/master/go/decimal/main.go)
 */
type Decimal string

func NewDecimal(in string) (Decimal, error) {
	_, err := decimal.NewFromString(in)
	if err != nil {
		return "", err
	}
	return Decimal(in), nil
}

func getDecimal(a Decimal) decimal.Decimal {
	if string(a) == "" {
		return decimal.Decimal{}
	}
	r, err := decimal.NewFromString(string(a))
	if err != nil {
		panic(err)
	}
	return r
}

func (left Decimal) Add(right Decimal) Decimal {
	l := getDecimal(left)
	r := getDecimal(right)
	return Decimal(l.Add(r).String())
}

func (left Decimal) Sub(right Decimal) Decimal {
	l := getDecimal(left)
	r := getDecimal(right)
	return Decimal(l.Sub(r).String())
}

func (left Decimal) Mul(right Decimal) Decimal {
	l := getDecimal(left)
	r := getDecimal(right)
	return Decimal(l.Mul(r).String())
}

func (left Decimal) Div(right Decimal) Decimal {
	l := getDecimal(left)
	r := getDecimal(right)
	return Decimal(l.Div(r).String())
}

func (left Decimal) Round(precision int) Decimal {
	l := getDecimal(left)
	return Decimal(l.Round(int32(precision)).String())
}

func (left Decimal) Cmp(right Decimal) int {
	l := getDecimal(left)
	r := getDecimal(right)
	return l.Cmp(r)
}

func (left Decimal) Equal(right Decimal) bool {
	l := getDecimal(left)
	r := getDecimal(right)
	return l.Equal(r)
}

func (left Decimal) Sign() int {
	l := getDecimal(left)
	return l.Sign()
}

func (left Decimal) Abs() Decimal {
	l := getDecimal(left)
	return Decimal(l.Abs().String())
}

func (left Decimal) String() string {
	l := getDecimal(left)
	return l.String()
}
