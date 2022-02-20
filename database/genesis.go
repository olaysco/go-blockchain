package database

import (
	"encoding/json"
	"io/ioutil"
)

type genesis struct {
	Balances map[Account]uint `json"balances"`
}

// Loads the genesis.json file from the spoecified path and conveerts into json.
func loadGenesis(path string) (genesis, error) {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return genesis{}, err
	}

	var loadedGenesis genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return genesis{}, err
	}

	return loadedGenesis, nil
}
