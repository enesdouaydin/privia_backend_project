# Yapılacaklar (TODO) REST API Uygulaması

Bu uygulama, Go ve Gin framework kullanılarak geliştirilmiş, Temiz Mimari prensiplerini uygulayan basit bir Yapılacaklar (TODO) REST API uygulamasıdır.

## Özellikler

- JWT ile kullanıcı kimlik doğrulama
- İki tür kullanıcı: Normal ve Admin
- Yapılacaklar Listesi yönetimi (Oluşturma, Okuma, Güncelleme, Silme)
- Yapılacak Öğe yönetimi (Oluşturma, Okuma, Güncelleme, Silme)
- Verilerin yumuşak silinmesi
- Mock repository'ler ile bellek içi veri saklama

## Proje Yapısı

Proje aşağıdaki dosyalardan oluşmaktadır:

1. `main.go`: Uygulamanın giriş noktası, rota tanımlarını ve ara yazılımları içerir
2. `models.go`: Uygulamada kullanılan veri yapılarını içerir
3. `repositories.go`: Bellek içi depolama ile veri erişim katmanını uygular
4. `services.go`: Uygulama için iş mantığını içerir
5. `README.md`: Uygulama için dokümantasyon

## API Uç Noktaları

### Kimlik Doğrulama
- `POST /giris`: JWT token almak için kullanıcı adı ve şifre ile giriş yapma

### Yapılacaklar Listeleri
- `GET /api/listeler`: Tüm yapılacaklar listelerini getir (Admin tüm listeleri, Normal kullanıcılar kendi listelerini görür)
- `POST /api/listeler`: Yeni bir yapılacaklar listesi oluştur
- `GET /api/listeler/:id`: Belirli bir yapılacaklar listesini getir
- `PUT /api/listeler/:id`: Belirli bir yapılacaklar listesini güncelle
- `DELETE /api/listeler/:id`: Belirli bir yapılacaklar listesini sil (yumuşak silme)

### Yapılacak Öğeler
- `GET /api/listeler/:id/ogeler`: Bir listedeki tüm öğeleri getir
- `POST /api/listeler/:id/ogeler`: Bir listeye yeni bir öğe ekle
- `PUT /api/ogeler/:id`: Belirli bir öğeyi güncelle
- `DELETE /api/ogeler/:id`: Belirli bir öğeyi sil (yumuşak silme)

## Uygulamayı Çalıştırma

1. Go'nun yüklü olduğundan emin olun
2. Bu repository'yi klonlayın
3. Bağımlılıkları yükleyin: `go mod tidy`
4. Uygulamayı çalıştırın: `go run *.go`
5. API `http://localhost:8080` adresinde kullanılabilir olacaktır

## Test Kullanıcıları

Test amaçlı olarak uygulama iki önceden tanımlanmış kullanıcı ile gelir:

1. Normal Kullanıcı:
   - Kullanıcı adı: normal
   - Şifre: sifre

2. Admin Kullanıcı:
   - Kullanıcı adı: admin
   - Şifre: admin123

## Kimlik Doğrulama

Kimlik doğrulamak için `/giris` adresine kullanıcı adı ve şifre ile bir POST isteği gönderin:

```json
{
  "username": "normal",
  "password": "sifre"
}
```

Yanıt, sonraki isteklerde Authorization başlığında Bearer token olarak eklenmesi gereken bir JWT token içerecektir:

```
Authorization: Bearer <token>
```

## Örnek İstekler

### Yapılacaklar Listesi Oluşturma

```
POST /api/listeler
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Benim Yapılacaklar Listem"
}
```

### Listeye Öğe Ekleme

```
POST /api/listeler/:id/ogeler
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Market alışverişi",
  "completed": false
}
```

### Bir Öğeyi Tamamlandı Olarak İşaretleme

```
PUT /api/ogeler/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Market alışverişi",
  "completed": true
}
```

internet erisimi: https://priviabackendproject-production.up.railway.app/