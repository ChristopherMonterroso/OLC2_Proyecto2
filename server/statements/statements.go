package statements

import (
	"Server/generator"
	"fmt"
	"regexp"
	"strconv"
)

func PrintEnteros(value interface{}, gen *generator.Generator) {
	gen.AddPrintf("d", "(int)"+fmt.Sprintf("%v", value))
	gen.AddPrintf("c", "10")
	gen.AddBr()

}
func PrintString(value interface{}, gen *generator.Generator) {
	// Llama a la función que genera el código para imprimir strings
	gen.GeneratePrintString()

	// Agrega código en el main para imprimir el string
	newTemp1 := gen.NewTemp()
	newTemp2 := gen.NewTemp()
	size := strconv.Itoa(len(value.(string)))
	gen.AddExpression(newTemp1, "P", size, "+")       // Nuevo temporal en posición vacía
	gen.AddExpression(newTemp1, newTemp1, "1", "+")   // Se deja espacio de retorno
	gen.AddSetStack("(int)"+newTemp1, value.(string)) // Se coloca el string en el parámetro que se manda
	gen.AddExpression("P", "P", size, "+")            // Cambio de entorno
	gen.AddCall("dbrust_printString")                 // Llamada a la función de impresión de strings
	gen.AddGetStack(newTemp2, "(int)P")               // Obtención del retorno
	gen.AddExpression("P", "P", size, "-")            // Regreso del entorno
	gen.AddPrintf("c", "10")                          // Salto de línea
	gen.AddBr()
}

func PrintBoolean(value interface{}, gen *generator.Generator) {

}

func OpSuma(a interface{}, b interface{}, gen *generator.Generator) {
	newTemp := gen.NewTemp()
	gen.AddExpression(newTemp, a.(string), b.(string), "+")
	gen.AddBr()
}

func opDivision(a interface{}, b interface{}, gen *generator.Generator) {
	newTemp := gen.NewTemp()
	gen.AddExpression(newTemp, a.(string), b.(string), "/")
	gen.AddBr()
}
func SumarLabel(input string) string {
	re := regexp.MustCompile(`L(\d+)`)

	// Encontrar la coincidencia en la cadena
	match := re.FindStringSubmatch(input)

	if len(match) == 2 {
		// Convertir el número a entero
		numero, err := strconv.Atoi(match[1])
		if err != nil {
			fmt.Println("Error al convertir el número a entero:", err)
			return input
		}

		// Sumar 1 al número
		numero += 1

		// Construir el nuevo string
		nuevoString := fmt.Sprintf("L%d", numero)
		fmt.Println("Nuevo string:", nuevoString)
		return nuevoString
	} else {
		fmt.Println("No se encontró un formato válido.")
		return input
	}
}
