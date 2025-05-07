// models.go
package main

import (
	"time"
)

// Kullanıcı tipleri
const (
	RegularUserType = 1
	AdminUserType   = 2
)

// User sistemdeki bir kullanıcıyı temsil eder
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`        // Şifre JSON yanıtlarında gösterilmemelidir
	UserType int    `json:"userType"` // 1: Normal kullanıcı, 2: Admin kullanıcı
}

// TodoList sistemdeki bir yapılacaklar listesini temsil eder
type TodoList struct {
	ID                   string     `json:"id"`
	UserID               int        `json:"userId"`
	Name                 string     `json:"name" binding:"required"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	DeletedAt            *time.Time `json:"deletedAt,omitempty"`
	CompletionPercentage int        `json:"completionPercentage"`
}

// TodoItem bir listedeki yapılacak öğeyi temsil eder
type TodoItem struct {
	ID        string     `json:"id"`
	ListID    string     `json:"listId"`
	Content   string     `json:"content" binding:"required"`
	Completed bool       `json:"completed"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// Error tutarlı hata yanıtları için özel bir hata türünü temsil eder
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
