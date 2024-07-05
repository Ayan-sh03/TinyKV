package main

import (
	"log"
	"strconv"
	"sync"
)

var Handlers = map[string]func([]Value) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hsetHT,
	"HGET":    hgetHT,
	"HGETALL": hgetallHT,
	"HDEL":    hdelHT,
	"INCR":    incr,
	"DECR":    decr,
	"INCRBY":  incrBy,
	"DECRBY":  decrBy,
	"DEL":     del,
	"APPEND":  appendto,
	"LPUSH":   Lpush,
	"LRANGE":  Lrange,
	"LPOP":    Lpop,
	"RPOP":    Rpop,
	"RPUSH":   Rpush,
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG Mine"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].bulk
	value := args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func appendto(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'appendto' command"}
	}

	key := args[0].bulk
	value := args[1].bulk

	SETsMu.Lock()
	SETs[key] += value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func del(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'del' command"}
	}

	key := args[0].bulk

	SETsMu.Lock()
	delete(SETs, key)
	SETsMu.Unlock()

	return Value{typ: "string", str: "ok"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

var INCRsMU = sync.RWMutex{}

func incr(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'incr' command"}
	}

	key := args[0].bulk

	val := SETs[key]

	//convert val to integer
	i, err := strconv.Atoi(val)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}
	INCRsMU.Lock()
	i += 1
	SETs[key] = strconv.Itoa(i)
	INCRsMU.Unlock()

	return Value{typ: "string", str: "OK"}

}
func decr(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'decr' command"}
	}

	key := args[0].bulk

	val := SETs[key]

	//convert val to integer
	i, err := strconv.Atoi(val)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}
	INCRsMU.Lock()
	i -= 1
	SETs[key] = strconv.Itoa(i)
	INCRsMU.Unlock()

	return Value{typ: "string", str: "OK"}

}
func incrBy(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'incrby' command"}
	}

	key := args[0].bulk
	incrementval := args[1].bulk

	//convert incrementval to integer
	increment, err := strconv.Atoi(incrementval)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}

	val := SETs[key]

	//convert val to integer
	i, err := strconv.Atoi(val)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}

	INCRsMU.Lock()
	i += increment
	SETs[key] = strconv.Itoa(i)
	INCRsMU.Unlock()

	return Value{typ: "string", str: "OK"}

}
func decrBy(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'decrby' command"}
	}

	key := args[0].bulk
	decrementVal := args[1].bulk
	// convert decrementVal to int
	decrement, err := strconv.Atoi(decrementVal)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}

	val := SETs[key]

	//convert val to integer
	i, err := strconv.Atoi(val)
	if err != nil {
		return Value{typ: "error", str: "ERR: value is not an integer"}
	}
	INCRsMU.Lock()
	i -= decrement
	SETs[key] = strconv.Itoa(i)
	INCRsMU.Unlock()

	return Value{typ: "string", str: "OK"}

}

var hashTable = NewHashTable(100)

func hsetHT(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}

	hashKey := args[0].bulk
	field := args[1].bulk
	val := args[2].bulk

	hashTable.Set(hashKey, field, val)

	return Value{typ: "string", str: "OK"}

}

func hgetHT(args []Value) Value {

	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}
	}

	hashKey := args[0].bulk
	field := args[1].bulk

	val, ok := hashTable.Get(hashKey, field)
	log.Println(val)
	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: val}
}

func hdelHT(args []Value) Value {

	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hdel' command"}
	}

	hashKey := args[0].bulk
	field := args[1].bulk

	hashTable.Delete(hashKey, field)

	return Value{typ: "string", str: "OK"}
}

func hgetallHT(args []Value) Value {

	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	hash := args[0].bulk

	val, ok := hashTable.GetAll(hash)
	log.Println(val)
	if !ok {
		return Value{typ: "null"}
	}

	res := []Value{}

	for k, v := range val {

		res = append(res, Value{typ: "bulk", bulk: k})
		res = append(res, Value{typ: "bulk", bulk: v})

	}

	return Value{typ: "array", array: res}

}
