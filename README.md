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
- **Database**: MySQL with GORM
- **Authentication**: JWT
- **Configuration**: Environment-based with envconfig

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- MySQL
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
go run main.go
```

### Environment Variables

Create a `.env` file based on `.env.example`

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
# Public endpoints
GET    /api/v1/events                           # List approved events
GET    /api/v1/events/:event_id                 # Get event details
GET    /api/v1/events/:event_id/tickets         # Get event tickets (legacy)
GET    /api/v1/events/:event_id/grouped-tickets # Get grouped tickets
GET    /api/v1/events/:event_id/sales           # Get event sales

# Seller only
POST   /api/v1/seller/events                    # Create event
GET    /api/v1/seller/events                    # Get seller's events
PUT    /api/v1/seller/events/:event_id          # Update event
DELETE /api/v1/seller/events/:event_id          # Delete event
GET    /api/v1/seller/events/:event_id/grouped-tickets # Get seller's grouped tickets
```

### Ticket Endpoints

```http
# Customer endpoints
POST /api/v1/tickets/purchase             # Purchase individual ticket (legacy)
POST /api/v1/tickets/purchase-group       # Purchase tickets from group
GET  /api/v1/tickets/my                   # Get user's purchased tickets
POST /api/v1/tickets/transfer             # Initiate ticket transfer
GET  /api/v1/tickets/:ticket_id/download  # Download ticket PDF
GET  /api/v1/tickets/:ticket_id/view      # View ticket PDF

# Seller only
POST   /api/v1/seller/tickets                    # Create tickets
PUT    /api/v1/seller/events/:event_id/tickets   # Update tickets
DELETE /api/v1/seller/events/:event_id/tickets   # Delete tickets
```

### Transfer Endpoints

```http
GET  /api/v1/transfers/active              # Get active transfers
POST /api/v1/transfers/:transfer_id/accept # Accept transfer
POST /api/v1/transfers/:transfer_id/reject # Reject transfer
GET  /api/v1/transfers/history             # Get transfer history
```

### Payment Endpoints

```http
# User payments
GET /api/v1/payments/my        # Get user's payment history
GET /api/v1/payments/:id       # Get payment status

# Seller payments
GET /api/v1/seller/payments    # Get seller's revenue history
```

### Payment Methods Endpoints

```http
POST   /api/v1/payment-methods              # Create payment method
GET    /api/v1/payment-methods              # Get user's payment methods
GET    /api/v1/payment-methods/:id          # Get specific payment method
PUT    /api/v1/payment-methods/:id          # Update payment method
DELETE /api/v1/payment-methods/:id          # Delete payment method
POST   /api/v1/payment-methods/:id/set-default # Set default payment method
```

### Sales Endpoints

```http
# Public
GET /api/v1/sales/:sale_id     # Get sale details

# Seller only
POST   /api/v1/seller/sales           # Create sale
PUT    /api/v1/seller/sales/:sale_id  # Update sale
DELETE /api/v1/seller/sales/:sale_id  # Delete sale
```

### User Endpoints

```http
GET    /api/v1/users/profile     # Get user profile
PUT    /api/v1/users/profile     # Update profile
PUT    /api/v1/users/password    # Change password
DELETE /api/v1/users/profile     # Delete account
```

### Seller Endpoints

```http
GET    /api/v1/seller/profile    # Get seller profile
PUT    /api/v1/seller/profile    # Update seller profile
PUT    /api/v1/seller/password   # Change seller password
DELETE /api/v1/seller/profile    # Delete seller account
GET    /api/v1/seller/stats      # Get seller statistics
```

### Admin Endpoints

```http
GET  /api/v1/admin/events/pending        # Get pending events
POST /api/v1/admin/events/:event_id/approve  # Approve event
POST /api/v1/admin/events/:event_id/reject   # Reject event
GET  /api/v1/admin/stats                 # Get system statistics (not implemented)
```

### Health Check

```http
GET /health    # System health check
```

## 🔐 Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```http
Authorization: Bearer <your_jwt_token>
```

## 👥 User Roles

- **User (1)**: Can purchase tickets, transfer tickets, view payment history
- **Seller (2)**: Can create events, manage tickets and sales, view revenue
- **Admin (3)**: Can approve/reject events, view system statistics

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
- [ ] Event search and filtering
- [ ] User reviews and ratings
- [ ] Real payment integration
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

# Security Scanning & Code Quality

This project includes comprehensive security scanning using **Gosec** for Go-specific vulnerabilities and **SonarQube** for code quality analysis. All security findings are integrated into a single SonarQube dashboard for easy monitoring.

## 🚀 Quick Start

```bash
# Run complete analysis (tests + security + quality)
make all

# Quick security check during development
make quick-security

# View security summary
make security-summary
```

## 🛠️ Setup

### Prerequisites

- **Go 1.19+**
- **SonarQube server** (local or remote)
- **Docker** (optional, for SonarQube scanner)

### Environment Variables

```bash
# Required for SonarQube integration
export SONAR_HOST_URL="http://localhost:9000"
export SONAR_TOKEN="your-sonarqube-token"

