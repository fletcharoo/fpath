package runtime

import (
	"fmt"

	"github.com/fletcharoo/fpath/internal/parser"
)

type evalFunc func(parser.Expr, any) (any, error)

var evalMap map[int]evalFunc

func init() {
	evalMap = map[int]evalFunc{
		parser.ExprType_Undefined: evalUndefined,
		parser.ExprType_Number:    evalNumber,
	}
}

// Eval accepts a parsed expression and the query's input data and returns the
// evaluated result
func Eval(expr parser.Expr, input any) (result any, err error) {
	f, ok := evalMap[expr.Type()]
	if !ok {
		return evalUndefined(nil, nil)
	}

	return f(expr, input)
}

func evalUndefined(_ parser.Expr, _ any) (ret any, err error) {
	err = fmt.Errorf("failed to eval undefined expression")
	return
}

func evalNumber(ast parser.Expr, _ any) (ret any, err error) {
	exprNumber := ast.(parser.ExprNumber)
	return exprNumber.Value, nil
}
