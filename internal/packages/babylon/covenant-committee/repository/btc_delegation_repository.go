package repository

import (
	"context"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	indexerrepo "github.com/cosmostation/cvms/internal/common/indexer/repository"
	dbhelper "github.com/cosmostation/cvms/internal/helper/db"
	"github.com/cosmostation/cvms/internal/packages/babylon/covenant-committee/model"
	"github.com/uptrace/bun"
)

const SubIndexName = "babylon_btc_delegation"

type BtcDelegationRepository struct {
	sqlTimeout time.Duration
	*bun.DB
	indexerrepo.IMetaRepository
}

func NewBtcDelegationRepository(indexerDB common.IndexerDB, sqlTimeout time.Duration) BtcDelegationRepository {
	// Instantiate the meta repository
	metarepo := indexerrepo.NewMetaRepository(indexerDB)

	// Return a repository that implements both IMetaRepository and vote-specific logic
	return BtcDelegationRepository{sqlTimeout, indexerDB.DB, metarepo}
}

func (repo *BtcDelegationRepository) DeleteOldBtcDelegationList(chainID, retentionPeriod string) (
	/* deleted rows */ int64,
	/* unexpected error */ error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	// Parsing retention period
	duration, err := dbhelper.ParseRetentionPeriod(retentionPeriod)
	if err != nil {
		return 0, err
	}

	// Calculate cutoff time duration
	cutoffTime := time.Now().Add(duration)

	// Make partition table name
	partitionTableName := dbhelper.MakePartitionTableName(IndexName, chainID)

	// Query Execution
	res, err := repo.NewDelete().
		Model((*model.BabylonBtcDelegation)(nil)).
		ModelTableExpr(partitionTableName).
		Where("timestamp < ?", cutoffTime).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}
