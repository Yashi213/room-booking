package repository

import (
	"context"
	"errors"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomsRepository interface {
	ListRooms(ctx context.Context) ([]domain.Room, error)
	CreateRoom(ctx context.Context, input domain.RoomCreateInput) (domain.Room, error)
	GetRoomByID(ctx context.Context, roomID uuid.UUID) (domain.Room, error)
}

type roomsRepository struct {
	db *pgxpool.Pool
}

func NewRoomsRepository(db *pgxpool.Pool) RoomsRepository {
	return &roomsRepository{db: db}
}

func (r *roomsRepository) ListRooms(ctx context.Context) ([]domain.Room, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, capacity, created_at
		FROM rooms
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]domain.Room, 0)
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *roomsRepository) CreateRoom(ctx context.Context, input domain.RoomCreateInput) (domain.Room, error) {
	var room domain.Room

	err := r.db.QueryRow(ctx, `
		INSERT INTO rooms (name, description, capacity)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, capacity, created_at
	`, input.Name, input.Description, input.Capacity).
		Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
	if err != nil {
		return domain.Room{}, err
	}

	return room, nil
}

func (r *roomsRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (domain.Room, error) {
	var room domain.Room
	err := r.db.QueryRow(ctx, `
		SELECT id, name, description, capacity, created_at
		FROM rooms
		WHERE id = $1
	`, roomID).Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Room{}, ErrNotFound
		}
		return domain.Room{}, err
	}
	return room, nil
}
