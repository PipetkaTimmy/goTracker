package parcel

import (
	"database/sql"
	"errors"
	"time"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt time.Time
}

type ParcelStore struct {
	db *sql.DB
}

type ParcelService struct {
	store *ParcelStore
}

func (s *ParcelService) RegisterParcel(client int, address string) (int, error) {
	if address == "" {
		return 0, errors.New("адрес не может быть пустым")
	}
	result, err := s.store.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		client, ParcelStatusRegistered, address, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (s *ParcelService) GetClientParcels(client int) ([]Parcel, error) {
	rows, err := s.store.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		var createdAtStr string
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &createdAtStr)
		if err != nil {
			return nil, err
		}
		p.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}
	return parcels, nil
}

func (s *ParcelService) UpdateParcelStatus(number int, status string) error {
	if status != ParcelStatusSent && status != ParcelStatusDelivered {
		return errors.New("недопустимый статус")
	}
	_, err := s.store.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	return err
}

func (s *ParcelService) UpdateDeliveryAddress(number int, address string) error {
	if address == "" {
		return errors.New("адрес не может быть пустым")
	}
	result, err := s.store.db.Exec("UPDATE parcel SET address = ? WHERE number = ? AND status = ?",
		address, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("нельзя изменить адрес: посылка не в статусе 'registered' или не найдена")
	}
	return nil
}

func (s *ParcelService) DeleteParcel(number int) error {
	result, err := s.store.db.Exec("DELETE FROM parcel WHERE number = ? AND status = ?",
		number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("нельзя удалить посылку: не в статусе 'registered' или не найдена")
	}
	return nil
}

func NewParcelStore(db *sql.DB) *ParcelStore {
	return &ParcelStore{db: db}
}

func NewParcelService(store *ParcelStore) *ParcelService {
	return &ParcelService{store: store}
}
