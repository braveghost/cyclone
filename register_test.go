package cyclone

import (
	"fmt"
	"regexp"
	"testing"
)

func TestRegex(t *testing.T) {
	match, err := regexp.MatchString(IpPortRegex, "127.0.0.11:12")
	fmt.Println(match, err)
}
