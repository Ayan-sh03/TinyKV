package main

import (
	"strconv"
	"sync"
)

var SETsL = map[string][]string{}
var SETLsMu = sync.RWMutex{}

func Lpush(args []Value) Value {
	if len(args) < 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'lpush' command"}
	}

	key := args[0].bulk
	values := make([]string, len(args)-1)

	// Collect values to push into a slice
	size := len(args) - 1
	for i := 0; i < size; i++ {
		values[size-i-1] = args[i+1].bulk
	}

	SETLsMu.Lock()
	defer SETLsMu.Unlock()

	// Append values to the beginning of SETsL[key]
	SETsL[key] = append(values, SETsL[key]...)

	return Value{typ: "string", str: "OK"}
}

func Lrange(args []Value) Value {
	if len(args) < 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'lrange' command"}
	}

	key := args[0].bulk

	start := args[1].bulk
	end := args[2].bulk

	//convert start and end to int
	startInt, err := strconv.Atoi(start)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}
	endInt, err := strconv.Atoi(end)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}

	SETLsMu.RLock()

	value, ok := SETsL[key]

	SETLsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	if startInt < 0 {
		startInt = 0
	}
	if endInt >= len(value) || endInt < 0 {
		endInt = len(value) - 1
	}

	if startInt > endInt {
		return Value{typ: "null"}
	}
	result := make([]Value, endInt-startInt+1)

	for i := startInt; i <= endInt; i++ {
		result[i-startInt] = Value{typ: "bulk", bulk: value[i]}
	}

	return Value{typ: "array", array: result}

}

func Rpush(args []Value) Value {
	if len(args) < 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'rpush' command"}
	}

	key := args[0].bulk
	values := []string{}

	for i := 0; i < len(args)-1; i++ {
		values = append(values, args[i+1].bulk)
	}

	SETLsMu.Lock()
	SETsL[key] = append(SETsL[key], values...)
	SETLsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func Lpop(args []Value) Value {

	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'lpop' command"}
	}

	key := args[0].bulk
	SETLsMu.Lock()

	value, ok := SETsL[key]
	res := value[0]
	SETsL[key] = value[1:]

	SETLsMu.Unlock()
	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: res}
}
func Rpop(args []Value) Value {

	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'lpop' command"}
	}

	key := args[0].bulk

	SETLsMu.Lock()

	value, ok := SETsL[key]
	res := value[len(value)-1]
	SETsL[key] = value[:len(value)-1]
	SETLsMu.Unlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: res}
}
