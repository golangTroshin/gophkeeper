# GophKeeper 🛡️

GophKeeper is a **secure, encrypted data storage solution** built with **Go** and **gRPC**.  
It allows users to safely store and retrieve **sensitive data**, such as **login credentials, text, binary files, and payment card details**, using **strong encryption** with a **master seed**.

## 📖 Features
✅ **User Authentication** – Secure login system using **hashed passwords** and **JWT tokens**.  
✅ **Data Encryption** – All stored data is encrypted with **AES-GCM**, using a **master seed** per user.  
✅ **Multi-Format Support** – Supports **credentials, text, binary data, and card details**.  
✅ **gRPC API** – Efficient **Remote Procedure Call (RPC)** communication.  
✅ **TUI Interface** – Built-in **Terminal User Interface (TUI)** using `tview`.  

---

## 🛠️ Installation & Setup

### **Prerequisites**
- **Go 1.18+** installed
- **PostgreSQL** database
- **Protobuf Compiler** (`protoc`) installed for gRPC

### **Clone the Repository**
```sh
git clone https://github.com/golangTroshin/gophkeeper.git
cd gophkeeper
```
---

## 📜 Usage

### Authenticate & Manage Data
When you start the client, you will see options to:
- **Login** with an existing account.
- **Sign Up** to create a new account.
- **Set Up Master Seed** during sign-up.
- **Store & Retrieve Data** via gRPC.

### Supported Data Types
- **Credentials** – Store usernames & passwords securely.
- **Text** – Securely save notes and secrets.
- **Binary Data** – Encrypt and store files.
- **Card Details** – Save payment card information.

### Version & Build Date Display
You can check the build version directly from the TUI:
```sh
GophKeeper CLI
Version: 1.0.0
Build Date: 2024-02-05
```

---


## 📦 Deployment
### Build Cross-Platform Binaries TUI Cient
```sh
make build
```

This will generate packaged binaries for Windows, Linux, and macOS in the `build/` directory.

---

## 🛠 Environment Variables
GophKeeper uses environment variables for database and server configuration. Set them in a `.env` file:
```sh
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=gophkeeper
DB_SSLMODE=disable
```
