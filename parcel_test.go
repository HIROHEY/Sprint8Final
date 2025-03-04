package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db) // Клон ДБ
	parcel := getTestParcel()   // Клон посылки

	//Добавление новой посылки
	id, err := store.Add(parcel)
	require.NoError(t, err) // проверка отстуствия ошибки
	assert.NotZero(t, id)   // проверка что посылке присвоен номер

	// Получение посылки
	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)                                  // проверка отстуствия ошибки
	assert.Equal(t, parcel.Client, retrievedParcel.Client)   // сравнение полей клиента
	assert.Equal(t, parcel.Status, retrievedParcel.Status)   // Сравнение полей статуса
	assert.Equal(t, parcel.Address, retrievedParcel.Address) // Сравнение полей адреса

	// Удаление посылки
	store.Delete(id)
	require.NoError(t, err) // проверка отстуствия ошибки при удалении
	_, err = store.Get(id)  // получение удаленной посылки
	assert.Error(t, err)    // Проверка что вернулась ошибка

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	//  подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel) // Добавление новой посылки
	require.NoError(t, err)

	newAddress := "тестовый адресс"        // Новый адресc
	err = store.SetAddress(id, newAddress) // Обновление адреса
	require.NoError(t, err)                //проверка ошибки при обновление адреса

	retrievedParcel, err := store.Get(id)                // Получение обновленной посылки
	require.NoError(t, err)                              // Проверка ошибки про получении посылки
	assert.Equal(t, newAddress, retrievedParcel.Address) //Сравнение адрессов новой пыслки после обновления

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel) // Добавление посылки
	require.NoError(t, err)
	assert.NotZero(t, id)

	newStatus := "Тестовый статус" // Новый статус посылки

	err = store.SetStatus(id, newStatus) // Обновление статуса
	require.NoError(t, err)              // проверка ошибки обновления статуса

	retrievedParcel, err := store.Get(id) // Получение текущего статуса посылки
	require.NoError(t, err)               // Проверка ошибки при получении статуса
	assert.Equal(t, newStatus, retrievedParcel.Status)
}

// Проверка получения посылок по ИД клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

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
		// добавление новой послыки, проверка на отсутствие ошибки при добавление, проверка что посылка добавилась
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		assert.NotZero(t, id)

		// обновляем идентификатор добавленной посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получение всех посылок клиента
	require.NoError(t, err)                         // проверка что нет ошибки при получении посылок
	assert.Len(t, parcels, len(storedParcels))      // сравниваем  количество полученных посылок с количеством добавленных

	// check
	// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
	// убедитесь, что все посылки из storedParcels есть в parcelMap
	// убедитесь, что значения полей полученных посылок заполнены верно
	for _, parcel := range storedParcels { // перебираем слайс посылок
		// проверяем что все посылки из storedParcels есть в parcelMap и возвращаем ошибку если нет
		_, yesOrNot := parcelMap[parcel.Number]
		assert.True(t, yesOrNot, parcel.Number)
		assert.Equal(t, parcelMap[parcel.Number], parcel) // Проверяем, что значения полей совпадают

	}
}
