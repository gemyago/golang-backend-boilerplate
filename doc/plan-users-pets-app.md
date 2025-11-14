# Plan: Users & Pets App Implementation

## 1. Introduction/Overview

This plan outlines the implementation of a simple users and pets management application that demonstrates the architecture patterns of the golang-backend-boilerplate project. The application follows hexagonal architecture with CQRS principles where Commands handle data mutations and Queries handle data retrieval operations.

### Goal
Build a minimal yet complete example application that:
- Manages user records with CRUD operations via UserCommands
- Creates pets in external Petstore API and manages user-pet relationships via PetsCommands
- Provides data retrieval via separate UserQueries and PetsQueries
- Uses SQLite as the underlying storage with a pure Go driver (no cgo)
- Follows the project's hexagonal architecture with proper layer separation
- Demonstrates "accept interface, return struct" principle
- Shows how application layer defines ports (interfaces) for infrastructure dependencies

### Problem It Solves
This app serves as a canonical example demonstrating:
- How to implement CQRS pattern in the application layer with clear separation
- How to properly define ports (interfaces) in application layer for infrastructure dependencies
- How to create and integrate a database repository in the infrastructure layer
- How to integrate with external HTTP APIs (petstore)
- How to wire HTTP endpoints to application logic
- How to write comprehensive tests following TDD approach
- API-driven design where models are generated from OpenAPI spec
- Proper layer separation where infrastructure implementations satisfy app-defined interfaces

## 2. Business Logic

### User Management
- **Create User**: Accept user details (name, email) and create a new user record
  - Email must be unique across all users
  - Both name and email are required fields
  - Returns user ID only
- **Update User**: Modify existing user details (name and/or email)
  - User must exist (identified by ID)
  - Email uniqueness constraint still applies
  - Returns nothing (void)
- **Delete User**: Remove a user from the system
  - User must exist
  - Deleting a user also removes all their pet relationships
  - Returns nothing (void)
- **Get User by ID**: Retrieve a specific user's details
  - Return user ID, name, and email
- **List Users**: Retrieve all users
  - Return user ID, name, and email for each user
  - No pagination for simplicity

### Pet Management
- **Add Pet to User**: Create a new pet in Petstore API and associate it with a user
  - User must exist
  - Pet data (name, status, photo URLs) provided in request
  - Pet is created in Petstore API first
  - Then relationship is stored in local database
  - Returns pet ID only
- **Remove Pet from User**: Remove the relationship between a user and a pet
  - Both user and pet relationship must exist
  - Only removes the relationship, does not delete the pet from Petstore
  - Returns nothing (void)
- **List User Pets**: Retrieve all pets for a specific user
  - Fetch pet details from Petstore API using stored pet IDs
  - Return pet ID, name, status, and photo URLs

## 3. High Level Architecture

The application follows hexagonal architecture with three main layers:

### Layer 1: Incoming Adapters (HTTP API)
- Location: `internal/api/http`
- OpenAPI specification: `internal/api/http/v1routes.yaml`
- Generated models in `internal/api/http/v1routes/models`
- HTTP controllers in `internal/api/http/v1controllers`
- Controllers use application layer services (Commands/Queries) as concrete structs

### Layer 2: Application Layer (Business Logic - CQRS)
- Location: `internal/app`
- **UserCommands**: `users_commands.go` - User CRUD operations (concrete struct, not interface)
- **PetsCommands**: `pets_commands.go` - Pet creation and relationship management (concrete struct)
- **UserQueries**: `users_queries.go` - User data retrieval (concrete struct)
- **PetsQueries**: `pets_queries.go` - Pet data retrieval (concrete struct)
- **Ports (Interfaces)**: Application layer defines interfaces for infrastructure dependencies
  - `UsersRepository` interface - defined in application layer
  - `PetsRepository` interface - defined in application layer
  - `PetstoreClient` interface - defined in application layer (satisfied by existing petstore.Client)
- Domain entities and errors in the same files
- Services return concrete structs, accept interfaces as dependencies

### Layer 3: Outgoing Adapters (Infrastructure)
- Location: `internal/infrastructure`
- **UsersRepository implementation**: Concrete struct implementing app.UsersRepository interface
- **PetsRepository implementation**: Concrete struct implementing app.PetsRepository interface
- **Petstore Client**: Already exists in `internal/infrastructure/petstore` - structurally satisfies app.PetstoreClient interface
- Infrastructure provides implementations, application defines interfaces
- Uses `modernc.org/sqlite` (pure Go, no cgo)

### Key Architectural Principles
1. **Accept Interface, Return Struct**: Services accept interfaces as dependencies but return concrete structs
2. **Consumer Defines Interface**: Application layer (consumer) defines repository interfaces, not infrastructure (provider)
3. **Commands/Queries are Concrete Structs**: No interfaces for application services - they're concrete implementations
4. **Simple Domain Entities**: Single `User` entity with all attributes, no artificial DTO/domain separation
5. **Petstore Port Without Adapter**: App layer defines PetstoreClient interface, existing petstore.Client satisfies it structurally, no adapter needed
6. **Direct Petstore Types**: App layer uses petstore data types (Pet, AddPetRequest) directly in the interface
7. **Layer Boundaries**: Infrastructure types don't leak to app layer; repositories use domain entities
8. **CQRS Separation**: Clear distinction between write operations (Commands) and read operations (Queries)

### Data Flow
```
HTTP Request → Controller → Command/Query (concrete struct) → Repository Port (interface) 
                                                                      ↓
                                                           Infrastructure Implementation
                                                                      ↓
HTTP Response ← Transform ← App Layer ← Repository Implementation → SQLite
                                    ↘
                                      → Petstore Client → External API
```

## 4. Detailed Architecture

### 4.1 Database Schema (SQLite)

