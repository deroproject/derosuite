package p2p

import "testing"
import "encoding/hex"


func Test_Levin_Header_Deserialisation(t *testing.T) {

	// this structure has been manually pulled from wireshark stream
	raw_data := "0121010101010101e20000000000000001e9030000000000000100000001000000"
	raw_data_blob, _ := hex.DecodeString(raw_data)

	_ = raw_data_blob

	var lheader Levin_Header

	err := lheader.DeSerialize(raw_data_blob)

	if err != nil {
		t.Error("DeSerialize Levin header Failed\n")
		return
	}

	// now serialize once again

	serialised, _ := lheader.Serialize()
	if raw_data != hex.EncodeToString(serialised) {
		t.Error("Serialize Levin_Header Failed")
	}

}

func Test_Levin_Data_Header(t *testing.T) {

	// this structure has been manually pulled from wireshark stream
	raw_data := "01110101010102010100"
	raw_data_blob, _ := hex.DecodeString(raw_data)

	_ = raw_data_blob

	var lheader Levin_Data_Header

	err := lheader.DeSerialize(raw_data_blob)

	if err != nil {
		t.Error("DeSerialize Levin Data header Failed\n")
		return
	}

	// now serialize once again

	serialised, _ := lheader.Serialize()
	if raw_data != hex.EncodeToString(serialised) {
		t.Errorf("Serialize Levin_Data_Header Failed \n%s correct value \n%s Our value", raw_data, hex.EncodeToString(serialised))
	}
}
