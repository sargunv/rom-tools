// Package core provides shared types for ROM identification and platform definitions.
package core

// Platform represents a gaming platform.
type Platform string

const (
	PlatformNES     Platform = "famicom"
	PlatformSNES    Platform = "superfamicom"
	PlatformN64     Platform = "nintendo64"
	PlatformGC      Platform = "gamecube"
	PlatformWii     Platform = "wii"
	PlatformWiiU    Platform = "wiiu"
	PlatformSwitch  Platform = "switch"
	PlatformSwitch2 Platform = "switch2"

	PlatformGB     Platform = "gameboy"
	PlatformGBC    Platform = "gameboycolor"
	PlatformGBA    Platform = "gameboyadvance"
	PlatformNDS    Platform = "ds"
	PlatformDSi    Platform = "dsi"
	Platform3DS    Platform = "3ds"
	PlatformNew3DS Platform = "new3ds"

	PlatformPS1 Platform = "playstation"
	PlatformPS2 Platform = "playstation2"
	PlatformPS3 Platform = "playstation3"
	PlatformPS4 Platform = "playstation4"
	PlatformPS5 Platform = "playstation5"

	PlatformPSP    Platform = "psp"
	PlatformPSVita Platform = "psvita"

	PlatformMS        Platform = "mastersystem"
	PlatformMD        Platform = "megadrive"
	PlatformSegaCD    Platform = "segacd"
	Platform32X       Platform = "sega32x"
	PlatformSaturn    Platform = "saturn"
	PlatformDreamcast Platform = "dreamcast"

	PlatformGameGear Platform = "gamegear"

	PlatformXbox       Platform = "xbox"
	PlatformXbox360    Platform = "xbox360"
	PlatformXboxOne    Platform = "xboxone"
	PlatformXboxSeries Platform = "xboxseries"
)
