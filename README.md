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
