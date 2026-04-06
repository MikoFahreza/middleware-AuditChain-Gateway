# 🛡️ AuditChain Gateway

AuditChain Gateway adalah sebuah *middleware* dan API Gateway berskala *Enterprise* yang dirancang untuk menerima, memproses, dan mengunci log audit sistem (aktivitas pengguna, transaksi finansial, perubahan konfigurasi) secara *immutable* (tidak dapat diubah) ke dalam jaringan **Hyperledger Fabric Blockchain**.

Sistem ini menggunakan arsitektur *Asynchronous*, struktur data *Merkle Tree*, dan algoritma *Hashing* tingkat lanjut untuk memastikan integritas data (Anti-Tampering) dengan performa *High-Throughput*.

---

## ✨ Fitur Utama

- **🚀 High-Throughput Ingestion:** Mampu menerima ribuan log per detik melalui REST API dengan memanfaatkan **Redis Queue** untuk memisahkan proses penerimaan dan komputasi berat.
- **🌳 Merkle Tree Aggregation:** Menghemat biaya dan ruang *ledger* Blockchain dengan mengelompokkan ratusan log menjadi satu *batch* dan hanya mengirimkan *Merkle Root* ke jaringan Hyperledger Fabric.
- **🔗 Cryptographic Local Chaining:** Setiap log di dalam database lokal dikaitkan dengan log sebelumnya (`PreviousHash`) menggunakan **SHA3-256**, membentuk struktur *chain* lokal sebelum masuk ke Blockchain.
- **🕵️ 3-Layer Verification Engine:** Sistem verifikasi anti-retas yang mengecek keaslian data melalui 3 lapis: Eksistensi Database, Re-Hashing Integritas Lokal, dan Konsensus Hyperledger Fabric (The Ultimate Source of Truth).
- **📚 Interaktif API Docs:** Terintegrasi penuh dengan **Swagger UI** untuk pengujian dan dokumentasi *endpoint* yang mudah.

---

## 🏗️ Arsitektur Sistem

Sistem ini dibagi menjadi 4 lapisan utama:

1. **Ingestion Layer:** REST API yang menerima *raw log* (JSON), menormalisasikannya, dan memasukkannya ke dalam antrean Redis.
2. **Processing Layer (Background Workers):**
   - **Hasher Worker:** Mengambil log dari Redis, menghitung `SHA3-256`, dan menyimpannya ke PostgreSQL.
   - **Aggregator Worker:** Berjalan setiap 10 detik, mengambil log yang belum diagregasi, membuat *Merkle Tree*, dan menyiapkannya untuk *anchoring*.
3. **Storage Layer:** **PostgreSQL** digunakan untuk menyimpan detail log agar dapat dicari (*searchable*) secara cepat oleh Dashboard Auditor.
4. **Consensus Layer:** Menggunakan **Hyperledger Fabric Gateway SDK** untuk menyimpan *Merkle Root* ke dalam *Smart Contract* (Chaincode), menjadikannya bukti kriptografi permanen.

---

## 🛠️ Tech Stack

- **Bahasa Pemrograman:** Go (Golang)
- **Web Framework:** Gin Web Framework
- **Database:** PostgreSQL (dengan GORM)
- **Message Broker:** Redis
- **Blockchain:** Hyperledger Fabric v2.4+ (Gateway SDK)
- **Kriptografi:** SHA3-256
- **Dokumentasi API:** Swaggo / Swagger

---

## 🚀 Instalasi & Konfigurasi

### Prasyarat
Pastikan sistem Anda sudah terinstal:
- Go 1.20+
- PostgreSQL
- Redis
- Akses ke Node Hyperledger Fabric (Certificate, Private Key, MSP ID)

### Setup Environment
Buat file `.env` di *root* direktori dan sesuaikan nilainya:

```env
# Server
PORT=3000

# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=rahasia
DB_NAME=auditchain_db
DB_PORT=5432

# Redis
REDIS_HOST=localhost:6379
REDIS_PASSWORD=

# Hyperledger Fabric Gateway
FABRIC_MSP_ID=Org1MSP
FABRIC_PEER_ENDPOINT=localhost:7051
FABRIC_TLS_CERT_PATH=./crypto-config/tls/ca.crt
FABRIC_CERT_PATH=./crypto-config/users/Admin@org1/msp/signcerts/cert.pem
FABRIC_KEY_PATH=./crypto-config/users/Admin@org1/msp/keystore/priv_key.pem
FABRIC_CHANNEL=audit-channel
FABRIC_CHAINCODE=audit-contract
```

### Menjalankan Aplikasi

1. *Clone* repository ini.
2. Unduh semua *dependencies*:
   ```bash
   go mod tidy
   ```
3. Generate dokumentasi Swagger terbaru (opsional):
   ```bash
   swag init -g cmd/gateway/main.go
   ```
4. Jalankan server:
   ```bash
   go run cmd/gateway/main.go
   ```

Aplikasi akan berjalan di `http://localhost:3000`. 
Buka `http://localhost:3000/swagger/index.html` untuk melihat UI Dokumentasi.

---

## 📡 API Endpoints Utama

| Method | Endpoint | Deskripsi | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/logs` | Menerima raw log audit dan memasukkan ke antrean (Async) | Ya |
| `GET` | `/dashboard/verify/:hash` | Memverifikasi keaslian log (3-Layer Verification) | Ya |
| `POST` | `/auth/register` | Mendaftarkan akun Auditor baru | Tidak |
| `POST` | `/auth/login` | Login dan mendapatkan token JWT | Tidak |

---

## 🛡️ Threat Model & Keamanan (Mengapa ini aman?)

Sistem ini kebal terhadap berbagai jenis serangan pada level database:
- **Modifikasi Teks (Level 1):** Jika *hacker* mengubah isi kolom *Action* atau *Resource* di database, mesin Re-Hashing (Lapis 2) akan mendeteksi ketidaksesuaian *hash*.
- **Modifikasi Teks + Re-Hash (Level 2):** Jika *hacker* mengubah data dan menimpa *hash* barunya, verifikasi *Merkle Tree* akan rusak.
- **Modifikasi Full Database (Level 3):** Jika *hacker* meretas seluruh PostgreSQL beserta *Merkle Root*-nya, mesin Konsensus (Lapis 3) akan mendeteksi bahwa *Merkle Root* tersebut tidak diakui oleh *Buku Besar* Hyperledger Fabric. Hacker tertangkap basah!
