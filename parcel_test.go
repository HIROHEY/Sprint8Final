package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Ошибка при добавлении посылки: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Некорректный идентификатор посылки: %d", id)
	}
	t.Logf("Посылка добавлена с ID: %d", id)

	// Получение посылки
	retrievedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Ошибка при получении посылки: %v", err)
	}

	// Проверка данных посылки
	if parcel.Number == retrievedParcel.Number {
		return
	}
	if retrievedParcel.Client != parcel.Client {
		t.Errorf("Ожидаемый клиент: %d, полученный: %d", parcel.Client, retrievedParcel.Client)
	}
	if retrievedParcel.Status != parcel.Status {
		t.Errorf("Ожидаемый статус: %s, полученный: %s", parcel.Status, retrievedParcel.Status)
	}
	if retrievedParcel.Address != parcel.Address {
		t.Errorf("Ожидаемый адрес: %s, полученный: %s", parcel.Address, retrievedParcel.Address)
	}
	if retrievedParcel.CreatedAt != parcel.CreatedAt {
		t.Errorf("Ожидаемое время создания: %v, полученное: %v", parcel.CreatedAt, retrievedParcel.CreatedAt)
	}

	// Удаление посылки
	store.Delete(id)
	if err != nil {
		t.Fatalf("Ошибка при удалении посылки: %v", err)
	}
	t.Logf("Посылка с ID %d удалена", id)

	// Проверка, что посылка удалена
	_, err = store.Get(id)
	if err == nil {
		t.Fatalf("Ошибка, посылка с номером %d не была удалена", id)
	}
	t.Log("Посылка успешно удалена и не найдена в базе данных")

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close() // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Ошибка при добавлении посылки: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Некорректный идентификатор посылки: %d", id)
	}
	t.Logf("Посылка добавлена с ID: %d", id)

	// Новый адрес
	newAddress := "new test address"

	// Обновление адреса
	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Fatalf("Ошибка при обновлении адреса: %v", err)
	}
	t.Logf("Адрес посылки с ID %d обновлен на '%s'", id, newAddress)

	// Проверка обновления адреса
	retrievedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Ошибка при получении посылки: %v", err)
	}

	// Убедимся, что адрес обновился
	if retrievedParcel.Address != newAddress {
		t.Errorf("Ожидаемый адрес: %s, полученный: %s", newAddress, retrievedParcel.Address)
	} else {
		t.Logf("Адрес успешно обновлен: %s", retrievedParcel.Address)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("Ошибка при открытии базы данных: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление посылки
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Ошибка при добавлении посылки: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Некорректный идентификатор посылки: %d", id)
	}
	t.Logf("Посылка добавлена с ID: %d", id)

	// Новый статус
	newStatus := "in transit"

	// Обновление статуса
	err = store.SetStatus(id, newStatus)
	if err != nil {
		t.Fatalf("Ошибка при обновлении статуса: %v", err)
	}
	t.Logf("Статус посылки с ID %d обновлен на '%s'", id, newStatus)

	// Проверка обновления статуса
	retrievedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Ошибка при получении посылки: %v", err)
	}

	if retrievedParcel.Status != newStatus {
		t.Errorf("Ожидаемый статус: %s, полученный: %s", newStatus, retrievedParcel.Status)
	} else {
		t.Logf("Статус успешно обновлен: %s", retrievedParcel.Status)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("Ошибка при открытии базы данных: %v", err)
	}
	defer db.Close() // настройте подключение к БД

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		if err != nil {
			t.Fatalf("Ошибка при добавлении посылки: %v", err)
		}
		if id <= 0 {
			t.Fatalf("Некорректный идентификатор посылки: %d", id)
		}

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Fatalf("Ошибка при получении посылок по клиенту: %v", err)
	}

	// Проверка количества полученных посылок
	if len(storedParcels) != len(parcels) {
		t.Fatalf("Ожидалось %d посылок, получено %d", len(parcels), len(storedParcels))
	} // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// Проверяем, что посылка есть в parcelMap
		storedParcel, err := parcelMap[parcel.Number]
		if !err {
			t.Errorf("Посылка с номером %d не найдена в map", parcel.Number)
			continue
		}
		// Проверяем значения полей
		if parcel.Client != storedParcel.Client {
			t.Errorf("Ожидаемый клиент: %d, полученный: %d", storedParcel.Client, parcel.Client)
		}
		if parcel.Status != storedParcel.Status {
			t.Errorf("Ожидаемый статус: %s, полученный: %s", storedParcel.Status, parcel.Status)
		}
		if parcel.Address != storedParcel.Address {
			t.Errorf("Ожидаемый адрес: %s, полученный: %s", storedParcel.Address, parcel.Address)
		}
		if parcel.CreatedAt != storedParcel.CreatedAt {
			t.Errorf("Ожидаемое время создания: %v, полученное: %v", storedParcel.CreatedAt, parcel.CreatedAt)
		}
	}
}