**Users Table**
```sql
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**User-Pets Relationship Table**
```sql
CREATE TABLE IF NOT EXISTS user_pets (
    user_id TEXT NOT NULL,
    pet_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, pet_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 4.2 Application Layer Ports (Interfaces)

The application layer defines interfaces (ports) for its dependencies. Infrastructure provides implementations.

#### UsersRepository Port (defined in app layer)

```go
// internal/app/users_repository.go
package app

import (
	"context"
	"time"
)

// User represents a user entity with all attributes needed for persistence and business logic
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UsersRepository is the port (interface) for user persistence operations
// Defined in application layer, implemented by infrastructure layer
type UsersRepository interface {
	CreateUser(ctx context.Context, user User) error
	UpdateUser(ctx context.Context, user User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
}
```

#### PetsRepository Port (defined in app layer)

```go
// internal/app/pets_repository.go
package app

import (
	"context"
	"time"
)

// UserPet represents a user-pet relationship
type UserPet struct {
	UserID    string
	PetID     int64
	CreatedAt time.Time
}

// PetsRepository is the port (interface) for pet relationship persistence
// Defined in application layer, implemented by infrastructure layer
type PetsRepository interface {
	AddUserPet(ctx context.Context, userPet UserPet) error
	RemoveUserPet(ctx context.Context, userID string, petID int64) error
	GetUserPetIDs(ctx context.Context, userID string) ([]int64, error)
	HasUserPet(ctx context.Context, userID string, petID int64) (bool, error)
}
```

#### PetstoreClient Port (defined in app layer)

```go
// internal/app/petstore_client_port.go
package app

import (
	"context"
	
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
)

// PetstoreClient is the port (interface) for interacting with Petstore API
// Defined in application layer, describes capabilities needed by the app
// The existing petstore.Client struct satisfies this interface structurally
// We use petstore package types directly (Pet, AddPetRequest, etc.) - they're just DTOs
type PetstoreClient interface {
	AddPet(ctx context.Context, req *petstore.AddPetRequest) (*petstore.Pet, error)
	GetPetByID(ctx context.Context, petID int64) (*petstore.Pet, error)
}
```

### 4.3 Application Layer - Commands & Queries (Concrete Structs)

#### UserCommands

```go
// internal/app/users_commands.go
package app

import (
	"context"
	"errors"
	"log/slog"
	
	"github.com/gofrs/uuid/v5"
	"go.uber.org/dig"
)

// Command request/response types
type CreateUserRequest struct {
	Name  string
	Email string
}

type CreateUserResponse struct {
	UserID string
}

type UpdateUserRequest struct {
	UserID string
	Name   string
	Email  string
}

// Domain errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserEmailConflict = errors.New("user with this email already exists")
	ErrInvalidInput      = errors.New("invalid input")
)

// UserCommands is a concrete struct (not an interface)
// Controllers use this directly
type UserCommands struct {
	usersRepo UsersRepository
	logger    *slog.Logger
}

type UserCommandsDeps struct {
	dig.In
	
	UsersRepo  UsersRepository
	RootLogger *slog.Logger
}

// NewUserCommands returns a concrete struct (not an interface)
// This follows "accept interface, return struct" principle
func NewUserCommands(deps UserCommandsDeps) *UserCommands {
	return &UserCommands{
		usersRepo: deps.UsersRepo,
		logger:    deps.RootLogger.WithGroup("app.user-commands"),
	}
}

func (c *UserCommands) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// Validate input
	// Check email uniqueness
	// Generate UUID
	// Create user in repository
	// Return user ID
	return nil, errors.New("not implemented")
}

func (c *UserCommands) UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	// Validate input
	// Check user exists
	// Check email uniqueness (if changed)
	// Update user
	return errors.New("not implemented")
}

func (c *UserCommands) DeleteUser(ctx context.Context, userID string) error {
	// Check user exists
	// Delete user (cascade will delete relationships)
	return errors.New("not implemented")
}
```

#### PetsCommands

```go
// internal/app/pets_commands.go
package app

import (
	"context"
	"errors"
	"log/slog"
	
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.uber.org/dig"
)

type AddPetRequest struct {
	UserID    string
	Name      string
	Status    string
	PhotoUrls []string
}

type AddPetResponse struct {
	PetID int64
}

var (
	ErrPetCreationFailed = errors.New("failed to create pet in petstore")
	ErrUserPetNotFound   = errors.New("user-pet relationship not found")
)

// PetsCommands is a concrete struct
type PetsCommands struct {
	petsRepo       PetsRepository
	usersRepo      UsersRepository
	petstoreClient PetstoreClient // Use the interface defined in app layer
	logger         *slog.Logger
}

type PetsCommandsDeps struct {
	dig.In
	
	PetsRepo       PetsRepository
	UsersRepo      UsersRepository
	PetstoreClient PetstoreClient // Interface, satisfied by *petstore.Client
	RootLogger     *slog.Logger
}

func NewPetsCommands(deps PetsCommandsDeps) *PetsCommands {
	return &PetsCommands{
		petsRepo:       deps.PetsRepo,
		usersRepo:      deps.UsersRepo,
		petstoreClient: deps.PetstoreClient,
		logger:         deps.RootLogger.WithGroup("app.pets-commands"),
	}
}

func (c *PetsCommands) AddPet(ctx context.Context, req AddPetRequest) (*AddPetResponse, error) {
	// Validate input
	// Check user exists
	// Create pet in Petstore using petstore.AddPetRequest type
	// Add relationship to repository
	// Return pet ID
	return nil, errors.New("not implemented")
}

func (c *PetsCommands) RemovePet(ctx context.Context, userID string, petID int64) error {
	// Check user exists
	// Check relationship exists
	// Remove relationship
	return errors.New("not implemented")
}
```

#### UserQueries

```go
// internal/app/users_queries.go
package app

import (
	"context"
	"log/slog"
	
	"go.uber.org/dig"
)

// UserQueries is a concrete struct
type UserQueries struct {
	usersRepo UsersRepository
	logger    *slog.Logger
}

type UserQueriesDeps struct {
	dig.In
	
	UsersRepo  UsersRepository
	RootLogger *slog.Logger
}

func NewUserQueries(deps UserQueriesDeps) *UserQueries {
	return &UserQueries{
		usersRepo: deps.UsersRepo,
		logger:    deps.RootLogger.WithGroup("app.user-queries"),
	}
}

func (q *UserQueries) GetUserByID(ctx context.Context, userID string) (*User, error) {
	// Fetch from repository
	// Return user (without timestamps for API response)
	return nil, errors.New("not implemented")
}

func (q *UserQueries) ListUsers(ctx context.Context) ([]*User, error) {
	// Fetch all from repository
	return nil, errors.New("not implemented")
}
```

#### PetsQueries

```go
// internal/app/pets_queries.go
package app

import (
	"context"
	"log/slog"
	
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.uber.org/dig"
)

// PetsQueries is a concrete struct
type PetsQueries struct {
	petsRepo       PetsRepository
	usersRepo      UsersRepository
	petstoreClient PetstoreClient // Use the interface defined in app layer
	logger         *slog.Logger
}

type PetsQueriesDeps struct {
	dig.In
	
	PetsRepo       PetsRepository
	UsersRepo      UsersRepository
	PetstoreClient PetstoreClient // Interface, satisfied by *petstore.Client
	RootLogger     *slog.Logger
}

func NewPetsQueries(deps PetsQueriesDeps) *PetsQueries {
	return &PetsQueries{
		petsRepo:       deps.PetsRepo,
		usersRepo:      deps.UsersRepo,
		petstoreClient: deps.PetstoreClient,
		logger:         deps.RootLogger.WithGroup("app.pets-queries"),
	}
}

func (q *PetsQueries) ListUserPets(ctx context.Context, userID string) ([]*petstore.Pet, error) {
	// Verify user exists
	// Get pet IDs from repository
	// Fetch pet details from Petstore for each ID
	// Return pets (petstore.Pet type used directly)
	return nil, errors.New("not implemented")
}
```

### 4.4 Infrastructure Layer - Repository Implementations

Infrastructure provides concrete implementations of the ports defined in the application layer.

#### UsersRepository Implementation

```go
// internal/infrastructure/users_repository.go
package services

import (
	"context"
	"database/sql"
	
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/dig"
)

// sqliteUsersRepository implements app.UsersRepository interface
type sqliteUsersRepository struct {
	db   *sql.DB
	time TimeProvider
}

type UsersRepositoryDeps struct {
	dig.In
	
	DB   *Database
	Time TimeProvider
}

// NewUsersRepository returns a concrete implementation of app.UsersRepository
// This follows "accept interface, return struct" principle
func NewUsersRepository(deps UsersRepositoryDeps) *sqliteUsersRepository {
	return &sqliteUsersRepository{
		db:   deps.DB.instance,
		time: deps.Time,
	}
}

// Implement all methods from app.UsersRepository interface
// Methods use app.User entity directly
func (r *sqliteUsersRepository) CreateUser(ctx context.Context, user app.User) error {
	// Generate UUID if not provided
	// Set timestamps
	// INSERT into database
	// Handle unique constraint violations
	return nil
}

// ... other methods: UpdateUser, DeleteUser, GetUserByID, GetUserByEmail, ListUsers
```

#### PetsRepository Implementation

```go
// internal/infrastructure/pets_repository.go
package services

import (
	"context"
	"database/sql"
	
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"go.uber.org/dig"
)

// sqlitePetsRepository implements app.PetsRepository interface
type sqlitePetsRepository struct {
	db   *sql.DB
	time TimeProvider
}

type PetsRepositoryDeps struct {
	dig.In
	
	DB   *Database
	Time TimeProvider
}

// NewPetsRepository returns a concrete implementation of app.PetsRepository
func NewPetsRepository(deps PetsRepositoryDeps) *sqlitePetsRepository {
	return &sqlitePetsRepository{
		db:   deps.DB.instance,
		time: deps.Time,
	}
}

// Implement all methods from app.PetsRepository interface
// Methods use app.UserPet entity directly
```

### 4.5 HTTP API Layer

**OpenAPI Specification**: `internal/api/http/v1routes.yaml`

```yaml
openapi: 3.0.3
info:
  title: Users & Pets API
  version: 1.0.0

paths:
  /users:
    post:
      summary: Create a new user
      operationId: createUser
      tags: [users]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateUserResponse'
        default:
          $ref: '#/components/responses/ErrorResponse'
    get:
      summary: List all users
      operationId: listUsers
      tags: [users]
      responses:
        '200':
          description: List of users
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUsersResponse'
        default:
          $ref: '#/components/responses/ErrorResponse'

  /users/{userId}:
    parameters:
      - name: userId
        in: path
        required: true
        schema:
          type: string
    get:
      summary: Get user by ID
      operationId: getUserById
      tags: [users]
      responses:
        '200':
          description: User details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        default:
          $ref: '#/components/responses/ErrorResponse'
    put:
      summary: Update user
      operationId: updateUser
      tags: [users]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUserRequest'
      responses:
        '204':
          description: User updated
        default:
          $ref: '#/components/responses/ErrorResponse'
    delete:
      summary: Delete user
      operationId: deleteUser
      tags: [users]
      responses:
        '204':
          description: User deleted
        default:
          $ref: '#/components/responses/ErrorResponse'

  /users/{userId}/pets:
    parameters:
      - name: userId
        in: path
        required: true
        schema:
          type: string
    get:
      summary: List user's pets
      operationId: listUserPets
      tags: [pets]
      responses:
        '200':
          description: List of user's pets
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUserPetsResponse'
        default:
          $ref: '#/components/responses/ErrorResponse'
    post:
      summary: Add a pet to user
      operationId: addUserPet
      tags: [pets]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AddPetRequest'
      responses:
        '201':
          description: Pet added to user
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AddPetResponse'
        default:
          $ref: '#/components/responses/ErrorResponse'

  /users/{userId}/pets/{petId}:
    parameters:
      - name: userId
        in: path
        required: true
        schema:
          type: string
      - name: petId
        in: path
        required: true
        schema:
          type: integer
          format: int64
    delete:
      summary: Remove a pet from user
      operationId: removeUserPet
      tags: [pets]
      responses:
        '204':
          description: Pet removed from user
        default:
          $ref: '#/components/responses/ErrorResponse'

components:
  schemas:
    CreateUserRequest:
      type: object
      required: [name, email]
      properties:
        name:
          type: string
          minLength: 1
        email:
          type: string
          format: email

    CreateUserResponse:
      type: object
      required: [userId]
      properties:
        userId:
          type: string

    UpdateUserRequest:
      type: object
      required: [name, email]
      properties:
        name:
          type: string
          minLength: 1
        email:
          type: string
          format: email

    UserResponse:
      type: object
      required: [id, name, email]
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string

    AddPetRequest:
      type: object
      required: [name, status]
      properties:
        name:
          type: string
          minLength: 1
        status:
          type: string
          enum: [available, pending, sold]
        photoUrls:
          type: array
          items:
            type: string

    AddPetResponse:
      type: object
      required: [petId]
      properties:
        petId:
          type: integer
          format: int64

    PetResponse:
      type: object
      required: [id, name, status, photoUrls]
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        status:
          type: string
        photoUrls:
          type: array
          items:
            type: string

    ListUsersResponse:
      type: object
      required: [users]
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/UserResponse'

    ListUserPetsResponse:
      type: object
      required: [pets]
      properties:
        pets:
          type: array
          items:
            $ref: '#/components/schemas/PetResponse'

  responses:
    ErrorResponse:
      description: Error response
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

### 4.6 Configuration

Configuration in `internal/config/default.json`:

```json
{
  "database": {
    "dsn": "./data/app.db"
  }
}
```

## 5. Key Architectural Decisions

### 5.1 Application Layer Defines Ports (Interfaces)

**Decision**: Repository interfaces are defined in the application layer, not infrastructure.

**Rationale**:
- Follows dependency inversion principle
- Application layer is independent of infrastructure details
- Easier to test application layer with mocks
- Aligns with hexagonal architecture where ports are defined by the application

### 5.2 Commands/Queries are Concrete Structs (Not Interfaces)

**Decision**: UserCommands, PetsCommands, UserQueries, PetsQueries are concrete structs, not interfaces.

**Rationale**:
- Controllers consume these services, so if interfaces were needed, controllers would define them
- No need to define interfaces in the provider layer (app layer)
- Follows "accept interface, return struct" principle
- Simpler, more direct approach
- If API layer needs to mock them for testing, API layer can define interfaces

### 5.3 Single User Entity (Not Separate Domain/Repository Types)

**Decision**: Single `User` struct with all attributes (ID, Name, Email, CreatedAt, UpdatedAt) used throughout.

**Rationale**:
- Simpler and more pragmatic
- Timestamps are just part of the entity, not a separate concern
- No artificial separation between "domain" and "persistence" models
- Easier to understand and maintain
- Business logic can ignore timestamps if not needed, but they're available

### 5.4 Petstore Port Without Adapter

**Decision**: App layer defines PetstoreClient interface describing needed capabilities. Existing `*petstore.Client` satisfies this interface structurally. Petstore data types used directly in the interface.

**Rationale**:
- Follows "consumer defines interface" principle - app layer defines what it needs
- No adapter layer needed - `*petstore.Client` satisfies the interface through structural typing
- Petstore types (Pet, AddPetRequest, etc.) are just DTOs, safe to use directly in app layer
- Simpler than creating wrapper types or adapter
- If we need to swap Petstore implementation, the new client just needs to satisfy the interface
- App layer remains independent while avoiding unnecessary indirection

### 5.5 CQRS Separation with Dedicated Services

**Decision**: Separate Commands and Queries into distinct services with clear responsibilities.

**Rationale**:
- Clear separation of write vs read operations
- Different dependencies can be injected if needed
- Easier to optimize each side independently
- Better testability with focused services

### 5.6 Pet Creation in Petstore API

**Decision**: Pets are created in external Petstore API, not local database.

**Rationale**:
- Demonstrates external API integration
- Shows how to coordinate between local and external state
- Realistic scenario (using third-party service)

### 5.7 Lean Mutations

**Decision**: Update and Delete operations return void (error only), not the modified entity.

**Rationale**:
- Simpler API contracts
- Avoids unnecessary database queries
- Client can fetch entity if needed

### 5.8 No Pagination

**Decision**: List operations don't include pagination.

**Rationale**:
- Keeps example simple
- Focus on architecture patterns, not pagination complexity
- Production systems should add pagination

### 5.9 Database Native Timestamps

**Decision**: Use SQLite's DEFAULT CURRENT_TIMESTAMP for created_at, manual updates for updated_at.

**Rationale**:
- Leverages database capabilities
- Consistent timestamps even if multiple processes write
- Simple implementation

### 5.10 API-Driven Design

**Decision**: OpenAPI spec is source of truth, models are generated.

**Rationale**:
- Contract-first approach
- Ensures API consistency
- Reduces manual DTO mapping errors

### 5.11 Cascade Delete for User-Pet Relationships

**Decision**: SQLite foreign key with ON DELETE CASCADE for user_pets table.

**Rationale**:
- Automatic cleanup when user is deleted
- Database enforces referential integrity
- Simpler application logic

### 5.12 Package Naming - "services" for Infrastructure

**Decision**: Infrastructure layer uses `package services` (already in codebase).

**Rationale**:
- Follows existing codebase convention
- Avoids conflict with `infrastructure` being a directory name
- Clear that these are service implementations

## 6. Implementation Status

Based on task summaries in `doc/implementation/plan-users-pets-app/`:

### ✅ Completed
- **Phase 1**: Foundation (Tasks 1.1-1.3)
  - SQLite dependency added
  - Database configuration added
  
- **Phase 2**: Database Layer - Users Repository (Tasks 2.1-2.6)
  - UsersRepository structure created (currently as concrete struct, needs refactoring)
  - All CRUD operations implemented and tested
  - Schema initialization complete

### 🔄 Needs Refactoring
- **UsersRepository**: Currently defined as concrete struct in infrastructure without app layer interface
  - Need to: Define `UsersRepository` interface in `internal/app/users_repository.go`
  - Need to: Define `User` entity in app layer
  - Need to: Refactor implementation to use app layer interface and entity
  - Tests need updating to reflect new structure

### ⏳ Not Started
- **Phase 3**: Database Layer - Pets Repository
- **Phase 4**: Application Layer - User Commands
- **Phase 5**: Application Layer - Pets Commands
- **Phase 6**: Application Layer - User Queries
- **Phase 7**: Application Layer - Pets Queries
- **Phase 8**: HTTP Controllers - Users
- **Phase 9**: HTTP Controllers - Pets
- **Phase 10**: Integration & Final Verification

## 7. Related Files

### Files to Create (New)

**Application Layer**:
- `internal/app/users_repository.go` - UsersRepository interface and User entity
- `internal/app/pets_repository.go` - PetsRepository interface and UserPet entity
- `internal/app/petstore_client_port.go` - PetstoreClient interface (satisfied by petstore.Client)
- `internal/app/users_commands.go` - UserCommands struct (concrete, not interface)
- `internal/app/users_commands_test.go` - Tests for UserCommands
- `internal/app/users_queries.go` - UserQueries struct (concrete, not interface)
- `internal/app/users_queries_test.go` - Tests for UserQueries
- `internal/app/pets_commands.go` - PetsCommands struct (concrete, not interface)
- `internal/app/pets_commands_test.go` - Tests for PetsCommands
- `internal/app/pets_queries.go` - PetsQueries struct (concrete, not interface)
- `internal/app/pets_queries_test.go` - Tests for PetsQueries
- `internal/app/users_testing.go` - Test factories for user-related types
- `internal/app/pets_testing.go` - Test factories for pet-related types

**Infrastructure Layer**:
- `internal/infrastructure/pets_repository.go` - PetsRepository implementation
- `internal/infrastructure/pets_repository_test.go` - Tests for PetsRepository
- `internal/infrastructure/pets_testing.go` - Test factories for pets

**API Layer**:
- `internal/api/http/v1controllers/users.go` - Users HTTP controllers
- `internal/api/http/v1controllers/users_test.go` - Tests for users controllers
- `internal/api/http/v1controllers/pets.go` - Pets HTTP controllers
- `internal/api/http/v1controllers/pets_test.go` - Tests for pets controllers

### Files to Update (Existing)

- `internal/api/http/v1routes.yaml` - Add users and pets endpoints
- `internal/infrastructure/users_repository.go` - Refactor to implement app.UsersRepository
- `internal/infrastructure/users_repository_test.go` - Update for new interface
- `internal/infrastructure/users_testing.go` - Update for app.User entity
- `internal/infrastructure/database.go` - Add pets schema initialization
- `internal/infrastructure/register.go` - Register repository implementations
- `internal/app/register.go` - Register commands and queries services
- `internal/api/http/v1controllers/register.go` - Register controllers
- `internal/api/http/v1routes.go` - Wire up routes

### Generated Files (by apigen)

- `internal/api/http/v1routes/models/*.go` - Request/response models
- `internal/api/http/v1routes/handlers/*.go` - Handler interfaces

## 8. Task List

This implementation follows TDD approach as per [tdd-flow.md](../.context/tdd-flow.md). Each task should leave the codebase in a buildable state with all tests passing (`make test`).

### Phase 0: Refactor Existing Code to Match Architecture

**Task 0.1: Define UsersRepository interface and User entity in app layer**
- Create `internal/app/users_repository.go`
- Define `User` struct with ID, Name, Email, CreatedAt, UpdatedAt fields
- Define `UsersRepository` interface with all methods (using User entity)
- Add comment explaining this is a port defined by the application layer
- Run: `make test`
  - Verify compilation errors (infrastructure doesn't implement interface yet)
  - This is expected at this stage
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-0.1.md`
- Success criteria: File created, interface and entity defined, code compiles after next task

**Task 0.2: Refactor users repository implementation**
- Update `internal/infrastructure/users_repository.go`:
  - Import `github.com/gemyago/golang-backend-boilerplate/internal/app`
  - Change all method signatures to use `app.User` instead of local User struct
  - Remove local User struct definition
  - Keep `sqliteUsersRepository` unexported (lowercase)
  - Update `NewUsersRepository` to return `*sqliteUsersRepository` (concrete struct)
- Update `internal/infrastructure/users_repository_test.go`:
  - Import app package
  - Update all test code to use `app.User`
- Update `internal/infrastructure/users_testing.go`:
  - Update `NewRandomUser` to return `app.User`
  - Update all option functions to work with `app.User`
- Run affected tests: `go test -v ./internal/infrastructure/ --run TestUsersRepository`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-0.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 0.3: Update DI registration for users repository**
- Update `internal/infrastructure/register.go`:
  - Ensure `NewUsersRepository` is registered and provides `*sqliteUsersRepository`
  - The DI container should provide it as `app.UsersRepository` interface to consumers
- Verify schema initialization for user_pets table exists in `database.go`
  - If not present, add `initUserPetsSchema` function
  - Add to schema initialization in `newDBProvider`
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
  - Verify logs show database initialization
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-0.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 1: Foundation & API Definition

**Task 1.1: Define OpenAPI specification for users and pets endpoints**
- Update `internal/api/http/v1routes.yaml`:
  - Add all users endpoints (POST/GET /users, GET/PUT/DELETE /users/{userId})
  - Add all pets endpoints (GET/POST /users/{userId}/pets, DELETE /users/{userId}/pets/{petId})
  - Add all request/response schemas (CreateUserRequest, UserResponse, AddPetRequest, etc.)
  - Add proper error responses
  - Follow existing patterns from echo endpoints
- Run: `make lint`
  - Verify YAML validation passes
  - Fix any schema errors
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-1.1.md`
- Success criteria: As per completion protocol: `make lint` passes

**Task 1.2: Generate API models and handlers**
- Run: `go generate ./internal/api/http/v1routes.go`
  - This generates models and handler interfaces
- Verify generated files in `internal/api/http/v1routes/models/` and `internal/api/http/v1routes/handlers/`
- Fix any compilation errors in generated code if needed
- Run: `make test`
  - Verify no breaking changes to existing tests
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-1.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

### Phase 2: Database Layer - Pets Repository

**Task 2.1: Define PetsRepository interface and UserPet entity in app layer**
- Create `internal/app/pets_repository.go`
- Define `UserPet` struct with UserID, PetID, CreatedAt fields
- Define `PetsRepository` interface with all methods (AddUserPet, RemoveUserPet, GetUserPetIDs, HasUserPet)
- Add comment explaining this is a port defined by the application layer
- Run: `make test`
  - Verify code compiles (no implementation yet)
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-2.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 2.2: Create PetsRepository implementation structure**
- Create `internal/infrastructure/pets_repository.go`
- Define `sqlitePetsRepository` struct (unexported)
- Define `PetsRepositoryDeps` struct with DB and Time dependencies
- Add constructor `NewPetsRepository(deps PetsRepositoryDeps) *sqlitePetsRepository`
- Add stub implementations for all interface methods (return `errors.New("not implemented")`)
- Create `internal/infrastructure/pets_repository_test.go` with basic test structure
- Add `makeMockDeps` function following existing patterns
- Write test for schema initialization (verify user_pets table exists)
- Run affected tests: `go test -v ./internal/infrastructure/ --run TestPetsRepository`
  - Verify schema test passes
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-2.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 2.3: Implement and test AddUserPet**
- Create `internal/infrastructure/pets_testing.go` with factory `NewRandomUserPet(fake *faker.Faker, opts ...RandomUserPetOpt) *app.UserPet`
- In `pets_repository_test.go`, add test `TestPetsRepository/AddUserPet`:
  - Happy path: relationship is created and can be verified via GetUserPetIDs
  - Idempotent: adding same pet twice succeeds (use INSERT OR IGNORE)
- Run affected tests: `go test -v ./internal/infrastructure/ --run TestPetsRepository/AddUserPet`
  - Verify failures (not implemented yet)
  - Ensure test failures are expectation failures, not compilation errors
- Implement `AddUserPet` in `pets_repository.go`:
  - Set timestamp using time provider
  - Use `INSERT OR IGNORE` for idempotency
  - Execute INSERT statement
- Run affected tests: `go test -v ./internal/infrastructure/ --run TestPetsRepository/AddUserPet`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-2.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 2.4: Implement and test GetUserPetIDs and HasUserPet**
- In `pets_repository_test.go`, add test `TestPetsRepository/GetUserPetIDs`:
  - Happy path: returns all pet IDs for user in consistent order
  - Empty result: returns empty slice when user has no pets
- Add test `TestPetsRepository/HasUserPet`:
  - Returns true when relationship exists
  - Returns false when relationship doesn't exist
- Run affected tests: `go test -v ./internal/infrastructure/ --run "TestPetsRepository/(GetUserPetIDs|HasUserPet)"`
  - Verify failures (not implemented yet)
- Implement `GetUserPetIDs` in `pets_repository.go`:
  - Query user_pets table ordered by created_at
  - Return slice of pet IDs
- Implement `HasUserPet` in `pets_repository.go`:
  - Query for specific user_id/pet_id pair
  - Return true if row exists, false otherwise
- Run affected tests: `go test -v ./internal/infrastructure/ --run "TestPetsRepository/(GetUserPetIDs|HasUserPet)"`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-2.4.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 2.5: Implement and test RemoveUserPet**
- In `pets_repository_test.go`, add test `TestPetsRepository/RemoveUserPet`:
  - Happy path: relationship is removed and no longer appears in GetUserPetIDs
  - Idempotent: removing non-existent relationship succeeds (no error)
- Run affected tests: `go test -v ./internal/infrastructure/ --run TestPetsRepository/RemoveUserPet`
  - Verify failures (not implemented yet)
- Implement `RemoveUserPet` in `pets_repository.go`:
  - Execute DELETE statement
  - Don't error if no rows affected (idempotent)
- Run affected tests: `go test -v ./internal/infrastructure/ --run TestPetsRepository/RemoveUserPet`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-2.5.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 2.6: Test cascade delete behavior**
- In `pets_repository_test.go`, add integration test `TestPetsRepository/CascadeDelete`:
  - Create user via UsersRepository
  - Add multiple pets via PetsRepository
  - Delete user via UsersRepository
  - Verify all pet relationships are deleted (GetUserPetIDs returns empty)
- Run affected tests: `go test -v ./internal/infrastructure/ --run TestPetsRepository/CascadeDelete`
  - Verify test passes (CASCADE should work from schema)
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-2.6.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 2.7: Register PetsRepository in DI**
- Update `internal/infrastructure/register.go`:
  - Add `NewPetsRepository` to providers
  - Configure DI to provide `*sqlitePetsRepository` as `app.PetsRepository` interface
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
  - Verify both repositories are initialized
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-2.7.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 3: Petstore Client Port

**Task 3.1: Define PetstoreClient interface in app layer**
- Create `internal/app/petstore_client_port.go`
- Import `github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore`
- Define `PetstoreClient` interface with methods:
  - `AddPet(ctx context.Context, req *petstore.AddPetRequest) (*petstore.Pet, error)`
  - `GetPetByID(ctx context.Context, petID int64) (*petstore.Pet, error)`
- Add comment explaining this describes capabilities needed by app layer
- Add comment that existing `*petstore.Client` satisfies this interface structurally
- Run: `make test`
  - Verify code compiles
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-3.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

### Phase 4: Application Layer - User Commands

**Task 4.1: Create UserCommands structure**
- Create `internal/app/users_commands.go`
- Define request/response types: `CreateUserRequest`, `CreateUserResponse`, `UpdateUserRequest`
- Define error types: `ErrUserNotFound`, `ErrUserEmailConflict`, `ErrInvalidInput`
- Define `UserCommands` struct (concrete, not interface) with UsersRepository dependency
- Add `UserCommandsDeps` struct for DI
- Add constructor `NewUserCommands(deps UserCommandsDeps) *UserCommands`
- Add stub implementations for CreateUser, UpdateUser, DeleteUser (return `errors.New("not implemented")`)
- Create `internal/app/users_commands_test.go` with test structure
- Add `makeMockDeps` function that creates mock UsersRepository
- Run: `make test`
  - Verify code compiles
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-4.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 4.2: Implement and test CreateUser command**
- Create `internal/app/users_testing.go` with factory `NewRandomCreateUserRequest(fake *faker.Faker) *CreateUserRequest`
- In `users_commands_test.go`, add test `TestUserCommands/CreateUser`:
  - Happy path: user is created, returns valid UUID
  - Validation: empty name returns ErrInvalidInput
  - Validation: empty email returns ErrInvalidInput
  - Validation: invalid email format returns ErrInvalidInput
  - Conflict: duplicate email returns ErrUserEmailConflict
- Run affected tests: `go test -v ./internal/app/ --run TestUserCommands/CreateUser`
  - Verify failures (not implemented yet)
  - Ensure failures are expectation failures, not compilation errors
- Implement `CreateUser` in `users_commands.go`:
  - Validate input (name not empty, email format with simple regex)
  - Check email uniqueness via repository.GetUserByEmail (if not sql.ErrNoRows, return ErrUserEmailConflict)
  - Generate UUID using uuid.NewV4()
  - Create User entity with generated ID
  - Call repository.CreateUser
  - Return CreateUserResponse with user ID
- Run affected tests: `go test -v ./internal/app/ --run TestUserCommands/CreateUser`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-4.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 4.3: Implement and test UpdateUser command**
- Add `NewRandomUpdateUserRequest` to `users_testing.go`
- In `users_commands_test.go`, add test `TestUserCommands/UpdateUser`:
  - Happy path: user is updated successfully
  - Validation: empty name returns ErrInvalidInput
  - Validation: empty email returns ErrInvalidInput
  - Validation: invalid email format returns ErrInvalidInput
  - Not found: non-existent user returns ErrUserNotFound
  - Conflict: duplicate email returns ErrUserEmailConflict
  - Same email: updating user with their own email works (no conflict)
- Run affected tests: `go test -v ./internal/app/ --run TestUserCommands/UpdateUser`
  - Verify failures (not implemented yet)
- Implement `UpdateUser` in `users_commands.go`:
  - Validate input (name, email format)
  - Get existing user via repository.GetUserByID (if sql.ErrNoRows, return ErrUserNotFound)
  - If email changed, check uniqueness via repository.GetUserByEmail
  - Create updated User entity with new values
  - Call repository.UpdateUser
  - Handle unique constraint violations → ErrUserEmailConflict
- Run affected tests: `go test -v ./internal/app/ --run TestUserCommands/UpdateUser`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-4.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 4.4: Implement and test DeleteUser command**
- In `users_commands_test.go`, add test `TestUserCommands/DeleteUser`:
  - Happy path: user is deleted successfully
  - Not found: non-existent user returns ErrUserNotFound
- Run affected tests: `go test -v ./internal/app/ --run TestUserCommands/DeleteUser`
  - Verify failures (not implemented yet)
- Implement `DeleteUser` in `users_commands.go`:
  - Check user exists via repository.GetUserByID (if sql.ErrNoRows, return ErrUserNotFound)
  - Call repository.DeleteUser (CASCADE will delete pet relationships)
- Run affected tests: `go test -v ./internal/app/ --run TestUserCommands/DeleteUser`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-4.4.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 4.5: Register UserCommands in DI**
- Update `internal/app/register.go`:
  - Add `NewUserCommands` to providers
  - DI should provide `*UserCommands` (concrete struct) to consumers
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-4.5.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 5: Application Layer - Pets Commands

**Task 5.1: Create PetsCommands structure**
- Create `internal/app/pets_commands.go`
- Define request/response types: `AddPetRequest`, `AddPetResponse`
- Define error types: `ErrPetCreationFailed`, `ErrUserPetNotFound`
- Define `PetsCommands` struct with PetsRepository, UsersRepository, and PetstoreClient interface dependencies
- Add `PetsCommandsDeps` struct for DI
- Add constructor `NewPetsCommands(deps PetsCommandsDeps) *PetsCommands`
- Add stub implementations for AddPet, RemovePet (return `errors.New("not implemented")`)
- Create `internal/app/pets_commands_test.go` with test structure
- Add `makeMockDeps` function
- Run: `make test`
  - Verify code compiles
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-5.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 5.2: Implement and test AddPet command**
- Create `internal/app/pets_testing.go` with factory `NewRandomAddPetRequest(fake *faker.Faker) *AddPetRequest`
- In `pets_commands_test.go`, add test `TestPetsCommands/AddPet`:
  - Happy path: pet is created in petstore, relationship added, returns pet ID
  - Validation: empty name returns ErrInvalidInput
  - User not found: non-existent user returns ErrUserNotFound
  - Petstore failure: petstore creation fails returns ErrPetCreationFailed
- Run affected tests: `go test -v ./internal/app/ --run TestPetsCommands/AddPet`
  - Verify failures (not implemented yet)
- Implement `AddPet` in `pets_commands.go`:
  - Validate input (name not empty)
  - Check user exists via usersRepository.GetUserByID (if sql.ErrNoRows, return ErrUserNotFound)
  - Create petstore.AddPetRequest from AddPetRequest
  - Create pet in Petstore via petstoreClient.AddPet (wrap errors as ErrPetCreationFailed)
  - Add relationship via petsRepository.AddUserPet using petstore response ID
  - Return AddPetResponse with pet ID
- Run affected tests: `go test -v ./internal/app/ --run TestPetsCommands/AddPet`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-5.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 5.3: Implement and test RemovePet command**
- In `pets_commands_test.go`, add test `TestPetsCommands/RemovePet`:
  - Happy path: relationship is removed
  - User not found: non-existent user returns ErrUserNotFound
  - Relationship not found: non-existent relationship returns ErrUserPetNotFound
- Run affected tests: `go test -v ./internal/app/ --run TestPetsCommands/RemovePet`
  - Verify failures (not implemented yet)
- Implement `RemovePet` in `pets_commands.go`:
  - Check user exists via usersRepository.GetUserByID (if sql.ErrNoRows, return ErrUserNotFound)
  - Check relationship exists via petsRepository.HasUserPet (if false, return ErrUserPetNotFound)
  - Call petsRepository.RemoveUserPet
- Run affected tests: `go test -v ./internal/app/ --run TestPetsCommands/RemovePet`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-5.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 5.4: Register PetsCommands in DI**
- Update `internal/app/register.go`:
  - Add `NewPetsCommands` to providers
  - DI should provide `*PetsCommands` (concrete struct) to consumers
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-5.4.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 6: Application Layer - User Queries

**Task 6.1: Create UserQueries structure**
- Create `internal/app/users_queries.go`
- Define `UserQueries` struct with UsersRepository dependency
- Add `UserQueriesDeps` struct for DI
- Add constructor `NewUserQueries(deps UserQueriesDeps) *UserQueries`
- Add stub implementations for GetUserByID, ListUsers (return `errors.New("not implemented")`)
- Create `internal/app/users_queries_test.go` with test structure
- Add `makeMockDeps` function
- Run: `make test`
  - Verify code compiles
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-6.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 6.2: Implement and test GetUserByID query**
- In `users_queries_test.go`, add test `TestUserQueries/GetUserByID`:
  - Happy path: returns user data with all fields
  - Not found: returns ErrUserNotFound when user doesn't exist
- Run affected tests: `go test -v ./internal/app/ --run TestUserQueries/GetUserByID`
  - Verify failures (not implemented yet)
- Implement `GetUserByID` in `users_queries.go`:
  - Call repository.GetUserByID
  - If sql.ErrNoRows, return ErrUserNotFound
  - Return user entity
- Run affected tests: `go test -v ./internal/app/ --run TestUserQueries/GetUserByID`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-6.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 6.3: Implement and test ListUsers query**
- In `users_queries_test.go`, add test `TestUserQueries/ListUsers`:
  - Happy path: returns list of users with all data
  - Empty: returns empty slice when no users exist
- Run affected tests: `go test -v ./internal/app/ --run TestUserQueries/ListUsers`
  - Verify failures (not implemented yet)
- Implement `ListUsers` in `users_queries.go`:
  - Call repository.ListUsers
  - Return list of user entities
- Run affected tests: `go test -v ./internal/app/ --run TestUserQueries/ListUsers`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-6.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 6.4: Register UserQueries in DI**
- Update `internal/app/register.go`:
  - Add `NewUserQueries` to providers
  - DI should provide `*UserQueries` (concrete struct) to consumers
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-6.4.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 7: Application Layer - Pets Queries

**Task 7.1: Create PetsQueries structure**
- Create `internal/app/pets_queries.go`
- Define `PetsQueries` struct with PetsRepository, UsersRepository, and PetstoreClient interface dependencies
- Add `PetsQueriesDeps` struct for DI
- Add constructor `NewPetsQueries(deps PetsQueriesDeps) *PetsQueries`
- Add stub implementation for ListUserPets (return `errors.New("not implemented")`)
- Create `internal/app/pets_queries_test.go` with test structure
- Add `makeMockDeps` function
- Run: `make test`
  - Verify code compiles
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-7.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 7.2: Implement and test ListUserPets query**
- In `pets_queries_test.go`, add test `TestPetsQueries/ListUserPets`:
  - Happy path: returns list of pets from petstore
  - User not found: non-existent user returns ErrUserNotFound
  - Empty: returns empty slice when user has no pets
  - Missing pets: gracefully handles when pet doesn't exist in petstore (skip it, log warning)
- Run affected tests: `go test -v ./internal/app/ --run TestPetsQueries/ListUserPets`
  - Verify failures (not implemented yet)
- Implement `ListUserPets` in `pets_queries.go`:
  - Verify user exists via usersRepository.GetUserByID (if sql.ErrNoRows, return ErrUserNotFound)
  - Get pet IDs via petsRepository.GetUserPetIDs
  - For each pet ID, fetch details from petstore via petstoreClient.GetPetByID
  - If pet not found in petstore, log warning and skip it
  - Return slice of petstore.Pet objects
- Run affected tests: `go test -v ./internal/app/ --run TestPetsQueries/ListUserPets`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-7.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 7.3: Register PetsQueries in DI**
- Update `internal/app/register.go`:
  - Add `NewPetsQueries` to providers
  - DI should provide `*PetsQueries` (concrete struct) to consumers
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-7.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 8: HTTP Controllers - Users

**Task 8.1: Create UsersController structure**
- Create `internal/api/http/v1controllers/users.go`
- Define `UsersController` struct with UserCommands and UserQueries dependencies (concrete structs)
- Add constructor `newUsersController(commands *app.UserCommands, queries *app.UserQueries) *UsersController`
- Add stub implementations for all handlers: CreateUser, GetUserById, UpdateUser, DeleteUser, ListUsers
- Ensure controller implements the generated handler interface
- Create `internal/api/http/v1controllers/users_test.go` with test structure
- Run: `make test`
  - Verify code compiles
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-8.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 8.2: Implement and test CreateUser endpoint**
- In `users_test.go`, add test `TestUsersController/CreateUser`:
  - Happy path: returns 201 with userId
  - Validation error: returns 400 for invalid input (ErrInvalidInput)
  - Conflict: returns 409 for duplicate email (ErrUserEmailConflict)
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/CreateUser`
  - Verify failures (not implemented yet)
- Implement `CreateUser` handler in `users.go`:
  - Create transformer to map API CreateUserRequest to app.CreateUserRequest
  - Call commands.CreateUser
  - Map errors: ErrUserEmailConflict → 409, ErrInvalidInput → 400
  - Transform response to API CreateUserResponse
  - Follow HandlerBuilder pattern from existing examples
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/CreateUser`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-8.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 8.3: Implement and test UpdateUser endpoint**
- In `users_test.go`, add test `TestUsersController/UpdateUser`:
  - Happy path: returns 204
  - Validation error: returns 400 (ErrInvalidInput)
  - Not found: returns 404 (ErrUserNotFound)
  - Conflict: returns 409 (ErrUserEmailConflict)
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/UpdateUser`
  - Verify failures (not implemented yet)
- Implement `UpdateUser` handler in `users.go`:
  - Transform API request to app.UpdateUserRequest (include userId from path)
  - Call commands.UpdateUser
  - Map errors appropriately
  - Return 204 on success
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/UpdateUser`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-8.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 8.4: Implement and test DeleteUser endpoint**
- In `users_test.go`, add test `TestUsersController/DeleteUser`:
  - Happy path: returns 204
  - Not found: returns 404 (ErrUserNotFound)
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/DeleteUser`
  - Verify failures (not implemented yet)
- Implement `DeleteUser` handler in `users.go`:
  - Call commands.DeleteUser with userId from path
  - Map errors: ErrUserNotFound → 404
  - Return 204 on success
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/DeleteUser`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-8.4.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 8.5: Implement and test GetUserById endpoint**
- In `users_test.go`, add test `TestUsersController/GetUserById`:
  - Happy path: returns 200 with user data (id, name, email only, no timestamps in API response)
  - Not found: returns 404 (ErrUserNotFound)
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/GetUserById`
  - Verify failures (not implemented yet)
- Implement `GetUserById` handler in `users.go`:
  - Call queries.GetUserByID with userId from path
  - Map errors: ErrUserNotFound → 404
  - Transform app.User to API UserResponse (map id, name, email only)
  - Return 200 with response
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/GetUserById`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-8.5.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 8.6: Implement and test ListUsers endpoint**
- In `users_test.go`, add test `TestUsersController/ListUsers`:
  - Happy path: returns 200 with list of users
  - Empty: returns 200 with empty array
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/ListUsers`
  - Verify failures (not implemented yet)
- Implement `ListUsers` handler in `users.go`:
  - Call queries.ListUsers
  - Transform each app.User to API UserResponse
  - Return 200 with ListUsersResponse
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/ListUsers`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-8.6.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 8.7: Register UsersController**
- Update `internal/api/http/v1controllers/register.go`:
  - Add `newUsersController` to providers
- Update `internal/api/http/v1routes.go`:
  - Add `*v1controllers.UsersController` to V1RoutesDeps
  - Call appropriate registration method to wire up routes
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
  - Verify routes are registered (check logs)
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-8.7.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 9: HTTP Controllers - Pets

**Task 9.1: Create PetsController structure**
- Create `internal/api/http/v1controllers/pets.go`
- Define `PetsController` struct with PetsCommands and PetsQueries dependencies (concrete structs)
- Add constructor `newPetsController(commands *app.PetsCommands, queries *app.PetsQueries) *PetsController`
- Add stub implementations for all handlers: AddUserPet, RemoveUserPet, ListUserPets
- Ensure controller implements the generated handler interface
- Create `internal/api/http/v1controllers/pets_test.go` with test structure
- Run: `make test`
  - Verify code compiles
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-9.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 9.2: Implement and test AddUserPet endpoint**
- In `pets_test.go`, add test `TestPetsController/AddUserPet`:
  - Happy path: returns 201 with petId
  - Validation error: returns 400 (ErrInvalidInput)
  - User not found: returns 404 (ErrUserNotFound)
  - Petstore failure: returns appropriate error (ErrPetCreationFailed)
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/AddUserPet`
  - Verify failures (not implemented yet)
- Implement `AddUserPet` handler in `pets.go`:
  - Transform API AddPetRequest to app.AddPetRequest (include userId from path)
  - Call commands.AddPet
  - Map errors: ErrUserNotFound → 404, ErrInvalidInput → 400, ErrPetCreationFailed → 502
  - Transform response to API AddPetResponse
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/AddUserPet`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-9.2.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 9.3: Implement and test RemoveUserPet endpoint**
- In `pets_test.go`, add test `TestPetsController/RemoveUserPet`:
  - Happy path: returns 204
  - User not found: returns 404 (ErrUserNotFound)
  - Relationship not found: returns 404 (ErrUserPetNotFound)
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/RemoveUserPet`
  - Verify failures (not implemented yet)
- Implement `RemoveUserPet` handler in `pets.go`:
  - Extract userId and petId from path parameters
  - Call commands.RemovePet
  - Map errors: ErrUserNotFound → 404, ErrUserPetNotFound → 404
  - Return 204 on success
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/RemoveUserPet`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-9.3.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 9.4: Implement and test ListUserPets endpoint**
- In `pets_test.go`, add test `TestPetsController/ListUserPets`:
  - Happy path: returns 200 with list of pets
  - User not found: returns 404 (ErrUserNotFound)
  - Empty: returns 200 with empty array
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/ListUserPets`
  - Verify failures (not implemented yet)
- Implement `ListUserPets` handler in `pets.go`:
  - Extract userId from path
  - Call queries.ListUserPets
  - Map errors: ErrUserNotFound → 404
  - Transform each petstore.Pet to API PetResponse
  - Return 200 with ListUserPetsResponse
- Run affected tests: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/ListUserPets`
  - Verify all tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-9.4.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 9.5: Register PetsController**
- Update `internal/api/http/v1controllers/register.go`:
  - Add `newPetsController` to providers
- Update `internal/api/http/v1routes.go`:
  - Add `*v1controllers.PetsController` to V1RoutesDeps
  - Call appropriate registration method to wire up routes
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
  - Verify pet routes are registered (check logs)
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-9.5.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes, server starts successfully

### Phase 10: Integration & Final Verification

**Task 10.1: Add comprehensive integration tests**
- In `users_test.go` and `pets_test.go`, add end-to-end integration tests:
  - Full user lifecycle: create → get → update → list → delete
  - Pet relationships: create user → add pet → list pets → remove pet → delete user
  - Multiple pets: create user → add multiple pets → verify all returned
  - Cascade delete: create user → add pets → delete user → verify relationships deleted
  - Error scenarios: duplicate email conflict, non-existent user/pet operations
- Run: `go test -v ./internal/api/http/v1controllers/`
  - Verify all integration tests pass
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-10.1.md`
- Success criteria: As per completion protocol: `make test` passes, `make lint` passes

**Task 10.2: Final verification and documentation**
- Run: `make lint`
  - Verify no linting errors
  - Fix any issues found
- Run: `make test`
  - Verify all tests pass
  - Verify coverage meets threshold (>90%)
- Update `.mockery.yaml` if needed to include new interfaces:
  - `app.UsersRepository`
  - `app.PetsRepository`
  - `app.PetstoreClient`
- Run: `go run ./cmd/server start --env local --noop`
  - Verify server starts successfully
  - Verify all routes registered
  - Verify database initialization
- Write summary to `doc/implementation/plan-users-pets-app/summary-task-10.2.md`
- Success criteria: As per completion protocol: `make lint` passes, `make test` passes with >90% coverage, server starts successfully
- In `pets_test.go`, add test `TestPetsController/AddUserPet`:
  - Happy path: returns 201 with petId
  - Validation error: returns 400 (ErrInvalidInput)
  - User not found: