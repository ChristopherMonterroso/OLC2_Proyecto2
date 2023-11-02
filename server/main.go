package main

import (
	"Server/generator"
	"Server/parser"
	parsermatrix "Server/parserMatrix"
	"Server/statements"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var Generator generator.Generator
var Cont_Stack = 0
var Cont_Heap = 0
var HeapPointer = 0
var LexerErrors []LexerError
var SyntaxErrors []SyntaxErrs
var SemanticErrors []SemantcErrs
var Symbols []ArraySimbols

type Environment struct {
	parent    *Environment
	variables map[string]Variable
	functions map[string]funcion
	size_var  map[string]int
}
type ArraySimbols struct {
	Id         string
	TipoSymbol string
	Tipo       string
	Ambit      string
	Line       int
	Column     int
}

type LexerError struct {
	*antlr.DefaultErrorListener
	ErrorMessage string
	Line         string
	Column       string
}
type SyntaxErrs struct {
	*antlr.DefaultErrorListener
	ErrorMessage string
	Line         string
	Column       string
}
type SemantcErrs struct {
	ErrorMessage string
	Line         string
	Column       string
}
type Value struct {
	value      interface{}
	Tipo       string
	Tipo2      string
	temp       interface{}
	TrueLabel  []string
	FalseLabel []string
}
type Variable struct {
	scope         string
	Ambit         string
	PrimitiveType string
	value         interface{}
	isVector      bool
	isMatrix      bool
	Line          string
	Column        string
	Stack         int
	Heap          int
}
type param struct {
	IDExterno     string
	IDInterno     string
	Inout         bool
	idVarToInout  string
	primitiveType string
	isVector      bool
	value         interface{}
}
type funcion struct {
	primitiveType string
	params        []param
	value         antlr.ParseTree
	hasReturn     bool
	exprToRetunr  antlr.ParseTree
}

type Visitor struct {
	currentAmbit string
	currentEnv   *Environment
	parser.ControlVisitor
	memory         map[string]Value
	exprToSwitch   interface{}
	exitSwitch     bool
	outputBuilder  strings.Builder
	SemanticErrors strings.Builder
	continueFlag   bool
	breakFlag      bool
	returnFlag     bool
	returnExpr     interface{}
	tmpList        []interface{}
	tmpListParams  []param
	tmpMtrx2       bool
	tmpMtrx3       bool
	tmpListaMtrx   [][]interface{}
	tmp_xprSwitch  string
	activate_case  bool
}

func AgregarErrorSemantico(mensaje string, lines string, columna string) {
	newError := SemantcErrs{
		ErrorMessage: mensaje,
		Line:         lines,
		Column:       columna,
	}
	SemanticErrors = append(SemanticErrors, newError)

}
func AgregarSimbolo(id string, tipoSimbolo string, tipo string, ambit string, line int, column int) {
	newSymbol := ArraySimbols{
		Id:         id,
		TipoSymbol: tipo,
		Tipo:       tipoSimbolo,
		Ambit:      ambit,
		Line:       line,
		Column:     column,
	}
	Symbols = append(Symbols, newSymbol)

}
func (e *LexerError) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, lineE int, columnE int, message string, er antlr.RecognitionException) {
	msg := fmt.Sprintf("Lexer error at linea %d, columna: %d %s", lineE, columnE, message)
	fmt.Println(msg)

	newError := LexerError{
		ErrorMessage: message,
		Line:         strconv.Itoa(lineE),
		Column:       strconv.Itoa(columnE),
	}
	LexerErrors = append(LexerErrors, newError)
}
func (e *SyntaxErrs) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, lineE int, columnE int, message string, er antlr.RecognitionException) {

	msg := fmt.Sprintf("Syntax error at: linea %d, columna: %d %s", lineE, columnE, message)
	fmt.Println(msg)
	newError := SyntaxErrs{
		ErrorMessage: "Syntax error: " + message,
		Line:         strconv.Itoa(lineE),
		Column:       strconv.Itoa(columnE),
	}
	SyntaxErrors = append(SyntaxErrors, newError)
}

func (v *Visitor) Visit(tree antlr.ParseTree) Value {
	switch val := tree.(type) {
	case *parser.ProgContext:
		return v.VisitProg(val)
	case *parser.BlockContext:
		return v.VisitBlock(val)
	case *parser.StmtContext:
		return v.VisitStmt(val)
	case *parser.ReturnstmtContext:
		return v.VisitReturn(val)
	case *parser.AsignacionContext:
		return v.VisitAsignacion(val)
	case *parser.AsignacionNoExprContext:
		return v.VisitAsignacionNoExpr(val)
	case *parser.AsignacionNoPrimitiveContext:
		return v.VisitAsignacionNoPrimitive(val)
	case *parser.ReasignacionContext:
		return v.VisitReasignacion(val)
	case *parser.IncrementoContext:
		return v.visitIncremento(val)
	case *parser.DecrementoContext:
		return v.VisitDecremento(val)
	case *parser.IfNormalContext:
		return v.VisitIfNormal(val)
	case *parser.ElseContext:
		return v.VisitIfElse(val)
	case *parser.Else_ifContext:
		return v.VisitElse_If(val)

	case *parser.AsignacionVectorVacioContext:
		return v.VisitAsignacionVectorVacio(val)
	case *parser.AsignacionVectorContext:
		return v.VisitAsignacionVector(val)
	case *parser.OneExprContext:
		return v.VisitOneExpr(val)
	case *parser.NumExprContext:
		return v.VisitNumExpr(val)
	case *parser.VectorAppendContext:
		return v.VisitVectorAppend(val)
	case *parser.VectorRemoveLastContext:
		return v.VisitVectorRemoveLast(val)
	case *parser.VectorCountContext:
		return v.visitVectorCount(val)
	case *parser.VectorIsEmptyContext:
		return v.VisitVectorIsEmpty(val)
	case *parser.VectorGetElementContext:
		return v.VisitVectorGetElement(val)
	case *parser.VectorRemoveAtContext:
		return v.VisitRemoveAt(val)
	case *parser.ReasignacionVectorContext:
		return v.VisitReasignacionVector(val)
	case *parser.ReasignacionMatrixTwoDContext:
		return v.VisitReasignacionMatrixTwoD(val)

	case *parser.SwitchstmtContext:
		return v.VisitSwitchstmt(val)
	case *parser.CasesContext:
		return v.VisitCases(val)
	case *parser.UnCaseContext:
		return v.VisitUnCase(val)
	case *parser.UnDefaultContext:
		return v.VisitUnDefault(val)

	case *parser.WhilestmtContext:
		return v.VisitWhilestmt(val)
	case *parser.AccesoMatrixTwoDContext:
		return v.VisitAccesoMatrixTwoD(val)
	case *parser.MatrixTwoDContext:
		return v.VisitMatrixTwoD(val)
	case *parser.MatrixThreeDContext:
		return v.VisitMatrixThreeD(val)
	case *parser.DefMatrixContext:
		return v.VisitDefMatrix(val)
	case *parser.ListaValores_MatContext:
		return v.VisitListaValores_Mat(val)
	case *parser.ListaValores_Mat2Context:
		return v.VisitListaValores_Mat2(val)
	case *parser.Lista_ExpresionesContext:
		return v.VisitListaExpresiones(val)
	case *parser.ForNormalContext:
		return v.VisitForNormal(val)
	case *parser.ForEachContext:
		return v.VisitForEach(val)

	case *parser.GuardstmtContext:
		return v.VisitGuardstmt(val)

	case *parser.PrintstmtContext:
		return v.VisitPrintstmt(val)

	case *parser.Func_sinRetornoContext:
		return v.VisitFunc_sinRetorno(val)
	case *parser.CallFunctionContext:
		return v.VisitCallFunction(val)
	case *parser.ListaParamsCallContext:
		return v.VisitListaParamsCall(val)
	case *parser.Func_conRetorno_conTipoContext:
		return v.Visitfunc_conRetorno_conTipo(val)
	case *parser.CallFuncAsExprContext:
		return v.VisitcallFuncAsExpr(val)

	case *parser.OneParamContext:
		return v.VisitOneParam(val)
	case *parser.NumParamsContext:
		return v.VisitNumParams(val)
	case *parser.OpExprContext:
		return v.VisitOpExpr(val)
	case *parser.NilExprContext:
		return v.VisitNilExpr(val)
	case *parser.IntExprContext:
		return v.VisitIntExpr(val)
	case *parser.FloatExprContext:
		return v.VisitFloatExpr(val)
	case *parser.IdExprContext:
		return v.VisitIdExpr(val)
	case *parser.StrExprContext:
		return v.VisitStrExpr(val)
	case *parser.CharExprContext:
		return v.VisitCharExpr(val)
	case *parser.BoolExprContext:
		return v.VisitBoolExpr(val)
	case *parser.Var_typeContext:
		return v.VisitVar_type(val)
	case *parser.PrimitivoContext:
		return v.VisitPrimitivo(val)
	case *parser.Transfer_sentenceContext:
		return v.VisitTransferSentence(val)
	case *parser.ToStringContext:
		return v.VisitToString(val)
	case *parser.ToIntContext:
		return v.VisitToInt(val)
	case *parser.ToFloatContext:
		return v.VisitToFloat(val)
	case *parser.ParExprContext:
		return v.VisitParExpr(val)
	case *parser.NotExprContext:
		return v.VisitNotExpr(val)
	case *parser.NegExprContext:
		return v.VisitNegExpr(val)
	default:
		return Value{value: nil}
	}

}

func (v *Visitor) VisitProg(ctx *parser.ProgContext) Value {
	v.currentAmbit = "global"
	localEnv := &Environment{
		parent:    v.currentEnv,
		variables: make(map[string]Variable),
		functions: make(map[string]funcion),
		size_var:  make(map[string]int),
	}
	localEnv.size_var["size"] = 0
	previousEnv := v.currentEnv
	v.currentEnv = localEnv
	defer func() {
		v.currentEnv = previousEnv
	}()
	return v.Visit(ctx.Block())
}

func (v *Visitor) VisitBlock(ctx *parser.BlockContext) Value {
	for i := 0; ctx.Stmt(i) != nil; i++ {

		v.Visit(ctx.Stmt(i))
		if v.continueFlag || v.breakFlag {
			v.continueFlag = false

			break
		}
		if v.returnFlag {
			break
		}

	}
	v.currentAmbit = "global"
	return Value{value: true}
}

func (v *Visitor) VisitStmt(ctx *parser.StmtContext) Value {
	if ctx.Assignstmt() != nil {

		return v.Visit(ctx.Assignstmt())
	}
	if ctx.Ifstmt() != nil {
		return v.Visit(ctx.Ifstmt())
	}
	if ctx.Printlnstmt() != nil {
		return v.Visit(ctx.Printlnstmt())
	}
	if ctx.Printstmt() != nil {
		return v.Visit(ctx.Printstmt())
	}
	if ctx.Whilestmt() != nil {
		return v.Visit(ctx.Whilestmt())
	}
	if ctx.Switchstmt() != nil {
		return v.Visit(ctx.Switchstmt())
	}
	if ctx.Forstmt() != nil {
		return v.Visit(ctx.Forstmt())
	}
	if ctx.Guardstmt() != nil {

		return v.Visit(ctx.Guardstmt())
	}
	if ctx.VectorPpts() != nil {
		return v.Visit(ctx.VectorPpts())
	}
	if ctx.Matrixstmt() != nil {
		return v.Visit(ctx.Matrixstmt())
	}
	if ctx.Funcstmt() != nil {
		return v.Visit(ctx.Funcstmt())
	}
	if ctx.CallFunction() != nil {
		return v.Visit(ctx.CallFunction())
	}
	if ctx.Returnstmt() != nil {
		return v.Visit(ctx.Returnstmt())
	}
	return Value{value: true}
}

