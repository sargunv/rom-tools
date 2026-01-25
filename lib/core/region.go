package core

// Region represents a geographic region for ROM/asset matching.
// Regions form a hierarchy (e.g., Germany -> Europe -> World) used for
// fallback matching when exact region assets aren't available.
type Region string

const (
	RegionUnknown Region = ""

	// Top level regions (parent is World, or no parent for World itself)
	RegionWorld      Region = "wor"
	RegionEurope     Region = "eu"
	RegionAsia       Region = "asi"
	RegionAmerica    Region = "ame" // American Continent (North + South)
	RegionOceania    Region = "oce"
	RegionMiddleEast Region = "mor"
	RegionAfrica     Region = "afr"

	// Europe children
	RegionGermany     Region = "de"
	RegionFrance      Region = "fr"
	RegionUK          Region = "uk"
	RegionSpain       Region = "sp"
	RegionItaly       Region = "it"
	RegionNetherlands Region = "nl"
	RegionSweden      Region = "se"
	RegionDenmark     Region = "dk"
	RegionFinland     Region = "fi"
	RegionNorway      Region = "no"
	RegionPortugal    Region = "pt"
	RegionPoland      Region = "pl"
	RegionCzech       Region = "cz"
	RegionHungary     Region = "hu"
	RegionSlovakia    Region = "sk"
	RegionBulgaria    Region = "bg"
	RegionGreece      Region = "gr"
	RegionRussia      Region = "ru"

	// Asia children
	RegionJapan  Region = "jp"
	RegionChina  Region = "cn"
	RegionKorea  Region = "kr"
	RegionTaiwan Region = "tw"

	// America children
	RegionUSA    Region = "us"
	RegionCanada Region = "ca"
	RegionBrazil Region = "br"
	RegionMexico Region = "mex"
	RegionChile  Region = "cl"
	RegionPeru   Region = "pe"

	// Oceania children
	RegionAustralia  Region = "au"
	RegionNewZealand Region = "nz"

	// Middle East children
	RegionIsrael Region = "il"
	RegionTurkey Region = "tr"
	RegionKuwait Region = "kw"
	RegionUAE    Region = "ae"

	// Africa children
	RegionSouthAfrica Region = "za"
)

// regionParents maps each region to its parent in the hierarchy.
var regionParents = map[Region]Region{
	// Continental regions -> World
	RegionEurope:     RegionWorld,
	RegionAsia:       RegionWorld,
	RegionAmerica:    RegionWorld,
	RegionOceania:    RegionWorld,
	RegionMiddleEast: RegionWorld,
	RegionAfrica:     RegionWorld,

	// Europe children
	RegionGermany:     RegionEurope,
	RegionFrance:      RegionEurope,
	RegionUK:          RegionEurope,
	RegionSpain:       RegionEurope,
	RegionItaly:       RegionEurope,
	RegionNetherlands: RegionEurope,
	RegionSweden:      RegionEurope,
	RegionDenmark:     RegionEurope,
	RegionFinland:     RegionEurope,
	RegionNorway:      RegionEurope,
	RegionPortugal:    RegionEurope,
	RegionPoland:      RegionEurope,
	RegionCzech:       RegionEurope,
	RegionHungary:     RegionEurope,
	RegionSlovakia:    RegionEurope,
	RegionBulgaria:    RegionEurope,
	RegionGreece:      RegionEurope,
	RegionRussia:      RegionEurope,

	// Asia children
	RegionJapan:  RegionAsia,
	RegionChina:  RegionAsia,
	RegionKorea:  RegionAsia,
	RegionTaiwan: RegionAsia,

	// America children
	RegionUSA:    RegionAmerica,
	RegionCanada: RegionAmerica,
	RegionBrazil: RegionAmerica,
	RegionMexico: RegionAmerica,
	RegionChile:  RegionAmerica,
	RegionPeru:   RegionAmerica,

	// Oceania children
	RegionAustralia:  RegionOceania,
	RegionNewZealand: RegionOceania,

	// Middle East children
	RegionIsrael: RegionMiddleEast,
	RegionTurkey: RegionMiddleEast,
	RegionKuwait: RegionMiddleEast,
	RegionUAE:    RegionMiddleEast,

	// Africa children
	RegionSouthAfrica: RegionAfrica,
}

// Parent returns this region's parent in the hierarchy.
// Returns RegionWorld for top-level continental regions.
// Returns RegionUnknown for RegionWorld and RegionUnknown.
func (r Region) Parent() Region {
	if parent, ok := regionParents[r]; ok {
		return parent
	}
	return RegionUnknown
}

// Ancestors returns the chain of ancestors from this region up to World.
// For example, RegionGermany.Ancestors() returns [RegionEurope, RegionWorld].
// Returns nil for RegionWorld, RegionUnknown, or top-level regions.
func (r Region) Ancestors() []Region {
	var ancestors []Region
	for p := r.Parent(); p != RegionUnknown; p = p.Parent() {
		ancestors = append(ancestors, p)
	}
	return ancestors
}

// IsAncestorOf returns true if r is an ancestor of other in the hierarchy,
// along with the distance (number of hops from other to r).
// For example, RegionEurope.IsAncestorOf(RegionGermany) returns (true, 1).
// Returns (false, -1) if r is not an ancestor of other.
func (r Region) IsAncestorOf(other Region) (bool, int) {
	dist := 0
	for p := other.Parent(); p != RegionUnknown; p = p.Parent() {
		dist++
		if p == r {
			return true, dist
		}
	}
	return false, -1
}

// IsDescendantOf returns true if r is a descendant of other in the hierarchy,
// along with the distance (number of hops from r to other).
// For example, RegionGermany.IsDescendantOf(RegionEurope) returns (true, 1).
// Returns (false, -1) if r is not a descendant of other.
func (r Region) IsDescendantOf(other Region) (bool, int) {
	return other.IsAncestorOf(r)
}
