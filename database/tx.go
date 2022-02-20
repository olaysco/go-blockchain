package database

type Account string

func NewAccount(value string) Account {
	return Account(value)
}

// The database state changes are called Transactions (TX).
// Transactions are old fashion Events representing actions within the system.
type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
}

func NewTx(from Account, to Account, value uint, data string) Tx {
	return Tx{from, to, value, data}
}

func (tx Tx) isReward() bool {
	return tx.Data == "reward"
}
