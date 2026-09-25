# DevStream

Go ile yazılmış backend projesi. REST API, PostgreSQL ve JWT tabanlı kimlik doğrulama içerir.

## Teknolojiler

- **Go 1.27**
- **chi** — router
- **pgx v5** — PostgreSQL sürücüsü (connection pool)
- **golang-jwt** — JWT access token
- **bcrypt** — şifre hash'leme

## Proje Yapısı

```
backend/
├── main.go                      # entrypoint, router ve server
└── internal/
    ├── handlers/auth/           # auth endpoint'leri
    │   ├── handlers.go          # Register, Login
    │   └── helpers.go           # refresh token üretimi ve hash'leme
    ├── models/                  # request/claim modelleri
    │   └── auth.go
    └── store/                   # veritabanı katmanı
        └── postre.go            # bağlantı, pool config, tablo migration
```

## Kurulum

```bash
git clone https://github.com/coderian/DevStream.git
cd DevStream/backend
```

`backend/.env` dosyası oluştur:

```env
DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/devstream?sslmode=disable
JWT_SECRET=<openssl rand -hex 32 çıktısı>
```

Veritabanını oluştur ve çalıştır:

```bash
createdb devstream
go run main.go
```

Sunucu `:8080` portunda başlar, tablolar (`users`, `refresh_tokens`) açılışta otomatik oluşturulur.

## Endpoints

| Method | Path                  | Açıklama                        |
|--------|-----------------------|---------------------------------|
| POST   | `/api/v1/auth/register` | Yeni kullanıcı kaydı          |
| POST   | `/api/v1/auth/login`    | Giriş, access + refresh token |

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "coderian", "email": "test@example.com", "password": "secret"}'
```

`201 Created` → `{"id": 1, "username": "coderian", "email": "test@example.com"}`

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "coderian", "password": "secret"}'
```

`200 OK` → `{"token": "<JWT>", "refresh_token": "<random>"}`

## Güvenlik Notları

- Şifreler bcrypt ile hash'lenerek saklanır
- Refresh token veritabanında yalnızca SHA256 hash olarak tutulur
- Access token 24 saat, refresh token 7 gün geçerlidir
- `.env` dosyası asla repoya eklenmez

## Lisans

[MIT](LICENSE)
