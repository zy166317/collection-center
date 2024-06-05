package dao

import (
	"collection-center/library/utils"
	"collection-center/service/db"
	"time"

	"golang.org/x/xerrors"
)

//type Wallet struct {
//	Address      string    `xorm:"pk not null default '' varchar(42) 'address'"`
//	Encryptedkey string    `xorm:"default '' varchar(128) 'encryptedkey'"`
//	Id           int64     `xorm:"default NULL int 'id'"`
//	CreatedAt    time.Time `xorm:"default 'CURRENT_TIMESTAMP' timestamp 'created_at'"`
//	UpdatedAt    time.Time `xorm:"default 'CURRENT_TIMESTAMP' timestamp 'updated_at'"`
//}

type Wallets struct {
	Id           int64     `json:"id" form:"id"`
	Address      string    `json:"address" form:"address"`
	EncryptedKey string    `json:"encrypted_key" form:"encrypted_key"`
	CreatedAt    time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
}

func (w *Wallets) TableName() string {
	return "wallets"
}

func InsertWallet(wallet *Wallets) (int64, error) {
	row, err := db.Client().InsertOne(wallet)
	if err != nil {
		return 0, err
	}
	if row != 1 {
		return 0, xerrors.New("Insert failed")
	}
	return wallet.Id, nil
}

func SelectOrderByAddr(addr string) (*Wallets, error) {
	data := Wallets{}
	_, err := db.Client().Where("address = ?", addr).Get(&data)
	if err != nil {
		return nil, err
	}
	encryedKey, err := utils.Decrypt(data.EncryptedKey)
	if err != nil {
		return nil, err
	}
	data.EncryptedKey = encryedKey
	return &data, nil
}
