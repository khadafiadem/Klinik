Ya. Ini saya buatkan versi **final** yang memang ditujukan untuk **langsung ditempel sebagai `AGENTS.md` di root project OpenCode**. Saya buat cukup detail supaya agent punya aturan kerja yang jelas, tetapi tidak terlalu panjang.

# AGENTS.md

# KLINIK MANAGEMENT SYSTEM

You are the AI coding agent responsible for building and maintaining this clinic management application.

The application is intended for clinics in Indonesia and must be secure, reliable, simple to use, and easy to maintain.

---

## 1. MAIN OBJECTIVE

Build a complete web-based Clinic Management System from scratch.

The system must support:

* Patient management
* Patient registration
* Queue management
* Doctor management
* Medical records
* Medical examinations
* Diagnoses
* Treatments
* Prescriptions
* Pharmacy / medicine inventory
* Cashier / payments
* Reports
* Users and roles
* Clinic settings
* Audit logs

Build the system incrementally.

Do not build everything at once.

Always make sure the current feature works before continuing to the next feature.

---

# 2. TECHNOLOGY STACK

Use the following stack unless there is a strong technical reason to change it.

### Backend

* Go (Golang)
* REST API
* Clean and modular architecture

### Database

* PostgreSQL
* Supabase can be used as the PostgreSQL database and supporting platform.

### Frontend

* HTML
* CSS
* JavaScript
* Use a lightweight frontend approach.
* Avoid unnecessary frontend frameworks unless required.

### Authentication

* JWT
* Secure password hashing
* Role Based Access Control (RBAC)

### Development

* Git
* Environment variables
* Database migrations

---

# 3. LANGUAGE

The application interface must use Indonesian.

Examples:

* Login
* Dashboard
* Pasien
* Dokter
* Perawat
* Rekam Medis
* Pemeriksaan
* Resep
* Obat
* Apotek
* Kasir
* Pembayaran
* Laporan
* Pengaturan

Code, database table names, API names, and variable names should use English consistently.

---

# 4. DEVELOPMENT PRINCIPLES

Follow these principles:

1. Security first.
2. Data integrity second.
3. Correctness third.
4. Maintainability fourth.
5. UI/UX simplicity fifth.

Never sacrifice security or medical data integrity for speed.

Do not create unnecessary complexity.

Prefer simple, readable, maintainable solutions.

---

# 5. PROJECT STRUCTURE

Use a structure similar to:

```text
clinic-app/
│
├── AGENTS.md
├── README.md
├── .env.example
├── go.mod
├── go.sum
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── config/
│   ├── database/
│   ├── middleware/
│   ├── auth/
│   ├── users/
│   ├── patients/
│   ├── doctors/
│   ├── registrations/
│   ├── queues/
│   ├── medical_records/
│   ├── prescriptions/
│   ├── medicines/
│   ├── pharmacy/
│   ├── payments/
│   ├── reports/
│   ├── audit/
│   └── settings/
│
├── migrations/
│
├── web/
│   ├── templates/
│   ├── static/
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│
└── tests/
```

Adjust the structure when necessary, but keep the application modular.

---

# 6. DATABASE

Use PostgreSQL.

Every important table should normally contain:

* id
* created_at
* updated_at

Use foreign keys for relationships.

Use indexes for frequently searched fields.

Use database constraints wherever possible.

Never rely only on frontend validation.

Backend validation is mandatory.

---

# 7. DATABASE MIGRATIONS

All database structure changes must use migration files.

Never manually modify production database structure without migration.

Migration names should be clear, for example:

```text
001_create_users.sql
002_create_roles.sql
003_create_patients.sql
004_create_doctors.sql
```

Never delete existing migrations unless explicitly instructed.

---

# 8. CORE DATABASE ENTITIES

The initial database should be designed around these entities:

### Authentication

* users
* roles
* permissions
* user_roles

### Clinic

* clinic_settings
* staff
* doctors

### Patients

* patients
* patient_addresses
* patient_contacts

### Registration

* registrations
* queues

