/*
 * MIT License - see LICENSE file for details
 */

package sources

import "fmt"

// CBytes takes a byte string and casts it to a comma separated string representation of that byte string
func CBytes(bytes []byte) string {
	var (
		result   = "["
		numBytes = len(bytes)
	)
	for i := 0; i < numBytes-1; i++ {
		result += fmt.Sprintf("%v", int(bytes[i])) + ","
	}
	result += fmt.Sprintf("%v", int(bytes[numBytes-1]))
	return result + "]"
}
