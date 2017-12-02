package scripting

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLua(t *testing.T) {
	var script luaScript
	script.argPairs = []argPair{}
	script.returns = []string{"int"}
	script.script = `function abc(i) print("function happened") return i end print("Something happened") return abc(5) `
	script.args = convertToInterface(script.argPairs)

	var script2 luaScript
	script2.argPairs = []argPair{argPair{name: "i", value: "4"}}
	script2.returns = []string{"int", "int", "string", "float"}
	script2.script = `function abc(n) return n end return 4,5,"six",7.1 `
	script2.args = convertToInterface(script2.argPairs)

	var script3 luaScript
	script3.argPairs = []argPair{argPair{name: "i", value: "5"}}
	script3.returns = []string{"int"}
	script3.script = `function abc(f) return f end return abc(i)`
	script3.args = convertToInterface(script3.argPairs)

	if !validReturns(script3.returns) {
		fmt.Println("Failed returns")
		return
	}

	v3, _ := callScript(&script3)

	fmt.Printf("%+v\n", script)

	err := verifyResults(v3, script3.returns)
	if err != nil {
		fmt.Println("Failed")
	} else {
		fmt.Println("Okay")
	}

	for k, v := range v3 {
		fmt.Println(k)
		fmt.Println(v)
		fmt.Println(reflect.TypeOf(v))
	}

	v1, _ := callScript(&script)
	v2, _ := callScript(&script2)

	t.Logf("%v\n%v", v1, v2)

	err = verifyResults(v2, script2.returns)
	if err != nil {
		fmt.Println("Failed")
		return
	}
	fmt.Println("Okay")

	for k, v := range v2 {
		fmt.Println(k)
		fmt.Println(v)
		fmt.Println(reflect.TypeOf(v))
	}
}