# Optional: Custom project settings
export SONAR_PROJECT_KEY="my-go-project"
export SONAR_PROJECT_NAME="My Go Project"
```

### Initial Setup

```bash
# Install all dependencies and security tools
make dev-setup
```

## 📊 Available Commands

### Development Workflow

| Command | Description |
|---------|-------------|
| `make all` | Complete analysis pipeline |
| `make dev-setup` | Initial development environment setup |
| `make quick-security` | Fast security check (gosec only) |
| `make test` | Run tests with coverage |
| `make clean` | Clean reports and artifacts |

### Security Scanning

| Command | Description |
|---------|-------------|
| `make gosec-scan` | Run gosec security analysis |
| `make security-scan` | Comprehensive security scan (gosec + staticcheck + govulncheck) |
| `make security-summary` | Display security findings summary |

### Code Quality

| Command | Description                        |
|---------|------------------------------------|
| `make sonar-scan` | Run SonarQube analysis with Docker |
| `make coverage` | Open coverage report in browser    |

## 🔍 What Gets Scanned

### Security Vulnerabilities (Gosec)

- **Hardcoded credentials** (passwords, API keys, tokens)
- **SQL injection** vulnerabilities
- **Command injection** risks
- **Weak cryptography** usage
- **Unsafe file operations**
- **Memory safety** issues
- **30+ additional security rules**

### Code Quality (SonarQube)

- **Code smells** and maintainability issues
- **Bugs** and reliability problems
- **Test coverage** analysis
- **Complexity** metrics
- **Duplicated code** detection

### Dependency Vulnerabilities

- **Known CVEs** in dependencies (govulncheck)
- **Outdated packages** with security fixes
- **License compliance** issues

## 📈 Viewing Results

### SonarQube Dashboard

After running `make all`, visit your SonarQube server to see:

1. **Security** tab - All gosec findings and vulnerabilities
2. **Issues** tab - Code quality and maintainability issues
3. **Coverage** tab - Test coverage metrics and line-by-line analysis
4. **Activity** tab - Historical trends and quality gate status

### Local Reports

Generated reports are available in the `reports/` directory:

```
reports/
├── gosec-report.json          # SonarQube integration format
├── gosec-detailed.json        # Detailed security findings
├── gosec-text.txt            # Human-readable security report
├── coverage.html             # Interactive coverage report
├── staticcheck-report.json   # Static analysis results
├── govulncheck-report.json   # Dependency vulnerabilities
└── security-summary.md       # Executive summary
```

## 🔧 Configuration

### Project Configuration

Edit `sonar-project.properties` to customize:

```properties
# Basic settings
sonar.projectKey=my-go-project
sonar.projectName=My Go Project

# Exclude files from analysis
sonar.exclusions=vendor/**,**/*_test.go

# Security scan integration
sonar.go.gosec.reportPaths=reports/gosec-report.json
```

### Custom Security Rules

To ignore specific security issues:

```properties
# Ignore security rules in test files
sonar.issue.ignore.multicriteria=tests
sonar.issue.ignore.multicriteria.tests.ruleKey=*
sonar.issue.ignore.multicriteria.tests.resourceKey=**/*_test.go
```

## 🎯 Quality Gates

The project enforces these quality standards:

- **Security Rating**: A (no vulnerabilities)
- **Reliability Rating**: A (no bugs)
- **Test Coverage**: >80%
- **Code Smells**: <10 per 1k lines
- **Duplicated Lines**: <3%

## 🔄 Development Workflow

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
make quick-security
if [ $? -ne 0 ]; then
    echo "❌ Security issues found. Fix before committing."
    exit 1
fi
```

### IDE Integration

Most IDEs support SonarQube integration:

- **VS Code**: SonarLint extension
- **GoLand**: SonarLint plugin
- **Vim/Neovim**: Use Language Server Protocol (LSP)

## 🆘 Troubleshooting

### Common Issues

**SonarQube connection issues:**
```bash
# Check connectivity
curl -u $SONAR_TOKEN: $SONAR_HOST_URL/api/system/status

# Use Docker scanner
make sonar-scan-docker
```

**Permission denied errors:**
```bash
# Fix Go binary path
export PATH=$PATH:$(go env GOPATH)/bin
```

### Getting Help

```bash
# Show all available commands
make help

# View detailed logs
make all VERBOSE=1
```

## 📚 Additional Resources

- [Gosec Rules Documentation](https://github.com/securego/gosec)
- [SonarQube Go Plugin](https://docs.sonarqube.org/latest/analysis/languages/go/)
- [Go Security Best Practices](https://golang.org/doc/security)

---

> 💡 **Pro Tip**: Run `make quick-security` frequently during development to catch security issues early, and `make all` before creating pull requests to ensure code quality standards are met.

---

**Built with ❤️ for educational purposes**