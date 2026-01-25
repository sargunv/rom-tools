package core

// GameInfo is implemented by all platform-specific ROM info structs.
// It provides common identification fields while allowing type assertion
// for platform-specific details.
type GameInfo interface {
	GamePlatform() Platform
	GameTitle() string  // May be empty if format doesn't have title
	GameSerial() string // May be empty if format doesn't have serial
	GameRegions() []Region
}