// Funciones de asignacion
func (v *Visitor) VisitAsignacion(ctx *parser.AsignacionContext) Value {
	scope := v.Visit(ctx.Var_type()).value //Local o global
	varID := ctx.ID().GetText()
	primitivetype := v.Visit(ctx.Primitivo())
	tmp_heap := Cont_Heap

	val := v.Visit(ctx.Expr())
	if v.DoesVarExist(varID) {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable "+varID+" ya existe"))
		AgregarErrorSemantico("La variable "+varID+" ya existe", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	} else {
		if primitivetype.value.(string) == getType(val.value) { // Confirmar que el tipo primitivo asignado sea igual al tipo primitivo de la expresión

			newVar := Variable{
				scope:         scope.(string),
				PrimitiveType: primitivetype.value.(string),
				value:         val.value,
				Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
				Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
				Ambit:         v.currentAmbit,
				Stack:         Cont_Stack,
				Heap:          tmp_heap,
			}
			AgregarSimbolo(varID, "Variable", primitivetype.value.(string), v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())
			v.currentEnv.variables[varID] = newVar
			valueVar := fmt.Sprintf("%v", val.value)
			v.currentEnv.size_var["size"] = v.currentEnv.size_var["size"] + 1
			if primitivetype.value.(string) == "String" {
				Generator.AddComment("Agregando una declaración")
				heap := fmt.Sprintf("%v", newVar.Heap)
				Generator.AddSetStack(strconv.Itoa(newVar.Stack), heap)
				Cont_Stack++
				return Value{value: true, temp: valueVar, Tipo: primitivetype.value.(string), Tipo2: "String"}
			} else if primitivetype.value.(string) == "Bool" {
				Generator.AddComment("Agregando una declaración")
				if valueVar == "true" {
					valueVar = "1"
				} else {
					valueVar = "0"
				}
				newLabel := Generator.NewLabel()
				for _, lvl := range val.TrueLabel {
					Generator.AddLabel(lvl)
				}
				Generator.AddSetStack(strconv.Itoa(newVar.Stack), "1")
				Generator.AddGoto(newLabel)
				for _, lvl := range val.FalseLabel {
					Generator.AddLabel(lvl)
				}
				Generator.AddSetStack(strconv.Itoa(newVar.Stack), "0")
				Generator.AddGoto(newLabel)
				Generator.AddLabel(newLabel)
				Cont_Stack++
				return Value{value: true, temp: valueVar, Tipo: primitivetype.value.(string), Tipo2: "Bool"}

			}

			valueVar = fmt.Sprintf("%v", val.value)
			Generator.AddSetStack(strconv.Itoa(newVar.Stack), valueVar)
			Cont_Stack++
			Generator.AddBr()
			return Value{value: true}
		} else {
			AgregarErrorSemantico("No coincide la variable "+varID+" con el tipo primitivo"+primitivetype.value.(string), strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "No coincide la variable "+varID+" con el tipo primitivo"+primitivetype.value.(string)))
			return Value{value: false}
		}
	}

}

func (v *Visitor) VisitAsignacionNoExpr(ctx *parser.AsignacionNoExprContext) Value {
	scope := v.Visit(ctx.Var_type()).value //Local o global
	varID := ctx.ID().GetText()
	primitivetype := v.Visit(ctx.Primitivo())

	v.currentEnv.variables[varID] = Variable{
		scope:         scope.(string),
		PrimitiveType: primitivetype.value.(string),
		value:         nil,
		Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
		Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
		Ambit:         v.currentAmbit,
	}

	AgregarSimbolo(varID, "Variable", primitivetype.value.(string), v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())
	if primitivetype.value.(string) == "Bool" {
		Generator.AddComment("Agregando una declaración")

	}
	return Value{value: true}

}

func (v *Visitor) VisitAsignacionNoPrimitive(ctx *parser.AsignacionNoPrimitiveContext) Value {
	scope := v.Visit(ctx.Var_type()).value //Local o global
	varID := ctx.ID().GetText()
	value := v.Visit(ctx.Expr())
	v.currentEnv.variables[varID] = Variable{
		scope:         scope.(string),
		PrimitiveType: getType(value.value),
		value:         value.value,
		Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
		Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
		Ambit:         v.currentAmbit,
	}
	AgregarSimbolo(varID, "Variable", getType(value.value), v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

	return Value{value: true}

}

func (v *Visitor) VisitReasignacion(ctx *parser.ReasignacionContext) Value {
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var currentValue interface{}
	tmp_heap := Cont_Heap

	valueToAssign := v.Visit(ctx.Expr())
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			scope = variable.scope
			primitivetype = variable.PrimitiveType
			currentValue = variable.value

			if scope == "let" {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Variable tipo let, no puede ser modificada"+varID))
				AgregarErrorSemantico("Variable tipo let, no puede ser modificada"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				if primitivetype == getType(valueToAssign.value) || currentValue == nil {
					// Realiza la reasignación en el ámbito actual

					tipo := current.variables[varID].PrimitiveType
					if tipo == "Int" || tipo == "Float" {
						Generator.AddComment("Reasignando una variable")
						valueVar := fmt.Sprintf("%v", valueToAssign.temp)
						Generator.AddSetStack(strconv.Itoa(current.variables[varID].Stack), valueVar)
						Generator.AddBr()
					}
					if tipo == "Bool" {
						Generator.AddComment("Reasignando una variable")
						valueVar := fmt.Sprintf("%v", valueToAssign.value)
						if valueVar == "true" {
							valueVar = "1"
						} else {
							valueVar = "0"
						}

						newLabel := Generator.NewLabel()
						for _, lvl := range valueToAssign.TrueLabel {
							Generator.AddLabel(lvl)
						}
						Generator.AddSetStack(strconv.Itoa(current.variables[varID].Stack), "1")
						Generator.AddGoto(newLabel)
						for _, lvl := range valueToAssign.FalseLabel {
							Generator.AddLabel(lvl)
						}
						Generator.AddSetStack(strconv.Itoa(current.variables[varID].Stack), "0")
						Generator.AddGoto(newLabel)
						Generator.AddLabel(newLabel)

					}
					if tipo == "String" {
						Generator.AddComment("Reasignando una variable")
						tmp_stack := Cont_Stack
						heap := fmt.Sprintf("%v", tmp_heap)
						Generator.AddSetStack(strconv.Itoa(tmp_stack), heap)

						current.variables[varID] = Variable{
							scope:         scope.(string),
							Ambit:         current.variables[varID].Ambit,
							PrimitiveType: primitivetype.(string),
							value:         valueToAssign.value,
							isVector:      false,
							Line:          current.variables[varID].Line,
							Column:        current.variables[varID].Column,
							Stack:         Cont_Stack,
							Heap:          tmp_heap,
						}
						return Value{value: true}

					}
					current.variables[varID] = Variable{
						scope:         scope.(string),
						Ambit:         current.variables[varID].Ambit,
						PrimitiveType: primitivetype.(string),
						value:         valueToAssign.value,
						isVector:      false,
						Line:          current.variables[varID].Line,
						Column:        current.variables[varID].Column,
						Stack:         current.variables[varID].Stack,
						Heap:          current.variables[varID].Heap,
					}
					return Value{value: true}
				} else {
					v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "No coincide la variable "+varID+" con el tipo primitivo"+primitivetype.(string)))
					AgregarErrorSemantico("No coincide la variable "+varID+" con el tipo primitivo"+primitivetype.(string), strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
					return Value{value: false}
				}
			}
		}
		current = current.parent
	}

	return Value{value: true}
}

