// SPDX-License-Identifier: MIT

package core

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RateLimiter", func() {

	Describe("hashIP", func() {
		Context("with various IP addresses", func() {
			It("should produce a 64-character hash for IPv4", func() {
				hash := hashIP("192.168.1.1")
				Expect(hash).To(HaveLen(64))
			})

			It("should produce a 64-character hash for IPv6", func() {
				hash := hashIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
				Expect(hash).To(HaveLen(64))
			})

			It("should produce a 64-character hash for localhost", func() {
				hash := hashIP("127.0.0.1")
				Expect(hash).To(HaveLen(64))
			})

			It("should produce a 64-character hash for empty string", func() {
				hash := hashIP("")
				Expect(hash).To(HaveLen(64))
			})
		})

		Context("deterministic behavior", func() {
			It("should produce the same hash for the same IP", func() {
				ip := "192.168.1.100"
				hash1 := hashIP(ip)
				hash2 := hashIP(ip)
				Expect(hash1).To(Equal(hash2))
			})
		})

		Context("uniqueness", func() {
			It("should produce different hashes for different IPs", func() {
				ip1 := "192.168.1.1"
				ip2 := "192.168.1.2"
				hash1 := hashIP(ip1)
				hash2 := hashIP(ip2)
				Expect(hash1).NotTo(Equal(hash2))
			})
		})
	})

	Describe("NewRateLimiter", func() {
		Context("with valid configuration", func() {
			It("should create a rate limiter with specified values", func() {
				rl := NewRateLimiter(100, 1*time.Minute)
				Expect(rl.GetLimit()).To(Equal(100))
				Expect(rl.GetPeriod()).To(Equal(1 * time.Minute))
			})
		})

		Context("with zero requests per period", func() {
			It("should use default requests per period", func() {
				rl := NewRateLimiter(0, 1*time.Minute)
				Expect(rl.GetLimit()).To(Equal(DefaultRateLimitRequests))
				Expect(rl.GetPeriod()).To(Equal(1 * time.Minute))
			})
		})

		Context("with negative requests per period", func() {
			It("should use default requests per period", func() {
				rl := NewRateLimiter(-10, 1*time.Minute)
				Expect(rl.GetLimit()).To(Equal(DefaultRateLimitRequests))
				Expect(rl.GetPeriod()).To(Equal(1 * time.Minute))
			})
		})

		Context("with zero period", func() {
			It("should use default period", func() {
				rl := NewRateLimiter(100, 0)
				Expect(rl.GetLimit()).To(Equal(100))
				Expect(rl.GetPeriod()).To(Equal(DefaultRateLimitPeriod))
			})
		})
	})

	Describe("Allow", func() {
		Context("within rate limit", func() {
			It("should allow requests up to the limit", func() {
				rl := NewRateLimiter(5, 10*time.Second)
				key := "test-ip-hash"

				// First 5 requests should be allowed
				for i := 0; i < 5; i++ {
					Expect(rl.Allow(key)).To(BeTrue(), "request %d should be allowed", i+1)
				}
			})
		})

		Context("exceeding rate limit", func() {
			It("should deny requests that exceed the limit", func() {
				rl := NewRateLimiter(5, 10*time.Second)
				key := "test-ip-hash"

				// Use up all tokens
				for i := 0; i < 5; i++ {
					rl.Allow(key)
				}

				// 6th request should be denied
				Expect(rl.Allow(key)).To(BeFalse(), "6th request should be denied")
			})
		})

		Context("remaining tokens after limit exceeded", func() {
			It("should have 0 remaining tokens after exceeding limit", func() {
				rl := NewRateLimiter(5, 10*time.Second)
				key := "test-ip-hash"

				// Use up all tokens
				for i := 0; i < 5; i++ {
					rl.Allow(key)
				}

				// Check remaining tokens
				Expect(rl.GetRemaining(key)).To(Equal(0))
			})
		})
	})

	Describe("TokenReplenishment", func() {
		Context("after waiting", func() {
			It("should allow requests after tokens replenish", func() {
				rl := NewRateLimiter(2, 1*time.Second)
				key := "test-ip-hash"

				// Use up all tokens
				rl.Allow(key) // 1 token remaining
				rl.Allow(key) // 0 tokens remaining

				// Next request should be denied
				Expect(rl.Allow(key)).To(BeFalse())

				// Wait for tokens to replenish (half the period = 0.5 seconds)
				time.Sleep(600 * time.Millisecond)

				// Should have ~1 token now (0.5 seconds * 2 tokens/second = 1 token)
				Expect(rl.Allow(key)).To(BeTrue(), "request should be allowed after token replenishment")
			})
		})
	})

	Describe("MultipleKeys", func() {
		Context("with different keys", func() {
			It("should isolate rate limits per key", func() {
				rl := NewRateLimiter(2, 1*time.Minute)
				key1 := "ip-hash-1"
				key2 := "ip-hash-2"

				// Key1: use up all tokens
				rl.Allow(key1)
				rl.Allow(key1)
				Expect(rl.Allow(key1)).To(BeFalse(), "key1 should be rate limited")

				// Key2: should still have tokens
				Expect(rl.Allow(key2)).To(BeTrue(), "key2 should be allowed (separate from key1)")

				// Verify remaining tokens
				Expect(rl.GetRemaining(key1)).To(Equal(0))
				Expect(rl.GetRemaining(key2)).To(Equal(1))
			})
		})
	})

	Describe("Cleanup", func() {
		Context("removing old entries", func() {
			It("should remove entries that haven't been accessed in a while", func() {
				rl := NewRateLimiter(10, 1*time.Minute)

				// Add some visitors
				rl.Allow("key1")
				rl.Allow("key2")
				rl.Allow("key3")

				// Verify they exist
				Expect(rl.GetRemaining("key1")).NotTo(Equal(0))

				// Manually set lastAccess to a long time ago
				rl.mu.Lock()
				for _, v := range rl.visitors {
					v.lastAccess = time.Now().Add(-10 * time.Minute)
				}
				rl.mu.Unlock()

				// Run cleanup
				rl.Cleanup()

				// Verify old entries are removed
				Expect(rl.GetRemaining("key1")).To(Equal(0))
			})
		})
	})

	Describe("ConcurrentAccess", func() {
		Context("with multiple goroutines", func() {
			It("should handle concurrent access safely", func() {
				rl := NewRateLimiter(100, 1*time.Minute)

				// Test concurrent access from multiple goroutines
				done := make(chan bool)
				for i := 0; i < 10; i++ {
					go func(id int) {
						key := hashIP("192.168.1." + string(rune(id)))
						for j := 0; j < 20; j++ {
							rl.Allow(key)
							rl.GetRemaining(key)
						}
						done <- true
					}(i)
				}

				// Wait for all goroutines to complete
				for i := 0; i < 10; i++ {
					<-done
				}

				// If we get here without panicking, concurrent access is safe
				// No assertion needed - the test passes if it doesn't panic
			})
		})
	})

	Describe("GetRemaining", func() {
		Context("for new key", func() {
			It("should return 0 for non-existent key", func() {
				rl := NewRateLimiter(10, 1*time.Minute)
				remaining := rl.GetRemaining("new-key")
				Expect(remaining).To(Equal(0))
			})

			It("should return correct remaining tokens after first request", func() {
				rl := NewRateLimiter(10, 1*time.Minute)
				rl.Allow("new-key")
				remaining := rl.GetRemaining("new-key")
				Expect(remaining).To(Equal(9))
			})
		})
	})
})
