package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type RATRestrictionInformation struct {
	Value aper.BitString `aper:"sizeExt,sizeLB:8,sizeUB:8"`
}