func (v *Visitor) VisitAsignacionVectorVacio(ctx *parser.AsignacionVectorVacioContext) Value {
	varID := ctx.ID().GetText()
	primitivetype := v.Visit(ctx.Primitivo())

	// Crear un slice vacío basado en el tipo primitivo
	var emptySlice interface{}
	switch primitivetype.value.(string) {
	case "Int":
		emptySlice = []int64{}
	case "Float":
		emptySlice = []float64{}
	case "String":
		emptySlice = []string{}
	// Agrega más tipos aquí si es necesario
	default:
		// Tipo desconocido, manejar el error apropiadamente
	}

	v.currentEnv.variables[varID] = Variable{
		scope:         "var",
		PrimitiveType: primitivetype.value.(string),
		value:         emptySlice,
		Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
		Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
		Ambit:         v.currentAmbit,
	}
	AgregarSimbolo(varID, "Variable", primitivetype.value.(string), v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

	return Value{value: true}
}

func (v *Visitor) VisitAsignacionVector(ctx *parser.AsignacionVectorContext) Value {
	varID := ctx.ID().GetText()
	primitivoType := v.Visit(ctx.Primitivo()).value.(string)
	v.Visit(ctx.ListaExp())
	// Parse the tmpList based on the primitive type
	var convertedSlice interface{}
	switch primitivoType {
	case "Int":
		intSlice := make([]int64, len(v.tmpList))
		for i, val := range v.tmpList {
			intSlice[i] = val.(int64)
		}
		convertedSlice = intSlice
	case "String":
		stringSlice := make([]string, len(v.tmpList))
		for i, val := range v.tmpList {
			stringSlice[i] = val.(string)
		}
		convertedSlice = stringSlice
	case "Float":
		floatSlice := make([]float64, len(v.tmpList))
		for i, val := range v.tmpList {
			floatSlice[i] = val.(float64)
		}
		convertedSlice = floatSlice
	// Add more cases for other primitive types if needed
	default:
		// Handle unsupported types
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "No coincide la variable "+varID+" con el tipo primitivo"+primitivoType))
		AgregarErrorSemantico("No coincide la variable "+varID+" con el tipo primitivo"+primitivoType, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		fmt.Println("Unsupported primitive type:", primitivoType)
		convertedSlice = nil
		return Value{value: false}
	}
	if v.DoesVarExist(varID) {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable ya existe"+varID))
		AgregarErrorSemantico("La variable ya existe"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

		return Value{value: false}

	} else {
		v.currentEnv.variables[varID] = Variable{
			scope:         "var",
			PrimitiveType: primitivoType,
			value:         convertedSlice,
			isVector:      true,
			Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
			Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
			Ambit:         v.currentAmbit,
		}
		fmt.Println("vector agregado", v.tmpList)
		AgregarSimbolo(varID, "Variable", primitivoType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

		v.tmpList = nil
	}
	return Value{value: true}
}
func (v *Visitor) VisitReasignacionVector(ctx *parser.ReasignacionVectorContext) Value {
	fmt.Println("en reasignacion")
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var value interface{}
	var ambit interface{}
	var line interface{}
	var column interface{}
	var isVector bool
	valueToAssign := v.Visit(ctx.Expr(1))
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			scope = variable.scope
			ambit = variable.Ambit
			isVector = variable.isVector
			line = variable.Line
			column = variable.Column
			break
		}
		current = current.parent
	}
	var int64Value int64 = v.Visit(ctx.Expr(0)).value.(int64)
	index := int(int64Value)
	if isVector {
		switch primitivetype.(string) {
		case "String":
			sliceValue, ok := value.([]string)
			if !ok || index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para el vector "+varID))
				AgregarErrorSemantico("Index incompatible para el vector "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue[index] = valueToAssign.value.(string)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		case "Int":
			sliceValue, ok := value.([]int64)
			if !ok || index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para el vector "+varID))
				AgregarErrorSemantico("Index incompatible para el vector "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue[index] = valueToAssign.value.(int64)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		case "Float":
			sliceValue, ok := value.([]float64)
			if !ok || index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para el vector "+varID))
				AgregarErrorSemantico("Index incompatible para el vector "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue[index] = valueToAssign.value.(float64)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		default:
			fmt.Println("tipo primitivo desconocido")
			return Value{value: false}
		}

		//fmt.Printf("Elemento en el índice %d removido del slice de %s: %v\n", index, varContent.PrimitiveType, varContent.value)
		return Value{value: true}
	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable no es un vector "))
		AgregarErrorSemantico("La variable no es un vector", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	}

}
func (v *Visitor) VisitReasignacionMatrixTwoD(ctx *parser.ReasignacionMatrixTwoDContext) Value {
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var value interface{}
	var ambit interface{}
	var line interface{}
	var column interface{}
	var isMatrix bool
	valueToAssign := v.Visit(ctx.Expr(2))
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			scope = variable.scope
			ambit = variable.Ambit
			isMatrix = variable.isMatrix
			line = variable.Line
			column = variable.Column
			break
		}
		current = current.parent
	}
	var int64Value int64 = v.Visit(ctx.Expr(0)).value.(int64)
	index1 := int(int64Value)
	int64Value = v.Visit(ctx.Expr(1)).value.(int64)
	index2 := int(int64Value)
	if isMatrix {
		switch primitivetype.(string) {
		case "String":
			sliceValue, ok := value.([][]string)
			if !ok || index1 < 0 || index1 >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para la matriz "+varID))
				AgregarErrorSemantico("Index incompatible para la matriz "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue[index1][index2] = valueToAssign.value.(string)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isMatrix:      isMatrix,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		case "Int":
			sliceValue, ok := value.([][]int64)
			if !ok || index1 < 0 || index1 >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para la matriz "+varID))
				AgregarErrorSemantico("Index incompatible para la matriz "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue[index1][index2] = valueToAssign.value.(int64)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isMatrix:      isMatrix,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		default:
			fmt.Println("tipo primitivo desconocido")
			return Value{value: false}
		}

		//fmt.Printf("Elemento en el índice %d removido del slice de %s: %v\n", index, varContent.PrimitiveType, varContent.value)
		return Value{value: true}
	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable no es una matriz "))
		AgregarErrorSemantico("La variable no es una matriz", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	}

}
func (v *Visitor) visitIncremento(ctx *parser.IncrementoContext) Value {
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var currentValue interface{}
	var ambit interface{}
	var line interface{}
	var column interface{}
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			currentValue = variable.value
			scope = variable.scope
			ambit = variable.Ambit
			line = variable.Line
			column = variable.Column
			break
		}
		current = current.parent
	}
	valueToAssign := v.Visit(ctx.Expr()).value

	if v.encontrarVariable(varID) {
		if scope == "let" {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Variable tipo let, no puede ser modificada"+varID))
			AgregarErrorSemantico("Variable tipo let, no puede ser modificada"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
			return Value{value: false}

		} else {
			if primitivetype == getType(currentValue) {
				if primitivetype == "Int" {

					current.variables[varID] = Variable{
						scope:         scope.(string),
						PrimitiveType: primitivetype.(string),
						Line:          line.(string),
						Column:        column.(string),
						Ambit:         ambit.(string),
						value:         currentValue.(int64) + valueToAssign.(int64),
					}
				} else if primitivetype == "Float" {
					current.variables[varID] = Variable{
						scope:         scope.(string),
						PrimitiveType: primitivetype.(string),
						Line:          line.(string),
						Column:        column.(string),
						Ambit:         ambit.(string),
						value:         currentValue.(float64) + valueToAssign.(float64),
					}
				} else if primitivetype == "String" {
					current.variables[varID] = Variable{
						scope:         scope.(string),
						PrimitiveType: primitivetype.(string),
						Line:          line.(string),
						Column:        column.(string),
						Ambit:         ambit.(string),
						value:         currentValue.(string) + valueToAssign.(string),
					}
				}
			} else {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Tipo de variable no coincide con valor a asignar"))
				AgregarErrorSemantico("Tipo de variable no coincide con valor a asignar", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			}
		}

	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable "+varID+" no existe"))
		AgregarErrorSemantico("La variable "+varID+" no existe", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

		return Value{value: false}
	}
	return Value{value: true}

}

func (v *Visitor) VisitDecremento(ctx *parser.DecrementoContext) Value {
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var currentValue interface{}
	var ambit interface{}
	var line interface{}
	var column interface{}
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			currentValue = variable.value
			scope = variable.scope
			ambit = variable.Ambit
			line = variable.Line
			column = variable.Column
			break
		}
		current = current.parent
	}
	valueToAssign := v.Visit(ctx.Expr()).value

	if v.encontrarVariable(varID) {
		if scope == "let" {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Variable tipo let, no puede ser modificada "+varID))
			AgregarErrorSemantico("Variable tipo let, no puede ser modificada"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
			return Value{value: false}

		} else {
			if primitivetype == getType(valueToAssign) {
				if primitivetype == "Int" {

					current.variables[varID] = Variable{
						scope:         scope.(string),
						PrimitiveType: primitivetype.(string),
						Line:          line.(string),
						Column:        column.(string),
						Ambit:         ambit.(string),
						value:         currentValue.(int64) - valueToAssign.(int64),
					}
				} else if primitivetype == "Float" {
					current.variables[varID] = Variable{
						scope:         scope.(string),
						PrimitiveType: primitivetype.(string),
						Line:          line.(string),
						Column:        column.(string),
						Ambit:         ambit.(string),
						value:         currentValue.(float64) - valueToAssign.(float64),
					}
				}
			} else {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Tipo de variable no coincide con valor a asignar"))
				AgregarErrorSemantico("Tipo de variable no coincide con valor a asignar", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			}
		}

	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable "+varID+" no existe"))
		AgregarErrorSemantico("La variable "+varID+" no existe", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

		return Value{value: false}
	}
	return Value{value: true}
}

func (v *Visitor) VisitIfNormal(ctx *parser.IfNormalContext) Value {
	v.currentAmbit = "if"
	Generator.AddComment("Generando if")
	localEnv := &Environment{
		parent:    v.currentEnv,
		variables: make(map[string]Variable),
		functions: make(map[string]funcion),
		size_var:  make(map[string]int),
	}

	previousEnv := v.currentEnv
	v.currentEnv = localEnv
	defer func() {
		v.currentEnv = previousEnv
	}()
	localEnv.size_var["size"] = 0
	exp := v.Visit(ctx.Expr())
	for _, lvl := range exp.TrueLabel {
		Generator.AddLabel(lvl)
	}

	_return := v.Visit(ctx.Block())
	for _, lvl := range exp.FalseLabel {
		Generator.AddLabel(lvl)
	}
	return _return
}

func (v *Visitor) VisitIfElse(ctx *parser.ElseContext) Value {
	v.currentAmbit = "if"
	Generator.AddComment("Generando if")
	localEnv := &Environment{
		parent:    v.currentEnv,
		variables: make(map[string]Variable),
		functions: make(map[string]funcion),
		size_var:  make(map[string]int),
	}

	previousEnv := v.currentEnv
	v.currentEnv = localEnv
	defer func() {
		v.currentEnv = previousEnv
	}()
	localEnv.size_var["size"] = 0
	exp := v.Visit(ctx.Expr())
	fmt.Println("expresion", exp)
	for _, lvl := range exp.TrueLabel {
		Generator.AddLabel(lvl)
	}
	v.Visit(ctx.Block(0))
	trueLabel := Generator.NewLabel()
	Generator.AddGoto(trueLabel)
	for _, lvl := range exp.FalseLabel {
		Generator.AddLabel(lvl)
	}
	v.Visit(ctx.Block(1))
	Generator.AddLabel(trueLabel)
	return Value{value: true}

}
func (v *Visitor) VisitElse_If(ctx *parser.Else_ifContext) Value {
	v.currentAmbit = "if"
	Generator.AddComment("Generando if")
	localEnv := &Environment{
		parent:    v.currentEnv,
		variables: make(map[string]Variable),
		functions: make(map[string]funcion),
		size_var:  make(map[string]int),
	}

	previousEnv := v.currentEnv
	v.currentEnv = localEnv
	defer func() {
		v.currentEnv = previousEnv
	}()
	localEnv.size_var["size"] = 0
	exp := v.Visit(ctx.Expr())
	for _, lvl := range exp.TrueLabel {
		Generator.AddLabel(lvl)
	}

	_return := v.Visit(ctx.Block())
	trueLabel := Generator.NewLabel()
	Generator.AddGoto(trueLabel)
	for _, lvl := range exp.FalseLabel {
		Generator.AddLabel(lvl)
	}
	v.Visit(ctx.Ifstmt())
	Generator.AddLabel(trueLabel)
	return _return

}

func (v *Visitor) VisitSwitchstmt(ctx *parser.SwitchstmtContext) Value {
	v.currentAmbit = "switch"
	expr := v.Visit(ctx.Expr())
	v.exprToSwitch = expr.value
	v.tmp_xprSwitch = expr.temp.(string)
	if ctx.Cases() != nil {
		return v.Visit(ctx.Cases())
	}
	return Value{value: true}

}
func (v *Visitor) VisitCases(ctx *parser.CasesContext) Value {
	salida := Generator.NewLabel()
	Generator.AddComment("Generando switch")
	_lenCases := len(ctx.AllCaseblock())
	for i := 0; i <= _lenCases-1; i++ {
		v.exitSwitch = false
		value := v.Visit(ctx.Caseblock(i))
		caseValue := value.value
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		if i == _lenCases-1 {
			v.activate_case = true
			v.Visit(ctx.Caseblock(i))
			Generator.AddLabel(falseLabel)
		} else {

			op := fmt.Sprintf("%v", v.tmp_xprSwitch)
			valCmp := fmt.Sprintf("%v", caseValue)
			fmt.Println("op", v.exprToSwitch)
			fmt.Println("valCmp", valCmp)

			Generator.AddIf(op, valCmp, "==", trueLabel)
			Generator.AddGoto(falseLabel)
			Generator.AddLabel(trueLabel)
			v.activate_case = true
			v.Visit(ctx.Caseblock(i))
			Generator.AddGoto(salida)
			Generator.AddLabel(falseLabel)

		}
		/*if v.exitSwitch {
			v.exitSwitch = false
			break
		}*/
	}
	Generator.AddLabel(salida)

	return Value{value: true}
}
func (v *Visitor) VisitUnCase(ctx *parser.UnCaseContext) Value {

	if v.activate_case {
		v.activate_case = false
		return v.Visit(ctx.Block())
	}
	return v.Visit(ctx.Expr())

}
func (v *Visitor) VisitUnDefault(ctx *parser.UnDefaultContext) Value {

	return v.Visit(ctx.Block())

}
func (v *Visitor) VisitPrintstmt(ctx *parser.PrintstmtContext) Value {
	visit := v
	contenido := ""
	if ctx.AllExpr() != nil {
		for i := 0; i < len(ctx.AllExpr()); i++ {
			if i > 1 {
				contenido += " "
			}

			val := v.Visit(ctx.Expr(i))
			value := val.value
			fmt.Println("Valor a imprimir", value)
			fmt.Println(val)
			switch v := value.(type) {
			case [][]int64:
				value = fmt.Sprintf("%d", v)

				contenido += value.(string) + " "
				statements.PrintEnteros(val.temp, &Generator)
			case [][][]int64:
				value = fmt.Sprintf("%d", v)
				contenido += value.(string) + " "
				statements.PrintEnteros(val.temp, &Generator)
			case []int64:
				value = fmt.Sprintf("%d", v)
				contenido += value.(string) + " "
				statements.PrintEnteros(val.temp, &Generator)
			case [][]string:
				contenido += value.(string) + " "
			case [][][]string:
				contenido += value.(string) + " "
			case []string:
				contenido += value.(string) + " "
			case int64:
				value = fmt.Sprintf("%d", v)
				contenido += value.(string) + " "
				statements.PrintEnteros(val.temp, &Generator)
			case float64:
				value = fmt.Sprintf("%f", v)
				contenido += value.(string) + " "
				statements.PrintEnteros(val.temp, &Generator)
			case string:
				contenido += value.(string) + " "
				Generator.GeneratePrintString()
				newTemp1 := Generator.NewTemp()
				newTemp2 := Generator.NewTemp()
				size := strconv.Itoa(visit.currentEnv.size_var["size"])
				Generator.AddExpression(newTemp1, "P", size, "+")          // Nuevo temporal en posición vacía
				Generator.AddExpression(newTemp1, newTemp1, "1", "+")      // Se deja espacio de retorno
				Generator.AddSetStack("(int)"+newTemp1, val.temp.(string)) // Se coloca el string en el parámetro que se manda
				Generator.AddExpression("P", "P", size, "+")               // Cambio de entorno
				Generator.AddCall("dbrust_printString")                    // Llamada a la función de impresión de strings
				Generator.AddGetStack(newTemp2, "(int)P")                  // Obtención del retorno
				Generator.AddExpression("P", "P", size, "-")               // Regreso del entorno
				Generator.AddPrintf("c", "10")                             // Salto de línea
				Generator.AddBr()
			case rune:
				value = asciiToChar(value)
				contenido += value.(string) + " "
			case bool:

				newLabel := Generator.NewLabel()
				for _, lvl := range val.TrueLabel {
					Generator.AddLabel(lvl)

				}

				Generator.AddPrintf("c", "(char)116") // 't'
				Generator.AddPrintf("c", "(char)114") // 'r'
				Generator.AddPrintf("c", "(char)117") // 'u'
				Generator.AddPrintf("c", "(char)101") // 'e'
				Generator.AddGoto(newLabel)

				for _, lvl := range val.FalseLabel {
					Generator.AddLabel(lvl)

				}
				Generator.AddPrintf("c", "(char)102") // 'f'
				Generator.AddPrintf("c", "(char)97")  // 'a'
				Generator.AddPrintf("c", "(char)108") // 'l'
				Generator.AddPrintf("c", "(char)115") // 's'
				Generator.AddPrintf("c", "(char)101") // 'e'
				Generator.AddLabel(newLabel)
				Generator.AddPrintf("c", "10") // Salto de línea
				Generator.AddBr()

				value = fmt.Sprintf("%t", v)

				contenido += value.(string) + " "
			case nil:
				value := 9999999
				Generator.AddPrintf("d", "(int)"+fmt.Sprintf("%v", value))
				Generator.AddPrintf("c", "10") // Salto de línea
				Generator.AddBr()
				contenido += "nil" + " "
			default:

				fmt.Println("Tipo desconocido")
			}
		}
	}
	v.outputBuilder.WriteString(fmt.Sprintf("%v\n", contenido))
	return Value{value: true}
}

func (v *Visitor) VisitWhilestmt(ctx *parser.WhilestmtContext) Value {
	v.currentAmbit = "while"
	localEnv := &Environment{
		parent:    v.currentEnv,
		variables: make(map[string]Variable),
		functions: make(map[string]funcion),
		size_var:  make(map[string]int),
	}
	previousEnv := v.currentEnv
	v.currentEnv = localEnv
	localEnv.size_var["size"] = 0
	defer func() {
		v.currentEnv = previousEnv
	}()

	trueLabel := Generator.NewLabel()
	Generator.AddLabel(trueLabel)
	Generator.AddComment("Generando while")
	condition := v.Visit(ctx.Expr())
	fmt.Println()
	for _, lvl := range condition.TrueLabel {
		Generator.AddLabel(lvl)
	}
	_return := v.Visit(ctx.Block())
	Generator.AddGoto(trueLabel)
	for _, lvl := range condition.FalseLabel {
		Generator.AddLabel(lvl)
	}
	return _return

}
func (v *Visitor) VisitForNormal(ctx *parser.ForNormalContext) Value {
	v.currentAmbit = "for"
	varID := ctx.ID().GetText()
	exp1 := v.Visit(ctx.Expr(0))
	exp2 := v.Visit(ctx.Expr(1))

	if exp1.value.(int64) > exp2.value.(int64) {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Operación incompatible, la expresión "+exp1.value.(string)+"es mayor a "+exp2.value.(string)))
		AgregarErrorSemantico("Operación incompatible, la expresión "+exp1.value.(string)+" es mayor a "+exp2.value.(string), strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	} else {

		localEnv := &Environment{
			parent:    v.currentEnv,
			variables: make(map[string]Variable),
			functions: make(map[string]funcion),
			size_var:  make(map[string]int),
		}

		localEnv.size_var["size"] = 0
		previousEnv := v.currentEnv
		v.currentEnv = localEnv
		defer func() {
			v.currentEnv = previousEnv
		}()
		iterador := v.currentEnv.variables[varID]
		iterador = Variable{
			scope:         "var",
			PrimitiveType: "Int",
			value:         exp1.value.(int64),
			Stack:         Cont_Stack,
		}

		label := Generator.NewLabel()
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		localEnv.size_var["size"]++
		valorIterador := fmt.Sprintf("%v", exp1.temp.(int64))
		Generator.AddSetStack(strconv.Itoa(Cont_Stack), valorIterador)
		newTmp := Generator.NewTemp()
		Generator.AddLabel(label)
		Generator.AddGetStack(newTmp, strconv.Itoa(Cont_Stack))
		exp2.temp = fmt.Sprintf("%v", exp2.temp.(int64))
		Generator.AddIf(newTmp, exp2.temp.(string), "<=", trueLabel)
		Generator.AddGoto(falseLabel)
		Generator.AddLabel(trueLabel)
		Cont_Stack++
		v.Visit(ctx.Block())
		Generator.AddGetStack(newTmp, strconv.Itoa(iterador.Stack))
		Generator.AddExpression(newTmp, newTmp, "1", "+")
		Generator.AddSetStack(strconv.Itoa(iterador.Stack), newTmp)
		Generator.AddGoto(label)
		Generator.AddLabel(falseLabel)

		return Value{value: true}
	}
}
func (v *Visitor) VisitForEach(ctx *parser.ForEachContext) Value {
	v.currentAmbit = "forE"
	varID := ctx.ID().GetText()
	expr := v.Visit(ctx.Expr()).value
	if getType(expr) == "String" {
		v.currentEnv.variables[varID] = Variable{
			scope:         "var",
			PrimitiveType: "Character",
			value:         nil,
		}
		for _, char := range expr.(string) {
			v.currentEnv.variables[varID] = Variable{
				scope:         "var",
				PrimitiveType: "Character",
				value:         char,
			}
			localEnv := &Environment{
				parent:    v.currentEnv,
				variables: make(map[string]Variable),
				functions: make(map[string]funcion),
				size_var:  make(map[string]int),
			}
			localEnv.size_var["size"] = 0
			previousEnv := v.currentEnv
			v.currentEnv = localEnv
			defer func() {
				v.currentEnv = previousEnv
			}()
			v.Visit(ctx.Block())
			if v.breakFlag { // si encuentra un break se termina el ciclo
				v.breakFlag = false
				break
			}

		}
		return Value{value: true}
	}
	return Value{value: false}
}
func (v *Visitor) VisitGuardstmt(ctx *parser.GuardstmtContext) Value {
	v.currentAmbit = "guard"
	value, ok := v.Visit(ctx.Expr()).value.(bool)
	if !value && ok {
		v.Visit(ctx.Block())
		hasTransferStc := ctx.Transfer_sentence().GetText()
		if hasTransferStc == "continue" {
			v.continueFlag = true //Se encontró el continue xd
		}
		if hasTransferStc == "break" {
			v.breakFlag = true //
		}
	}
	return Value{value: false}
}

func (v *Visitor) VisitVectorAppend(ctx *parser.VectorAppendContext) Value {
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var value interface{}
	var ambit interface{}
	var line interface{}
	var column interface{}
	var isVector bool
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			scope = variable.scope
			ambit = variable.Ambit
			isVector = variable.isVector
			line = variable.Line
			column = variable.Column
			break
		}
		current = current.parent
	}
	exprValue := v.Visit(ctx.Expr()).value
	primitive := getType(exprValue)

	if isVector {
		if primitivetype == primitive {
			switch primitivetype {
			case "String":
				sliceValue, ok := value.([]string)
				if !ok {
					return Value{value: false}
				}
				newValue := exprValue.(string)
				sliceValue = append(sliceValue, newValue)

				// Actualizar el valor en el mapa
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}

			case "Int":
				sliceValue, ok := value.([]int64)
				if !ok {
					return Value{value: false}
				}
				newValue := exprValue.(int64)
				sliceValue = append(sliceValue, newValue)

				// Actualizar el valor en el mapa
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			case "Float":
				sliceValue, ok := value.([]float64)
				if !ok {
					return Value{value: false}
				}
				newValue := exprValue.(float64)
				sliceValue = append(sliceValue, newValue)

				// Actualizar el valor en el mapa
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}

			default:
				fmt.Println("tipo primitivo desconocido")
				return Value{value: false}
			}

			//fmt.Printf("Nuevo valor agregado al slice de %s: %v\n", varContent.PrimitiveType, varContent.value)
			return Value{value: true}
		} else {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "No coincide el tipo de expresión con el tipo del vector"))
			AgregarErrorSemantico("No coincide el tipo de expresión con el tipo del vector("+primitivetype.(string)+")", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

			return Value{value: false}
		}
	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable no es un slice"))
		AgregarErrorSemantico("La variable no es un vector", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

		return Value{value: false}
	}
}

func (v *Visitor) VisitVectorRemoveLast(ctx *parser.VectorRemoveLastContext) Value {
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var value interface{}
	var ambit interface{}
	var line interface{}
	var column interface{}
	var isVector bool
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			scope = variable.scope
			ambit = variable.Ambit
			isVector = variable.isVector
			line = variable.Line
			column = variable.Column
			break
		}
		current = current.parent
	}
	if isVector {
		switch primitivetype {
		case "String":
			sliceValue, ok := value.([]string)
			if !ok || len(sliceValue) == 0 {
				return Value{value: false}
			}
			sliceValue = sliceValue[:len(sliceValue)-1]
			current.variables[varID] = Variable{
				scope:         scope.(string),
				PrimitiveType: primitivetype.(string),
				value:         sliceValue,
				isVector:      isVector,
				Line:          line.(string),
				Column:        column.(string),
				Ambit:         ambit.(string),
			}

		case "Int":
			sliceValue, ok := value.([]int64)
			if !ok || len(sliceValue) == 0 {
				return Value{value: false}
			}
			sliceValue = sliceValue[:len(sliceValue)-1]
			current.variables[varID] = Variable{
				scope:         scope.(string),
				PrimitiveType: primitivetype.(string),
				value:         sliceValue,
				isVector:      isVector,
				Line:          line.(string),
				Column:        column.(string),
				Ambit:         ambit.(string),
			}

		case "Float":
			sliceValue, ok := value.([]float64)
			if !ok || len(sliceValue) == 0 {
				return Value{value: false}
			}
			sliceValue = sliceValue[:len(sliceValue)-1]
			current.variables[varID] = Variable{
				scope:         scope.(string),
				PrimitiveType: primitivetype.(string),
				value:         sliceValue,
				isVector:      isVector,
				Line:          line.(string),
				Column:        column.(string),
				Ambit:         ambit.(string),
			}

		default:
			fmt.Println("tipo primitivo desconocido")
			return Value{value: false}
		}

		//fmt.Printf("Último valor removido del slice de %s: %v\n", varContent.PrimitiveType, varContent.value)
		return Value{value: true}
	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable no es un vector"))
		AgregarErrorSemantico("La variable no es un vector", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	}

}

func (v *Visitor) VisitRemoveAt(ctx *parser.VectorRemoveAtContext) Value {
	varID := ctx.ID().GetText()
	var scope interface{}
	var primitivetype interface{}
	var value interface{}
	var ambit interface{}
	var line interface{}
	var column interface{}
	var isVector bool
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			scope = variable.scope
			ambit = variable.Ambit
			isVector = variable.isVector
			line = variable.Line
			column = variable.Column
			break
		}
		current = current.parent
	}
	var int64Value int64 = v.Visit(ctx.Expr()).value.(int64)
	index := int(int64Value)
	if isVector {
		switch primitivetype.(string) {
		case "String":
			sliceValue, ok := value.([]string)
			if !ok || index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para el vector "+varID))
				AgregarErrorSemantico("Index incompatible para el vector "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue = append(sliceValue[:index], sliceValue[index+1:]...)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		case "Int":
			sliceValue, ok := value.([]int64)
			if !ok || index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para el vector "+varID))
				AgregarErrorSemantico("Index incompatible para el vector "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue = append(sliceValue[:index], sliceValue[index+1:]...)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		case "Float":
			sliceValue, ok := value.([]float64)
			if !ok || index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Index incompatible para el vector "+varID))
				AgregarErrorSemantico("Index incompatible para el vector "+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				return Value{value: false}
			} else {
				sliceValue = append(sliceValue[:index], sliceValue[index+1:]...)
				current.variables[varID] = Variable{
					scope:         scope.(string),
					PrimitiveType: primitivetype.(string),
					value:         sliceValue,
					isVector:      isVector,
					Line:          line.(string),
					Column:        column.(string),
					Ambit:         ambit.(string),
				}
			}

		default:
			fmt.Println("tipo primitivo desconocido")
			return Value{value: false}
		}

		//fmt.Printf("Elemento en el índice %d removido del slice de %s: %v\n", index, varContent.PrimitiveType, varContent.value)
		return Value{value: true}
	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable no es un slice"))
		AgregarErrorSemantico("La variable no es un vector", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	}
}

func (v *Visitor) visitVectorCount(ctx *parser.VectorCountContext) Value {
	varID := ctx.ID().GetText()
	var primitivetype interface{}
	var value interface{}
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			break
		}
		current = current.parent
	}
	getSliceSize := func(value interface{}, primitiveType string) (int, error) {
		switch primitiveType {
		case "String":
			sliceValue, ok := value.([]string)
			if !ok {
				return 0, fmt.Errorf("tipo de slice incorrecto")
			}
			return len(sliceValue), nil
		case "Int":
			sliceValue, ok := value.([]int64)
			if !ok {
				return 0, fmt.Errorf("tipo de slice incorrecto")
			}
			return len(sliceValue), nil
		case "Float":
			sliceValue, ok := value.([]float64)
			if !ok {
				return 0, fmt.Errorf("tipo de slice incorrecto")
			}
			return len(sliceValue), nil
		default:
			return 0, fmt.Errorf("tipo primitivo desconocido")
		}
	}

	// Obtener el tamaño de los slices
	sizeSlice, errSlice := getSliceSize(value, primitivetype.(string))
	if errSlice != nil {
		fmt.Println(errSlice)
	}
	int64Value := int64(sizeSlice)
	return Value{value: int64Value}
}

func (v *Visitor) VisitVectorIsEmpty(ctx *parser.VectorIsEmptyContext) Value {
	varID := ctx.ID().GetText()
	var primitivetype interface{}
	var value interface{}
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			break
		}
		current = current.parent
	}
	getSliceSize := func(value interface{}, primitiveType string) (int, error) {
		switch primitiveType {
		case "String":
			sliceValue, ok := value.([]string)
			if !ok {
				return 0, fmt.Errorf("tipo de slice incorrecto")
			}
			return len(sliceValue), nil
		case "Int":
			sliceValue, ok := value.([]int64)
			if !ok {
				return 0, fmt.Errorf("tipo de slice incorrecto")
			}
			return len(sliceValue), nil
		case "Float":
			sliceValue, ok := value.([]float64)
			if !ok {
				return 0, fmt.Errorf("tipo de slice incorrecto")
			}
			return len(sliceValue), nil
		default:
			return 0, fmt.Errorf("tipo primitivo desconocido")
		}
	}

	// Obtener el tamaño de los slices
	sizeSlice, errSlice := getSliceSize(value, primitivetype.(string))
	if errSlice != nil {
		fmt.Println(errSlice)
	}
	if sizeSlice == 0 {
		return Value{value: true}
	}
	return Value{value: false}

}

