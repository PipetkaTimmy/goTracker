package parcel

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestRegisterParcel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	id, err := service.RegisterParcel(1, "ул. Пример, 1")
	if err != nil {
		t.Errorf("Ошибка регистрации: %v", err)
	}
	if id <= 0 {
		t.Error("Некорректный ID посылки")
	}

	var status string
	err = db.QueryRow("SELECT status FROM parcel WHERE number = ?", id).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}
	if status != ParcelStatusRegistered {
		t.Errorf("Неверный статус: ожидался %s, получен %s", ParcelStatusRegistered, status)
	}
}

func TestGetClientParcels(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	_, err := service.RegisterParcel(1, "ул. Пример, 1")
	if err != nil {
		t.Fatal(err)
	}

	parcels, err := service.GetClientParcels(1)
	if err != nil {
		t.Errorf("Ошибка получения списка: %v", err)
	}
	if len(parcels) == 0 {
		t.Error("Список посылок пуст")
	}
}

func TestUpdateParcelStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	id, err := service.RegisterParcel(1, "ул. Пример, 1")
	if err != nil {
		t.Fatal(err)
	}

	err = service.UpdateParcelStatus(id, ParcelStatusSent)
	if err != nil {
		t.Errorf("Ошибка изменения статуса: %v", err)
	}

	var status string
	err = db.QueryRow("SELECT status FROM parcel WHERE number = ?", id).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}
	if status != ParcelStatusSent {
		t.Errorf("Неверный статус: ожидался %s, получен %s", ParcelStatusSent, status)
	}
}

func TestUpdateDeliveryAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	id, err := service.RegisterParcel(1, "ул. Пример, 1")
	if err != nil {
		t.Fatal(err)
	}

	err = service.UpdateDeliveryAddress(id, "ул. Новый адрес, 2")
	if err != nil {
		t.Errorf("Ошибка изменения адреса: %v", err)
	}

	var address string
	err = db.QueryRow("SELECT address FROM parcel WHERE number = ?", id).Scan(&address)
	if err != nil {
		t.Fatal(err)
	}
	if address != "ул. Новый адрес, 2" {
		t.Errorf("Неверный адрес: ожидался 'ул. Новый адрес, 2', получен %s", address)
	}
}

func TestDeleteParcel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	id, err := service.RegisterParcel(1, "ул. Пример, 1")
	if err != nil {
		t.Fatal(err)
	}

	err = service.UpdateParcelStatus(id, ParcelStatusSent)
	if err != nil {
		t.Fatal(err)
	}

	err = service.DeleteParcel(id)
	if err == nil {
		t.Error("Удаление должно быть запрещено, так как статус не 'registered'")
	}

	_, err = db.Exec("UPDATE parcel SET status = ? WHERE number = ?", ParcelStatusRegistered, id)
	if err != nil {
		t.Fatal(err)
	}

	err = service.DeleteParcel(id)
	if err != nil {
		t.Errorf("Ошибка удаления: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM parcel WHERE number = ?", id).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Error("Посылка не удалена")
	}
}
