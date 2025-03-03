package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, fmt.Errorf("ошибка при добавлении посылки: %w", err)
	}

	// верните идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении ID: %w", err)
	}

	return int(id), nil

}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	p := Parcel{}
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number", sql.Named("number", number))
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если запись не найдена, возвращаем ошибку
			return Parcel{}, fmt.Errorf("посылка с номером %d не найдена", number)
		}
		// Возвращаем другие ошибки
		return Parcel{}, fmt.Errorf("ошибка при сканировании данных: %w", err)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE Client = :client", sql.Named("client", client))
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	// срез для хранения результатов
	var parcels []Parcel

	// Читаем данные из rows
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных: %w", err)
		}
		parcels = append(parcels, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET Status = :status WHERE Number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("ошибка при обновлении статуса посылки: %w", err)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Получаем текущий статус посылки
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE Number = :number", sql.Named("number", number)).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("посылка с номером %d не найдена", number)
		}
		return fmt.Errorf("ошибка при получении статуса посылки: %w", err)
	}

	// Проверяем, что статус равен "registered"
	if currentStatus != "registered" {
		return fmt.Errorf("нельзя изменить адрес: статус посылки должен быть 'registered', текущий статус: '%s'", currentStatus)
	}

	// Обновляем адрес, если статус равен "registered"
	_, err = s.db.Exec("UPDATE parcel SET Address = :address WHERE Number = :number",
		sql.Named("address", address),
		sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("ошибка при обновлении адреса посылки: %w", err)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered

	// Получаем текущий статус посылки
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number", sql.Named("number", number)).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("посылка с номером %d не найдена", number)
		}
		return fmt.Errorf("ошибка при получении статуса посылки: %w", err)
	}

	// Проверяем, что статус равен "registered"
	if currentStatus == "registered" {
		_, err = s.db.Exec("DELETE FROM parcel WHERE Number = :number", sql.Named("number", number))
		if err != nil {
			return fmt.Errorf("ошибка при удалении посылки: %w", err)
		}
	}
	return nil
}