func (v *Visitor) VisitVectorGetElement(ctx *parser.VectorGetElementContext) Value {
	varID := ctx.ID().GetText()
	var primitivetype interface{}
	var value interface{}
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			break
		}
		current = current.parent
	}
	getSliceValue := func(value interface{}, primitiveType string, index int) (interface{}, error) {
		switch primitiveType {
		case "String":
			sliceValue, ok := value.([]string)
			if !ok {
				return nil, fmt.Errorf("tipo de slice incorrecto")
			}
			if index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Indice fuera de rango"))
				AgregarErrorSemantico("Indice fuera de rango", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

				return nil, fmt.Errorf("índice fuera de rango")
			}
			return sliceValue[index], nil
		case "Float":
			sliceValue, ok := value.([]float64)
			if !ok {
				return nil, fmt.Errorf("tipo de slice incorrecto")
			}
			if index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Indice fuera de rango"))
				AgregarErrorSemantico("Indice fuera de rango", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

				return nil, fmt.Errorf("índice fuera de rango")
			}
			return sliceValue[index], nil
		case "Int":
			sliceValue, ok := value.([]int64)
			if !ok {
				return nil, fmt.Errorf("tipo de slice incorrecto")
			}
			if index < 0 || index >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Indice fuera de rango"))
				AgregarErrorSemantico("Indice fuera de rango", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

				return nil, fmt.Errorf("índice fuera de rango")
			}
			return sliceValue[index], nil
		default:
			return nil, fmt.Errorf("tipo primitivo desconocido")
		}
	}
	var int64Value int64 = v.Visit(ctx.Expr()).value.(int64)
	intValue := int(int64Value)

	sliceValu, errSliceValue := getSliceValue(value, primitivetype.(string), intValue)
	if errSliceValue != nil {
		fmt.Println(errSliceValue)
	}
	return Value{value: sliceValu}
}

