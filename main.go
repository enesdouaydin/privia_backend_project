package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT gizli anahtarı - Gerçek ortamda güvenli şekilde saklanmalıdır
var jwtSecretKey = []byte("guvenli_gizli_anahtar")

// Ana giriş noktası - Rotaları ayarlar ve sunucuyu başlatır
func main() {
	router := gin.Default()

	// HTML template yükleme (Docker içi yol)
	router.LoadHTMLFiles(filepath.Join("/app", "son.html"))

	// Ana sayfa endpoint'i
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "son.html", nil)
	})

	// PORT ayarı (Railway uyumlu)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Sunucuyu başlat
	router.Run(":" + port)

	// Ana sayfa için route ekleyin:
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "son.html", nil) // "son.html" dosyasını render et
	})

	// Sabit CORS yapılandırması - OPTIONS metodunu ve uygun başlıkları ekledik
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Servisleri ve repository'leri başlat
	kullaniciRepo := NewUserRepository()
	yapilacakRepo := NewTodoRepository()

	kullaniciService := NewUserService(kullaniciRepo)
	yapilacakService := NewTodoService(yapilacakRepo)

	// Genel erişim rotaları

	router.POST("/giris", func(c *gin.Context) {
		var girisIstegi struct {
			KullaniciAdi string `json:"username" binding:"required"`
			Sifre        string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&girisIstegi); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"hata": "Geçersiz istek formatı"})
			return
		}

		kullanici, err := kullaniciService.Authenticate(girisIstegi.KullaniciAdi, girisIstegi.Sifre)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"hata": "Geçersiz kullanıcı adı veya şifre"})
			return
		}

		// Token oluştur
		claims := jwt.MapClaims{
			"id":       kullanici.ID,
			"username": kullanici.Username,
			"userType": kullanici.UserType,
			"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token 24 saat sonra geçersiz olur
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtSecretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": "Token oluşturma başarısız"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	// Tüm API endpoint'leri için OPTIONS isteklerini ele al
	router.OPTIONS("/api/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Status(http.StatusNoContent)
	})

	// Korumalı rotalar
	api := router.Group("/api")
	api.Use(authMiddleware())

	// Yapılacaklar listesi endpoint'leri
	api.GET("/listeler", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		var listeler []TodoList
		var err error

		if kullaniciTipi.(int) == AdminUserType {
			listeler, err = yapilacakService.GetAllLists()
		} else {
			listeler, err = yapilacakService.GetUserLists(kullaniciID.(int))
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		c.JSON(http.StatusOK, listeler)
	})

	api.POST("/listeler", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")

		var yapilacakListesi TodoList
		if err := c.ShouldBindJSON(&yapilacakListesi); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"hata": "Geçersiz istek formatı"})
			return
		}

		yapilacakListesi.UserID = kullaniciID.(int)
		yapilacakListesi.CreatedAt = time.Now()
		yapilacakListesi.UpdatedAt = time.Now()

		yeniListe, err := yapilacakService.CreateList(yapilacakListesi)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, yeniListe)
	})

	api.GET("/listeler/:id", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		listeID := c.Param("id")

		liste, err := yapilacakService.GetList(listeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"hata": "Liste bulunamadı"})
			return
		}

		// Kullanıcının bu listeye erişim hakkı var mı kontrol et
		if liste.UserID != kullaniciID.(int) && kullaniciTipi.(int) != AdminUserType {
			c.JSON(http.StatusForbidden, gin.H{"hata": "Erişim reddedildi"})
			return
		}

		c.JSON(http.StatusOK, liste)
	})

	api.PUT("/listeler/:id", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		listeID := c.Param("id")

		mevcutListe, err := yapilacakService.GetList(listeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"hata": "Liste bulunamadı"})
			return
		}

		// Kullanıcının bu listeye erişim hakkı var mı kontrol et
		if mevcutListe.UserID != kullaniciID.(int) && kullaniciTipi.(int) != AdminUserType {
			c.JSON(http.StatusForbidden, gin.H{"hata": "Erişim reddedildi"})
			return
		}

		var guncellenmisListe TodoList
		if err := c.ShouldBindJSON(&guncellenmisListe); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"hata": "Geçersiz istek formatı"})
			return
		}

		guncellenmisListe.ID = mevcutListe.ID
		guncellenmisListe.UserID = mevcutListe.UserID
		guncellenmisListe.CreatedAt = mevcutListe.CreatedAt
		guncellenmisListe.UpdatedAt = time.Now()

		sonuc, err := yapilacakService.UpdateList(guncellenmisListe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		c.JSON(http.StatusOK, sonuc)
	})

	api.DELETE("/listeler/:id", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		listeID := c.Param("id")

		liste, err := yapilacakService.GetList(listeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"hata": "Liste bulunamadı"})
			return
		}

		// Kullanıcının bu listeye erişim hakkı var mı kontrol et
		if liste.UserID != kullaniciID.(int) && kullaniciTipi.(int) != AdminUserType {
			c.JSON(http.StatusForbidden, gin.H{"hata": "Erişim reddedildi"})
			return
		}

		err = yapilacakService.DeleteList(listeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	})

	// Yapılacak öğe endpoint'leri
	api.GET("/listeler/:id/ogeler", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		listeID := c.Param("id")

		liste, err := yapilacakService.GetList(listeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"hata": "Liste bulunamadı"})
			return
		}

		// Kullanıcının bu listeye erişim hakkı var mı kontrol et
		if liste.UserID != kullaniciID.(int) && kullaniciTipi.(int) != AdminUserType {
			c.JSON(http.StatusForbidden, gin.H{"hata": "Erişim reddedildi"})
			return
		}

		ogeler, err := yapilacakService.GetListItems(listeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ogeler)
	})

	api.POST("/listeler/:id/ogeler", func(c *gin.Context) {
		listeID := c.Param("id")

		// Eklenecek kontrol:
		if listeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"hata": "Liste ID boş olamaz"})
			return
		}

		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		liste, err := yapilacakService.GetList(listeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"hata": "Liste bulunamadı"})
			return
		}

		// Kullanıcının bu listeye erişim hakkı var mı kontrol et
		if liste.UserID != kullaniciID.(int) && kullaniciTipi.(int) != AdminUserType {
			c.JSON(http.StatusForbidden, gin.H{"hata": "Erişim reddedildi"})
			return
		}

		var yapilacakOge TodoItem
		if err := c.ShouldBindJSON(&yapilacakOge); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"hata": "Geçersiz istek formatı"})
			return
		}

		yapilacakOge.ListID = liste.ID
		yapilacakOge.CreatedAt = time.Now()
		yapilacakOge.UpdatedAt = time.Now()

		yeniOge, err := yapilacakService.CreateItem(yapilacakOge)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		// Liste tamamlanma yüzdesini güncelle
		yapilacakService.UpdateListCompletionPercentage(listeID)

		c.JSON(http.StatusCreated, yeniOge)
	})

	api.PUT("/ogeler/:id", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		ogeID := c.Param("id")

		// Mevcut öğeyi al
		mevcutOge, err := yapilacakService.GetItem(ogeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"hata":  "Öğe bulunamadı",
				"detay": err.Error(),
			})
			return
		}

		// Kullanıcının bu listeye erişim hakkı kontrolü
		liste, err := yapilacakService.GetList(mevcutOge.ListID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"hata":  "Liste bilgisi alınamadı",
				"detay": err.Error(),
			})
			return
		}

		if liste.UserID != kullaniciID.(int) && kullaniciTipi.(int) != AdminUserType {
			c.JSON(http.StatusForbidden, gin.H{
				"hata":  "Erişim reddedildi",
				"detay": "Bu öğeyi güncellemek için yetkiniz yok",
			})
			return
		}

		// Gelen veriyi bağla
		var guncellemeData struct {
			Content   *string `json:"content"`
			Completed *bool   `json:"completed"`
		}

		if err := c.ShouldBindJSON(&guncellemeData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"hata":  "Geçersiz istek formatı",
				"detay": err.Error(),
			})
			return
		}

		// Sadece sağlanan alanları güncelle
		if guncellemeData.Content != nil {
			mevcutOge.Content = *guncellemeData.Content
		}

		if guncellemeData.Completed != nil {
			mevcutOge.Completed = *guncellemeData.Completed
		}

		mevcutOge.UpdatedAt = time.Now()

		// Öğeyi güncelle
		guncellenmisOge, err := yapilacakService.UpdateItem(mevcutOge)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"hata":  "Öğe güncelleme hatası",
				"detay": err.Error(),
			})
			return
		}

		// Liste tamamlanma yüzdesini güncelle
		if err := yapilacakService.UpdateListCompletionPercentage(mevcutOge.ListID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"hata":  "Tamamlanma yüzdesi güncellenemedi",
				"detay": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, guncellenmisOge)
	})

	api.DELETE("/ogeler/:id", func(c *gin.Context) {
		kullaniciID, _ := c.Get("userID")
		kullaniciTipi, _ := c.Get("userType")

		ogeID := c.Param("id")

		oge, err := yapilacakService.GetItem(ogeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"hata": "Öğe bulunamadı"})
			return
		}

		// Kullanıcının bu listeye erişim hakkı var mı kontrol et
		liste, err := yapilacakService.GetList(oge.ListID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		if liste.UserID != kullaniciID.(int) && kullaniciTipi.(int) != AdminUserType {
			c.JSON(http.StatusForbidden, gin.H{"hata": "Erişim reddedildi"})
			return
		}

		err = yapilacakService.DeleteItem(ogeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"hata": err.Error()})
			return
		}

		// Liste tamamlanma yüzdesini güncelle
		yapilacakService.UpdateListCompletionPercentage(oge.ListID)

		c.Status(http.StatusNoContent)
	})

	// Sunucuyu başlat
	log.Println("Sunucu 8080 portunda başlatılıyor...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Rotaları korumak için kimlik doğrulama ara katmanı
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"hata": "Yetkilendirme başlığı eksik"})
			c.Abort()
			return
		}

		// Doğru Bearer formatını kontrol et
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"hata": "Geçersiz yetkilendirme formatı, 'Bearer TOKEN' bekleniyordu"})
			c.Abort()
			return
		}

		// Authorization başlığından token'ı çıkar (Bearer token)
		tokenString := authHeader[7:] // "Bearer " önekini kaldır

		// Token'ı ayrıştır ve doğrula
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// İmzalama yöntemini doğrula
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("beklenmeyen imzalama yöntemi: %v", token.Header["alg"])
			}
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"hata": "Geçersiz veya süresi dolmuş token"})
			c.Abort()
			return
		}

		// Talepleri çıkar
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"hata": "Token talepleri ayrıştırılamadı"})
			c.Abort()
			return
		}

		// Gerekli talepleri doğrula
		kullaniciID, ok1 := claims["id"].(float64)
		kullaniciAdi, ok2 := claims["username"].(string)
		kullaniciTipi, ok3 := claims["userType"].(float64)

		if !ok1 || !ok2 || !ok3 {
			c.JSON(http.StatusUnauthorized, gin.H{"hata": "Geçersiz token talep formatı"})
			c.Abort()
			return
		}

		// Kullanıcı bilgilerini bağlama ayarla
		c.Set("userID", int(kullaniciID))
		c.Set("username", kullaniciAdi)
		c.Set("userType", int(kullaniciTipi))

		c.Next()
	}
}
