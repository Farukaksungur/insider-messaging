# Insider Messaging System

Bu proje, otomatik mesaj gönderme sistemi için geliştirilmiş bir Go uygulamasıdır. Sistem, veritabanındaki mesajları belirli aralıklarla toplu olarak webhook endpoint'lerine gönderir.

## 🎯 Proje Hakkında

Bu sistem şu şekilde çalışır:
- Mesajlar REST API üzerinden oluşturulur ve veritabanına kaydedilir
- Scheduler (zamanlayıcı) her 2 dakikada bir çalışır ve gönderilmemiş mesajları alır
- Her batch'te 2 mesaj (ayarlanabilir) webhook URL'ine gönderilir
- Gönderilen mesajlar veritabanında işaretlenir ve tekrar gönderilmez
- Mesaj ID'leri ve gönderme zamanları Redis'te cache'lenir (bonus özellik)

## 🚀 Hızlı Başlangıç

### Gereksinimler

- Docker ve Docker Compose yüklü olmalı
- Webhook.site URL'i (test için)

### Adım 1: Webhook.site URL'i Alın

1. Tarayıcınızda https://webhook.site adresine gidin
2. Yeni bir webhook URL'i oluşturun (örnek: `https://webhook.site/c3f13233-1ed4-429e-9649-8133b3b9c9cd`)
3. **Önemli:** Webhook.site'da "Edit" butonuna tıklayın ve şu ayarları yapın:
   - **Status code:** `202` (veya `200`)
   - **Content type:** `application/json` (mutlaka!)
   - **Content (Response body):**
     ```json
     {
       "message": "Accepted",
       "messageId": "{{uuid}}"
     }
     ```
   - **Save** butonuna tıklayın

**Not:** `{{uuid}}` yazabilirsiniz veya boş bırakabilirsiniz. Uygulama otomatik olarak benzersiz bir UUID oluşturacaktır.

### Adım 2: Projeyi Çalıştırın

Proje klasörüne gidin ve şu komutu çalıştırın:

```bash
docker-compose up --build
```

Bu komut şunları yapar:
- MariaDB veritabanını başlatır
- Redis cache'i başlatır
- Go uygulamasını derler ve çalıştırır

**İlk çalıştırmada biraz zaman alabilir** çünkü Docker image'ları indirilir ve uygulama derlenir.

### Adım 3: Servislerin Hazır Olduğunu Kontrol Edin

Terminal çıktısında şunu görmelisiniz:
```
insider-messaging-app-1 | Connected to MySQL
insider-messaging-app-1 | Connected to Redis
insider-messaging-app-1 | server started on :8080
```

Eğer bu mesajları görüyorsanız, her şey hazır demektir!

### Adım 4: Webhook URL'ini Yapılandırın

`docker-compose.yml` dosyasını açın ve şu satırı bulun:
```yaml
WEBHOOK_URL: ${WEBHOOK_URL:-https://webhook.site/YOUR-ID}
```

Bu satırı kendi webhook URL'inizle değiştirin:
```yaml
WEBHOOK_URL: https://webhook.site/c3f13233-1ed4-429e-9649-8133b3b9c9cd
```

**Alternatif:** `.env` dosyası oluşturup orada da tanımlayabilirsiniz:
```env
WEBHOOK_URL=https://webhook.site/c3f13233-1ed4-429e-9649-8133b3b9c9cd
WEBHOOK_AUTH_KEY=INS.me1x9uMcyYG1hKKQVPoc.b03j9aZwRTOCA2Ywo
API_KEY=your-secret-api-key-here
```

Değişiklik yaptıktan sonra servisleri yeniden başlatın:
```bash
docker-compose restart app
```

## 📖 Sistem Nasıl Çalışır?

### 1. Mesaj Oluşturma

Mesajlar REST API üzerinden oluşturulur. Her mesaj veritabanına kaydedilir ve `sent=false` olarak işaretlenir.

**Örnek:**
```bash
curl -X POST "http://localhost:8080/api/messages" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '{
    "to": "+905551111111",
    "content": "Hello, this is a test message"
  }'
```

### 2. Otomatik Gönderme

