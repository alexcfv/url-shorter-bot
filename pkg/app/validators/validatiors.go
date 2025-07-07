package validators

import (
	"hash/fnv"
	"regexp"
)

func IsValidURL(url string) bool {
	re := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9.-]+$`)
	return re.MatchString(url)
}

func ShortToHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
