package context

import (
	"encoding/hex"
	"fmt"
	"my5G-RANTester/lib/UeauCommon"
	"my5G-RANTester/lib/milenage"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"my5G-RANTester/lib/openapi/models"
	"regexp"
)

type RanUeContext struct {
	Supi               string
	RanUeNgapId        int64
	AmfUeNgapId        int64
	mcc                string
	mnc                string
	ULCount            security.Count
	DLCount            security.Count
	CipheringAlg       uint8
	IntegrityAlg       uint8
	Snn                string
	KnasEnc            [16]uint8
	KnasInt            [16]uint8
	Kamf               []uint8
	AuthenticationSubs models.AuthenticationSubscription
	Suci               nasType.MobileIdentity5GS
	Snssai             models.Snssai
	UeIp               string
	ueTeid             uint8
}

func (ue *RanUeContext) NewRanUeContext(imsi string, ranUeNgapId int64, cipheringAlg, integrityAlg uint8, k, opc, op, amf, mcc, mnc string, sst int32, sd string) {

	// added Ran UE NGAP ID.
	ue.RanUeNgapId = ranUeNgapId

	// added SUPI.
	ue.Supi = imsi

	// TODO ue.amfUENgap is received by AMF in authentication request.(? changed this).
	// ue.AmfUeNgapId = ranUeNgapId

	// added ciphering algorithm.
	ue.CipheringAlg = cipheringAlg

	// added integrity algorithm.
	ue.IntegrityAlg = integrityAlg

	// added key, AuthenticationManagementField and opc or op.
	ue.SetAuthSubscription(k, opc, op, amf)

	// added suci
	suciV1, suciV2, suciV3 := ue.EncodeUeSuci()

	// added mcc and mnc
	ue.mcc = mcc
	ue.mnc = mnc

	// added network slice
	ue.Snssai.Sd = sd
	ue.Snssai.Sst = sst

	// encode mcc and mnc for mobileIdentity5Gs.
	resu := ue.GetMccAndMncInOctets()

	// added suci to mobileIdentity5GS
	if len(ue.Supi) == 18 {
		ue.Suci = nasType.MobileIdentity5GS{
			Len:    12, // suci
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], 0xf0, 0xff, 0x00, 0x00, 0x00, suciV3, suciV2, suciV1},
		}
	} else {
		ue.Suci = nasType.MobileIdentity5GS{
			Len:    13, // suci
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], 0xf0, 0xff, 0x00, 0x00, 0x00, 0x00, suciV3, suciV2, suciV1},
		}
	}

	// added snn.
	ue.Snn = ue.deriveSNN()
}

func (ue *RanUeContext) GetUeTeid() uint8 {
	return ue.ueTeid
}

func (ue *RanUeContext) SetUeTeid(teid uint8) {
	ue.ueTeid = teid
}

