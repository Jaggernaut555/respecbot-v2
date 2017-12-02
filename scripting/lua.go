package scripting

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Jaggernaut555/respecbot-v2/logging"
	"github.com/Jaggernaut555/respecbot-v2/types"
	lua "github.com/Shopify/go-lua"
	"github.com/Shopify/goluago/util"
)

type argPair struct {
	name  string
	value string
}

type luaScript struct {
	script   string
	argPairs []argPair
	args     []interface{}
	returns  []string
}

/*
usage:
%lua [variables] return [return types]
```lua
script goes here
```
script must include lua tag, only returned values will be printed. returned values must align to type specified in command. Will run until either 500 instructions have gone by or 10MB of memory has been allocated. io and os commands shouldn't be able to do anything.

At least one return MUST be specified

variables must be `name=value`, either int, float, bool, or string
return types must be int/float/bool/string
*/

func Lua(api types.API, message *types.Message, args []string) {
	if len(args) < 1 {
		api.ReplyTo("Not enough arguments", message)
		return
	}
	script := getScript(args)
	if script == nil {
		api.ReplyTo("Invalid script", message)
		return
	}

	if !validReturns(script.returns) {
		api.ReplyTo("Not valid returns", message)
		return
	}

	returns, err := callScript(script)
	if err != nil {
		api.ReplyTo(err.Error(), message)
		logging.Err(err)
		return
	}

	err = verifyResults(returns, script.returns)
	if err != nil {
		api.ReplyTo(err.Error(), message)
		logging.Err(err)
		return
	}

	reply := fmt.Sprintf("%+v", returns)
	api.ReplyTo(reply, message)
}

func callScript(script *luaScript) (returnValues []interface{}, err error) {
	l := lua.NewState()

	lua.Require(l, "_G", lua.BaseOpen, true)
	l.Pop(1)
	//lua.Require(l, "package", lua.PackageOpen, true)
	//l.Pop(1)
	lua.Require(l, "string", lua.StringOpen, true)
	l.Pop(1)
	lua.Require(l, "table", lua.TableOpen, true)
	l.Pop(1)
	lua.Require(l, "math", lua.MathOpen, true)
	l.Pop(1)
	lua.Require(l, "bit32", lua.Bit32Open, true)
	l.Pop(1)

	/*
		Ability to save scripts to run later (with args too)

		if err := lua.LoadFile(l, "scripting/hello.lua", "text"); err != nil {
			logging.Err(err)
		}
	*/

	if err := lua.LoadString(l, script.script); err != nil {
		return nil, err
	}

	for k, v := range script.args {
		util.DeepPush(l, v)
		l.SetGlobal(script.argPairs[k].name)
	}

	var iCount int
	f := func(state *lua.State, activationRecord lua.Debug) {
		iCount += 10
		if activationRecord.Event == lua.HookCount {
			if iCount > 500 {
				lua.Errorf(state, "More than 500 instructions")
			}
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			if mem.TotalAlloc > 10000000 {
				lua.Errorf(state, "10 MB memory limit reached")
			}
			return
		}
		state.Error()
	}
	lua.SetDebugHook(l, f, lua.MaskCount, 10)

	if err := l.ProtectedCall(0, len(script.returns), 0); err != nil {
		return nil, err
	}

	res, err := util.PullVarargs(l, l.Top()-len(script.returns)+1)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func validReturns(args []string) bool {
	for _, v := range args {
		switch v {
		case "int":
		case "bool":
		case "float":
		case "string":
		default:
			return false
		}
	}
	return true
}

func getScript(args []string) *luaScript {
	var script luaScript
	var startFound, endFound, returnFound bool
	var start, end int
	for k, v := range args {
		if !returnFound && strings.Contains(v, "=") {
			s := strings.Split(v, "=")
			if len(s) < 2 {
				return nil
			}
			script.argPairs = append(script.argPairs, argPair{name: s[0], value: s[1]})
		} else if !returnFound && strings.Contains(v, "return") {
			returnFound = true
		} else if returnFound && !startFound && !strings.Contains(v, "```lua") {
			script.returns = append(script.returns, v)
		} else if !startFound && strings.Contains(v, "```lua") {
			start = k
			startFound = true
		} else if startFound && !endFound && strings.Contains(v, "```") {
			end = k
			endFound = true
		}
		// add an else { break }?
	}
	if !startFound || !endFound || (start >= end) || !returnFound || len(script.returns) == 0 {
		return nil
	}
	script.script = strings.Join(args[start+1:end], " ")
	script.args = convertToInterface(script.argPairs)
	return &script
}

// float/string/int/bool
func verifyResults(returns []interface{}, types []string) error {
	if len(returns) != len(types) {
		return fmt.Errorf("Did not return correct amount")
	}
	for k, v := range returns {
		switch v.(type) {
		case float64:
			if types[k] == "int" || types[k] == "float" {
			} else {
				return fmt.Errorf("Return value %v is not a(n) %v", v, types[k])
			}
		case string:
			if types[k] == "string" {
			} else {
				return fmt.Errorf("Return value %v is not a(n) %v", v, types[k])
			}
		case bool:
			if types[k] == "bool" {
			} else {
				return fmt.Errorf("Return value %v is not a(n) %v", v, types[k])
			}
		default:
			return fmt.Errorf("Returned value #%v is not valid", k+1)
		}
	}
	return nil
}

func convertToInterface(args []argPair) []interface{} {
	res := make([]interface{}, len(args))
	var i int
	var f float64
	var b bool
	var err error
	for k, v := range args {
		if i, err = strconv.Atoi(v.value); err == nil {
			res[k] = i
		} else if f, err = strconv.ParseFloat(v.value, 64); err == nil {
			res[k] = f
		} else if b, err = strconv.ParseBool(v.value); err == nil {
			res[k] = b
		} else {
			res[k] = v.value
		}
	}
	return res
}
