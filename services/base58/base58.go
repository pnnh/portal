package base58

import (
	"fmt"

	"github.com/dromara/dongle"
)

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
