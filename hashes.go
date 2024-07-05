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

    ht.set(hashKey, field, val)
}

func (ht *HashTable) set(hashKey, field, val string) {
    kv := &KeyValue{Key: hashKey, Field: field, Value: val}
    for attempt := 0; attempt < 10; attempt++ {
        for i := 0; i < 2; i++ {
            bucketIndex := ht.hashFunc[i](kv.Key)

            if ht.buckets[bucketIndex] == nil {
                ht.buckets[bucketIndex] = []*KeyValue{kv}
                ht.count++
                return
            }

            for j, existingKv := range ht.buckets[bucketIndex] {
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
    ht.set(hashKey, field, val)
}

func (ht *HashTable) resize() {
    newSize := ht.size * 2
    newBuckets := make([][]*KeyValue, newSize)
    oldBuckets := ht.buckets

    ht.buckets = newBuckets
    ht.size = newSize
    ht.count = 0

    for _, bucket := range oldBuckets {
        if bucket != nil {
            for _, kv := range bucket {
                if kv != nil {
                    ht.set(kv.Key, kv.Field, kv.Value)
                }
            }
        }
    }
}

func (ht *HashTable) Get(hashKey, field string) (string, bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	for i := 0; i < 2; i++ {
		hash := ht.hashFunc[i](hashKey)
		bucket := ht.buckets[hash]

		if bucket != nil {
			for _, kv := range bucket {
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
		for j := bucketIndex; ; j = (j + 1) % len(ht.buckets) {
			bucket := ht.buckets[j]

			if len(bucket) == 0 {
				return
			}

			for k, existingKv := range bucket {
				if existingKv.Key == hashKey && existingKv.Field == field {
					log.Println("Deleted key-value pair:", existingKv.Key, existingKv.Field)
					bucket = append(bucket[:k], bucket[k+1:]...)
					ht.buckets[j] = bucket

					for l := (j + 1) % len(ht.buckets); l != bucketIndex; l = (l + 1) % len(ht.buckets) {
						nextBucket := ht.buckets[l]

						if len(nextBucket) == 0 {
							break
						}

						kv := nextBucket[0]
						if ht.hashFunc[i](kv.Key) != bucketIndex {
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

	return result, found
}
