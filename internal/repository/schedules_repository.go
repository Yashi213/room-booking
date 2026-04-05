package repository

import (
	"context"
	"errors"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SchedulesRepository interface {
	CreateSchedule(ctx context.Context, input domain.ScheduleCreateInput) (domain.Schedule, error)
	GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, error)
	ListAllSchedules(ctx context.Context) ([]domain.Schedule, error)
}

type schedulesRepository struct {
	db *pgxpool.Pool
}

func NewSchedulesRepository(db *pgxpool.Pool) SchedulesRepository {
	return &schedulesRepository{db: db}
}

func (r *schedulesRepository) CreateSchedule(ctx context.Context, input domain.ScheduleCreateInput) (domain.Schedule, error) {
	var schedule domain.Schedule

	err := r.db.QueryRow(ctx, `
		INSERT INTO schedules (room_id, days_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, days_of_week, start_time, end_time
	`, input.RoomID, input.DaysOfWeek, input.StartTime, input.EndTime).
		Scan(&schedule.ID, &schedule.RoomID, &schedule.DaysOfWeek, &schedule.StartTime, &schedule.EndTime)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Schedule{}, ErrConflict
		}
		return domain.Schedule{}, err
	}

	return schedule, nil
}

func (r *schedulesRepository) GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, error) {
	var schedule domain.Schedule

	err := r.db.QueryRow(ctx, `
		SELECT id, room_id, days_of_week, start_time, end_time
		FROM schedules
		WHERE room_id = $1
	`, roomID).Scan(&schedule.ID, &schedule.RoomID, &schedule.DaysOfWeek, &schedule.StartTime, &schedule.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Schedule{}, ErrNotFound
		}
		return domain.Schedule{}, err
	}

	return schedule, nil
}

func (r *schedulesRepository) ListAllSchedules(ctx context.Context) ([]domain.Schedule, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, room_id, days_of_week, start_time, end_time
		FROM schedules
		ORDER BY room_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Schedule
	for rows.Next() {
		var s domain.Schedule
		if err := rows.Scan(&s.ID, &s.RoomID, &s.DaysOfWeek, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
