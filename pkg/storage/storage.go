package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)


func GetClearPath() (string, error) {
    dir, err := os.UserConfigDir()

    if err != nil {
        return "", err
    }

    appDir := filepath.Join(dir, "CryMeVpn")

    if err := os.MkdirAll(appDir, 0700); err != nil {
        return "", err
    }

    return appDir, nil
}

// func GetConfigPath() (string, error) {
// 	dir, err := os.UserConfigDir()

// 	if err != nil {
// 		return "", err
// 	}

// 	appDir  := filepath.Join(dir, "CryMeVpn")

// 	if err := os.MkdirAll(appDir, 0700); err != nil {
// 		return "",err
// 	}

// 	return filepath.Join(appDir, "stroage.enc"), nil
// }


func GetConfigPath() (string, error) {
	dir, err := GetClearPath()

	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "stroage.enc"), nil
}


// func GetJsonPath() (string, error) { 
// 	dir, err := os.UserConfigDir()

// 	if err != nil {
// 		return "", err
// 	}

// 	appDir  := filepath.Join(dir, "CryMeVpn")

// 	if err := os.MkdirAll(appDir, 0700); err != nil {
// 		return "",err
// 	}

// 	return filepath.Join(appDir, "config.json"), nil
// }

func GetJsonPath() (string, error) {
	dir, err := GetClearPath()

	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.json"), nil
}


func (a *AppStorage) Save(key []byte) error {
	jsonData, err := json.Marshal(a)

	if err != nil {
		return err
	}

	encryptedJson, err := Encrypt(jsonData, key)

	if err != nil {
		return err
	}

	path, err := GetConfigPath()

	if err != nil {
		return err
	}

	return os.WriteFile(path, encryptedJson, 0600)
}

func Load(key []byte) (*AppStorage, error) {
	path, err := GetConfigPath()

	if err != nil {
		return nil, err
	}

	encryptedData, err := os.ReadFile(path)

	if err != nil {
		if os.IsNotExist(err) {
			return &AppStorage{Servers: []ServerData{}},nil
		}
		return nil, err
	}

	jsonData, err := Decrypt(encryptedData, key)

	if err != nil {
		return nil,err
	}

	var app AppStorage
	if err := json.Unmarshal(jsonData,&app); err != nil {
		return nil, err
	}

	return &app, nil
} 