func (v *Visitor) VisitMatrixTwoD(ctx *parser.MatrixTwoDContext) Value {
	varID := ctx.ID().GetText()
	primitivoType := v.Visit(ctx.Primitivo()).value.(string)
	v.Visit(ctx.DefMatrix())
	matrix := v.tmpListaMtrx
	v.tmpListaMtrx = nil
	if primitivoType == "Int" {
		var matriz [][]int64

		if v.DoesVarExist(varID) {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable ya existe"+varID))
			AgregarErrorSemantico("La variable ya existe"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

			return Value{value: false}

		} else {
			matriz = parsermatrix.ConvertirMatrizToInt(matrix)
			v.currentEnv.variables[varID] = Variable{
				scope:         "var",
				PrimitiveType: primitivoType,
				value:         matriz,
				isVector:      false,
				isMatrix:      true,
				Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
				Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
				Ambit:         v.currentAmbit,
			}
			AgregarSimbolo(varID, "Variable", primitivoType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

		}

	} else if primitivoType == "String" {
		var matriz [][]string
		if v.DoesVarExist(varID) {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable ya existe"+varID))
			AgregarErrorSemantico("La variable ya existe"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

			return Value{value: false}

		} else {
			matriz = parsermatrix.ConvertirMatrizToString(matrix)
			v.currentEnv.variables[varID] = Variable{
				scope:         "var",
				PrimitiveType: primitivoType,
				value:         matriz,
				isVector:      false,
				isMatrix:      true,
				Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
				Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
				Ambit:         v.currentAmbit,
			}
			AgregarSimbolo(varID, "Variable", primitivoType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

		}
	}
	return Value{value: true}
}

func (v *Visitor) VisitDefMatrix(ctx *parser.DefMatrixContext) Value {
	if ctx.ListaValores_Mat() != nil {
		v.Visit(ctx.ListaValores_Mat())
		return Value{value: true}
	} else {
		return Value{value: nil}
	}

}
func (v *Visitor) VisitListaValores_Mat(ctx *parser.ListaValores_MatContext) Value {

	if ctx.ListaValores_Mat2() != nil {
		v.Visit(ctx.ListaValores_Mat2())

		return Value{value: true}
	} else {
		return Value{value: nil}
	}

}
func (v *Visitor) VisitListaValores_Mat2(ctx *parser.ListaValores_Mat2Context) Value {

	if ctx.ListaValores_Mat2() != nil {
		v.Visit(ctx.ListaValores_Mat2())

	}
	if ctx.ListaValores_Mat() != nil {
		v.Visit(ctx.ListaValores_Mat())

	}
	if ctx.Lista_Expresiones() != nil {
		v.Visit(ctx.Lista_Expresiones())

	}
	return Value{value: true}
}
func (v *Visitor) VisitListaExpresiones(ctx *parser.Lista_ExpresionesContext) Value {
	var lista []interface{}
	if ctx.AllExpr() != nil {
		for i := 0; i < len(ctx.AllExpr()); i++ {

			lista = append(lista, v.Visit(ctx.Expr(i)).value)

		}
		v.tmpListaMtrx = append(v.tmpListaMtrx, lista)
	}
	return Value{value: lista}
}
func (v *Visitor) VisitMatrixThreeD(ctx *parser.MatrixThreeDContext) Value {
	var matrixThreeD [][][]interface{}
	varID := ctx.ID().GetText()
	primitivoType := v.Visit(ctx.Primitivo()).value.(string)
	v.tmpMtrx3 = true
	v.Visit(ctx.DefMatrix())
	v.tmpMtrx3 = false
	matrix := v.tmpListaMtrx
	matrixThreeD = append(matrixThreeD, matrix)
	v.tmpListaMtrx = nil
	if primitivoType == "Int" {
		var matriz [][][]int64

		if v.DoesVarExist(varID) {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable ya existe"+varID))
			AgregarErrorSemantico("La variable ya existe"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

			return Value{value: false}

		} else {
			matriz = parsermatrix.ConvertirMatriz3ToInt(matrixThreeD)
			v.currentEnv.variables[varID] = Variable{
				scope:         "var",
				PrimitiveType: primitivoType,
				value:         matriz,
				isVector:      false,
				isMatrix:      true,
				Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
				Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
				Ambit:         v.currentAmbit,
			}
			AgregarSimbolo(varID, "Variable", primitivoType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

		}

	} else if primitivoType == "String" {
		var matriz [][][]string
		if v.DoesVarExist(varID) {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La variable ya existe"+varID))
			AgregarErrorSemantico("La variable ya existe"+varID, strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

			return Value{value: false}

		} else {
			matriz = parsermatrix.ConvertirMatriz3ToString(matrixThreeD)
			v.currentEnv.variables[varID] = Variable{
				scope:         "var",
				PrimitiveType: primitivoType,
				value:         matriz,
				isVector:      false,
				isMatrix:      true,
				Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
				Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn()),
				Ambit:         v.currentAmbit,
			}
			AgregarSimbolo(varID, "Variable", primitivoType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

		}
	}
	return Value{value: true}

}
func (v *Visitor) VisitAccesoMatrixTwoD(ctx *parser.AccesoMatrixTwoDContext) Value {
	varID := ctx.ID().GetText()
	var primitivetype interface{}
	var value interface{}
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			primitivetype = variable.PrimitiveType
			value = variable.value
			break
		}
		current = current.parent
	}
	getSliceValue := func(value interface{}, primitiveType string, index1 int, index2 int) (interface{}, error) {
		switch primitiveType {
		case "String":
			sliceValue, ok := value.([][]string)
			if !ok {
				return nil, fmt.Errorf("tipo de slice incorrecto")
			}
			if index1 < 0 || index1 >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Indice fuera de rango"))
				AgregarErrorSemantico("Indice fuera de rango", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

				return nil, fmt.Errorf("índice fuera de rango")
			}
			return sliceValue[index1][index2], nil
		case "Int":
			sliceValue, ok := value.([][]int64)
			if !ok {
				return nil, fmt.Errorf("tipo de slice incorrecto")
			}
			if index1 < 0 || index1 >= len(sliceValue) {
				v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Indice fuera de rango"))
				AgregarErrorSemantico("Indice fuera de rango", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))

				return nil, fmt.Errorf("índice fuera de rango")
			}
			return sliceValue[index1][index2], nil
		default:
			return nil, fmt.Errorf("tipo primitivo desconocido")
		}
	}
	var int64Value int64 = v.Visit(ctx.Expr(0)).value.(int64)
	index1 := int(int64Value)
	int64Value = v.Visit(ctx.Expr(1)).value.(int64)
	index2 := int(int64Value)
	sliceValu, errSliceValue := getSliceValue(value, primitivetype.(string), index1, index2)
	if errSliceValue != nil {
		fmt.Println(errSliceValue)
	}
	return Value{value: sliceValu}
}

func (v *Visitor) VisitFunc_sinRetorno(ctx *parser.Func_sinRetornoContext) Value {
	block := ctx.Block()
	funcID := ctx.ID().GetText()
	if v.DoesFuncExist(funcID) {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La función "+funcID+" ya existe"))
		AgregarErrorSemantico("La función "+funcID+" ya existe", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	} else {

		//Agregar a la lista temporal como en los vectores :)
		if ctx.ListaParams() != nil {
			v.Visit(ctx.ListaParams())
		}
		v.currentEnv.functions[funcID] = funcion{
			primitiveType: "",
			params:        v.tmpListParams,
			value:         block,
			hasReturn:     false,
		}
		AgregarSimbolo(funcID, "Función", "--", v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

		v.tmpListParams = nil
		return Value{value: true}
	}

}

func (v *Visitor) Visitfunc_conRetorno_conTipo(ctx *parser.Func_conRetorno_conTipoContext) Value {
	block := ctx.Block()
	funcID := ctx.ID().GetText()
	primitivo := ctx.Primitivo().GetText()
	var valueExpr antlr.ParseTree
	if ctx.Expr() != nil {
		valueExpr = ctx.Expr()
	} else {
		valueExpr = nil
	}
	if ctx.ListaParams() != nil {
		v.Visit(ctx.ListaParams())
	}
	if v.DoesFuncExist(funcID) {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La función "+funcID+" ya existe"))
		AgregarErrorSemantico("La función "+funcID+" ya existe", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	} else {
		v.currentEnv.functions[funcID] = funcion{
			primitiveType: "",
			params:        v.tmpListParams,
			value:         block,
			hasReturn:     true,
			exprToRetunr:  valueExpr,
		}
		AgregarSimbolo(funcID, "Función", primitivo, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

		v.tmpListParams = nil
		return Value{value: true}
	}
}

func (v *Visitor) VisitCallFunction(ctx *parser.CallFunctionContext) Value {
	funcID := ctx.ID().GetText()
	v.currentAmbit = funcID
	var parametros []param
	if v.encontrarFunc(funcID) {
		current := v.currentEnv
		for current != nil {
			if function, ok := current.functions[funcID]; ok {
				if len(function.params) > 0 {
					if ctx.ListaParamsCall() != nil {
						parametros = v.Visit(ctx.ListaParamsCall()).value.([]param)

					}
					if len(parametros) != len(function.params) {
						fmt.Println("tamaño parametros que entran", len(parametros))
						fmt.Println("tamaño parametros existentes", len(function.params))
						v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Error en declaración de parametros, parametros de la función: "+strconv.Itoa(len(function.params))))
						AgregarErrorSemantico("Error en declaración de parametros, parametros de la función: "+strconv.Itoa(len(function.params)), strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
							strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
						return Value{value: nil}
					}
					localEnv := &Environment{
						parent:    v.currentEnv,
						variables: make(map[string]Variable),
						functions: make(map[string]funcion),
						size_var:  make(map[string]int),
					}
					localEnv.size_var["size"] = 0
					previousEnv := v.currentEnv
					v.currentEnv = localEnv
					defer func() {
						v.currentEnv = previousEnv
					}()
					for i := range parametros {
						if function.params[i].IDExterno == "_" {
							if function.params[i].Inout {
								function.params[i].idVarToInout = parametros[i].idVarToInout
							}
							newVar := Variable{
								scope:         "var",
								Ambit:         v.currentAmbit,
								PrimitiveType: function.params[i].primitiveType,
								value:         parametros[i].value,
								isVector:      function.params[i].isVector,
								Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
								Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn() + i),
							}
							v.currentEnv.variables[function.params[i].IDInterno] = newVar
							AgregarSimbolo(function.params[i].IDInterno, "Variable", function.params[i].primitiveType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

						} else if function.params[i].IDExterno == parametros[i].IDInterno {
							if function.params[i].Inout {

								function.params[i].idVarToInout = parametros[i].idVarToInout
							}
							newVar := Variable{
								scope:         "var",
								Ambit:         v.currentAmbit,
								PrimitiveType: function.params[i].primitiveType,
								value:         parametros[i].value,
								isVector:      function.params[i].isVector,
								Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
								Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn() + i),
							}
							v.currentEnv.variables[function.params[i].IDInterno] = newVar
							AgregarSimbolo(function.params[i].IDInterno, "Variable", function.params[i].primitiveType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

						}

					}

					v.Visit(function.value)
					// después de haber ejecutado todas las sentencias recorro de nuevo los parametros buscar los inout y reasignar valores
					for i := range parametros {
						if function.params[i].Inout {

							varID := parametros[i].idVarToInout
							tmp := v.currentEnv

							for tmp != nil {
								if variable, ok := tmp.variables[varID]; ok {
									tmp.variables[varID] = Variable{
										scope:         variable.scope,
										Ambit:         variable.Ambit,
										PrimitiveType: variable.PrimitiveType,
										value:         v.currentEnv.variables[function.params[i].IDInterno].value,
										isVector:      variable.isVector,
										Line:          variable.Line,
										Column:        variable.Column,
									}
									break
								}

								tmp = tmp.parent
							}
						}
					}
					v.returnFlag = false

					return Value{value: nil}
				} else {
					localEnv := &Environment{
						parent:    v.currentEnv,
						variables: make(map[string]Variable),
						functions: make(map[string]funcion),
						size_var:  make(map[string]int),
					}
					localEnv.size_var["size"] = 0
					previousEnv := v.currentEnv
					v.currentEnv = localEnv
					defer func() {
						v.currentEnv = previousEnv
					}()
					v.Visit(function.value)
					v.returnFlag = false
					return Value{value: true}
				}
			}
			current = current.parent
		}
	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La funcion "+funcID+" no existe"))
		AgregarErrorSemantico("La funcion "+funcID+" no exoste", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: false}
	}
	return Value{value: true}
}

func (v *Visitor) VisitcallFuncAsExpr(ctx *parser.CallFuncAsExprContext) Value {
	v.returnFlag = false
	funcID := ctx.ID().GetText()
	v.currentAmbit = funcID
	var parametros []param
	if v.encontrarFunc(funcID) {
		current := v.currentEnv
		for current != nil {
			if function, ok := current.functions[funcID]; ok {
				if function.hasReturn {
					if len(function.params) > 0 {
						if ctx.ListaParamsCall() != nil {
							parametros = v.Visit(ctx.ListaParamsCall()).value.([]param)

						}
						if len(parametros) != len(function.params) {
							fmt.Println("tamaño parametros que entran", len(parametros))
							fmt.Println("tamaño parametros existentes", len(function.params))
							v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "Error en declaración de parametros, parametros de la función: "+strconv.Itoa(len(function.params))))
							AgregarErrorSemantico("Error en declaración de parametros, parametros de la función: "+strconv.Itoa(len(function.params)), strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
								strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
							return Value{value: nil}
						}
						localEnv := &Environment{
							parent:    v.currentEnv,
							variables: make(map[string]Variable),
							functions: make(map[string]funcion),
							size_var:  make(map[string]int),
						}
						localEnv.size_var["size"] = 0
						previousEnv := v.currentEnv
						v.currentEnv = localEnv
						defer func() {
							v.currentEnv = previousEnv
						}()
						for i := range parametros {
							if function.params[i].IDExterno == "_" {
								if function.params[i].Inout {
									function.params[i].idVarToInout = parametros[i].idVarToInout
								}
								newVar := Variable{
									scope:         "var",
									Ambit:         v.currentAmbit,
									PrimitiveType: function.params[i].primitiveType,
									value:         parametros[i].value,
									isVector:      function.params[i].isVector,
									Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
									Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn() + i),
								}
								v.currentEnv.variables[function.params[i].IDInterno] = newVar
								AgregarSimbolo(function.params[i].IDInterno, "Variable", function.params[i].primitiveType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

							} else if function.params[i].IDExterno == parametros[i].IDInterno {
								if function.params[i].Inout {

									function.params[i].idVarToInout = parametros[i].idVarToInout
								}
								newVar := Variable{
									scope:         "var",
									Ambit:         v.currentAmbit,
									PrimitiveType: function.params[i].primitiveType,
									value:         parametros[i].value,
									isVector:      function.params[i].isVector,
									Line:          strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
									Column:        strconv.Itoa(ctx.ID().GetSymbol().GetColumn() + i),
								}
								v.currentEnv.variables[function.params[i].IDInterno] = newVar
								AgregarSimbolo(function.params[i].IDInterno, "Variable", function.params[i].primitiveType, v.currentAmbit, ctx.ID().GetSymbol().GetLine(), ctx.ID().GetSymbol().GetColumn())

							}
						}

						v.Visit(function.value)
						// después de haber ejecutado todas las sentencias recorro de nuevo los parametros buscar los inout y reasignar valores
						for i := range parametros {
							if function.params[i].Inout {

								varID := parametros[i].idVarToInout
								tmp := v.currentEnv

								for tmp != nil {
									if variable, ok := tmp.variables[varID]; ok {
										tmp.variables[varID] = Variable{
											scope:         variable.scope,
											Ambit:         variable.Ambit,
											PrimitiveType: variable.PrimitiveType,
											value:         v.currentEnv.variables[function.params[i].IDInterno].value,
											isVector:      variable.isVector,
											Line:          variable.Line,
											Column:        variable.Column,
										}
										break
									}

									tmp = tmp.parent
								}
							}
						}

						if v.returnExpr != nil {
							v.returnFlag = false
							return Value{value: v.returnExpr}
						}
						v.returnFlag = false
						return Value{value: nil}
					} else {
						localEnv := &Environment{
							parent:    v.currentEnv,
							variables: make(map[string]Variable),
							functions: make(map[string]funcion),
							size_var:  make(map[string]int),
						}
						localEnv.size_var["size"] = 0
						previousEnv := v.currentEnv
						v.currentEnv = localEnv
						defer func() {
							v.currentEnv = previousEnv
						}()
						v.Visit(function.value)
						if v.returnExpr != nil {
							v.returnFlag = false
							return Value{value: v.returnExpr}
						}
						v.returnFlag = false
						return Value{value: nil}
					}
				} else {
					v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La funcion "+funcID+" no devuelve una expresión"))
					AgregarErrorSemantico("La funcion "+funcID+" no devuelve una expresión", strconv.Itoa(ctx.ID().GetSymbol().GetLine()),
						strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
				}
			}
			current = current.parent
		}
	} else {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "La funcion "+funcID+" no existe"))
		AgregarErrorSemantico("La funcion "+funcID+" no exoste", strconv.Itoa(ctx.ID().GetSymbol().GetLine()), strconv.Itoa(ctx.ID().GetSymbol().GetColumn()))
		return Value{value: nil}
	}
	return Value{value: nil}
}

func (v *Visitor) VisitIntExpr(ctx *parser.IntExprContext) Value {

	i, _ := strconv.ParseInt(ctx.GetText(), 10, 64)
	return Value{value: i, temp: i, Tipo: "Int"}
}
func (v *Visitor) VisitNilExpr(ctx *parser.NilExprContext) Value {

	return Value{value: nil}
}

func (v *Visitor) VisitFloatExpr(ctx *parser.FloatExprContext) Value {
	f, _ := strconv.ParseFloat(ctx.GetText(), 64)
	return Value{value: f, temp: f, Tipo: "Float"}
}

func (v *Visitor) VisitStrExpr(ctx *parser.StrExprContext) Value {
	HeapPointer = Cont_Heap
	newTmp := Generator.NewTemp()
	Generator.AddAssign(newTmp, "H")
	value := strings.Trim(ctx.GetText(), "\"")

	for _, asc := range value {
		//se agrega ascii al heap
		Generator.AddSetHeap("(int)H", strconv.Itoa(int(asc)))
		//suma heap pointer
		Generator.AddExpression("H", "H", "1", "+")
		Cont_Heap++
	}
	fmt.Println("Contador de heap: ", Cont_Heap)
	Cont_Heap++
	//caracteres de escape
	Generator.AddSetHeap("(int)H", "-1")
	Generator.AddExpression("H", "H", "1", "+")
	Generator.AddBr()
	return Value{value: value, temp: newTmp, Tipo: "String"}
}
func (v *Visitor) VisitCharExpr(ctx *parser.CharExprContext) Value {
	value := rune(ctx.GetText()[1]) // Obtiene el caracter entre las comillas simples
	return Value{value: value}      // Suponiendo que Value tiene un campo charValue de tipo rune
}

func (v *Visitor) VisitIdExpr(ctx *parser.IdExprContext) Value {
	varID := ctx.ID().GetText()
	Generator.AddComment("LLamando a variable: " + varID)
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			result := Value{}
			newTmp := Generator.NewTemp()
			newTmp2 := Generator.NewTemp()
			if Generator.MainCode {
				Generator.AddGetStack(newTmp2, strconv.Itoa(variable.Stack))
			} else {
				Generator.AddExpression(newTmp, "P", strconv.Itoa(variable.Stack), "+")
				Generator.AddGetStack(newTmp2, "(int)"+newTmp)
			}
			if variable.PrimitiveType == "Bool" {
				trueLabel := Generator.NewLabel()
				falseLabel := Generator.NewLabel()
				Generator.AddIf(newTmp2, "1", "==", trueLabel)
				Generator.AddGoto(falseLabel)

				result.TrueLabel = append(result.TrueLabel, trueLabel)
				result.FalseLabel = append(result.FalseLabel, falseLabel)

			}
			return Value{value: variable.value, temp: newTmp2, Tipo: variable.PrimitiveType, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
		}

		current = current.parent
	}
	return Value{value: nil}
}

// Funciones para asignaciones
func (v *Visitor) VisitVar_type(ctx *parser.Var_typeContext) Value {
	return Value{value: ctx.GetText()}
}
func (v *Visitor) VisitPrimitivo(ctx *parser.PrimitivoContext) Value {
	return Value{value: ctx.GetText()}
}

func (v *Visitor) VisitBoolExpr(ctx *parser.BoolExprContext) Value {
	value, _ := strconv.ParseBool(ctx.GetText())

	Generator.AddComment("Primitivo Bool")
	trueLabel := Generator.NewLabel()
	falseLabel := Generator.NewLabel()
	if value {
		Generator.AddGoto(trueLabel)
	} else {
		Generator.AddGoto(falseLabel)
	}
	result := Value{}
	result.TrueLabel = append(result.TrueLabel, trueLabel)
	result.FalseLabel = append(result.FalseLabel, falseLabel)
	return Value{value: value, temp: value, Tipo: "Bool", TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
}
func (v *Visitor) VisitTransferSentence(ctx *parser.Transfer_sentenceContext) Value {
	return Value{value: ctx.GetText()}
}
func (v *Visitor) VisitOneExpr(ctx *parser.OneExprContext) Value {
	exprValue := v.Visit(ctx.Expr()).value.(interface{})
	v.tmpList = append(v.tmpList, exprValue)
	return Value{value: true}
}

func (v *Visitor) VisitListaParamsCall(ctx *parser.ListaParamsCallContext) Value {
	paramID := ""
	Parametros := make([]param, 0)
	varID := ""
	if ctx.AllExpr() != nil {
		for index := 0; index < len(ctx.AllExpr()); index++ {
			if ctx.ID(index) != nil {
				paramID = ctx.ID(index).GetText()
			}
			if ctx.GetRef() != nil {
				varID = ctx.Expr(index).GetText()
			}

			exprValue := v.Visit(ctx.Expr(index)).value
			primiteType := getType(exprValue)
			param := param{
				IDInterno:     paramID,
				value:         exprValue,
				primitiveType: primiteType,
				idVarToInout:  varID,
			}
			Parametros = append(Parametros, param)
		}
	}

	return Value{value: Parametros}
}

func (v *Visitor) VisitOneParam(ctx *parser.OneParamContext) Value {
	IDExterno := ""
	IDInterno := ""
	hasInout := false
	primitiveType := v.Visit(ctx.Primitivo()).value.(string)
	if ctx.GetExt() != nil {
		IDExterno = ctx.GetExt().GetText()
		if IDExterno == "_" {
			IDInterno = ctx.ID(0).GetText()
		} else {
			IDInterno = ctx.ID(1).GetText()
		}

	}
	if ctx.GetIno() != nil {
		hasInout = true
	}

	param := param{
		IDExterno:     IDExterno,
		IDInterno:     IDInterno,
		Inout:         hasInout,
		primitiveType: primitiveType,
	}

	v.tmpListParams = append(v.tmpListParams, param)
	return Value{value: true}
}

func (v *Visitor) VisitNumParams(ctx *parser.NumParamsContext) Value {
	IDExterno := ""
	IDInterno := ""
	hasInout := false
	isVector := false
	primitiveType := v.Visit(ctx.Primitivo()).value.(string)
	if ctx.GetExt() != nil {
		IDExterno = ctx.GetExt().GetText()
		if IDExterno == "_" {
			IDInterno = ctx.ID(0).GetText()
		} else {
			IDInterno = ctx.ID(1).GetText()
		}

	}
	if ctx.GetIno() != nil {
		hasInout = true
	}
	if ctx.GetIsVector() != nil {
		isVector = true
	}
	firstParam := param{
		IDExterno:     IDExterno,
		IDInterno:     IDInterno,
		Inout:         hasInout,
		primitiveType: primitiveType,
		isVector:      isVector,
	}

	v.tmpListParams = append(v.tmpListParams, firstParam)
	// Visita el resto de los parámetros, si existen
	if ctx.ListaParams() != nil {
		v.Visit(ctx.ListaParams())
	}

	return Value{value: true}
}

func (v *Visitor) VisitNumExpr(ctx *parser.NumExprContext) Value {
	exprValue := v.Visit(ctx.Expr()).value.(interface{})
	v.tmpList = append(v.tmpList, exprValue) // Add value to tmpList
	if ctx.ListaExp() != nil {
		v.Visit(ctx.ListaExp())
	}
	return Value{value: true}
}
func (v *Visitor) VisitToString(ctx *parser.ToStringContext) Value {
	value := v.Visit(ctx.Expr()).value

	switch v := value.(type) {
	case int64:
		value = fmt.Sprintf("%d", v)
	case float64:
		value = fmt.Sprintf("%f", v)

	case string:
		break
	case rune:
		value = asciiToChar(value)

	case bool:
		value = fmt.Sprintf("%t", v)

	case nil:
		value = "nil"
	default:
		fmt.Println("Tipo desconocido")
	}
	return Value{value: value.(string)}
}
func (v *Visitor) VisitToFloat(ctx *parser.ToFloatContext) Value {
	value := v.Visit(ctx.Expr()).value

	floatValue, err := strconv.ParseFloat(value.(string), 64)
	if err != nil {
		v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "No pudo realizarse el truncamiento a float"))
		AgregarErrorSemantico("No pudo realizarse el truncamiento a float", "0", "0")
		return Value{value: nil} // O manejar el error de otra manera
	}
	value = floatValue
	return Value{value: value} //
}
func (v *Visitor) VisitToInt(ctx *parser.ToIntContext) Value {
	value := v.Visit(ctx.Expr()).value

	switch value.(type) {
	case float64:
		value = int64(value.(float64))
	case string:
		intValue, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			v.outputBuilder.WriteString(fmt.Sprintf("%v\n", "No pudo realizarse el truncamiento a int"))
			AgregarErrorSemantico("No pudo realizarse el truncamiento a int", "0", "0")

			return Value{value: nil} // O manejar el error de otra manera
		}
		value = intValue
	default:
		fmt.Println("Tipo desconocido")
	}
	return Value{value: value}
}

func (v *Visitor) VisitReturn(ctx *parser.ReturnstmtContext) Value {
	v.returnFlag = true
	expr := v.Visit(ctx.Expr()).value

	v.returnExpr = expr
	return Value{value: expr}
}
func (v *Visitor) VisitParExpr(ctx *parser.ParExprContext) Value {
	return v.Visit(ctx.Expr())

}
func (v *Visitor) VisitNotExpr(ctx *parser.NotExprContext) Value {
	info := v.Visit(ctx.Expr())
	value := info.value
	if val, ok := value.(bool); ok {

		newValue := Value{}
		newValue.TrueLabel = append(info.FalseLabel, newValue.TrueLabel...)
		newValue.FalseLabel = append(info.TrueLabel, newValue.FalseLabel...)
		return Value{value: val, TrueLabel: newValue.TrueLabel, FalseLabel: newValue.FalseLabel}

	}
	return Value{value: false}
}

func (v *Visitor) VisitNegExpr(ctx *parser.NegExprContext) Value {
	expr := v.Visit(ctx.Expr()).value

	return Value{value: 0 - expr.(int64)}
}
func (v *Visitor) VisitOpExpr(ctx *parser.OpExprContext) Value {

	op := ctx.GetOp().GetText()
	switch op {
	case "+":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		newTmp := Generator.NewTemp()
		Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "+")
		if type_left == "Float" && type_right == "Float" {
			return Value{value: l.(float64) + r.(float64), temp: newTmp}
		} else if type_left == "String" && type_right == "String" {
			return Value{value: l.(string) + r.(string), temp: newTmp}
		} else if type_left == "Int" && type_right == "Int" {
			return Value{value: l.(int64) + r.(int64), temp: newTmp}
		} else if type_left == "Float" && type_right == "Int" {
			return Value{value: l.(float64) + r.(float64), temp: newTmp}
		} else if type_left == "Int" && type_right == "Float" {
			return Value{value: l.(float64) + r.(float64), temp: newTmp}
		}
	case "-":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		newTmp := Generator.NewTemp()

		if type_left == "Float" && type_right == "Float" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "-")
			return Value{value: l.(float64) - r.(float64), temp: newTmp}
		} else if type_left == "Int" && type_right == "Int" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "-")
			return Value{value: l.(int64) - r.(int64), temp: newTmp}
		} else if type_left == "Float" && type_right == "Int" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "-")
			return Value{value: l.(float64) - r.(float64), temp: newTmp}
		} else if type_left == "Int" && type_right == "Float" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "-")
			return Value{value: l.(float64) - r.(float64), temp: newTmp}
		}
	case "*":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		newTmp := Generator.NewTemp()
		if type_left == "Float" && type_right == "Float" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "*")
			return Value{value: l.(float64) * r.(float64), temp: newTmp}
		} else if type_left == "Int" && type_right == "Int" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "*")
			return Value{value: l.(int64) * r.(int64), temp: newTmp}
		} else if type_left == "Float" && type_right == "Int" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "*")
			return Value{value: l.(float64) * r.(float64), temp: newTmp}
		} else if type_left == "Int" && type_right == "Float" {
			Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "*")
			return Value{value: l.(float64) * r.(float64), temp: newTmp}
		}
	case "/":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		newTmp := Generator.NewTemp()
		lvl1 := Generator.NewLabel()
		lvl2 := Generator.NewLabel()
		Generator.AddIf(tmpr_der, "0", "!=", lvl1)
		Generator.AddPrintf("c", "77")
		Generator.AddPrintf("c", "97")
		Generator.AddPrintf("c", "116")
		Generator.AddPrintf("c", "104")
		Generator.AddPrintf("c", "69")
		Generator.AddPrintf("c", "114")
		Generator.AddPrintf("c", "114")
		Generator.AddPrintf("c", "111")
		Generator.AddPrintf("c", "114")
		Generator.AddExpression(newTmp, "0", "", "")
		Generator.AddGoto(lvl2)
		Generator.AddLabel(lvl1)
		Generator.AddExpression(newTmp, tmp_izq, tmpr_der, "/")
		Generator.AddLabel(lvl2)
		if l.(int64) == 0 || r.(int64) == 0 {
			return Value{value: 0, temp: newTmp}
		}
		if type_left == "Float" && type_right == "Float" {

			return Value{value: l.(float64) / r.(float64), temp: newTmp}

		} else if type_left == "Int" && type_right == "Int" {
			return Value{value: l.(int64) / r.(int64), temp: newTmp}
		} else if type_left == "Float" && type_right == "Int" {
			return Value{value: l.(float64) / r.(float64), temp: newTmp}
		} else if type_left == "Int" && type_right == "Float" {
			return Value{value: l.(float64) / r.(float64), temp: newTmp}
		}
	case "%":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		newTmp := Generator.NewTemp()
		lvl1 := Generator.NewLabel()
		lvl2 := Generator.NewLabel()

		Generator.AddIf(tmpr_der, "0", "!=", lvl1) // Verifica si el divisor es cero
		Generator.AddPrintf("c", "77")
		Generator.AddPrintf("c", "97")
		Generator.AddPrintf("c", "116")
		Generator.AddPrintf("c", "104")
		Generator.AddPrintf("c", "69")
		Generator.AddPrintf("c", "114")
		Generator.AddPrintf("c", "114")
		Generator.AddPrintf("c", "111")
		Generator.AddPrintf("c", "114")
		Generator.AddExpression(newTmp, "0", "", "")
		Generator.AddGoto(lvl2)
		Generator.AddLabel(lvl1) // Inicia el código para el caso donde el divisor no es cero
		Generator.AddExpression(newTmp, "(int)"+tmp_izq, tmpr_der, "%")

		Generator.AddLabel(lvl2)
		if type_left == "Int" && type_right == "Int" {
			return Value{value: l.(int64) % r.(int64), temp: newTmp}
		}
		//Menor que  o menor igual que
	case "<":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		Generator.AddIf(tmp_izq, tmpr_der, "<", trueLabel)
		Generator.AddGoto(falseLabel)
		result := Value{}
		result.TrueLabel = append(result.TrueLabel, trueLabel)
		result.FalseLabel = append(result.FalseLabel, falseLabel)
		if type_left == "Int" && type_right == "Int" {
			if l.(int64) < r.(int64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Float" && type_right == "Float" {
			if l.(float64) < r.(float64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "String" && type_right == "String" {
			if len(l.(string)) < len(r.(string)) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		}
	case "<=":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		Generator.AddIf(tmp_izq, tmpr_der, "<=", trueLabel)
		Generator.AddGoto(falseLabel)
		result := Value{}
		result.TrueLabel = append(result.TrueLabel, trueLabel)
		result.FalseLabel = append(result.FalseLabel, falseLabel)
		if type_left == "Int" && type_right == "Int" {
			if l.(int64) <= r.(int64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Float" && type_right == "Float" {
			if l.(float64) <= r.(float64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "String" && type_right == "String" {
			if len(l.(string)) <= len(r.(string)) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		}
	case ">":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		Generator.AddIf(tmp_izq, tmpr_der, ">", trueLabel)
		Generator.AddGoto(falseLabel)
		result := Value{}
		result.TrueLabel = append(result.TrueLabel, trueLabel)
		result.FalseLabel = append(result.FalseLabel, falseLabel)

		if type_left == "Int" && type_right == "Int" {
			if l.(int64) > r.(int64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Float" && type_right == "Float" {
			if l.(float64) > r.(float64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "String" && type_right == "String" {
			if len(l.(string)) > len(r.(string)) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		}
	case ">=":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		Generator.AddIf(tmp_izq, tmpr_der, ">=", trueLabel)
		Generator.AddGoto(falseLabel)
		result := Value{}
		result.TrueLabel = append(result.TrueLabel, trueLabel)
		result.FalseLabel = append(result.FalseLabel, falseLabel)
		if type_left == "Int" && type_right == "Int" {
			if l.(int64) >= r.(int64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Float" && type_right == "Float" {
			if l.(float64) >= r.(float64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "String" && type_right == "String" {
			if len(l.(string)) >= len(r.(string)) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		}
		//----------------------------------------------------------------
	case "==":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		Generator.AddIf(tmp_izq, tmpr_der, "==", trueLabel)
		Generator.AddGoto(falseLabel)
		result := Value{}
		result.TrueLabel = append(result.TrueLabel, trueLabel)
		result.FalseLabel = append(result.FalseLabel, falseLabel)
		if type_left == "Int" && type_right == "Int" {
			if l.(int64) == r.(int64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Float" && type_right == "Float" {
			if l.(float64) == r.(float64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "String" && type_right == "String" {
			if l.(string) == r.(string) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Bool" && type_right == "Bool" {
			if l.(bool) == r.(bool) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else {
			fmt.Print("Los operandos no coinciden")
			return Value{value: false}
		}
	//----------------------------------------------------------------
	case "!=":
		left := v.Visit(ctx.GetLeft())
		right := v.Visit(ctx.GetRight())
		l := left.value
		r := right.value
		type_left := getType(left.value)
		type_right := getType(left.value)
		tmp_izq := fmt.Sprintf("%v", left.temp)
		tmpr_der := fmt.Sprintf("%v", right.temp)
		trueLabel := Generator.NewLabel()
		falseLabel := Generator.NewLabel()
		Generator.AddIf(tmp_izq, tmpr_der, "!=", trueLabel)
		Generator.AddGoto(falseLabel)
		result := Value{}
		result.TrueLabel = append(result.TrueLabel, trueLabel)
		result.FalseLabel = append(result.FalseLabel, falseLabel)
		if type_left == "Int" && type_right == "Int" {
			if l.(int64) != r.(int64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Float" && type_right == "Float" {
			if l.(float64) != r.(float64) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "String" && type_right == "String" {
			if l.(string) != r.(string) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else if type_left == "Bool" && type_right == "Bool" {
			if l.(bool) != r.(bool) {
				return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			} else {
				return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
			}
		} else {
			fmt.Print("Los operandos no coinciden")
			return Value{value: false}
		}
	case "&&":
		left := v.Visit(ctx.GetLeft())
		l := left.value

		for _, lvl := range left.TrueLabel {
			Generator.AddLabel(lvl)
		}
		right := v.Visit(ctx.GetRight())
		r := right.value
		result := Value{}
		result.TrueLabel = append(right.TrueLabel, result.TrueLabel...)
		result.FalseLabel = append(left.FalseLabel, result.FalseLabel...)
		result.FalseLabel = append(right.FalseLabel, result.FalseLabel...)
		if l.(bool) && r.(bool) {
			return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
		} else {
			return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
		}
	case "||":
		left := v.Visit(ctx.GetLeft())
		l := left.value

		for _, lvl := range left.FalseLabel {
			Generator.AddLabel(lvl)
		}
		right := v.Visit(ctx.GetRight())
		r := right.value
		result := Value{}
		result.TrueLabel = append(left.TrueLabel, result.TrueLabel...)
		result.TrueLabel = append(right.TrueLabel, result.TrueLabel...)
		result.FalseLabel = append(right.FalseLabel, result.FalseLabel...)
		if l.(bool) || r.(bool) {
			return Value{value: true, temp: true, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
		} else {
			return Value{value: false, temp: false, TrueLabel: result.TrueLabel, FalseLabel: result.FalseLabel}
		}

	}

	return Value{value: false}

}

func getType(value interface{}) string {
	switch v := value.(type) {
	case bool:
		return "Bool"
	case int:
		return "Int"
	case int64:
		return "Int"
	case string:
		return "String"
	case float64:
		return "Float"
	case rune:
		return "Character"
	case nil:
		return "nil"
	default:
		fmt.Println("valor de ", v, " desconocido")
		return "Unknown"
	}
}
func (v *Visitor) encontrarVariable(varID string) bool {
	current := v.currentEnv
	for current != nil {
		_, exists := current.variables[varID]
		if exists {
			return true
		}
		current = current.parent
	}
	return false
}
func (v *Visitor) encontrarFunc(varID string) bool {
	current := v.currentEnv
	for current != nil {
		_, exists := current.functions[varID]
		if exists {
			return true
		}
		current = current.parent
	}
	return false
}

func (v *Visitor) getContentVariable(varID string) Value {
	current := v.currentEnv
	for current != nil {
		if variable, ok := current.variables[varID]; ok {
			return Value{value: variable.value}
		}
		current = current.parent
	}
	return Value{value: nil}
}

func (v *Visitor) DoesVarExist(varID string) bool {

	_, exists := v.currentEnv.variables[varID]
	return exists
}
func (v *Visitor) DoesFuncExist(varID string) bool {

	_, exists := v.currentEnv.functions[varID]
	return exists
}
func asciiToChar(ascii interface{}) string {
	return string(ascii.(int32))
}

func main() {

	app := fiber.New()
	app.Use(cors.New())
	app.Post("/analyze", func(c *fiber.Ctx) error {
		Cont_Heap = 0
		Cont_Stack = 0
		Generator.Reset()
		Generator.MainCode = true
		SemanticErrors = nil
		LexerErrors = nil
		SyntaxErrors = nil
		Symbols = nil
		prog := string(c.Body())

		input := antlr.NewInputStream(prog)
		lexer := parser.NewControlLexer(input)
		getLexerErrors := &LexerError{}
		lexer.AddErrorListener(getLexerErrors)
		tokens := antlr.NewCommonTokenStream(lexer, 0)
		getSyntaxsErrors := &SyntaxErrs{}
		p := parser.NewControlParser(tokens)
		p.RemoveErrorListeners()
		p.AddErrorListener(getSyntaxsErrors)
		p.BuildParseTrees = true
		tree := p.Prog()
		eval := Visitor{memory: make(map[string]Value)}
		eval.Visit(tree)
		//graph.CreateGraph(prog)

		Generator.GenerateFinalCode()
		var ConsoleOut = ""

		for _, item := range Generator.GetFinalCode() {
			ConsoleOut += item.(string)
		}
		return c.JSON(fiber.Map{
			"prints":         ConsoleOut,
			"LexerErrors":    LexerErrors,
			"SyntaxErrors":   SyntaxErrors,
			"SemanticErrors": SemanticErrors,
			"Symbols":        Symbols,
		})

	})

	app.Get("/graph", func(c *fiber.Ctx) error {
		// Lee el archivo SVG
		svgContent, err := ioutil.ReadFile("cst_graph_tree.svg")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error al leer el archivo SVG")
		}

		fmt.Println("SVG enviado...")
		return c.Type("svg").Send(svgContent)
	})

	app.Listen(":8080")
}
