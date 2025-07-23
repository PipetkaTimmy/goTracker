package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"goTracker/internal/parcel"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	)`)

	if err != nil {
		log.Fatal(err)
	}

	store := parcel.NewParcelStore(db)
	service := parcel.NewParcelService(store)

	clientID := 1
	address := "ул. Пример, 1"

	number, err := service.RegisterParcel(clientID, address)
	if err != nil {
		log.Fatalf("Ошибка регистрации: %v", err)
	}
	fmt.Printf("Посылка зарегистрирована, номер: %d\n", number)

	parcels, err := service.GetClientParcels(clientID)
	if err != nil {
		log.Fatalf("Ошибка получения списка: %v", err)
	}
	fmt.Println("Список посылок клиента:")
	for _, p := range parcels {
		fmt.Printf("Номер: %d, Статус: %s, Адрес: %s, Дата: %s\n", p.Number, p.Status, p.Address, p.CreatedAt.Format(time.RFC3339))
	}

	err = service.UpdateParcelStatus(number, parcel.ParcelStatusSent)
	if err != nil {
		log.Fatalf("Ошибка изменения статуса: %v", err)
	}
	fmt.Printf("Статус посылки %d изменен на 'sent'\n", number)

	err = service.UpdateDeliveryAddress(number, "ул. Новый адрес, 2")
	if err != nil {
		fmt.Printf("Ошибка изменения адреса: %v (ожидаемо, статус уже не 'registered')\n", err)
	}

	err = service.DeleteParcel(number)
	if err != nil {
		fmt.Printf("Ошибка удаления: %v (ожидаемо, статус уже не 'registered')\n", err)
	}
}
