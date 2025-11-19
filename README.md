# Daily Greens Backend API

A RESTful API backend service for Daily Greens application built with Go and Gin framework. This service provides comprehensive endpoints for managing products, users, authentication, and file uploads with integrated Swagger documentation.

## ERD Daily Greens

```mermaid
erDiagram
    users {
        serial id PK
        varchar(255) email UK
        varchar(20) role
        text password
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    profiles {
        serial id PK
        int user_id FK
        varchar(255) full_name
        text profile_photo
        varchar(255) address
        varchar(20) phone_number
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    categories {
        serial id PK
        varchar(100) name UK
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    sizes {
        serial id PK
        varchar(10) name UK
        numeric size_cost
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    variants {
        serial id PK
        varchar(10) name UK
        numeric variant_cost
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    order_methods {
        serial id PK
        varchar(10) name UK
        numeric delivery_fee
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    payment_methods {
        serial id PK
        varchar(10) name UK
        numeric admin_fee
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    status {
        serial id PK
        varchar(10) name UK
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    products {
        serial id PK
        varchar(255) name UK
        text description
        numeric price
        numeric discount_percent
        numeric rating
        bool is_flash_sale
        int stock
        bool is_active
        bool is_favourite
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    product_images {
        serial id PK
        int product_id FK
        text product_image
        bool is_primary
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    products_sizes {
        serial id PK
        int product_id FK
        int size_id FK
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    product_categories {
        serial id PK
        int product_id FK
        int category_id FK
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    product_variants {
        serial id PK
        int product_id FK
        int variant_id FK
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    sessions {
        serial id PK
        int user_id FK
        timestamp login_time
        timestamp logout_time
        timestamp expired_at
        varchar(30) ip_address
        varchar(255) device
        bool is_active
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    password_resets {
        serial id PK
        int user_id FK
        char(6) token_reset UK
        timestamp expired_at
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    testimonies {
        serial id PK
        int user_id FK
        varchar(100) position
        numeric rating
        text testimonial
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    carts {
        serial id PK
        int user_id FK
        int product_id FK
        int size_id FK
        int variant_id FK
        int amount
        numeric subtotal
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    transactions {
        serial id PK
        int user_id FK
        varchar(255) no_invoice
        timestamp date_transaction
        varchar(255) full_name
        varchar(255) email
        varchar(255) address
        varchar(20) phone
        int payment_method_id FK
        int order_method_id FK
        int status_id FK
        numeric delivery_fee
        numeric admin_fee
        numeric tax
        numeric total_transaction
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    transaction_items {
        serial id PK
        int transaction_id FK
        int product_id FK
        varchar(255) product_name
        numeric product_price
        numeric discount_percent
        numeric discount_price
        varchar(10) size
        numeric size_cost
        varchar(50) variant
        numeric variant_cost
        int amount
        numeric subtotal
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    coupons {
        serial id PK
        varchar(100) title UK
        text description
        numeric discount_percent
        numeric min_purchase
        text coupon_image
        varchar(20) bg_color
        timestamp valid_until
        bool is_active
        timestamp created_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    coupon_usage {
        serial id PK
        int user_id FK
        int coupon_id FK
        int transaction_id FK
        numeric discount_amount
        timestamp used_at
        timestamp updated_at
        int created_by FK
        int updated_by FK
    }

    users ||--o| profiles : has
    users ||--o{ sessions : has
    users ||--o{ password_resets : requests
    users ||--o{ testimonies : writes
    users ||--o{ carts : creates
    users ||--o{ transactions : places
    users ||--o{ coupon_usage : uses

    users ||--o{ users : manages

    products ||--o{ product_images : has
    products ||--o{ products_sizes : has
    products ||--o{ product_categories : belongs_to
    products ||--o{ product_variants : has
    products ||--o{ carts : added_to
    products ||--o{ transaction_items : ordered_in

    categories ||--o{ product_categories : includes

    sizes ||--o{ products_sizes : used_in
    sizes ||--o{ carts : selected_in

    variants ||--o{ product_variants : used_in
    variants ||--o{ carts : selected_in

    payment_methods ||--o{ transactions : used_in
    order_methods ||--o{ transactions : used_in
    status ||--o{ transactions : assigned_to

    transactions ||--o{ transaction_items : contains
    transactions ||--o{ coupon_usage : applied_to

    coupons ||--o{ coupon_usage : applied_in

    users ||--o{ categories : manages
    users ||--o{ sizes : manages
    users ||--o{ variants : manages
    users ||--o{ order_methods : manages
    users ||--o{ payment_methods : manages
    users ||--o{ status : manages
    users ||--o{ products : manages
    users ||--o{ product_images : manages
    users ||--o{ products_sizes : manages
    users ||--o{ product_categories : manages
    users ||--o{ product_variants : manages
    users ||--o{ profiles : manages
    users ||--o{ coupons : manages
    users ||--o{ transaction_items : manages
```

## Tech Stack