func (ue *RanUeContext) SetIp(ip [12]uint8) {
	ue.UeIp = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (ue *RanUeContext) GetIp() string {
	return ue.UeIp
}

func (ue *RanUeContext) SetAmfNgapId(amfUeId int64) {
	// fmt.Println(amfUeId)
	ue.AmfUeNgapId = amfUeId
}

func (ue *RanUeContext) deriveSNN() string {
	// 5G:mnc093.mcc208.3gppnetwork.org
	var resu string
	if len(ue.mnc) == 2 {
		resu = "5G:mnc0" + ue.mnc + ".mcc" + ue.mcc + ".3gppnetwork.org"
	} else {
		resu = "5G:mnc" + ue.mnc + ".mcc" + ue.mcc + ".3gppnetwork.org"
	}

	return resu
}

func (ue *RanUeContext) GetMccAndMncInOctets() []byte {

	// reverse mcc and mnc
	mcc := reverse(ue.mcc)
	mnc := reverse(ue.mnc)

	// include mcc and mnc in octets
	oct5 := mcc[1:3]
	var oct6 string
	var oct7 string
	if len(ue.mnc) == 2 {
		oct6 = "f" + string(mcc[0])
		oct7 = mnc
	} else {
		oct6 = string(mnc[0]) + string(mcc[0])
		oct7 = mnc[1:3]
	}

	// changed for bytes.
	resu, err := hex.DecodeString(oct5 + oct6 + oct7)
	if err != nil {
		fmt.Println(err)
	}

	return resu
}

func (ue *RanUeContext) EncodeUeSuci() (uint8, uint8, uint8) {

	// reverse imsi string.
	aux := reverse(ue.Supi)

	// calculate decimal value.
	suci, error := hex.DecodeString(aux[:6])
	if error != nil {
		return 0, 0, 0
	}

	// return decimal value
	// Function worked fine.
	return uint8(suci[0]), uint8(suci[1]), uint8(suci[2])
}

func (ue *RanUeContext) deriveSQN(autn []byte, ak []uint8) []byte {
	sqn := make([]byte, 6)

	// get SQNxorAK
	SQNxorAK := autn[0:6]

	// get sqn
	for i := 0; i < len(SQNxorAK); i++ {
		sqn[i] = SQNxorAK[i] ^ ak[i]
	}

	// return sqn
	return sqn
}

func (ue *RanUeContext) DeriveRESstarAndSetKey(authSubs models.AuthenticationSubscription, RAND []byte, snNmae string, AUTN []byte) []byte {

	// SQN, _ := hex.DecodeString(authSubs.SequenceNumber)

	// get management field.
	AMF, _ := hex.DecodeString(authSubs.AuthenticationManagementField)

	// Run milenage
	// TODO: verify MAC
	MAC_A, MAC_S := make([]byte, 8), make([]byte, 8)
	CK, IK := make([]byte, 16), make([]byte, 16)
	RES := make([]byte, 8)
	AK, AKstar := make([]byte, 6), make([]byte, 6)

	// generate OPC, K.
	OPC, _ := hex.DecodeString(authSubs.Opc.OpcValue)
	K, _ := hex.DecodeString(authSubs.PermanentKey.PermanentKeyValue)

	// Generate RES, CK, IK, AK, AKstar
	milenage.F2345_Test(OPC, K, RAND, RES, CK, IK, AK, AKstar)

	// Generate SQN.
	SQN := ue.deriveSQN(AUTN, AK)

	// Generate MAC_A, MAC_S
	milenage.F1_Test(OPC, K, RAND, SQN, AMF, MAC_A, MAC_S)

	// Generate RES, CK, IK, AK, AKstar
	milenage.F2345_Test(OPC, K, RAND, RES, CK, IK, AK, AKstar)

	// derive RES*
	key := append(CK, IK...)
	FC := UeauCommon.FC_FOR_RES_STAR_XRES_STAR_DERIVATION
	P0 := []byte(snNmae)
	P1 := RAND
	P2 := RES

	ue.DerivateKamf(key, snNmae, SQN, AK)
	ue.DerivateAlgKey()
	kdfVal_for_resStar := UeauCommon.GetKDFValue(key, FC, P0, UeauCommon.KDFLen(P0), P1, UeauCommon.KDFLen(P1), P2, UeauCommon.KDFLen(P2))
	return kdfVal_for_resStar[len(kdfVal_for_resStar)/2:]

}

func (ue *RanUeContext) DerivateKamf(key []byte, snName string, SQN, AK []byte) {

	FC := UeauCommon.FC_FOR_KAUSF_DERIVATION
	P0 := []byte(snName)
	SQNxorAK := make([]byte, 6)
	for i := 0; i < len(SQN); i++ {
		SQNxorAK[i] = SQN[i] ^ AK[i]
	}
	P1 := SQNxorAK
	Kausf := UeauCommon.GetKDFValue(key, FC, P0, UeauCommon.KDFLen(P0), P1, UeauCommon.KDFLen(P1))
	P0 = []byte(snName)
	Kseaf := UeauCommon.GetKDFValue(Kausf, UeauCommon.FC_FOR_KSEAF_DERIVATION, P0, UeauCommon.KDFLen(P0))

	supiRegexp, _ := regexp.Compile("(?:imsi|supi)-([0-9]{5,15})")
	groups := supiRegexp.FindStringSubmatch(ue.Supi)

	P0 = []byte(groups[1])
	L0 := UeauCommon.KDFLen(P0)
	P1 = []byte{0x00, 0x00}
	L1 := UeauCommon.KDFLen(P1)

	ue.Kamf = UeauCommon.GetKDFValue(Kseaf, UeauCommon.FC_FOR_KAMF_DERIVATION, P0, L0, P1, L1)
}

// Algorithm key Derivation function defined in TS 33.501 Annex A.9
func (ue *RanUeContext) DerivateAlgKey() {
	// Security Key
	P0 := []byte{security.NNASEncAlg}
	L0 := UeauCommon.KDFLen(P0)
	P1 := []byte{ue.CipheringAlg}
	L1 := UeauCommon.KDFLen(P1)

	kenc := UeauCommon.GetKDFValue(ue.Kamf, UeauCommon.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	copy(ue.KnasEnc[:], kenc[16:32])

	// Integrity Key
	P0 = []byte{security.NNASIntAlg}
	L0 = UeauCommon.KDFLen(P0)
	P1 = []byte{ue.IntegrityAlg}
	L1 = UeauCommon.KDFLen(P1)

	kint := UeauCommon.GetKDFValue(ue.Kamf, UeauCommon.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	copy(ue.KnasInt[:], kint[16:32])
}

func (ue *RanUeContext) SetAuthSubscription(k, opc, op, amf string) {
	ue.AuthenticationSubs.PermanentKey = &models.PermanentKey{
		PermanentKeyValue: k,
	}
	ue.AuthenticationSubs.Opc = &models.Opc{
		OpcValue: opc,
	}
	ue.AuthenticationSubs.Milenage = &models.Milenage{
		Op: &models.Op{
			OpValue: op,
		},
	}
	ue.AuthenticationSubs.AuthenticationManagementField = amf

	//ue.AuthenticationSubs.SequenceNumber = TestGenAuthData.MilenageTestSet19.SQN
	ue.AuthenticationSubs.AuthenticationMethod = models.AuthMethod__5_G_AKA
}

func SetUESecurityCapability(ue *RanUeContext) (UESecurityCapability *nasType.UESecurityCapability) {
	UESecurityCapability = &nasType.UESecurityCapability{
		Iei:    nasMessage.RegistrationRequestUESecurityCapabilityType,
		Len:    8,
		Buffer: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	switch ue.CipheringAlg {
	case security.AlgCiphering128NEA0:
		UESecurityCapability.SetEA0_5G(1)
	case security.AlgCiphering128NEA1:
		UESecurityCapability.SetEA1_128_5G(1)
	case security.AlgCiphering128NEA2:
		UESecurityCapability.SetEA2_128_5G(1)
	case security.AlgCiphering128NEA3:
		UESecurityCapability.SetEA3_128_5G(1)
	}

	switch ue.IntegrityAlg {
	case security.AlgIntegrity128NIA0:
		UESecurityCapability.SetIA0_5G(1)
	case security.AlgIntegrity128NIA1:
		UESecurityCapability.SetIA1_128_5G(1)
	case security.AlgIntegrity128NIA2:
		UESecurityCapability.SetIA2_128_5G(1)
	case security.AlgIntegrity128NIA3:
		UESecurityCapability.SetIA3_128_5G(1)
	}

	return
}
