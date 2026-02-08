package decoder

import (
	"io"
	
	"github.com/hdt3213/rdb/parser"
)

// DecodeWithHDT uses the HDT3213 parser to decode RDB file
// This replaces the old github.com/919927181/rdb parser
func (d *Decoder) DecodeWithHDT(file io.Reader) error {
	decoder := parser.NewDecoder(file)

	err := decoder.Parse(func(obj parser.RedisObject) bool {
		// Convert RedisObject to Entry using adapter
		entry := ConvertToEntry(obj)
		
		// IMPORTANT: Create a copy to avoid all entries sharing the same pointer
		// This prevents memory leak when entries are stored in heaps/maps
		entryCopy := *entry
		
		// Send copy to channel for processing
		d.Entries <- &entryCopy
		
		// Return true to continue parsing
		return true
	})

	// Close channel to signal Count() goroutine that parsing is complete
	close(d.Entries)
	
	return err
}
