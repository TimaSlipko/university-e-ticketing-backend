# E-Ticketing System

It is a ticket-selling web application built with Go, featuring event management, ticket purchasing, and user authentication. This project is designed for educational purposes and follows clean architecture principles.

## 🚀 Features

### Core Functionality
- **User Management**: Registration, login, profile management
- **Event Management**: Create, edit, view events (sellers)
- **Ticket System**: Purchase, transfer, and manage tickets
- **Payment Processing**: Mocked payment system for development
- **Admin Panel**: Event approval, user management, statistics

### Technical Features
- **JWT Authentication**: Secure token-based authentication
- **Role-Based Access Control**: User, Seller, Admin roles
- **RESTful API**: Clean API design with proper HTTP methods
- **Database Migrations**: Automatic schema management with GORM
- **Rate Limiting**: Protection against abuse
- **CORS Support**: Frontend integration ready

## 🛠️ Tech Stack

- **Backend**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT
- **Configuration**: Environment-based with envconfig

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 6+

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
# Go to repository directory
```

2. **Setup environment**
```bash
cp .env.example .env
# Edit .env with your credentials
```

3. **Install dependencies**
```bash
go mod tidy
```

4. **Run the application**
```bash
make run
```

## 🔧 Development

### Available Commands

```bash
# Development
make run              # Start the application
make dev              # Complete development setup
make restart          # Restart the application

# Database
make migrate          # Run migrations
make seed             # Seed with sample data
make dev-db           # Setup development database

# Building
make build            # Build binary
make build-linux      # Build for Linux

# Testing & Quality
make test             # Run tests
make test-coverage    # Run tests with coverage
make lint             # Run linter
make fmt              # Format code
make check            # Run all checks

# Docker
make docker-build     # Build Docker image
make docker-run       # Run with Docker
make docker-stop      # Stop Docker containers
```

### Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=e_ticketing_dev

# JWT Configuration
JWT_SECRET=your-super-secret-key
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h

# Payment Configuration
PAYMENT_MOCK_PAYMENTS=true
```

## 📚 API Documentation

### Authentication Endpoints

```http
POST /api/v1/auth/register    # User registration
POST /api/v1/auth/login       # User login
POST /api/v1/auth/refresh     # Refresh token
POST /api/v1/auth/logout      # User logout
```

### Event Endpoints

```http
GET    /api/v1/events              # List approved events
GET    /api/v1/events/:id          # Get event details
GET    /api/v1/events/:id/tickets  # Get event tickets

# Seller only
POST   /api/v1/seller/events       # Create event
GET    /api/v1/seller/events       # Get seller's events
PUT    /api/v1/seller/events/:id   # Update event
DELETE /api/v1/seller/events/:id   # Delete event
```

### Ticket Endpoints

```http
POST /api/v1/tickets/purchase    # Purchase tickets
GET  /api/v1/tickets/my          # Get user's tickets
POST /api/v1/tickets/transfer    # Transfer ticket
```

### User Endpoints

```http
GET    /api/v1/users/profile     # Get user profile
PUT    /api/v1/users/profile     # Update profile
PUT    /api/v1/users/password    # Change password
DELETE /api/v1/users/profile     # Delete account
```

### Admin Endpoints

```http
GET  /api/v1/admin/events/pending      # Get pending events
POST /api/v1/admin/events/:id/approve  # Approve event
POST /api/v1/admin/events/:id/reject   # Reject event
GET  /api/v1/admin/stats               # Get system statistics
```

## 🔐 Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```http
Authorization: Bearer <your_jwt_token>
```

### User Roles

- **User**: Can view events, purchase tickets, manage profile
- **Seller**: Can create and manage events, view sales statistics
- **Admin**: Can approve/reject events, manage users, view system stats

## 💳 Payment System

The payment system is **mocked** for development purposes:

- All payments are simulated with 90% success rate
- No real money transactions occur
- Payment methods supported: Card, PayPal, Google Pay
- Failed payments return appropriate error messages

## 🏗️ Architecture

### Clean Architecture

The project follows clean architecture principles:

1. **Models**: Core business entities
2. **Repositories**: Data access layer with interfaces
3. **Services**: Business logic layer
4. **Handlers**: HTTP presentation layer
5. **Middleware**: Cross-cutting concerns

### Database Schema

Key relationships:
- Users can have multiple roles (User, Seller, Admin)
- Sellers can create multiple Events
- Events can have multiple Tickets through Sales
- Users can purchase Tickets (PurchasedTickets)
- Tickets can be transferred between users

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test <path>
```

### Test Coverage

The project aims for high test coverage across all layers:
- Unit tests for services and utilities
- Integration tests for repositories
- HTTP tests for handlers

## 📈 Performance Considerations

### Rate Limiting
- Default: 100 requests per minute per IP
- Configurable in middleware
- Prevents abuse and ensures fair usage

### Database Optimization
- Connection pooling configured
- Proper indexing on foreign keys
- Pagination for list endpoints

### Caching Strategy
- Redis caching
- JWT token validation caching
- Event data caching for high-traffic scenarios

## 🔒 Security Features

### Input Validation
- Email format validation
- Password strength requirements
- Username format validation
- Request payload validation

### Authentication Security
- JWT tokens with expiration
- Refresh token mechanism
- Password hashing with bcrypt
- Role-based access control

### API Security
- CORS configuration
- Rate limiting
- Request logging
- Panic recovery

## 🚧 Future Enhancements

### Phase 2 Features
- [ ] Email notifications for events
- [ ] Ticket QR code generation
- [ ] Event search and filtering
- [ ] User reviews and ratings
- [ ] Real payment integration (Stripe)
- [ ] Event categories and tags
- [ ] Bulk ticket operations

### Technical Improvements
- [ ] Redis caching implementation
- [ ] Comprehensive admin dashboard
- [ ] API documentation with Swagger
- [ ] Performance monitoring
- [ ] Advanced logging with structured logs
- [ ] Database query optimization
- [ ] WebSocket for real-time updates

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Write tests for new features
- Update documentation for API changes
- Run `make check` before committing
- Use conventional commit messages

## 📊 Project Status

- ✅ Core authentication system
- ✅ Event management (CRUD)
- ✅ Ticket purchasing system
- ✅ Mocked payment processing
- ✅ Role-based access control
- ✅ Database migrations and seeding
- ✅ Docker deployment setup
- 🚧 Admin panel (basic implementation)
- 🚧 Ticket transfer system (structure ready)
- ⏳ Email notifications
- ⏳ Real payment integration
- ⏳ Frontend application

---

**Built with ❤️ for educational purposes**