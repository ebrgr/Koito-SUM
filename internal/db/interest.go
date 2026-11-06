package db

import "fmt"

const (
	DefaultInterestBuckets = 16
	MaxInterestBuckets     = 512
)

func ValidateInterestBuckets(count int) error {
	if count < 1 || count > MaxInterestBuckets {
		return fmt.Errorf("buckets must be between 1 and %d", MaxInterestBuckets)
	}
	return nil
}
