### How to use the lua scripting

The code block used must be specified with `lua` syntax
````
%lua return bool
```lua
return false
```
````


Any return values must be specified in the command
````
%lua return string
```lua
return "this is a string"
```
````
Multiple values may be returned 
````
%lua return int bool
```lua
return 1, true
```
````
Valid return types are `int`,`bool`,`float`, and `string`


The lua command optionally takes in variables to use in the script in the form `name=value`
````
%lua a=123 b=false c=test d=1.2 return int bool string float
```lua
return a,b,c,d
```
````

#### String argument bug
String arguments cannot include a space right now

