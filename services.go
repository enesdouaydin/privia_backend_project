// services.go
package main

import (
	"errors"
)

// KullaniciService kullanıcı ile ilgili iş mantığını ele alır
type UserService struct {
	userRepo *UserRepository
}

// Yeni bir kullanıcı servisi oluştur
func NewUserService(userRepo *UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Kullanıcı adı ve şifre ile kullanıcıyı doğrula
func (s *UserService) Authenticate(kullaniciAdi, sifre string) (*User, error) {
	kullanici, err := s.userRepo.FindByUsername(kullaniciAdi)
	if err != nil {
		return nil, errors.New("geçersiz kullanıcı adı veya şifre")
	}

	// Basit şifre kontrolü (gerçek bir uygulamada şifre hashleme kullanırdık)
	if kullanici.Password != sifre {
		return nil, errors.New("geçersiz kullanıcı adı veya şifre")
	}

	return kullanici, nil
}

// TodoService yapılacaklar ile ilgili iş mantığını ele alır
type TodoService struct {
	todoRepo *TodoRepository
}

// Yeni bir yapılacaklar servisi oluştur
func NewTodoService(todoRepo *TodoRepository) *TodoService {
	return &TodoService{
		todoRepo: todoRepo,
	}
}

// Tüm yapılacaklar listelerini getir (admin kullanıcılar için)
func (s *TodoService) GetAllLists() ([]TodoList, error) {
	return s.todoRepo.GetAllLists()
}

// Belirli bir kullanıcı için yapılacaklar listelerini getir
func (s *TodoService) GetUserLists(kullaniciID int) ([]TodoList, error) {
	return s.todoRepo.GetListsByUserID(kullaniciID)
}

// Yeni bir yapılacaklar listesi oluştur
func (s *TodoService) CreateList(liste TodoList) (TodoList, error) {
	return s.todoRepo.CreateList(liste)
}

// ID'ye göre liste getir
func (s *TodoService) GetList(id string) (TodoList, error) {
	return s.todoRepo.GetList(id)
}

// Listeyi güncelle
func (s *TodoService) UpdateList(liste TodoList) (TodoList, error) {
	return s.todoRepo.UpdateList(liste)
}

// Listeyi sil
func (s *TodoService) DeleteList(id string) error {
	return s.todoRepo.DeleteList(id)
}

// Yeni bir yapılacak öğe oluştur
func (s *TodoService) CreateItem(oge TodoItem) (TodoItem, error) {
	return s.todoRepo.CreateItem(oge)
}

// Bir liste için tüm öğeleri getir
func (s *TodoService) GetListItems(listeID string) ([]TodoItem, error) {
	return s.todoRepo.GetItemsByListID(listeID)
}

// ID'ye göre öğe getir
func (s *TodoService) GetItem(id string) (TodoItem, error) {
	return s.todoRepo.GetItem(id)
}

// Öğe güncelle
func (s *TodoService) UpdateItem(oge TodoItem) (TodoItem, error) {
	return s.todoRepo.UpdateItem(oge)
}

// Öğe sil
func (s *TodoService) DeleteItem(id string) error {
	return s.todoRepo.DeleteItem(id)
}

// Liste tamamlanma yüzdesini güncelle
func (s *TodoService) UpdateListCompletionPercentage(listeID string) error {
	// Öncelikle listenin var olup olmadığını kontrol et
	liste, err := s.GetList(listeID)
	if err != nil {
		return err
	}

	ogeler, err := s.GetListItems(listeID)
	if err != nil {
		return err
	}

	// Tamamlanma yüzdesini hesapla
	toplam := len(ogeler)
	if toplam == 0 {
		// Öğe yok, liste %0 tamamlandı
		liste.CompletionPercentage = 0
		_, err = s.UpdateList(liste)
		return err
	}

	tamamlanan := 0
	for _, oge := range ogeler {
		if oge.Completed {
			tamamlanan++
		}
	}

	yuzde := (tamamlanan * 100) / toplam

	// Listeyi güncelle
	liste.CompletionPercentage = yuzde
	_, err = s.UpdateList(liste)
	return err
}