### Medical

* medical_records
* examinations
* diagnoses
* treatments
* medical_record_diagnoses

### Pharmacy

* medicines
* medicine_categories
* medicine_units
* medicine_stock
* prescriptions
* prescription_items

### Finance

* invoices
* invoice_items
* payments

### System

* audit_logs

Do not create unnecessary tables before they are required.

---

# 9. PATIENT DATA

Patient data is sensitive.

Patient records must never be exposed to unauthorized users.

Important patient fields may include:

* Medical record number
* Patient name
* NIK
* Gender
* Date of birth
* Address
* Phone number
* Emergency contact
* Insurance information
* Blood type
* Allergies

Do not expose sensitive information unnecessarily.

Use access control on every patient-related API.

---

# 10. MEDICAL RECORDS

Medical records are critical data.

A medical record should be associated with:

* Patient
* Doctor
* Registration
* Examination date
* Chief complaint
* Vital signs
* Anamnesis
* Physical examination
* Diagnosis
* Treatment
* Prescription
* Notes

Never permanently delete medical records through a normal UI.

Use status/soft delete where appropriate.

Important changes should be recorded in audit logs.

Do not allow unauthorized users to edit medical records.

---

# 11. USERS AND ROLES

The system must support:

### ADMIN

Full system access.

### DOCTOR

Access to:

* Patient data
* Queue
* Medical examination
* Medical records
* Diagnosis
* Treatment
* Prescription

### NURSE

Access to:

* Patient registration
* Queue
* Basic examination
* Vital signs
* Patient information

### PHARMACIST

Access to:

* Prescriptions
* Medicines
* Stock
* Pharmacy transactions

### CASHIER

Access to:

* Invoices
* Payments
* Transactions

### OWNER / MANAGEMENT

Access to:

* Dashboard
* Reports
* Financial reports
* Operational reports

Do not assume every role has full access.

Authorization must be enforced on the backend.

---

# 12. AUTHENTICATION

Implement:

* Login
* Logout
* Password hashing
* JWT authentication
* Token validation
* Role-based authorization

Never store plaintext passwords.

Never store passwords in logs.

Never put JWT secrets or database passwords directly in source code.

Use environment variables.

Example:

```env
DATABASE_URL=
JWT_SECRET=
APP_ENV=
APP_PORT=
```

Provide `.env.example`.

Never commit `.env`.

---

# 13. API RULES

Use REST API.

Example:

```text
POST   /api/auth/login
POST   /api/auth/logout

GET    /api/patients
POST   /api/patients
GET    /api/patients/:id
PUT    /api/patients/:id

GET    /api/doctors
POST   /api/doctors

GET    /api/registrations
POST   /api/registrations

GET    /api/queues

GET    /api/medical-records
POST   /api/medical-records

GET    /api/prescriptions
POST   /api/prescriptions

GET    /api/medicines
POST   /api/medicines

GET    /api/payments
POST   /api/payments
```

Use appropriate HTTP status codes.

Examples:

```text
200 OK
201 Created
400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
409 Conflict
422 Unprocessable Entity
500 Internal Server Error
```

API responses must be consistent.

---

# 14. VALIDATION

Validate all user input on the backend.

Validate:

* Required fields
* Email
* Phone number
* Date
* Numeric values
* IDs
* Duplicate records
* Medicine quantities
* Payment amounts

Never trust frontend validation.

---

# 15. SECURITY

Protect against:

* SQL Injection
* XSS
* CSRF where applicable
* Broken authentication
* Broken authorization
* IDOR
* Credential leakage
* Sensitive information exposure

Use parameterized queries or a safe database abstraction.

Never concatenate raw user input into SQL queries.

Never expose internal errors to users.

Never return unnecessary sensitive patient information.

---

# 16. AUDIT LOG

Important actions must be recorded.

Examples:

* Login
* Logout
* Create patient
* Update patient
* Create medical record
* Update medical record
* Create prescription
* Change medicine stock
* Create payment
* Cancel transaction
* Change user role

