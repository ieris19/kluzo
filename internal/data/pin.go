package data

import (
	"fmt"
	"strconv"
)

// PinLevel is how many leading version segments a candidate tag must share
type PinLevel int

const (
	// PinFreeze skips the upstream check entirely.
	// Deliberately outside the order below: frozen containers should never compare
	PinFreeze PinLevel = -1
	// PinChannel is the default: only tags matching the current channel (Extra)
	PinChannel PinLevel = 0
	// PinMajor additionally only considers tags sharing the first segment.
	PinMajor PinLevel = 1
	// PinMinor additionally only considers tags sharing the first two segments.
	PinMinor PinLevel = 2
)

func (p PinLevel) String() string {
	switch p {
	case PinFreeze:
		return "freeze"
	case PinChannel:
		return "channel"
	case PinMajor:
		return "major"
	case PinMinor:
		return "minor"
	default:
		// Any other depth has no name, only its number
		if p > 0 {
			return strconv.Itoa(int(p))
		}
	}
	return "invalid"
}

// Parses the PinLevel string into its corresponding enum value.
func ParsePinLevel(s string) (PinLevel, error) {
	switch s {
	// Empty string represents an omitted Pin key, defaulting to PinChannel.
	case "", "channel":
		return PinChannel, nil
	case "major":
		return PinMajor, nil
	case "minor":
		return PinMinor, nil
	case "freeze":
		return PinFreeze, nil
	}
	// Depths past the named ones are spelled as a plain segment count.
	depth, err := strconv.Atoi(s)
	// Freeze is spelled "freeze", so we reject negative numbers.
	if err != nil || depth < 0 {
		return 0, fmt.Errorf("invalid pin level %q", s)
	}
	return PinLevel(depth), nil
}
