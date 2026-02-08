package decoder

import (
	"github.com/hdt3213/rdb/parser"
)

// ConvertToEntry converts HDT3213 RedisObject to our Entry format
// This adapter allows us to use the new parser with existing analysis code
func ConvertToEntry(obj parser.RedisObject) *Entry {
	entry := &Entry{
		Key:   obj.GetKey(),
		Type:  obj.GetType(),
		Bytes: uint64(obj.GetSize()),
		Db:    obj.GetDBIndex(),
	}

	// Handle expiration (pointer to time.Time)
	if exp := obj.GetExpiration(); exp != nil {
		entry.Expiration = exp.Unix() * 1000 // Convert to milliseconds
	} else {
		entry.Expiration = 0
	}

	// Extract NumOfElem based on type
	switch o := obj.(type) {
	case *parser.ListObject:
		entry.NumOfElem = uint64(len(o.Values))
		// Find largest element
		if len(o.Values) > 0 {
			maxLen := 0
			for _, v := range o.Values {
				if len(v) > maxLen {
					maxLen = len(v)
				}
			}
			entry.LenOfLargestElem = uint64(maxLen)
		}

	case *parser.SetObject:
		entry.NumOfElem = uint64(len(o.Members))
		// Find largest member
		if len(o.Members) > 0 {
			maxLen := 0
			for _, m := range o.Members {
				if len(m) > maxLen {
					maxLen = len(m)
				}
			}
			entry.LenOfLargestElem = uint64(maxLen)
		}

	case *parser.HashObject:
		entry.NumOfElem = uint64(len(o.Hash))
		// Find largest field
		if len(o.Hash) > 0 {
			maxLen := 0
			var maxField string
			for k, v := range o.Hash {
				if len(v) > maxLen {
					maxLen = len(v)
					maxField = k
				}
			}
			entry.LenOfLargestElem = uint64(maxLen)
			entry.FieldOfLargestElem = maxField
		}

	case *parser.ZSetObject:
		entry.NumOfElem = uint64(len(o.Entries))
		// Find largest member
		if len(o.Entries) > 0 {
			maxLen := 0
			for _, e := range o.Entries {
				if len(e.Member) > maxLen {
					maxLen = len(e.Member)
				}
			}
			entry.LenOfLargestElem = uint64(maxLen)
		}

	case *parser.StreamObject:
		// Stream entries count
		totalEntries := 0
		for _, e := range o.Entries {
			totalEntries += len(e.Fields)
		}
		entry.NumOfElem = uint64(totalEntries)

	case *parser.StringObject:
		// For strings, NumOfElem is the length of the string value
		// This is important for length distribution charts
		entry.NumOfElem = uint64(len(o.Value))
		
	default:
		entry.NumOfElem = 0
	}

	// Encoding is not exposed by HDT3213, leave empty
	entry.Encoding = ""

	return entry
}
