package base58

import (
	"fmt"

	"github.com/dromara/dongle"
)

func EncodeBase58String(rawStr string) string {
	encoded := dongle.Encode.FromBytes([]byte(rawStr)).ByBase58().ToString()
	//encoded := base58.CheckEncode([]byte(rawStr), 0x00)
	return encoded
}

func DecodeBase58String(encodedStr string) (string, error) {

	result := dongle.Decode.FromString(encodedStr).ByBase58()
	if result.Error != nil {
		return "", fmt.Errorf("dongle.Decode.FromString: %w", result.Error)
	}
	decoded := result.ToBytes()
	//decoded, _, err := base58.CheckDecode(encodedStr)
	//if err != nil {
	//	return "", fmt.Errorf("base58.CheckDecode: %w", err)
	//}

	return string(decoded), nil
}
