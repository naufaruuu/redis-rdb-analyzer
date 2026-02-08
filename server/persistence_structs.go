package server

import (
    "github.com/naufaruuu/redis-rdb-analyzer/decoder"
)

// CounterDTO is the exportable version of Counter for JSON Marshaling
type CounterDTO struct {
	LargestEntries     []*decoder.Entry       `json:"LargestEntries"`
	LargestKeyPrefixes []*PrefixEntry         `json:"LargestKeyPrefixes"`
	LengthLevel0       uint64                 `json:"LengthLevel0"`
	LengthLevel1       uint64                 `json:"LengthLevel1"`
	LengthLevel2       uint64                 `json:"LengthLevel2"`
	LengthLevel3       uint64                 `json:"LengthLevel3"`
	LengthLevel4       uint64                 `json:"LengthLevel4"`
	LengthLevelBytes   map[string]uint64      `json:"LengthLevelBytes"` // keys as string representation of typeKey
	LengthLevelNum     map[string]uint64      `json:"LengthLevelNum"`
	KeyPrefixBytes     map[string]uint64      `json:"KeyPrefixBytes"`
	KeyPrefixNum       map[string]uint64      `json:"KeyPrefixNum"`
	KeyPrefixDb        map[string]string      `json:"KeyPrefixDb"`
	TypeBytes          map[string]uint64      `json:"TypeBytes"`
	TypeNum            map[string]uint64      `json:"TypeNum"`
	SlotBytes          map[int]uint64         `json:"SlotBytes"`
	SlotNum            map[int]uint64         `json:"SlotNum"`
}

// Helper to convert complex map keys to string for JSON
func typeKeyToString(t, k string) string {
    // Simple serialization format: Type|Key
    return t + "|" + k
}

func stringToTypeKey(s string) (string, string) {
    // rudimentary split
    // Assumes | is separator. If keys contain |, this is flawed. 
    // Ideally we'd use a struct as key but JSON doesn't support struct keys in maps.
    // Given rdr context, Type is usually string enum, Key varies.
    // Let's use customized Marshal/Unmarshal if strictly needed, or just serialize key properly.
    // For simplicity, let's use a struct array for serialization if maps are problematic?
    // Or just "Type:Key".
    parts := jsonStringSplit(s)
    if len(parts) >= 2 {
        return parts[0], parts[1]
    }
    return "", ""
}

func jsonStringSplit(s string) []string {
    // Basic split. 
    // In rdr typeKey is {Type, Key}.
    // We'll trust our serializer.
    // Let's implement ToDTO method on Counter
    return []string{} // placeholder
}

// Convert Counter to DTO
func (c *Counter) ToDTO() *CounterDTO {
    dto := &CounterDTO{
        LengthLevel0:     c.lengthLevel0,
        LengthLevel1:     c.lengthLevel1,
        LengthLevel2:     c.lengthLevel2,
        LengthLevel3:     c.lengthLevel3,
        LengthLevel4:     c.lengthLevel4,
        TypeBytes:        c.typeBytes,
        TypeNum:          c.typeNum,
        SlotBytes:        c.slotBytes,
        SlotNum:          c.slotNum,
        LengthLevelBytes: make(map[string]uint64),
        LengthLevelNum:   make(map[string]uint64),
        KeyPrefixBytes:   make(map[string]uint64),
        KeyPrefixNum:     make(map[string]uint64),
        KeyPrefixDb:      make(map[string]string),
    }

    // Convert heaps to slices
    dto.LargestEntries = c.GetLargestEntries(500, 0)
    dto.LargestKeyPrefixes = c.GetLargestKeyPrefixes()

    // Convert maps with struct keys
    for k, v := range c.lengthLevelBytes {
        dto.LengthLevelBytes[k.Type+"|"+k.Key] = v
    }
    for k, v := range c.lengthLevelNum {
        dto.LengthLevelNum[k.Type+"|"+k.Key] = v
    }
    for k, v := range c.keyPrefixBytes {
        dto.KeyPrefixBytes[k.Type+"|"+k.Key] = v
    }
    for k, v := range c.keyPrefixNum {
        dto.KeyPrefixNum[k.Type+"|"+k.Key] = v
    }
    for k, v := range c.keyPrefixDb {
        dto.KeyPrefixDb[k.Type+"|"+k.Key] = v
    }

    return dto
}

// Convert DTO back to Counter
func (dto *CounterDTO) ToCounter() *Counter {
    c := NewCounter()
    // Restore basic fields
    c.lengthLevel0 = dto.LengthLevel0
    c.lengthLevel1 = dto.LengthLevel1
    c.lengthLevel2 = dto.LengthLevel2
    c.lengthLevel3 = dto.LengthLevel3
    c.lengthLevel4 = dto.LengthLevel4
    c.typeBytes = dto.TypeBytes
    c.typeNum = dto.TypeNum
    c.slotBytes = dto.SlotBytes
    c.slotNum = dto.SlotNum

    // Restore heaps
    for _, e := range dto.LargestEntries {
        c.countLargestEntries(e, 500)
    }
    // Note: largestKeyPrefixes is derived, but can be restored.
    // However, heap logic usually rebuilds. 
    // Actually Counter.count() builds heaps incrementally.
    // But here we have the final list.
    // We can push them back into the heap.


    // Wait, DTO contains `LargestKeyPrefixes` which is a slice.
    // We can just iterate and push.
    // Problem: `PrefixEntry` might be missing `Bytes` etc if JSON didn't save them? No, DTO has them.
    
    // Actually, simply populating the keyPrefix maps might be safer if we want to support further ops?
    // But `calcuLargestKeyPrefix` consumes the maps.
    // If we are just viewing, we need the heap populated for `GetLargestKeyPrefixes()` to work.
    
    // Let's populate the heap directly.
    for _, pe := range dto.LargestKeyPrefixes {
        // We need to Push to heap.
        // We can't use heap.Push directly on unexported type if we were external, but we are internal.
        *c.largestKeyPrefixes = append(*c.largestKeyPrefixes, pe)
    }
    // Re-heapify?
    // sort.Sort will be called by Getter anyway.

    // Restore Maps
    restoreMap(dto.LengthLevelBytes, c.lengthLevelBytes)
    restoreMap(dto.LengthLevelNum, c.lengthLevelNum)
    restoreMap(dto.KeyPrefixBytes, c.keyPrefixBytes)
    restoreMap(dto.KeyPrefixNum, c.keyPrefixNum)
    
    // KeyPrefixDb
    for kStr, v := range dto.KeyPrefixDb {
        t, k := parseTypeKey(kStr)
        c.keyPrefixDb[typeKey{Type: t, Key: k}] = v
    }

    return c
}

func restoreMap(src map[string]uint64, dst map[typeKey]uint64) {
    for kStr, v := range src {
        t, k := parseTypeKey(kStr)
        dst[typeKey{Type: t, Key: k}] = v
    }
}

func parseTypeKey(s string) (string, string) {
    // Split by first |
    // If key contains |, this is risky. 
    // Counter.go separators are ":;,_- ". Key shouldn't contain | usually?
    // Redis keys CAN contain anything.
    // We should use a safer separator or JSON array.
    // But for now, let's assume | is safe enough or find first occurrence.
    // type is usually limited set (string, hash, etc). Key is explicit.
    // Type is "string", "hash", "list"...
    // So we can find first |.
    /*
    typeKey struct {
	    Type string
	    Key  string
    }
    */
    // Type doesn't contain |.
    for i, c := range s {
        if c == '|' {
            return s[:i], s[i+1:]
        }
    }
    return "", s 
}