- **Go** 1.25.3 - Programming language
- **Gin** v1.11.0 - HTTP web framework
- **PostgreSQL** (via pgx/v5) - Primary database
- **Redis** v9.16.0 - Caching and session management
- **JWT** v5.3.0 - Authentication tokens
- **Cloudinary** v2.14.0 - Cloud-based image management
- **Swagger** v1.16.6 - API documentation
- **Argon2** v1.4.1 - Password hashing

## Key Features

- **RESTful API** - Clean and intuitive API design
- **JWT Authentication** - Secure token-based authentication
- **Password Security** - Argon2 password hashing
- **File Upload** - Image upload via Cloudinary
- **Redis Caching** - Fast data retrieval and session management
- **API Documentation** - Auto-generated Swagger/OpenAPI docs
- **Environment Config** - dotenv configuration management
- **Database Connection Pool** - Efficient PostgreSQL connection handling

## Prerequisites

Before running this application, make sure you have:

- **Go 1.25.3 or higher** installed on your system
- **PostgreSQL** database server
- **Redis** server
- **Cloudinary** account (for file uploads)
- **Git** (for cloning the repository)

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
ENVIRONMENT=development

# origin url for cors
ORIGIN_URL=http://localhost:3000

# app secret jwt
APP_SECRET=<your_app_secret_key>

# connection string postgres
DATABASE_URL=postgresql://user:password@localhost:5432/your_database

# connection string redis
REDIS_URL=redis://default:<PASSWORD>@<HOST>:<PORT>

# cloudinary configuration
CLOUDINARY_NAME=<CLOUDINARY_NAME>
CLOUDINARY_API_SECRET=<CLOUDINARY_API_SECRET>
CLOUDINARY_API_KEY=<CLOUDINARY_API_KEY>

# environment docker compose
DATABASE_URL_DOCKER=postgresql://user:password@localhost:5432/your_database
REDIS_URL_DOCKER=redis://host:port/db
POSTGRES_PASSWORD_DOCKER=<your_posgtres_password>
POSTGRES_DB_DOCKER=<db_name>
POSTGRES_USER_DOCKER=<postgres_user>

# smtp configuration
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=port_smtp
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_email_password
```

## Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/backend-daily-greens.git
cd backend-daily-greens
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Database Setup

Create PostgreSQL database and run migrations:

```bash
# Create database
psql -U postgres -c "CREATE DATABASE daily_greens;"

# Run migrations (if you have migration files)
# psql -U postgres -d daily_greens -f migrations/schema.sql
```

### 4. Generate Swagger Documentation

```bash
swag init
```

This will generate Swagger docs in the `docs/` directory.

### 5. Run the Application

**Development mode:**

```bash
go run main.go
```

**Production mode:**

```bash
go build -o daily-greens-api
./daily-greens-api
```

The server will start on `http://localhost:8080` (or your configured PORT).

## API Documentation

Once the application is running, access the Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

## Key Dependencies

| Package                  | Version | Purpose               |
| ------------------------ | ------- | --------------------- |
| gin-gonic/gin            | v1.11.0 | Web framework         |
| jackc/pgx/v5             | v5.7.6  | PostgreSQL driver     |
| redis/go-redis/v9        | v9.16.0 | Redis client          |
| golang-jwt/jwt/v5        | v5.3.0  | JWT authentication    |
| cloudinary-go/v2         | v2.14.0 | Image upload service  |
| swaggo/swag              | v1.16.6 | Swagger documentation |
| matthewhartstonge/argon2 | v1.4.1  | Password hashing      |
| joho/godotenv            | v1.5.1  | Environment config    |

## How to Contribute

### 1. Fork the Repository

Click the **Fork** button at the top right of this page.

### 2. Clone Your Fork

```bash
git clone https://github.com/yourusername/backend-daily-greens.git
cd backend-daily-greens
```

### 3. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 4. Make Your Changes

- Follow Go best practices and conventions
- Write unit tests for new features
- Update Swagger documentation if adding new endpoints
- Keep code clean and well-documented

### 5. Run Tests

```bash
go test ./...
go vet ./...
```

### 6. Commit Your Changes

```bash
git add .
git commit -m "Add: description of your changes"
```

**Commit Message Convention:**

- `Add:` for new features
- `Fix:` for bug fixes
- `Update:` for improvements
- `Refactor:` for code refactoring
- `Docs:` for documentation changes
- `Test:` for adding tests

### 7. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 8. Create a Pull Request

1. Go to the original repository
2. Click **New Pull Request**
3. Select your feature branch
4. Describe your changes clearly
5. Submit the PR

## Contact

For questions or support, please contact:

- Email: your.email@example.com
- GitHub: [@yourusername](https://github.com/yourusername)

# Tabel Perbandingan Penggunaan Cache

| Kondisi                          | Foto                                                         |
| -------------------------------- | ------------------------------------------------------------ |
| Waktu Request Sebelum Cache (ms) | ![Waktu request sebelum cache](/assets/img/before_cache.png) |
| Waktu Request Setelah Cache (ms) | ![Waktu request sebelum cache](/assets/img/after_cache.png)  |