Audit log should record at minimum:

* User
* Action
* Entity
* Entity ID
* Timestamp
* IP address where appropriate
* Description

---

# 17. UI / UX

The interface must be simple for clinic staff.

Prioritize:

* Fast patient search
* Fast registration
* Clear queue status
* Easy medical record access
* Easy prescription entry
* Easy pharmacy processing
* Easy payment processing

Use:

* Sidebar navigation
* Dashboard cards
* Tables
* Search
* Filters
* Pagination
* Modal/form where appropriate
* Confirmation dialog for destructive actions

Responsive design is required.

The system should work on:

* Desktop
* Laptop
* Tablet
* Mobile browser

---

# 18. DASHBOARD

Dashboard should depend on the user's role.

Possible information:

* Total patients
* Today's registrations
* Waiting patients
* Patients being examined
* Completed visits
* Today's revenue
* Low-stock medicines

Do not show financial information to users without permission.

---

# 19. PATIENT REGISTRATION

Registration flow:

```text
Search Patient
      ↓
Existing Patient?
   ↓         ↓
 YES        NO
 ↓           ↓
Register   Create Patient
      ↓
Select Doctor
      ↓
Create Queue
      ↓
Patient Waiting
```

Patient registration must prevent accidental duplicate registration.

---

# 20. QUEUE

Queue should support:

* Queue number
* Patient
* Doctor
* Registration
* Status
* Registration time

Statuses may include:

```text
WAITING
CALLED
IN_EXAMINATION
COMPLETED
CANCELLED
```

Queue changes must be handled consistently.

---

# 21. PHARMACY

Medicine management should support:

* Medicine name
* Code
* Category
* Unit
* Purchase price
* Selling price
* Stock
* Minimum stock
* Expiration date
* Batch number

Stock must be updated through controlled transactions.

Never directly modify stock without recording the reason/transaction.

---

# 22. PRESCRIPTION

Prescription must contain:

* Patient
* Doctor
* Date
* Medicine
* Quantity
* Dosage
* Frequency
* Duration
* Instructions

Pharmacist should be able to process prescription status.

Example:

```text
PENDING
PROCESSING
COMPLETED
CANCELLED
```

---

# 23. PAYMENT

Payment should support:

* Invoice number
* Patient
* Registration
* Invoice items
* Total
* Discount if applicable
* Payment amount
* Payment method
* Payment status

Possible payment methods:

* CASH
* BANK_TRANSFER
* QRIS
* DEBIT
* CREDIT_CARD
* OTHER

Never mark an invoice as paid unless the payment is valid.

---

# 24. REPORTS

Reports may include:

* Patient report
* Registration report
* Doctor activity
* Medical visit report
* Medicine stock report
* Medicine usage
* Revenue report
* Payment report

Reports should support:

* Date filter
* Search
* Pagination where applicable
* Export where required

---

# 25. ERROR HANDLING

Never ignore errors.

Bad:

```go
result, _ := db.Query(...)
```

Prefer:

```go
result, err := db.Query(...)
if err != nil {
    return err
}
```

Errors must be logged internally but sensitive details must not be exposed to users.

---

# 26. LOGGING

Logs should help debugging.

Include:

* Timestamp
* Level
* Request
* User where appropriate
* Error

Never log:

* Password
* JWT secret
* API secret
* Sensitive medical information unnecessarily

---

# 27. TESTING

Create tests for important business logic.

At minimum test:

* Authentication
* Authorization
* Patient creation
* Duplicate patient prevention
* Registration
* Queue
* Medical record
* Prescription
* Medicine stock
* Payment

Before considering a feature complete:

```bash
go test ./...
go build ./...
```

Fix all build errors.

Do not consider a feature finished while tests or build are failing.

---

# 28. DEVELOPMENT ORDER

Build in this order.

## Phase 1 — Foundation

1. Project structure
2. Configuration
3. Database connection
4. Migration system
5. Logging
6. Basic REST API
7. Error handling

## Phase 2 — Authentication

