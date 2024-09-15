# MyGoShare - File Sharing & Management System 📂
**Mohamed Sameer**  
**21BAI1109**  
**VIT Chennai**  
*Done as a part of Trademarkia's recruitment process.*
---

Welcome to **MyGoShare**, a powerful and efficient file-sharing and management system built with **GoLang**! This project is designed to provide secure file uploads, metadata storage, user authentication, file sharing via public links, caching using Redis, and optimized search for large datasets.

## 🌟 Features Overview

- **🔐 User Authentication with JWT**
- **📤 Secure File Uploads**
- **📝 File Metadata Storage in PostgreSQL**
- **🔗 Public File Sharing**
- **⚡ In-Memory Caching with Redis for Fast Metadata Access**
- **🔍 File Search by Name, Date, and Type**
- **⏱️ Hosted on AWS EC2 and uses a S3 bucket**

---

## 📑 API Endpoints
You can try out the application hosted on the following public IP address:

- **Public IPv4 Address**: [http://16.171.136.21:8080/]

Here’s a list of all the available API endpoints for **MyGoShare**:

### 🔐 User Authentication
1. **Register**:  
   - **URL**: `POST /register`  
   - **Description**: Register a new user with email and password.

2. **Login**:  
   - **URL**: `POST /login`  
   - **Description**: Authenticate the user and return a JWT token.

### 📤 File Upload & Management
3. **Upload File**:  
   - **URL**: `POST /upload`  
   - **Description**: Upload a file. Requires JWT authentication.
   - **Authorization**: Bearer Token  
   - **Body**: Multipart form with a `file` field.

4. **Get Files**:  
   - **URL**: `GET /files`  
   - **Description**: Retrieve the list of files for the authenticated user.
   - **Authorization**: Bearer Token  

5. **Search Files**:  
   - **URL**: `GET /search/files?name=&date=&file_type=`  
   - **Description**: Search the user’s files by name, upload date, and file type.
   - **Authorization**: Bearer Token  

