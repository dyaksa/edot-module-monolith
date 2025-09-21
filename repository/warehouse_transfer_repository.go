package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type warehouseTransferRepository struct {
	db pqsql.Client
}

// Create implements domain.WarehouseTransferRepository.
func (wtr *warehouseTransferRepository) Create(ctx context.Context, tx *sql.Tx, transfer *domain.WarehouseTransfer) error {
	query := sq.Insert("warehouse_transfers").
		Columns("id", "from_warehouse_id", "to_warehouse_id", "status", "created_at").
		Values(transfer.ID, transfer.FromWarehouseID, transfer.ToWarehouseID, transfer.Status, "now()").
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// GetByID implements domain.WarehouseTransferRepository.
func (wtr *warehouseTransferRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WarehouseTransfer, error) {
	query := sq.Select("id", "from_warehouse_id", "to_warehouse_id", "status", "created_at").
		From("warehouse_transfers").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var transfer domain.WarehouseTransfer
	err = wtr.db.Database().QueryRowContext(ctx, q, args...).Scan(
		&transfer.ID,
		&transfer.FromWarehouseID,
		&transfer.ToWarehouseID,
		&transfer.Status,
		&transfer.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load items
	items, err := wtr.GetItemsByTransferID(ctx, id)
	if err != nil {
		return nil, err
	}
	transfer.Items = items

	return &transfer, nil
}

// UpdateStatus implements domain.WarehouseTransferRepository.
func (wtr *warehouseTransferRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status domain.TransferStatus) error {
	query := sq.Update("warehouse_transfers").
		Set("status", status).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// GetByWarehouse implements domain.WarehouseTransferRepository.
func (wtr *warehouseTransferRepository) GetByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]domain.WarehouseTransfer, int, error) {
	// Get total count
	countQuery := sq.Select("COUNT(*)").
		From("warehouse_transfers").
		Where(sq.Or{
			sq.Eq{"from_warehouse_id": warehouseID},
			sq.Eq{"to_warehouse_id": warehouseID},
		}).
		PlaceholderFormat(sq.Dollar)

	cq, cargs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = wtr.db.Database().QueryRowContext(ctx, cq, cargs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get transfers
	query := sq.Select("id", "from_warehouse_id", "to_warehouse_id", "status", "created_at").
		From("warehouse_transfers").
		Where(sq.Or{
			sq.Eq{"from_warehouse_id": warehouseID},
			sq.Eq{"to_warehouse_id": warehouseID},
		}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := wtr.db.Database().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transfers []domain.WarehouseTransfer
	for rows.Next() {
		var transfer domain.WarehouseTransfer
		err := rows.Scan(
			&transfer.ID,
			&transfer.FromWarehouseID,
			&transfer.ToWarehouseID,
			&transfer.Status,
			&transfer.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		transfers = append(transfers, transfer)
	}

	return transfers, total, nil
}

// GetActiveTransfersByWarehouse implements domain.WarehouseTransferRepository.
func (wtr *warehouseTransferRepository) GetActiveTransfersByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]domain.WarehouseTransfer, error) {
	query := sq.Select("id", "from_warehouse_id", "to_warehouse_id", "status", "created_at").
		From("warehouse_transfers").
		Where(sq.And{
			sq.Or{
				sq.Eq{"from_warehouse_id": warehouseID},
				sq.Eq{"to_warehouse_id": warehouseID},
			},
			sq.Eq{"status": []domain.TransferStatus{
				domain.TransferStatusRequested,
				domain.TransferStatusApproved,
				domain.TransferStatusInTransit,
			}},
		}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := wtr.db.Database().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []domain.WarehouseTransfer
	for rows.Next() {
		var transfer domain.WarehouseTransfer
		err := rows.Scan(
			&transfer.ID,
			&transfer.FromWarehouseID,
			&transfer.ToWarehouseID,
			&transfer.Status,
			&transfer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// CreateItems implements domain.WarehouseTransferRepository.
func (wtr *warehouseTransferRepository) CreateItems(ctx context.Context, tx *sql.Tx, items []domain.WarehouseTransferItem) error {
	if len(items) == 0 {
		return nil
	}

	query := sq.Insert("warehouse_transfer_items").
		Columns("id", "transfer_id", "product_id", "qty")

	for _, item := range items {
		query = query.Values(item.ID, item.TransferID, item.ProductID, item.Qty)
	}

	query = query.PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// GetItemsByTransferID implements domain.WarehouseTransferRepository.
func (wtr *warehouseTransferRepository) GetItemsByTransferID(ctx context.Context, transferID uuid.UUID) ([]domain.WarehouseTransferItem, error) {
	query := sq.Select("id", "transfer_id", "product_id", "qty").
		From("warehouse_transfer_items").
		Where(sq.Eq{"transfer_id": transferID}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := wtr.db.Database().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.WarehouseTransferItem
	for rows.Next() {
		var item domain.WarehouseTransferItem
		err := rows.Scan(&item.ID, &item.TransferID, &item.ProductID, &item.Qty)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func NewWarehouseTransferRepository(db pqsql.Client) domain.WarehouseTransferRepository {
	return &warehouseTransferRepository{db: db}
}
