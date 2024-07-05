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

type HashTable struct {
	mu       sync.RWMutex
	buckets  [][]*KeyValue
	size     int
	hashFunc [2]func(string) int
	seed     maphash.Seed
	count    int
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

// These two functions, `hashFunc1` and `hashFunc2`, are hash functions used in the `HashTable` struct
// to calculate the index of the bucket where a key-value pair should be stored or retrieved.
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

// This `Set` method in the `HashTable` struct is responsible for adding or updating key-value pairs in
// the hash table. Here's a breakdown of what the method does:
func (ht *HashTable) Set(hashKey, field, val string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	kv := &KeyValue{Key: hashKey, Field: field, Value: val}
	original :=kv
	for attempt := 0; attempt < 10; attempt++ {
		// limiting the number of relocations to avoid infinite loops
		for i := 0; i < 2; i++ {
			bucketIndex := ht.hashFunc[i](kv.Key)

			if ht.buckets[bucketIndex] == nil {
				ht.buckets[bucketIndex] = []*KeyValue{kv}
				ht.count++
				return
			}

			for j,existingKv := range ht.buckets[bucketIndex]{
				if existingKv.Key == kv.Key && existingKv.Field == kv.Field {
					ht.buckets[bucketIndex][j] = kv
					return
				}
			}

			evictedKv := ht.buckets[bucketIndex][0]
			ht.buckets[bucketIndex][0] = kv
			kv = evictedKv
		}
	}
	ht.resize()
	ht.Set(original.Key, original.Field, original.Value)
	
}
func (ht *HashTable) resize() {
	newSize := ht.size * 2
	newBuckets := make([][]*KeyValue, newSize)
	oldBuckets := ht.buckets

	// Update the size and bucket reference
	ht.buckets = newBuckets
	ht.size = newSize
	ht.count = 0

	// Rehash all the elements into the new bucket array
	for _, bucket := range oldBuckets {
		if bucket != nil {
			for _, kv := range bucket {
				if kv != nil {
					ht.Set(kv.Key, kv.Field, kv.Value)
				}
			}
		}
	}
}

// The `Get` method in the `HashTable` struct is responsible for retrieving a value associated with a
// specific key and field from the hash table. Here's a breakdown of what the method does:
func (ht *HashTable) Get(hashKey, field string) (string, bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	for i := 0; i < 2; i++ {
		hash := ht.hashFunc[i](hashKey)
		bucket := ht.buckets[hash]
		
		if bucket != nil{
			for _,kv := range bucket{
				if kv.Key == hashKey && kv.Field == field {
					return kv.Value, true
				}
			}
		}
	}

	return "", false
}

// The `Delete` method in the `HashTable` struct is responsible for removing a specific key-value pair
// from the hash table. Here's a breakdown of what the method does:
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

func (ht *HashTable) GetAll(hashKey string) (map[string]string,bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	result := make(map[string]string)
	found:=false

	for i := 0; i < 2; i++ {
		bucketIndex := ht.hashFunc[i](hashKey)
		bucket := ht.buckets[bucketIndex]

		if bucket != nil {
			for _, kv := range bucket {
				if kv.Key == hashKey {
					found = true
					result[kv.Field] = kv.Value
				}
			}
		}
	}

	return result,found
}
