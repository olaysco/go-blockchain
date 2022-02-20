package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type State struct {
	Balances  map[Account]uint
	txMempool []Tx

	dbFile *os.File
}

// Initialises the blockchain state from disk.
func NewStateFromDisk() (*State, error) {
	cw, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	// first, loads the genesis file from disk.
	// remember that our genesis file contains the initial balance for each participant.
	genesisFilePath := filepath.Join(cw, "database", "genesis.json")
	gen, err := loadGenesis(genesisFilePath)

	if err != nil {
		return nil, err
	}

	// now from the balances returned from the genesis file, map each account to their balance. {"ola": 100}
	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	// Open the transaction file for i/o.
	// the transaction file, keeps track of how money changes hand ammong each participant
	// you had see something like, {"from": "ola", "to": "sam", "value": 100, "data": "any extra-data"}
	txDbFilePath := filepath.Join(cw, "database", "tx.db")
	f, err := os.OpenFile(txDbFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	// the state is a runtime representaition of everything happening in the system
	// the state stores the balances loaded from the genesis file
	// and also the balance reflects what remains after applying a transaction
	//, then a slice that contains a struct off all transacrtions, loaded from the dtransaction file
	// and the one that occurs during runtime.
	state := &State{balances, make([]Tx, 0), f}
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var tx Tx
		json.Unmarshal(scanner.Bytes(), &tx)

		if err := state.apply(tx); err != nil {
			return nil, err
		}
	}

	return state, nil
}

// Add new transaction to the mempool.
func (s *State) Add(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

// Persist state transactions to DB, so we can revert the state if the system crashes or we retart the system. e.g. using the NewStateFromDisk() funcrion.
func (s *State) Persist() error {
	mempool := make([]Tx, len(s.txMempool))
	copy(mempool, s.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJson, err := json.Marshal(mempool[i])
		if err != nil {
			return err
		}

		if _, err := s.dbFile.Write(append(txJson, '\n')); err != nil {
			return err
		}

		// Remove the TX written to a file from the mempool
		s.txMempool = s.txMempool[1:]
	}

	return nil
}

// Apply and validate transaction.
func (s *State) apply(tx Tx) error {
	// If the transaction is a reward, e.g. reward for mining,
	// no need to check balance.
	if tx.isReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	// If it's a proper transaction, i.e. sending token from one user to another,
	// then check is the sender has enough balance.
	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("insufficient balance")
	}

	// Debit the sender.
	s.Balances[tx.From] -= tx.Value

	// Credit thte receiver.
	s.Balances[tx.To] += tx.Value

	return nil
}

// Close all open connection to our transacrtion file.
func (s *State) Close() {
	// the transaction file is not closed immediately after opening in the NewStateFromDisk() function
	// because we will need to write to it when we call the persist function.
	s.dbFile.Close()
}
