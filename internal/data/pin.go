package data

import "fmt"

type PinLevel int

const (
	// PinFreeze skips the upstream check entirely.
	// Deliberately outside the order below: frozen containers should never compare
	PinFreeze PinLevel = -1
	// PinChannel is the default: only tags matching the current channel (Extra)
	PinChannel PinLevel = 0
	// PinMajor additionally only considers tags sharing the current Major version.
	PinMajor PinLevel = 1
	// PinMinor additionally only considers tags sharing the current Major.Minor version.
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
		return "invalid"
	}
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
	default:
		return 0, fmt.Errorf("invalid pin level %q", s)
	}
}
