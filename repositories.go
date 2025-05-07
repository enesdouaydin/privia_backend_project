// repositories.go
package main

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// UserRepository kullanıcı veri erişiminden sorumludur
type UserRepository struct {
	users map[int]User
	mutex sync.RWMutex
}

// Önceden tanımlanmış kullanıcılarla yeni bir kullanıcı deposu oluşturur
func NewUserRepository() *UserRepository {
	repo := &UserRepository{
		users: make(map[int]User),
	}

	// Önceden tanımlanmış kullanıcıları ekle (görev gereği)
	repo.users[1] = User{ID: 1, Username: "normal", Password: "sifre", UserType: RegularUserType}
	repo.users[2] = User{ID: 2, Username: "admin", Password: "admin123", UserType: AdminUserType}

	return repo
}

// Kullanıcı adına göre bir kullanıcı bul
func (r *UserRepository) FindByUsername(username string) (*User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, errors.New("kullanıcı bulunamadı")
}

// TodoRepository yapılacaklar listeleri ve öğeleri veri erişiminden sorumludur
type TodoRepository struct {
	lists         map[string]TodoList
	items         map[string]TodoItem
	listsMutex    sync.RWMutex
	itemsMutex    sync.RWMutex
	listIDCounter int
	itemIDCounter int
}

// Yeni bir yapılacaklar deposu oluştur
func NewTodoRepository() *TodoRepository {
	return &TodoRepository{
		lists:         make(map[string]TodoList),
		items:         make(map[string]TodoItem),
		listIDCounter: 0,
		itemIDCounter: 0,
	}
}

// Yeni bir yapılacaklar listesi oluştur
func (r *TodoRepository) CreateList(list TodoList) (TodoList, error) {
	r.listsMutex.Lock()
	defer r.listsMutex.Unlock()

	// ID oluştur
	r.listIDCounter++
	list.ID = strconv.Itoa(r.listIDCounter)

	// Tamamlanma yüzdesini başlangıçta 0 olarak ayarla
	list.CompletionPercentage = 0

	r.lists[list.ID] = list
	return list, nil
}

// Tüm yapılacaklar listelerini al
func (r *TodoRepository) GetAllLists() ([]TodoList, error) {
	r.listsMutex.RLock()
	defer r.listsMutex.RUnlock()

	var aktifListeler []TodoList
	for _, list := range r.lists {
		// Sadece silinmemiş listeleri dahil et
		if list.DeletedAt == nil {
			aktifListeler = append(aktifListeler, list)
		}
	}

	return aktifListeler, nil
}

// Kullanıcı ID'sine göre listeleri al
func (r *TodoRepository) GetListsByUserID(userID int) ([]TodoList, error) {
	r.listsMutex.RLock()
	defer r.listsMutex.RUnlock()

	var kullaniciListeleri []TodoList
	for _, list := range r.lists {
		// Sadece kullanıcıya ait ve silinmemiş listeleri dahil et
		if list.UserID == userID && list.DeletedAt == nil {
			kullaniciListeleri = append(kullaniciListeleri, list)
		}
	}

	return kullaniciListeleri, nil
}

// ID'ye göre bir liste al
func (r *TodoRepository) GetList(id string) (TodoList, error) {
	r.listsMutex.RLock()
	defer r.listsMutex.RUnlock()

	list, exists := r.lists[id]
	if !exists || list.DeletedAt != nil {
		return TodoList{}, errors.New("liste bulunamadı")
	}

	return list, nil
}

// Bir listeyi güncelle
func (r *TodoRepository) UpdateList(list TodoList) (TodoList, error) {
	r.listsMutex.Lock()
	defer r.listsMutex.Unlock()

	// Listenin var olup olmadığını kontrol et
	_, exists := r.lists[list.ID]
	if !exists {
		return TodoList{}, errors.New("liste bulunamadı")
	}

	r.lists[list.ID] = list
	return list, nil
}

// Bir listeyi sil (yumuşak silme)
func (r *TodoRepository) DeleteList(id string) error {
	r.listsMutex.Lock()
	defer r.listsMutex.Unlock()

	list, exists := r.lists[id]
	if !exists {
		return errors.New("liste bulunamadı")
	}

	// DeletedAt ayarlayarak yumuşak silme
	now := time.Now()
	list.DeletedAt = &now
	list.UpdatedAt = now
	r.lists[id] = list

	return nil
}

// Yeni bir yapılacak öğe oluştur
func (r *TodoRepository) CreateItem(item TodoItem) (TodoItem, error) {
	r.itemsMutex.Lock()
	defer r.itemsMutex.Unlock()

	// ID oluştur
	r.itemIDCounter++
	item.ID = strconv.Itoa(r.itemIDCounter)

	r.items[item.ID] = item
	return item, nil
}

// Bir liste için tüm öğeleri al
func (r *TodoRepository) GetItemsByListID(listID string) ([]TodoItem, error) {
	r.itemsMutex.RLock()
	defer r.itemsMutex.RUnlock()

	var listeOgeleri []TodoItem
	for _, item := range r.items {
		// Sadece belirtilen listeye ait ve silinmemiş öğeleri dahil et
		if item.ListID == listID && item.DeletedAt == nil {
			listeOgeleri = append(listeOgeleri, item)
		}
	}

	return listeOgeleri, nil
}

// ID'ye göre bir öğe al
func (r *TodoRepository) GetItem(id string) (TodoItem, error) {
	r.itemsMutex.RLock()
	defer r.itemsMutex.RUnlock()

	item, exists := r.items[id]
	if !exists || item.DeletedAt != nil {
		return TodoItem{}, errors.New("öğe bulunamadı")
	}

	return item, nil
}

// Bir öğeyi güncelle
func (r *TodoRepository) UpdateItem(item TodoItem) (TodoItem, error) {
	r.itemsMutex.Lock()
	defer r.itemsMutex.Unlock()

	// Öğenin var olup olmadığını kontrol et
	_, exists := r.items[item.ID]
	if !exists {
		return TodoItem{}, errors.New("öğe bulunamadı")
	}

	r.items[item.ID] = item
	return item, nil
}

// Bir öğeyi sil (yumuşak silme)
func (r *TodoRepository) DeleteItem(id string) error {
	r.itemsMutex.Lock()
	defer r.itemsMutex.Unlock()

	item, exists := r.items[id]
	if !exists {
		return errors.New("öğe bulunamadı")
	}

	// DeletedAt ayarlayarak yumuşak silme
	now := time.Now()
	item.DeletedAt = &now
	item.UpdatedAt = now
	r.items[id] = item

	return nil
}