Scheduler'ı başlattığınızda:
- Her 2 dakikada bir (varsayılan) çalışır
- Her batch'te 2 mesaj (varsayılan) gönderir
- Sadece `sent=false` olan mesajları gönderir
- Gönderilen mesajlar `sent=true` olarak işaretlenir

**Scheduler'ı başlatmak için:**
```bash
curl -X POST "http://localhost:8080/api/auto?action=start" \
  -H "X-API-Key: your-secret-api-key-here"
```

**Scheduler'ı durdurmak için:**
```bash
curl -X POST "http://localhost:8080/api/auto?action=stop" \
  -H "X-API-Key: your-secret-api-key-here"
```

### 3. Mesaj Gönderme Süreci

1. Scheduler gönderilmemiş mesajları veritabanından alır
2. Her mesaj için webhook.site'a POST isteği gönderir
3. Webhook'tan dönen `messageId` değerini alır
4. Mesajı veritabanında `sent=true` olarak işaretler
5. `messageId` ve gönderme zamanını Redis'te cache'ler

### 4. Gönderilen Mesajları Görüntüleme

```bash
curl -X GET "http://localhost:8080/api/sent" \
  -H "X-API-Key: your-secret-api-key-here"
```

## 🔧 Yapılandırma

### Environment Variables

`docker-compose.yml` dosyasında veya `.env` dosyasında şu değişkenleri ayarlayabilirsiniz:

| Değişken | Açıklama | Varsayılan |
|----------|----------|------------|
| `PORT` | Uygulama portu | `8080` |
| `DB_PASSWORD` | MariaDB root şifresi | `root` |
| `DB_NAME` | Veritabanı adı | `insider` |
| `WEBHOOK_URL` | Webhook endpoint URL'i | - |
| `WEBHOOK_AUTH_KEY` | Webhook authentication key | `INS.example` |
| `API_KEY` | API authentication key | `your-secret-api-key-here` |
| `SCHEDULE_SECONDS` | Scheduler aralığı (saniye) | `120` (2 dakika) |
| `MSG_PER_TICK` | Her batch'te gönderilecek mesaj sayısı | `2` |
| `MSG_CHAR_LIMIT` | Mesaj karakter limiti | `160` |

### Webhook.site Yapılandırması

Webhook.site'da mutlaka şu ayarları yapın:

1. **Edit** butonuna tıklayın
2. **Status code:** `202` veya `200`
3. **Content type:** `application/json` (çok önemli!)
4. **Content:**
   ```json
   {
     "message": "Accepted",
     "messageId": "{{uuid}}"
   }
   ```
5. **Save** butonuna tıklayın

Eğer `Content type` yanlış ayarlanırsa (örneğin `text/html`), uygulama hata verecektir.

## 📚 API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```
API key gerektirmez.

### Mesaj Oluştur
```bash
curl -X POST "http://localhost:8080/api/messages" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '{
    "to": "+905551111111",
    "content": "Test mesajı"
  }'
```

### Scheduler Başlat/Durdur
```bash
# Başlat
curl -X POST "http://localhost:8080/api/auto?action=start" \
  -H "X-API-Key: your-secret-api-key-here"

# Durdur
curl -X POST "http://localhost:8080/api/auto?action=stop" \
  -H "X-API-Key: your-secret-api-key-here"
```

### Gönderilen Mesajları Listele
```bash
curl -X GET "http://localhost:8080/api/sent" \
  -H "X-API-Key: your-secret-api-key-here"
```

## 🧪 Test Etme

### Swagger UI Kullanarak

En kolay yöntem Swagger UI kullanmaktır:

1. Tarayıcıda şu adresi açın: http://localhost:8080/swagger/
2. Endpoint'leri göreceksiniz
3. "Try it out" butonuna tıklayın
4. Gerekli bilgileri doldurun
5. "Execute" butonuna tıklayın

### Manuel Test Senaryosu

1. **Birkaç mesaj oluşturun:**
   ```bash
   curl -X POST "http://localhost:8080/api/messages" \
     -H "Content-Type: application/json" \
     -H "X-API-Key: your-secret-api-key-here" \
     -d '{"to": "+905551111111", "content": "Test 1"}'
   ```

