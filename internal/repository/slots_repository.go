package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotsRepository interface {
	GetSlotByID(ctx context.Context, slotID uuid.UUID) (domain.Slot, error)
	ListSlots(ctx context.Context, input domain.SlotsListInput) ([]domain.Slot, error)
	// EnsureSlot inserts a slot when (room_id, start_time) is new; otherwise no-op.
	EnsureSlot(ctx context.Context, slot domain.SlotCreateInput) error
}

type slotsRepository struct {
	db *pgxpool.Pool
}

func NewSlotsRepository(db *pgxpool.Pool) SlotsRepository {
	return &slotsRepository{db: db}
}

func (r *slotsRepository) GetSlotByID(ctx context.Context, slotID uuid.UUID) (domain.Slot, error) {
	var slot domain.Slot
	err := r.db.QueryRow(ctx, `
		SELECT id, room_id, start_time, end_time
		FROM slots
		WHERE id = $1
	`, slotID).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Slot{}, ErrNotFound
		}
		return domain.Slot{}, err
	}
	return slot, nil
}

func (r *slotsRepository) ListSlots(ctx context.Context, input domain.SlotsListInput) ([]domain.Slot, error) {
	dayStart := input.Date.UTC().Truncate(24 * time.Hour)
	dayEnd := dayStart.AddDate(0, 0, 1)

	rows, err := r.db.Query(ctx, `
		SELECT id, room_id, start_time, end_time
		FROM slots
		WHERE room_id = $1
		  AND start_time >= $2
		  AND start_time < $3
		ORDER BY start_time ASC
	`, input.RoomID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]domain.Slot, 0)
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slots, nil
}

func (r *slotsRepository) EnsureSlot(ctx context.Context, slot domain.SlotCreateInput) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO slots (room_id, start_time, end_time)
		SELECT $1, $2, $3
		WHERE NOT EXISTS (
			SELECT 1 FROM slots s WHERE s.room_id = $1 AND s.start_time = $2
		)
	`, slot.RoomID, slot.Start, slot.End)
	return err
}
