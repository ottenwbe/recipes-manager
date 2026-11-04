/*
 * MIT License - see LICENSE file for details
 */

package recipes

// UniqueSlice removes duplicates and ensures that a slice only comprises unique elements
func UniqueSlice(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	last := 0
	for _, v := range input {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		input[last] = v
		last++
	}
	return input[:last]
}