2. **Scheduler'ı başlatın:**
   ```bash
   curl -X POST "http://localhost:8080/api/auto?action=start" \
     -H "X-API-Key: your-secret-api-key-here"
   ```

3. **2 dakika bekleyin** (veya `SCHEDULE_SECONDS` değerinde)

4. **Webhook.site'da mesajları kontrol edin**

5. **Gönderilen mesajları listele:**
   ```bash
   curl -X GET "http://localhost:8080/api/sent" \
     -H "X-API-Key: your-secret-api-key-here"
   ```

## 🗄️ Veritabanı ve Redis

### MariaDB'ye Bağlanma

```bash
docker-compose exec mariadb mariadb -u root -proot insider
```

**Örnek sorgular:**
```sql
-- Tüm mesajları görüntüle
SELECT * FROM message_models;

-- Gönderilen mesajları görüntüle
SELECT * FROM message_models WHERE sent = 1;

-- Gönderilmemiş mesajları görüntüle
SELECT * FROM message_models WHERE sent = 0;
```

### Redis'e Bağlanma

```bash
docker-compose exec redis redis-cli
```

**Örnek komutlar:**
```redis
-- Tüm mesaj cache'lerini listele
KEYS message:*

-- Belirli bir mesajın cache'ini görüntüle
HGETALL message:1

-- Sadece webhook_id'yi görüntüle
HGET message:1 webhook_id
```

## 🐛 Sorun Giderme

### Uygulama başlamıyor

```bash
# Servislerin durumunu kontrol edin
docker-compose ps

# Logları kontrol edin
docker-compose logs app
```

### Mesajlar gönderilmiyor

1. Scheduler'ın başlatıldığından emin olun
2. Webhook.site URL'inin doğru olduğunu kontrol edin
3. Webhook.site'da response'un JSON formatında olduğunu kontrol edin
4. Logları kontrol edin: `docker-compose logs -f app`

### "failed to decode response" hatası

Bu, webhook.site'dan dönen response'un JSON formatında olmadığını gösterir. Webhook.site'da:
- Content type'ın `application/json` olduğundan emin olun
- Response body'nin geçerli JSON olduğundan emin olun

### Veritabanı bağlantı hatası

```bash
# MariaDB'nin çalıştığını kontrol edin
docker-compose ps mariadb

# MariaDB loglarını kontrol edin
docker-compose logs mariadb

# Servisleri yeniden başlatın
docker-compose restart
```

## 📁 Proje Yapısı

```
insider-messaging/
├── cmd/app/              # Uygulama giriş noktası
├── internal/
│   ├── application/      # İş mantığı (use cases)
│   ├── config/           # Yapılandırma
│   ├── domain/           # Domain modelleri ve interface'ler
│   ├── infrastructure/   # DB, Redis, Webhook, Scheduler
│   └── presentation/     # API handlers ve router
├── tests/                # Test dosyaları
├── docker-compose.yml    # Docker Compose yapılandırması
└── Dockerfile            # Docker image tanımı
```

## 🔒 Güvenlik

- Tüm API endpoint'leri (health ve swagger hariç) `X-API-Key` header'ı gerektirir
- Varsayılan API key: `your-secret-api-key-here` (production'da değiştirin!)
- Webhook authentication için `x-ins-auth-key` header'ı kullanılır

## 📝 Notlar

- Scheduler varsayılan olarak **otomatik başlamaz**. Manuel olarak `/api/auto?action=start` ile başlatmanız gerekir.
- Her batch'te varsayılan olarak **2 mesaj** gönderilir
- Mesajlar **FIFO** (First In First Out) sırasıyla gönderilir
- Bir mesaj bir kez gönderildikten sonra **tekrar gönderilmez**
- Redis cache opsiyoneldir ama önerilir

## 🆘 Yardım

Sorun yaşıyorsanız:

1. Logları kontrol edin: `docker-compose logs -f app`
2. Health check yapın: `curl http://localhost:8080/health`
3. Swagger UI'yi kullanın: http://localhost:8080/swagger/
4. Servislerin durumunu kontrol edin: `docker-compose ps`

## 📞 İletişim

Bu proje bir değerlendirme projesidir. Sorularınız için proje sahibiyle iletişime geçin.
