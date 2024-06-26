package main

import (
	"hash/maphash"
	"log"
	"sync"
)

type KeyValue struct {
	Key   string
	Field string
	Value string
}
type HashAlgorithm int

const (
	SHA256 HashAlgorithm = iota
	FNV1
)

type HashTable struct {
	mu       sync.RWMutex
	buckets  [][]*KeyValue
	size     int
	hashFunc [2]func(string) int
	seed     maphash.Seed
}

func NewHashTable(size int) *HashTable {
	ht := &HashTable{
		buckets: make([][]*KeyValue, size),
		size:    size,
		seed:    maphash.MakeSeed(),
	}

	ht.hashFunc[0] = ht.hashFunc1
	ht.hashFunc[1] = ht.hashFunc2
	return ht
}
func (cht *HashTable) hashFunc1(key string) int {
	var h maphash.Hash
	h.SetSeed(cht.seed)
	h.WriteString(key)
	return int(h.Sum64() % uint64(cht.size))
}

func (cht *HashTable) hashFunc2(key string) int {
	var h maphash.Hash
	h.SetSeed(cht.seed)
	h.WriteString(key)
	return int((h.Sum64() * 16777619) % uint64(cht.size))
}

func (ht *HashTable) Set(hashKey, field, val string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	kv := &KeyValue{Key: hashKey, Field: field, Value: val}

	for i := 0; i < 10; i++ { // limiting the number of relocations to avoid infinite loops
		bucketIndex := ht.hashFunc[i%2](kv.Key)

		// Use linear probing to find an empty slot in the bucket
		for j := bucketIndex; ; j = (j + 1) % len(ht.buckets) {
			bucket := ht.buckets[j]

			// Check if the bucket is empty
			if len(bucket) == 0 {
				ht.buckets[j] = []*KeyValue{kv}
				return
			}

			// Check if the key-value pair already exists in the bucket
			for k, existingKv := range bucket {
				if existingKv.Key == kv.Key && existingKv.Field == kv.Field {
					ht.buckets[j][k] = kv
					return
				}
			}
			
			
			if len(bucket) >= ht.size {
				bucket = bucket[1:] // Evict the oldest entry
			}

			// Swap the existing key-value pair with the new one
			bucket = append(bucket, kv)
			ht.buckets[j] = bucket
			kv = bucket[0]
			bucket = bucket[1:]
		}
	}

	// TODO: Resize or rehash the table
}


// func (ht *HashTable) Set(hashKey, field, val string) {
// 	ht.mu.Lock()
// 	defer ht.mu.Unlock()

// 	kv := &KeyValue{Key: hashKey, Field: field, Value: val}

// 	for i := 0; i < 10; i++ { //limiting the number of relocations to avoid infinite loops
// 		bucketIndex := ht.hashFunc[i%2](kv.Key)
// 		if ht.buckets[bucketIndex] == nil {
// 			ht.buckets[bucketIndex] = []*KeyValue{kv}
// 			return
// 		}

// 		for j, existingKv := range ht.buckets[bucketIndex] {
// 			if existingKv.Key == kv.Key && existingKv.Field == kv.Field {
// 				ht.buckets[bucketIndex][j] = kv
// 				return
// 			}
// 		}
// 		ht.buckets[bucketIndex] = append(ht.buckets[bucketIndex], kv)

// 		return
// 		// TODO: Resize or rehash the table

// 	}

// }

func (ht *HashTable) Get(hashKey, field string) (string, bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	for i := 0; i < 2; i++ {
		hash := ht.hashFunc[i](hashKey)
		if ht.buckets[hash] != nil {
			for _, kv := range ht.buckets[hash] {
				if kv.Key == hashKey && kv.Field == field {
					return kv.Value, true
				}
			}
		}
	}

	return "", false
}

func (ht *HashTable) Delete(hashKey, field string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()
	log.Println("Deleting key-value pair:", hashKey, field)
	for i := 0; i < 2; i++ {
		bucketIndex := ht.hashFunc[i](hashKey)
		log.Println("Bucket index:", bucketIndex)
		// Use linear probing to find the key-value pair in the bucket
		for j := bucketIndex; ; j = (j + 1) % len(ht.buckets) {
			bucket := ht.buckets[j]

			// Check if the bucket is empty
			if len(bucket) == 0 {
				
				return
			}

			// Check if the key-value pair exists in the bucket
			for k, existingKv := range bucket {
				if existingKv.Key == hashKey && existingKv.Field == field {
					// Remove the key-value pair from the bucket
					log.Println("Deleted key-value pair:", existingKv.Key, existingKv.Field)
					bucket = append(bucket[:k], bucket[k+1:]...)
					ht.buckets[j] = bucket

					// Shift any subsequent key-value pairs to fill the gap
					for l := (j + 1) % len(ht.buckets); l != bucketIndex; l = (l + 1) % len(ht.buckets) {
						nextBucket := ht.buckets[l]

						// Check if the next bucket is empty
						if len(nextBucket) == 0 {
							break
						}

						kv := nextBucket[0]
						if ht.hashFunc[i](kv.Key) != bucketIndex {
							// The key-value pair no longer belongs in this bucket, so move it to the gap
							bucket = append(bucket, kv)
							ht.buckets[j] = bucket
							nextBucket = nextBucket[1:]
							ht.buckets[l] = nextBucket
							j = l
						}
					}

					return
				}
			}
		}
	}
}

func (ht *HashTable) GetAll(hashKey string) (map[string]string, bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	result := make(map[string]string)

	found := false
	for i := 0; i < 2; i++ {
		hash := ht.hashFunc[i](hashKey)
		for _, kv := range ht.buckets[hash] {
			if kv.Key == hashKey {
				log.Println(kv)
				found = true
				result[kv.Field] = kv.Value
			}
		}
	}

	return result, found
}
