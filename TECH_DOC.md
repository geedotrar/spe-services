## 1. Ringkasan Project

Project ini adalah POC optimasi API integrasi pembayaran berbasis microservices untuk memenuhi requirement pada dokumen soal.
Scope POC ini bukan membangun core switching/payment engine end-to-end, tetapi fokus pada API layer yang diminta di dokumen:
- Autentikasi aman dan efisien.
- Proses notifikasi transaksi menggunakan mekanisme asynchronous.
- Pengecekan status transaksi tanpa membebani server.
- Menyiapkan praktik pengurangan bug ratio 
- Efisiensi deployment.

---

## 2. Tujuan Solusi

Menghadirkan sistem pembayaran elektronik dengan fokus pada:

✅ **Autentikasi Aman & Efisien** - Proses token generation yang terlindungi dari brute-force attack  
✅ **Notifikasi Transaksi Async** - Eliminasi timeout dengan memisahkan proses logging dari response handling  
✅ **Status Checking Optimal** - Pengecakan status transaksi tanpa beban berlebih di database  
✅ **Bug Ratio Lebih Rendah** - Struktur code yang jelas, didukung unit test, supaya error lebih mudah dicegah dan dilacak  
✅ **Deployment Otomatis** - Alur build dan deployment disiapkan reproducible lewat Docker Compose dan GitHub Actions  

---

## 3. High-Level Architecture

### 3.1 Komponen Sistem

<img width="3664" height="2812" alt="High-Level Architecture" src="https://github.com/user-attachments/assets/d4f2955d-e15f-48d2-85d7-7aad5f238fbd" />


### 3.2 Layering Architecture
<img width="1584" height="2150" alt="Layering Architecture" src="https://github.com/user-attachments/assets/81163991-3fd2-436c-9548-50d387c9617b" />

---

## 4. Flow Ringkas

### 4.1 Authentication Flow
1. Merchant mengirim `merchant_id` dan `merchant_secret` ke Auth Service
2. Auth Service memvalidasi credentials di database
3. Menghasilkan JWT token dengan TTL 15 menit
4. Token dikembalikan ke merchant

### 4.2 Transaction Notification Flow
1. Payment provider mengirim transaction notification ke Payment Service
2. Service memvalidasi signature HMAC SHA512
3. Data ditaruh ke Kafka queue untuk async processing
4. Response 200 OK langsung dikirim (tidak menunggu logging selesai)
5. Consumer (background worker) memproses notifikasi secara async

### 4.3 Check Transaction Status Flow
1. Merchant query status transaksi dengan `bill_number`
2. Payment Service mengecek cache Redis terlebih dahulu
3. Jika miss, query database dengan optimized indexes
4. Response dikembalikan dengan status transaksi

---

## 5. Flow Detail

### 5.1 Flow Autentikasi

<img width="6043" height="1017" alt="flow-autentikasi" src="https://github.com/user-attachments/assets/ab648784-6117-40db-aaa8-77aff3142320" />

**Key Points:**
- ✅ Rate limiting mencegah brute-force attack
- ✅ JWT token, hanya divalidasi di payment service
---

### 5.2 Flow Transaction Notification (dengan Async Business Processing)
<img width="6613" height="1017" alt="Flow Transaction Notification async" src="https://github.com/user-attachments/assets/232885c4-0761-4a5f-a1f4-1dc6f8ee990a" />

**Perilaku saat Kafka down:**
- Endpoint notifikasi tetap bisa mengembalikan `200 OK` selama validasi dan simpan transaksi utama ke database berhasil
- Event async untuk audit trail bisa tidak ter-publish (degraded mode) jika producer tidak tersambung atau queue penuh
- Service tetap berjalan, tetapi proses audit trail async menjadi best-effort sampai Kafka pulih
- Saat Kafka normal kembali, event baru diproses normal oleh consumer

**Desain ini memastikan:**
- ✅ Response time notification cepat (tidak tergantung logging)
- ✅ Scalability - proses background handling dapat di-scale independently
- ✅ Resilience - notification tidak terpengaruh jika external service down
- ✅ Consistency - database tetap ter-update dengan transaction state

---

### 5.3 Flow Check Status Transaksi

<img width="2123" height="4200" alt="Flow Check Status Transaksi" src="https://github.com/user-attachments/assets/2233202d-7b7e-4320-a808-897dce967c9e" />

**Optimasi Database:**
- Index pada `(bill_number, created_date)` untuk query cepat
- Cache TTL 5 menit untuk hot queries

---

### 5.4 Tujuan Tiap Route

| Route | Method | Tujuan | Input | Output |
|-------|--------|--------|-------|--------|
| `/api/v1/auth/token` | POST | Mendapatkan access token | merchant_id, merchant_secret | access_token, expires_in |
| `/api/v1/transaction-notification` | POST | Menerima notifikasi transaksi dari provider | request_id, amount, bill_number | {code, message} |
| `/api/v1/check-status` | POST | Mengecek status transaksi | request_id, bill_number | {code, message, status} |
| `/dev/signature` | POST | Generate signature untuk testing | payload | base64_signature |

---

## 6. Tech Stack dan Alasan Pemilihan

| Component | Technology | Alasan |
|-----------|-----------|--------|
| **Language** | Go 1.25+ | Fast, compiled, excellent goroutine support untuk concurrency |
| **Web Framework** | Gin | Lightweight, fast routing, built-in middleware support, popular di Go ecosystem |
| **Database** | PostgreSQL 16 | powerful indexing, reliable dan production-ready |
| **Cache** | Redis | Fast in-memory cache |
| **Message Queue** | Apache Kafka | Durable message persistence, consumer groups, fault tolerance |
| **JWT Auth** | Standard JWT | Stateless, tidak perlu session storage, mudah untuk distributed systems |
| **Containerization** | Docker & Docker Compose | Consistency across environments, easy deployment, reproducible setup |
| **Testing** | Go Testing + Postman | Native Go testing library, Postman untuk integration testing |

