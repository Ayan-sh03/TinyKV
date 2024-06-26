package main

import (
	"fmt"
	"log"
	"testing"
)

func TestLpush(t *testing.T) {
	// Reset the global state for tests
	log.Println("Testing LPush")
	// Test cases
	
	for k := range SETsL {
		delete(SETsL, k)
	}

	var tests = []struct {
		name     string
		args     []Value
		want     Value
		wantList []string
	}{
		{
			name:     "Single value push",
			args:     []Value{{bulk: "mylist"}, {bulk: "value1"}},
			want:     Value{typ: "string", str: "OK"},
			wantList: []string{"value1"},
		},
		{
			name:     "Multiple values push",
			args:     []Value{{bulk: "mylist"}, {bulk: "value2"}, {bulk: "value3"}},
			want:     Value{typ: "string", str: "OK"},
			wantList: []string{"value3", "value2", "value1"},
		},
		{
			name: "Wrong number of arguments",
			args: []Value{{bulk: "mylist"}},
			want: Value{typ: "error", str: "ERR wrong number of arguments for 'lpush' command"},
		},
	}
	

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tt.name)
			t.Logf("Initial state: %v", SETsL)

			got := Lpush(tt.args).str
			if got != tt.want.str {
				t.Errorf("Lpush() = %v, want %v", got, tt.want)
			}
			if tt.want.typ == "string" {
				SETLsMu.Lock()
				if list, exists := SETsL[tt.args[0].bulk]; exists {
					if !equal(list, tt.wantList) {
						t.Errorf("SETsL[%v] = %v, want %v", tt.args[0].bulk, list, tt.wantList)
					}
				} else {
					t.Errorf("SETsL[%v] does not exist", tt.args[0].bulk)
				}
				SETLsMu.Unlock()
			}
		})
	}
}

func equal(a, b []string) bool {

	log.Println(len(a),len(b))

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}



func BenchmarkLpush(b *testing.B) {
	// Benchmark case: Pushing to an empty list
	key := "benchmarklist"
	value := "value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SETsL = make(map[string][]string)
		Lpush([]Value{{bulk: key}, {bulk: value}})
	}
	b.StopTimer()
}

func BenchmarkLpushMultipleValues(b *testing.B) {
	// Benchmark case: Pushing multiple values to a list
	key := "benchmarklist"
	values := []Value{{bulk: key}, {bulk: "value1"}, {bulk: "value2"}, {bulk: "value3"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SETsL = make(map[string][]string)
		Lpush(values)
	}
	b.StopTimer()
}

func BenchmarkLpushLarge(b *testing.B) {
	// Benchmark case: Pushing a large number of values to a list
	key := "benchmarklist"
	values := []Value{{bulk: key}}
	for i := 0; i < 1000; i++ {
		values = append(values, Value{bulk: fmt.Sprintf("value%d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SETsL = make(map[string][]string)
		Lpush(values)
	}
	b.StopTimer()
}
func BenchmarkLpushExtraLarge(b *testing.B) {
	// Benchmark case: Pushing a large number of values to a list
	key := "benchmarklist"
	values := []Value{{bulk: key}}
	for i := 0; i < 100_000; i++ {
		values = append(values, Value{bulk: fmt.Sprintf("value%d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SETsL = make(map[string][]string)
		Lpush(values)
	}
	b.StopTimer()
}


func TestRpush(t *testing.T) {
	// Initialize the map


	tests := []struct {
		name     string
		args     []Value
		want     Value
		wantList []string
	}{
		{
			name:     "Single value push",
			args:     []Value{{bulk: "mylist"}, {bulk: "value1"}},
			want:     Value{typ: "string", str: "OK"},
			wantList: []string{"value1"},
		},
		{
			name:     "Multiple values push",
			args:     []Value{{bulk: "mylist"}, {bulk: "value2"}, {bulk: "value3"}},
			want:     Value{typ: "string", str: "OK"},
			wantList: []string{"value2", "value3"},
		},
		{
			name: "Wrong number of arguments",
			args: []Value{{bulk: "mylist"}},
			want: Value{typ: "error", str: "ERR wrong number of arguments for 'rpush' command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the map before each test
			for k := range SETsL {
				delete(SETsL, k)
			}

			t.Logf("Running test case: %s", tt.name)
			t.Logf("Initial state: %v", SETsL)

			got := Rpush(tt.args)

			if got.typ != tt.want.typ || got.str != tt.want.str {
				t.Errorf("Rpush() = %v, want %v", got, tt.want)
			}

			if tt.want.typ == "string" {
				SETLsMu.Lock()
				if list, exists := SETsL[tt.args[0].bulk]; exists {
					if !equal(list, tt.wantList) {
						t.Errorf("SETsL[%v] = %v, want %v", tt.args[0].bulk, list, tt.wantList)
					}
				} else {
					t.Errorf("SETsL[%v] does not exist", tt.args[0].bulk)
				}
				SETLsMu.Unlock()
			}
		})
	}
}
func BenchmarkRpush(b *testing.B){
	key := "benchmarklist"
	value := "value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SETsL = make(map[string][]string)
		Rpush([]Value{{bulk: key}, {bulk: value}})
	}
	b.StopTimer()
}

func BenchmarkRpushMultipleValues(b *testing.B) {
	// Benchmark case: Pushing multiple values to a list
	key := "benchmarklist"
	values := []Value{{bulk: key}, {bulk: "value1"}, {bulk: "value2"}, {bulk: "value3"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SETsL = make(map[string][]string)
		Rpush(values)
	}
	b.StopTimer()
}

func BenchmarkRpushLarge(b *testing.B) {
	// Benchmark case: Pushing a large number of values to a list
	key := "benchmarklist"
	values := []Value{{bulk: key}}
	for i := 0; i < 1000; i++ {
		values = append(values, Value{bulk: fmt.Sprintf("value%d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SETsL = make(map[string][]string)
		Rpush(values)
	}
	b.StopTimer()
}