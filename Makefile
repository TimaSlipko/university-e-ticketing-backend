# Go + SonarQube + Gosec Integration Makefile

# variables
GOSEC_VERSION := latest
REPORTS_DIR := reports
COVERAGE_FILE := coverage.out

# colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # no color

# default target
.PHONY: all
all: clean deps test security-scan sonar-scan

# install dependencies and tools
.PHONY: deps
deps: install-gosec
	@echo "$(GREEN)✓ All dependencies installed$(NC)"

# install gosec if not present
.PHONY: install-gosec
install-gosec:
	@echo "$(YELLOW)Checking gosec installation...$(NC)"
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing gosec...$(NC)"; \
		go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION); \
		echo "$(GREEN)✓ gosec installed$(NC)"; \
	else \
		echo "$(GREEN)✓ gosec already installed$(NC)"; \
	fi

# create reports directory
.PHONY: setup-dirs
setup-dirs:
	@mkdir -p $(REPORTS_DIR)

# run tests with coverage
.PHONY: test
test: setup-dirs
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(REPORTS_DIR)/coverage.html
	@echo "$(GREEN)✓ Tests completed with coverage$(NC)"

# run gosec security scan
.PHONY: gosec-scan
gosec-scan: install-gosec setup-dirs
	@echo "$(YELLOW)Running gosec security scan...$(NC)"
	@gosec -fmt sonarqube -out $(REPORTS_DIR)/gosec-report.json ./... || true
	@gosec -fmt json -out $(REPORTS_DIR)/gosec-report-full.json ./... || true
	@gosec -fmt text ./... || true
	@echo "$(GREEN)✓ Gosec scan completed$(NC)"

# run comprehensive security analysis
.PHONY: security-scan
security-scan: gosec-scan
	@echo "$(YELLOW)Running comprehensive security analysis...$(NC)"
	@# Additional security tools can be added here
	@if command -v staticcheck >/dev/null 2>&1; then \
		echo "$(YELLOW)Running staticcheck...$(NC)"; \
		staticcheck -f json ./... > $(REPORTS_DIR)/staticcheck-report.json 2>/dev/null || true; \
	fi
	@if command -v govulncheck >/dev/null 2>&1; then \
		echo "$(YELLOW)Running govulncheck...$(NC)"; \
		govulncheck -json ./... > $(REPORTS_DIR)/govulncheck-report.json 2>/dev/null || true; \
	fi
	@echo "$(GREEN)✓ Security scan completed$(NC)"

# run SonarQube analysis (with Docker)
.PHONY: sonar-scan
sonar-scan: test security-scan
	@echo "$(YELLOW)Running SonarQube analysis with Docker...$(NC)"
	@docker run --rm \
		-e SONAR_HOST_URL="$${SONAR_HOST_URL:-http://localhost:9000}" \
		-e SONAR_TOKEN="$$SONAR_TOKEN" \
		-v "$$(pwd):/usr/src" \
		sonarsource/sonar-scanner-cli
	@echo "$(GREEN)✓ SonarQube analysis completed$(NC)"

# install additional security tools
.PHONY: install-security-tools
install-security-tools:
	@echo "$(YELLOW)Installing additional security tools...$(NC)"
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "$(GREEN)✓ Additional security tools installed$(NC)"

# quick security check (gosec only)
.PHONY: quick-security
quick-security: install-gosec
	@echo "$(YELLOW)Running quick security check...$(NC)"
	@gosec ./...

# clean reports and build artifacts
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning up...$(NC)"
	@rm -rf $(REPORTS_DIR)
	@rm -f $(COVERAGE_FILE)
	@go clean -testcache
	@echo "$(GREEN)✓ Cleanup completed$(NC)"

# show security report summary
.PHONY: security-summary
security-summary:
	@echo "$(YELLOW)Security Scan Summary:$(NC)"
	@if [ -f $(REPORTS_DIR)/gosec-report.json ]; then \
		echo "$(GREEN)✓ Gosec report: $(REPORTS_DIR)/gosec-report.json$(NC)"; \
		jq -r '.Issues | length' $(REPORTS_DIR)/gosec-report.json 2>/dev/null | xargs -I {} echo "  Issues found: {}"; \
	fi
	@if [ -f $(REPORTS_DIR)/staticcheck-report.json ]; then \
		echo "$(GREEN)✓ Staticcheck report: $(REPORTS_DIR)/staticcheck-report.json$(NC)"; \
	fi
	@if [ -f $(REPORTS_DIR)/govulncheck-report.json ]; then \
		echo "$(GREEN)✓ Govulncheck report: $(REPORTS_DIR)/govulncheck-report.json$(NC)"; \
	fi

# view coverage report
.PHONY: coverage
coverage:
	@if [ -f $(REPORTS_DIR)/coverage.html ]; then \
		echo "$(GREEN)Opening coverage report...$(NC)"; \
		open $(REPORTS_DIR)/coverage.html || xdg-open $(REPORTS_DIR)/coverage.html; \
	else \
		echo "$(RED)Coverage report not found. Run 'make test' first.$(NC)"; \
	fi

# help target
.PHONY: help
help:
	@echo "$(GREEN)Available targets:$(NC)"
	@echo "  $(YELLOW)all$(NC)                 - Run complete analysis (test + security + sonar)"
	@echo "  $(YELLOW)deps$(NC)                - Install gosec"
	@echo "  $(YELLOW)test$(NC)                - Run tests with coverage"
	@echo "  $(YELLOW)gosec-scan$(NC)          - Run gosec security scan only"
	@echo "  $(YELLOW)security-scan$(NC)       - Run comprehensive security analysis"
	@echo "  $(YELLOW)sonar-scan$(NC)          - Run SonarQube analysis"
	@echo "  $(YELLOW)sonar-scan-docker$(NC)   - Run SonarQube analysis with Docker"
	@echo "  $(YELLOW)quick-security$(NC)      - Quick gosec security check"
	@echo "  $(YELLOW)security-summary$(NC)    - Show security scan summary"
	@echo "  $(YELLOW)coverage$(NC)            - Open coverage report in browser"
	@echo "  $(YELLOW)clean$(NC)               - Clean reports and build artifacts"
	@echo "  $(YELLOW)install-security-tools$(NC) - Install additional security tools"
	@echo ""
	@echo "$(GREEN)Environment variables:$(NC)"
	@echo "  $(YELLOW)SONAR_HOST_URL$(NC)      - SonarQube server URL (default: http://localhost:9000)"
	@echo "  $(YELLOW)SONAR_TOKEN$(NC)         - SonarQube authentication token"

# development workflow targets
.PHONY: dev-setup
dev-setup: deps install-security-tools
	@echo "$(GREEN)✓ Development environment setup complete$(NC)"

.PHONY: ci
ci: clean all security-summary
	@echo "$(GREEN)✓ CI pipeline completed$(NC)"