1. Users
2. Roles
3. Login
4. Logout
5. JWT
6. Middleware
7. RBAC

## Phase 3 — Clinic Core

1. Clinic settings
2. Doctors
3. Staff
4. Patients
5. Patient search

## Phase 4 — Registration

1. Registration
2. Queue
3. Doctor schedule
4. Queue dashboard

## Phase 5 — Medical

1. Examination
2. Medical records
3. Diagnosis
4. Treatment
5. Prescription

## Phase 6 — Pharmacy

1. Medicine master
2. Medicine stock
3. Batch
4. Expiration
5. Prescription processing

## Phase 7 — Finance

1. Invoice
2. Payment
3. Payment methods
4. Transaction history

## Phase 8 — Reports

1. Patient reports
2. Medical reports
3. Pharmacy reports
4. Financial reports

## Phase 9 — Security & Production

1. Audit logs
2. Security review
3. Database indexes
4. Error handling review
5. API security review
6. Backup strategy
7. Production configuration

---

# 29. AI AGENT WORKFLOW

Before changing code:

1. Inspect the existing project.
2. Read relevant files.
3. Check database schema.
4. Check existing API.
5. Check existing frontend.
6. Understand the current implementation.
7. Make the smallest appropriate change.
8. Test the change.
9. Fix errors.
10. Re-test the entire affected feature.

Never blindly overwrite files.

Never recreate existing functionality.

Never delete working functionality without explicit instruction.

---

# 30. WHEN SOMETHING IS UNCLEAR

Do not invent business rules.

If a requirement is unclear:

* inspect existing implementation;
* check related modules;
* use standard clinic workflow;
* choose the simplest reasonable implementation;
* document assumptions in code or README when necessary.

For medical decisions, do not invent clinical rules.

The software manages clinical data; it does not replace medical professionals.

---

# 31. GIT

Make changes in small logical steps.

Do not commit:

```text
.env
credentials
passwords
API keys
secrets
temporary files
build artifacts
```

Use meaningful commit messages.

Example:

```text
feat: add patient registration
feat: add JWT authentication
fix: prevent duplicate patient registration
feat: add prescription module
```

---

# 32. PERFORMANCE

The application should remain responsive with growing data.

Use:

* Database indexes
* Pagination
* Efficient queries
* Proper joins
* Avoid unnecessary API requests

Do not load thousands of records into the browser at once.

---

# 33. IMPORTANT AI RULES

The AI agent MUST NOT:

* Delete the database without explicit instruction.
* Delete migrations.
* Remove authentication.
* Disable authorization.
* Expose passwords.
* Expose secrets.
* Hardcode credentials.
* Delete medical records casually.
* Modify production configuration without instruction.
* Replace the entire project architecture unnecessarily.
* Install unnecessary dependencies.
* Rewrite working modules without a clear reason.

The AI agent MUST:

* Preserve existing functionality.
* Keep code readable.
* Follow the existing architecture.
* Validate input.
* Handle errors.
* Test changes.
* Protect patient data.
* Maintain database integrity.

---

# 34. DEFINITION OF DONE

A feature is complete only when:

* Backend works.
* Frontend works.
* Database migration works.
* API works.
* Authentication works where required.
* Authorization works.
* Validation exists.
* Error handling exists.
* Tests pass.
* Build succeeds.
* Existing functionality is not broken.
* No secrets are exposed.
* UI is usable on desktop and mobile.

---

# 35. FIRST TASK

When starting this project for the first time, DO NOT immediately build all modules.

First:

1. Inspect the project directory.
2. Create the initial project structure.
3. Initialize the Go application.
4. Configure PostgreSQL/Supabase.
5. Create database migrations.
6. Create users and roles.
7. Implement authentication.
8. Implement JWT middleware.
9. Implement RBAC.
10. Create a basic Indonesian login page.
11. Create role-based dashboard.
12. Run tests.
13. Run build.
14. Only after authentication is stable, continue to the Patient module.

Always work incrementally.

The application must remain runnable after every development step.

# END OF AGENTS.md
