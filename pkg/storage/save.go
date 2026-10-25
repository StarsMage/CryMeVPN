package storage

import (
	"os"
)

func SaveConfig(s string)error {
	path, err := GetJsonPath()
	if err != nil {
		return err
	}


	return os.WriteFile(path, []byte(s),0600)
}