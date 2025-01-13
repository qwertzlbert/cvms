package model

import (
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type BabylonVoteExtension struct {
	bun.BaseModel         `bun:"table:babylon_checkpoint"`
	ID                    int64     `bun:"id,pk,autoincrement"`
	ChainInfoID           int64     `bun:"chain_info_id,pk,notnull"`
	Epoch                 int64     `bun:"epoch,notnull"`
	Height                int64     `bun:"height,notnull"`
	Timestamp             time.Time `bun:"timestamp,notnull"`
	ValidatorHexAddressID int64     `bun:"validator_hex_address_id,notnull"`
	Status                int64     `bun:"status,notnull"`
}

func (bve BabylonVoteExtension) String() string {
	return fmt.Sprintf("BabylonExtensionVote<%d %d %d %d %d %d %d>",
		bve.ID,
		bve.ChainInfoID,
		bve.Epoch,
		bve.Height,
		bve.Timestamp.Unix(),
		bve.ValidatorHexAddressID,
		bve.Status,
	)
}

type RecentBabylonVoteExtension struct {
	Moniker string `bun:"moniker"`
	Epoch   int64  `bun:"epoch"`
	Height  int64  `bun:"height"`
	Status  int64  `bun:"status"`
}