---

## 7. Pemetaan ke Problem Statement

### **Problem 1: Brute-force Autentikasi** 

**Masalah:**
- Tidak ada batasan login attempts
- Attacker bisa brute-force MerchantID/Secret
- Query ke database meningkat drastis

**Solusi :**
```go
// Rate Limiting Middleware (5 attempt)
├─ Store attempt counter di Redis
├─ Return 429 Too Many Requests jika exceed limit
└─ TTL auto-reset
```

**Impact:**
- ✅ Brute-force attack terhenti otomatis
- ✅ Query reduction (dengan rate limiting)
- ✅ CPU usage stabil

---

### **Problem 2: Notifikasi Timeout (Synchronous Logging)** 

**Masalah:**
- Proses logging berjalan synchronous dalam response handler
- External service timeout = API timeout → Failed notification
- Response time = processing time + logging time

**Solusi :**
<img width="6006" height="305" alt="Notifikasi Timeout (Synchronous Logging" src="https://github.com/user-attachments/assets/eaca967f-798e-4abb-b561-70018b30213f" />


**Teknologi:**
- Kafka untuk message durability
- Consumer group untuk parallel processing
- Exponential backoff untuk retry

**Impact:**
- ✅ Response time stabil
- ✅ No timeout pada logging failure

---

### **Problem 3: Check Status - DB Overload** 

**Masalah:**
- Setiap check-status query ke database
- Banyak merchant checking frequently
- CPU usage tinggi

**Solusi :**
<img width="2945" height="2982" alt="Check Status - DB Overload" src="https://github.com/user-attachments/assets/c94da8b8-fae5-49a1-909d-054f89d0a34d" />


**Implementasi:**
```go
// Optimized Query dengan Index
SELECT id, status, amount 
FROM transactions 
WHERE bill_number = $1 
  AND DATE(created_at) = TODAY
INDEX: (bill_number, created_date)
```

**Impact:**
- ✅ DB CPU usage turun

---

### **Problem 4: Bug Ratio Tinggi di Development** 

**Masalah:**
- Architecture tidak jelas
- Code organization chaotic
- Sulit testing & malakukan debug

**Solusi :**

- Unit test pada auth service dan helper JWT/signature
- GitHub Actions workflow untuk menjalankan test dan build otomatis

**Impact:**
- ✅ Clear responsibility per layer
- ✅ Existing unit test membantu menangkap regression lebih awal
- ✅ Easy to debug karena trace flow jelas
- ✅ Code reusability (layering)

---

### **Problem 5: Deployment Manual** 

**Masalah:**
- Step-by-step manual deployment
- Easy untuk salah
- Tidak reproducible

**Solusi :**
```bash
# Docker + Docker Compose automation
Steps:
1. docker compose up -d postgres redis kafka
2. run migrations:
  docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/migrations/auth/001_init_auth.sql
  docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/migrations/payment/001_init_payment.sql
  docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/migrations/payment/002_optimize_indexes_payment.sql
  docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/migrations/payment/003_transaction_event_logs.sql
3. docker compose seed data
4. docker compose up auth-service payment-service
5. curl health check

# GitHub Actions workflow untuk CI/CD:
# - run unit test
# - build binary
# - publish artifact / image

# Single command untuk local stack:
docker compose up -d --build
```

**Keuntungan:**
- ✅ Reproducible across machines
- ✅ Single command untuk full stack
- ✅ CI pipeline otomatis menjalankan test dan build
- ✅ Easy rollback (just change image tag)

---

## 9. Environment Configuration

### **auth-services/.env**
```env
AUTH_PORT=9090
DATABASE_URL=postgres://postgres:postgres@localhost:5432/spe?sslmode=disable
REDIS_ADDR=localhost:6379
JWT_SECRET=super-secret-jwt-key
JWT_TTL_MINUTES=15
RATE_LIMIT_MAX_ATTEMPTS=5
RATE_LIMIT_WINDOW_MINUTES=10
AUTH_LOGIN_ENABLED=true
```

### **payment-services/.env**
```env
PAYMENT_PORT=9091
DATABASE_URL=postgres://postgres:postgres@localhost:5432/spe?sslmode=disable
REDIS_ADDR=localhost:6379
KAFKA_BROKERS=kafka:9092
JWT_SECRET=super-secret-jwt-key
MERCHANT_API_KEY=speskilltest
```

---

## 10. Security Best Practices

### **Implemented:**
- ✅ JWT Token Authentication (stateless)
- ✅ HMAC SHA512 Signature Validation
- ✅ Rate Limiting (Redis-based)
- ✅ Secure secret storage (hashed passwords)

### **Recommended Improvement:**
- Implement API Gateway dengan rate limiting global
- Audit logging untuk semua operations
- CORS configuration untuk web clients
- Tambahkan NoSQL (mis. MongoDB) untuk event log/history ber-volume tinggi agar query baca lebih cepat dan tidak membebani database transaksi utama

---

## 11. Kesimpulan

**SPE** adalah sistem pembayaran elektronik yang dirancang dengan **modern architecture** untuk menyelesaikan 5 masalah utama:

| # | Masalah | Solusi | Status |
|---|---------|--------|--------|
| 1 | Brute-force autentikasi | Rate limiting dengan Redis | ✅ |
| 2 | Notifikasi timeout | Async processing dengan Kafka | ✅ |
| 3 | DB overload pada check status | Caching + optimized indexes | ✅ |
| 4 | Bug ratio tinggi | Clean architecture + testing | ✅ |
| 5 | Deployment manual | Docker + automation scripts | ✅ |
---
