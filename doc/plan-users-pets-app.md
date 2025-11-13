# Plan: Users & Pets App Implementation

## 1. Introduction/Overview

This plan outlines the implementation of a simple users and pets management application that demonstrates the architecture patterns of the golang-backend-boilerplate project. The application will follow CQRS principles with Commands handling data mutations and Queries handling data retrieval operations.

### Goal
Build a minimal yet complete example application that:
- Manages user records with CRUD operations via UserCommands
- Creates pets in external Petstore API and manages user-pet relationships via PetsCommands
- Provides data retrieval via separate UserQueries and PetsQueries
- Uses SQLite as the underlying storage with a pure Go driver (no cgo)
- Follows the project's hexagonal architecture and testing best practices

### Problem It Solves
This app serves as a canonical example demonstrating:
- How to implement CQRS pattern in the application layer with clear separation
- How to create and integrate a database repository in the services layer
- How to integrate with external HTTP APIs (petstore)
- How to wire HTTP endpoints to application logic
- How to write comprehensive tests following TDD approach
- API-driven design where models are generated from OpenAPI spec

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
  - Return user ID, name, and email only (no pet data)
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

### Layer 2: Application Layer (Business Logic - CQRS)
- Location: `internal/app`
- **UserCommands**: `users_commands.go` - User CRUD operations
- **PetsCommands**: `pets_commands.go` - Pet creation and relationship management
- **UserQueries**: `users_queries.go` - User data retrieval
- **PetsQueries**: `pets_queries.go` - Pet data retrieval
- Domain models and errors in the same files

### Layer 3: Outgoing Adapters (Services)
- Location: `internal/infrastructure`
- **UsersRepository**: `users_repository.go` - User CRUD in SQLite
- **PetsRepository**: `pets_repository.go` - User-pet relationships in SQLite
- **Petstore Client**: Already exists in `internal/infrastructure/petstore` - External API integration
- Uses `modernc.org/sqlite` (pure Go, no cgo)

### Data Flow
```
HTTP Request → Controller → Command/Query → Repository + Petstore → SQLite + External API
                    ↓
HTTP Response ← Transform ← Result ← Repository + Petstore ← SQLite + External API
```

## 4. Detailed Architecture

### 4.1 Database Schema (SQLite)

**Users Table**
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**User Pets Table** (relationship only, pet data comes from Petstore API)
```sql
CREATE TABLE user_pets (
    user_id TEXT NOT NULL,
    pet_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, pet_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 4.2 Services Layer

**Files to Create:**
- `internal/infrastructure/users_repository.go` - User repository interface and implementation
- `internal/infrastructure/users_repository_test.go` - User repository tests
- `internal/infrastructure/pets_repository.go` - Pets repository interface and implementation
- `internal/infrastructure/pets_repository_test.go` - Pets repository tests
- `internal/infrastructure/users_testing.go` - Test helpers and factories for users
- `internal/infrastructure/pets_testing.go` - Test helpers and factories for pets

**UsersRepository Interface:**
```go
type UsersRepository interface {
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, userID string) error
    GetUserByID(ctx context.Context, userID string) (*User, error)
    GetUserByEmail(ctx context.Context, email string) (*User, error)
    ListUsers(ctx context.Context) ([]*User, error)
}

type User struct {
    ID        string
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**PetsRepository Interface:**
```go
type PetsRepository interface {
    AddUserPet(ctx context.Context, userID string, petID int64) error
    RemoveUserPet(ctx context.Context, userID string, petID int64) error
    GetUserPetIDs(ctx context.Context, userID string) ([]int64, error)
    HasUserPet(ctx context.Context, userID string, petID int64) (bool, error)
}

type UserPet struct {
    UserID    string
    PetID     int64
    CreatedAt time.Time
}
```

**Implementation Details:**
- Use `modernc.org/sqlite` driver with `database/sql`
- Initialize database with schema on first connection
- Use database native TIMESTAMP type with timezone support
- Handle unique constraint violations for email
- Generate UUIDs for user IDs using `github.com/gofrs/uuid/v5`

### 4.3 Application Layer

**Files to Create:**
- `internal/app/users_commands.go` - User commands implementation with models and errors
- `internal/app/users_commands_test.go` - User commands tests
- `internal/app/pets_commands.go` - Pets commands implementation with models and errors
- `internal/app/pets_commands_test.go` - Pets commands tests
- `internal/app/users_queries.go` - User queries implementation with models
- `internal/app/users_queries_test.go` - User queries tests
- `internal/app/pets_queries.go` - Pets queries implementation with models
- `internal/app/pets_queries_test.go` - Pets queries tests
- `internal/app/users_testing.go` - Test helpers
- `internal/app/pets_testing.go` - Test helpers

**UserCommands (in `users_commands.go`):**
```go
// Service interface
type UserCommands interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (string, error) // returns userID
    UpdateUser(ctx context.Context, req *UpdateUserRequest) error
    DeleteUser(ctx context.Context, userID string) error
}

// Request models
type CreateUserRequest struct {
    Name  string
    Email string
}

type UpdateUserRequest struct {
    UserID string
    Name   string
    Email  string
}

// Domain models
type User struct {
    ID    string
    Name  string
    Email string
}

// Error types
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrUserEmailConflict = errors.New("user with this email already exists")
    ErrInvalidInput      = errors.New("invalid input")
)
```

**PetsCommands (in `pets_commands.go`):**
```go
// Service interface
type PetsCommands interface {
    AddPet(ctx context.Context, req *AddPetRequest) (int64, error) // returns petID
    RemovePet(ctx context.Context, userID string, petID int64) error
}

// Request models
type AddPetRequest struct {
    UserID    string
    Name      string
    Status    string
    PhotoUrls []string
}

// Error types
var (
    ErrPetCreationFailed = errors.New("failed to create pet in petstore")
    ErrUserPetNotFound   = errors.New("user-pet relationship not found")
)
```

**UserQueries (in `users_queries.go`):**
```go
// Service interface
type UserQueries interface {
    GetUserByID(ctx context.Context, userID string) (*User, error)
    ListUsers(ctx context.Context) ([]*User, error)
}

// Domain models (same as in commands)
type User struct {
    ID    string
    Name  string
    Email string
}
```

**PetsQueries (in `pets_queries.go`):**
```go
// Service interface
type PetsQueries interface {
    ListUserPets(ctx context.Context, userID string) ([]*Pet, error)
}

// Domain models
type Pet struct {
    ID        int64
    Name      string
    Status    string
    PhotoUrls []string
}
```

**Business Logic:**

In UserCommands:
- Validate input (non-empty name, valid email format)
- Check for email uniqueness before create/update
- Generate UUIDs for new users

In PetsCommands:
- Validate input (non-empty pet name)
- Verify user exists before adding pet
- Create pet in Petstore API first
- Then create relationship in local database
- For remove: verify relationship exists

In UserQueries:
- Simple pass-through to repository

In PetsQueries:
- Fetch pet IDs from repository
- Fetch pet details from Petstore API
- Handle gracefully when pets exist in DB but not in Petstore

### 4.4 HTTP API Layer

**Files to Update:**
- `internal/api/http/v1routes.yaml` - Add OpenAPI definitions for users and pets endpoints

**Files to Create:**
- `internal/api/http/v1controllers/users.go` - Users controller
- `internal/api/http/v1controllers/users_test.go` - Users controller tests
- `internal/api/http/v1controllers/pets.go` - Pets controller
- `internal/api/http/v1controllers/pets_test.go` - Pets controller tests

**OpenAPI Endpoints to Add:**

```yaml
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
```

**Request/Response Schemas:**
```yaml
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
      required: [id, name, status]
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

**Controller Implementation:**
- Follow the pattern from `echo.go`
- Use transformer pattern to map between API models and domain models
- Handle errors and map to appropriate HTTP status codes
- Return proper error responses using the Error schema

**Files to Update for Registration:**
- `internal/infrastructure/register.go` - Register UsersRepository, PetsRepository, and DB connection
- `internal/app/register.go` - Register UserCommands, PetsCommands, UserQueries, PetsQueries
- `internal/api/http/v1controllers/register.go` - Register UsersController and PetsController
- `internal/api/http/v1routes.go` - Wire up users and pets routes

### 4.5 Configuration

**Files to Update:**
- `internal/config/default.json` - Add database configuration
```json
{
  "database": {
    "path": "./data/app.db"
  }
}
```

## 5. Key Architectural Decisions

### 5.1 CQRS Separation with Split Services
**Decision**: Separate Commands and Queries into distinct services, and further split by domain (Users vs Pets).

**Rationale**: 
- Clear separation of concerns between reads and writes
- Clear separation between user operations and pet operations
- Easier to reason about side effects
- Follows single responsibility principle
- Each service has focused, cohesive responsibilities

### 5.2 Pure Go SQLite Driver
**Decision**: Use `modernc.org/sqlite` instead of cgo-based drivers.

**Rationale**:
- No cgo dependency simplifies builds and cross-compilation
- Better portability across platforms
- Simpler deployment (no need for sqlite3 C library)
- Sufficient performance for this use case

### 5.3 Split Repository Pattern
**Decision**: Separate UsersRepository and PetsRepository in services layer.

**Rationale**:
- Each repository handles its own domain
- UsersRepository: user CRUD operations
- PetsRepository: user-pet relationship operations
- Clear separation of data access concerns
- Easier to test and maintain

### 5.4 Pet Creation in Petstore API
**Decision**: Create pets in Petstore API first, then store relationship locally.

**Rationale**:
- Demonstrates realistic integration with external system
- External system is source of truth for pet data
- Local database only maintains relationships
- If Petstore creation fails, no local state is created

### 5.5 Lean Mutations
**Decision**: Mutating endpoints return minimal data (just ID or void).

**Rationale**:
- Keeps commands lean and focused on writes
- Clients can GET full data if needed
- Reduces response payload size
- Follows command pattern principles
- Clear separation between write and read operations

### 5.6 No Pagination
**Decision**: List endpoints return all data without pagination.

**Rationale**:
- Simplifies implementation for example app
- Reduces complexity in initial version
- Pagination can be added later if needed
- Focus on core CQRS patterns

### 5.7 Database Native Timestamps
**Decision**: Use database native TIMESTAMP type instead of int64 Unix timestamps.

**Rationale**:
- More readable in database queries and tools
- Database handles timezone correctly
- Standard SQL type
- Better integration with database features
- Go's `time.Time` maps naturally to SQL TIMESTAMP

### 5.8 API-Driven Design
**Decision**: Define OpenAPI specification first, generate models from it.

**Rationale**:
- API contract is the source of truth
- Models are automatically generated and consistent
- Forces thinking about API design upfront
- Ensures API documentation is always up-to-date
- Follows API-first development approach

### 5.9 Cascade Delete for User-Pet Relationships
**Decision**: Use database CASCADE on foreign key to auto-delete relationships when user is deleted.

**Rationale**:
- Ensures data consistency
- Simplifies application code
- Leverages database capabilities
- Prevents orphaned relationship records
- Does NOT delete pets from Petstore (only relationships)

### 5.10 Flat File Structure
**Decision**: Keep files at package level without nested folders.

**Rationale**:
- Simpler structure for this example
- Easier navigation with fewer directories
- Models and logic can be in same files
- Consistent with keeping things "plain" for now

## 6. Uncertainties

The plan is comprehensive based on the existing architecture. Minor details that may need clarification during implementation:

1. **Email validation**: Use simple regex or more comprehensive validation? (Assumed simple for MVP)
2. **Missing Pets**: How to handle when a pet ID exists in our DB but not in Petstore? (Assumed: skip/filter out with warning log)
3. **Pet status validation**: Should we validate status enum at app layer or rely on Petstore? (Assumed: validate at app layer)
4. **Duplicate pet add**: Should adding same pet twice be idempotent or error? (Assumed: idempotent/ignore)
5. **Database location**: Database path configurable via config system (using existing pattern)

## 7. Related Files

### Files to Create (New)

**Services Layer:**
- `internal/infrastructure/users_repository.go` - Users repository interface and implementation
- `internal/infrastructure/users_repository_test.go` - Users repository tests
- `internal/infrastructure/pets_repository.go` - Pets repository interface and implementation
- `internal/infrastructure/pets_repository_test.go` - Pets repository tests
- `internal/infrastructure/users_testing.go` - Test helpers for users
- `internal/infrastructure/pets_testing.go` - Test helpers for pets

**Application Layer:**
- `internal/app/users_commands.go` - User commands implementation with models and errors
- `internal/app/users_commands_test.go` - User commands tests
- `internal/app/pets_commands.go` - Pets commands implementation with models and errors
- `internal/app/pets_commands_test.go` - Pets commands tests
- `internal/app/users_queries.go` - User queries implementation with models
- `internal/app/users_queries_test.go` - User queries tests
- `internal/app/pets_queries.go` - Pets queries implementation with models
- `internal/app/pets_queries_test.go` - Pets queries tests
- `internal/app/users_testing.go` - Test helpers for users
- `internal/app/pets_testing.go` - Test helpers for pets

**HTTP Controllers:**
- `internal/api/http/v1controllers/users.go` - Users controller
- `internal/api/http/v1controllers/users_test.go` - Users controller tests
- `internal/api/http/v1controllers/pets.go` - Pets controller
- `internal/api/http/v1controllers/pets_test.go` - Pets controller tests

### Files to Update (Existing)

- `internal/api/http/v1routes.yaml` - Add users and pets endpoints and schemas
- `internal/api/http/v1routes.go` - Register users and pets routes
- `internal/api/http/v1controllers/register.go` - Register UsersController and PetsController
- `internal/app/register.go` - Register all commands and queries services
- `internal/infrastructure/register.go` - Register repositories and DB connection
- `internal/config/default.json` - Add database configuration
- `go.mod` - Add `modernc.org/sqlite` dependency
- `.mockery.yaml` - Add interfaces for mocking

### Generated Files (by apigen)

After updating `v1routes.yaml` and running `go generate ./internal/api/http/v1routes.go`:
- `internal/api/http/v1routes/models/create_user_request.go`
- `internal/api/http/v1routes/models/create_user_response.go`
- `internal/api/http/v1routes/models/update_user_request.go`
- `internal/api/http/v1routes/models/user_response.go`
- `internal/api/http/v1routes/models/add_pet_request.go`
- `internal/api/http/v1routes/models/add_pet_response.go`
- `internal/api/http/v1routes/models/pet_response.go`
- `internal/api/http/v1routes/models/list_users_response.go`
- `internal/api/http/v1routes/models/list_user_pets_response.go`
- `internal/api/http/v1routes/handlers/*_users.go` (various generated handler files)
- `internal/api/http/v1routes/handlers/*_pets.go` (various generated handler files)

## 8. Task List

This implementation will follow TDD approach as per [tdd-flow.md](../.context/tdd-flow.md). Each task should leave the codebase in a buildable state with all tests passing (`make test`).

### Phase 1: Foundation & API Definition

**Task 1.1: Define OpenAPI specification**
- Update `internal/api/http/v1routes.yaml`:
  - Add all users endpoints paths (POST/GET /users, GET/PUT/DELETE /users/{userId})
  - Add all pets endpoints paths (GET/POST /users/{userId}/pets, DELETE /users/{userId}/pets/{petId})
  - Add all request/response schemas (see section 4.4)
  - Add proper error responses
  - Follow existing patterns from echo endpoints
- Run: `make lint` to verify YAML is valid
- Success criteria: `make lint` passes
- Implementation status: NOT STARTED

**Task 1.2: Generate API models and handlers**
- Run: `go generate ./internal/api/http/v1routes.go` to generate code
- Verify generated files in `internal/api/http/v1routes/models/` and `internal/api/http/v1routes/handlers/`
- Fix any compilation errors in generated code
- Run: `make test` to verify no breaking changes
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 1.3: Add SQLite dependency and database configuration**
- Add `modernc.org/sqlite` to `go.mod`: `go get modernc.org/sqlite`
- Update `internal/config/default.json` to add database configuration with path "./data/app.db"
- Run: `make test` to verify no breaking changes
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

### Phase 2: Database Layer - Users Repository

**Task 2.1: Create UsersRepository interface and schema**
- Create `internal/infrastructure/users_repository.go`
- Define `User` struct with ID, Name, Email, CreatedAt, UpdatedAt fields (using time.Time)
- Define `UsersRepository` interface with all methods
- Create `sqliteUsersRepository` struct implementing the interface
- Add constructor `NewUsersRepository(db *sql.DB) UsersRepository`
- Implement schema initialization (CREATE TABLE IF NOT EXISTS for users)
- Add stub implementations for all interface methods (return `errors.New("not implemented")`)
- Create `internal/infrastructure/users_repository_test.go` with basic test structure
- Write test for schema initialization (verify table is created)
- Run: `go test -v ./internal/infrastructure/ --run TestUsersRepository`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 2.2: Implement and test CreateUser**
- Create `internal/infrastructure/users_testing.go` with factory: `NewRandomUser(fake *faker.Faker, opts ...RandomUserOpt)`
- In `users_repository_test.go`, write tests for CreateUser:
  - Happy path: user is created and can be retrieved
  - Error case: duplicate email returns unique constraint error
  - Edge case: timestamps are set correctly
- Run: `go test -v ./internal/infrastructure/ --run TestUsersRepository/CreateUser`
  - Verify failures are expected (not implemented yet)
- Implement `CreateUser` in `users_repository.go`:
  - Generate UUID if ID is empty
  - Set timestamps (created_at, updated_at) to current time
  - Execute INSERT statement
  - Handle unique constraint violation for email
- Run: `go test -v ./internal/infrastructure/ --run TestUsersRepository/CreateUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 2.3: Implement and test GetUserByID and GetUserByEmail**
- In `users_repository_test.go`, write tests for GetUserByID:
  - Happy path: existing user is retrieved correctly with all fields
  - Error case: non-existent user returns error (sql.ErrNoRows)
- Write tests for GetUserByEmail:
  - Happy path: user found by email
  - Error case: non-existent email returns error
- Run tests to verify failures
- Implement `GetUserByID` and `GetUserByEmail` in `users_repository.go`
- Run: `go test -v ./internal/infrastructure/ --run "TestUsersRepository/(GetUserByID|GetUserByEmail)"`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 2.4: Implement and test UpdateUser**
- In `users_repository_test.go`, write tests for UpdateUser:
  - Happy path: user fields are updated correctly
  - Updated timestamp is changed, created timestamp stays same
  - Error case: non-existent user returns error
  - Error case: email conflict with another user returns unique constraint error
- Run tests to verify failures
- Implement `UpdateUser` in `users_repository.go`:
  - Update name and email
  - Update updated_at timestamp to current time
  - Handle unique constraint violation for email
  - Verify user exists (check affected rows)
- Run: `go test -v ./internal/infrastructure/ --run TestUsersRepository/UpdateUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 2.5: Implement and test ListUsers**
- In `users_repository_test.go`, write tests for ListUsers:
  - Happy path: returns all users
  - Empty result: returns empty slice when no users
  - Order: users returned in consistent order (e.g., by created_at)
- Run tests to verify failures
- Implement `ListUsers` in `users_repository.go`
- Run: `go test -v ./internal/infrastructure/ --run TestUsersRepository/ListUsers`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 2.6: Implement and test DeleteUser**
- In `users_repository_test.go`, write tests for DeleteUser:
  - Happy path: user is deleted successfully
  - Error case: non-existent user returns error
- Run tests to verify failures
- Implement `DeleteUser` in `users_repository.go`
  - Verify user exists before delete (or check affected rows)
- Run: `go test -v ./internal/infrastructure/ --run TestUsersRepository/DeleteUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

### Phase 3: Database Layer - Pets Repository

**Task 3.1: Create PetsRepository interface and schema**
- Create `internal/infrastructure/pets_repository.go`
- Define `UserPet` struct with UserID, PetID, CreatedAt fields (using time.Time)
- Define `PetsRepository` interface with all methods
- Create `sqlitePetsRepository` struct implementing the interface
- Add constructor `NewPetsRepository(db *sql.DB) PetsRepository`
- Implement schema initialization (CREATE TABLE IF NOT EXISTS for user_pets)
- Add stub implementations for all interface methods
- Create `internal/infrastructure/pets_repository_test.go` with basic test structure
- Write test for schema initialization
- Run: `go test -v ./internal/infrastructure/ --run TestPetsRepository`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 3.2: Implement and test AddUserPet**
- Create `internal/infrastructure/pets_testing.go` with factory: `NewRandomUserPet(fake *faker.Faker, opts ...RandomUserPetOpt)`
- In `pets_repository_test.go`, write tests for AddUserPet:
  - Happy path: relationship is created
  - Error case: non-existent user returns foreign key error
  - Idempotent: adding same pet twice succeeds (use INSERT OR IGNORE or check duplicate)
- Run: `go test -v ./internal/infrastructure/ --run TestPetsRepository/AddUserPet`
  - Verify failures are expected
- Implement `AddUserPet` in `pets_repository.go`:
  - Set timestamp to current time
  - Use INSERT OR IGNORE for idempotency
- Run: `go test -v ./internal/infrastructure/ --run TestPetsRepository/AddUserPet`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 3.3: Implement and test GetUserPetIDs and HasUserPet**
- In `pets_repository_test.go`, write tests for GetUserPetIDs:
  - Happy path: returns all pet IDs for user
  - Empty result: returns empty slice when user has no pets
  - Order: pet IDs returned in consistent order
- Write tests for HasUserPet:
  - Returns true when relationship exists
  - Returns false when relationship doesn't exist
- Run tests to verify failures
- Implement `GetUserPetIDs` and `HasUserPet` in `pets_repository.go`
- Run: `go test -v ./internal/infrastructure/ --run "TestPetsRepository/(GetUserPetIDs|HasUserPet)"`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 3.4: Implement and test RemoveUserPet**
- In `pets_repository_test.go`, write tests for RemoveUserPet:
  - Happy path: relationship is removed
  - Idempotent: removing non-existent relationship succeeds (or is idempotent)
- Run tests to verify failures
- Implement `RemoveUserPet` in `pets_repository.go`
- Run: `go test -v ./internal/infrastructure/ --run TestPetsRepository/RemoveUserPet`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 3.5: Test cascade delete**
- In `pets_repository_test.go`, write integration test:
  - Create user → Add pets → Delete user → Verify relationships are deleted
- Verify CASCADE works correctly
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 3.6: Register repositories in DI container**
- Update `internal/infrastructure/register.go`:
  - Add database connection provider that reads path from config
  - Add database initialization (open connection, create schemas)
  - Add UsersRepository provider using NewUsersRepository
  - Add PetsRepository provider using NewPetsRepository
- Add config provider that reads database.path from config
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors during startup
- Success criteria: `make test` passes and server starts with --noop
- Implementation status: NOT STARTED

### Phase 4: Application Layer - User Commands

**Task 4.1: Create UserCommands service structure**
- Create `internal/app/users_commands.go`
- Define request models: `CreateUserRequest`, `UpdateUserRequest`
- Define domain model: `User`
- Define error types: `ErrUserNotFound`, `ErrUserEmailConflict`, `ErrInvalidInput`
- Define `UserCommands` interface with all methods
- Create `userCommands` struct with UsersRepository dependency
- Add constructor `NewUserCommands(deps UserCommandsDeps) UserCommands`
- Add stub implementations for all methods
- Create `internal/app/users_commands_test.go` with test structure and `makeMockDeps`
- Run: `make test`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 4.2: Implement and test CreateUser command**
- Create `internal/app/users_testing.go` with factory: `NewRandomCreateUserRequest`
- In `users_commands_test.go`, write tests for CreateUser:
  - Happy path: user is created and returns UUID
  - Validation: empty name returns ErrInvalidInput
  - Validation: empty email returns ErrInvalidInput
  - Validation: invalid email format returns ErrInvalidInput
  - Conflict: duplicate email returns ErrUserEmailConflict
- Run: `go test -v ./internal/app/ --run TestUserCommands/CreateUser`
  - Verify failures (not implemented yet)
- Implement `CreateUser` in `users_commands.go`:
  - Validate input (name not empty, email format with simple regex)
  - Check email uniqueness via repository.GetUserByEmail
  - Generate UUID using uuid.NewV4()
  - Map to repository model
  - Call repository.CreateUser
  - Return user ID
  - Handle repository errors and wrap appropriately
- Run: `go test -v ./internal/app/ --run TestUserCommands/CreateUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 4.3: Implement and test UpdateUser command**
- Add `NewRandomUpdateUserRequest` to `users_testing.go`
- In `users_commands_test.go`, write tests for UpdateUser:
  - Happy path: user is updated successfully
  - Validation: empty name returns ErrInvalidInput
  - Validation: empty email returns ErrInvalidInput
  - Validation: invalid email format returns ErrInvalidInput
  - Not found: non-existent user returns ErrUserNotFound
  - Conflict: duplicate email returns ErrUserEmailConflict
  - Same email: updating user with their own email works
- Run tests to verify failures
- Implement `UpdateUser` in `users_commands.go`:
  - Validate input (name, email)
  - Get existing user to verify it exists
  - Check email uniqueness (only if email changed)
  - Map and call repository.UpdateUser
  - Handle errors appropriately
- Run: `go test -v ./internal/app/ --run TestUserCommands/UpdateUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 4.4: Implement and test DeleteUser command**
- In `users_commands_test.go`, write tests for DeleteUser:
  - Happy path: user is deleted successfully
  - Not found: non-existent user returns ErrUserNotFound
- Run tests to verify failures
- Implement `DeleteUser` in `users_commands.go`:
  - Check user exists via repository.GetUserByID
  - Call repository.DeleteUser (this will cascade delete relationships)
- Run: `go test -v ./internal/app/ --run TestUserCommands/DeleteUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 4.5: Register UserCommands in DI container**
- Update `internal/app/register.go`:
  - Add `NewUserCommands` provider
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors
- Success criteria: `make test` passes and server starts with --noop
- Implementation status: NOT STARTED

### Phase 5: Application Layer - Pets Commands

**Task 5.1: Create PetsCommands service structure**
- Create `internal/app/pets_commands.go`
- Define request model: `AddPetRequest` with UserID, Name, Status, PhotoUrls
- Define error types: `ErrPetCreationFailed`, `ErrUserPetNotFound`
- Define `PetsCommands` interface with AddPet and RemovePet methods
- Create `petsCommands` struct with PetsRepository, UsersRepository, and Petstore Client dependencies
- Add constructor `NewPetsCommands(deps PetsCommandsDeps) PetsCommands`
- Add stub implementations for all methods
- Create `internal/app/pets_commands_test.go` with test structure and `makeMockDeps`
- Run: `make test`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 5.2: Implement and test AddPet command**
- Create `internal/app/pets_testing.go` with factory: `NewRandomAddPetRequest`
- In `pets_commands_test.go`, write tests for AddPet:
  - Happy path: pet is created in petstore, relationship added, returns petID
  - Validation: empty name returns ErrInvalidInput
  - Validation: user must exist (returns ErrUserNotFound)
  - Error: petstore creation fails returns ErrPetCreationFailed
  - Idempotent: adding same pet twice succeeds
- Run: `go test -v ./internal/app/ --run TestPetsCommands/AddPet`
  - Verify failures
- Implement `AddPet` in `pets_commands.go`:
  - Validate input (name not empty, status valid)
  - Check user exists via usersRepository.GetUserByID
  - Create pet in Petstore API via petstoreClient.AddPet
  - If petstore creation succeeds, add relationship via petsRepository.AddUserPet
  - Return pet ID from petstore response
  - Handle errors appropriately
- Run: `go test -v ./internal/app/ --run TestPetsCommands/AddPet`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 5.3: Implement and test RemovePet command**
- In `pets_commands_test.go`, write tests for RemovePet:
  - Happy path: relationship is removed
  - Not found: non-existent relationship returns ErrUserPetNotFound
  - User not found: non-existent user returns ErrUserNotFound
- Run tests to verify failures
- Implement `RemovePet` in `pets_commands.go`:
  - Check user exists via usersRepository.GetUserByID
  - Check relationship exists via petsRepository.HasUserPet
  - Call petsRepository.RemoveUserPet
  - Handle errors appropriately
- Run: `go test -v ./internal/app/ --run TestPetsCommands/RemovePet`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 5.4: Register PetsCommands in DI container**
- Update `internal/app/register.go`:
  - Add `NewPetsCommands` provider
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors
- Success criteria: `make test` passes and server starts with --noop
- Implementation status: NOT STARTED

### Phase 6: Application Layer - User Queries

**Task 6.1: Create UserQueries service structure**
- Create `internal/app/users_queries.go`
- Define domain model: `User` (same as in commands)
- Define `UserQueries` interface with GetUserByID and ListUsers methods
- Create `userQueries` struct with UsersRepository dependency
- Add constructor `NewUserQueries(deps UserQueriesDeps) UserQueries`
- Add stub implementations for all methods
- Create `internal/app/users_queries_test.go` with test structure and `makeMockDeps`
- Run: `make test`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 6.2: Implement and test GetUserByID query**
- In `users_queries_test.go`, write tests for GetUserByID:
  - Happy path: returns user data
  - Not found: returns ErrUserNotFound when user doesn't exist
- Run tests to verify failures
- Implement `GetUserByID` in `users_queries.go`:
  - Call repository.GetUserByID
  - Map to domain model
  - Handle errors
- Run: `go test -v ./internal/app/ --run TestUserQueries/GetUserByID`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 6.3: Implement and test ListUsers query**
- In `users_queries_test.go`, write tests for ListUsers:
  - Happy path: returns list of users
  - Empty: returns empty list when no users
- Run tests to verify failures
- Implement `ListUsers` in `users_queries.go`:
  - Call repository.ListUsers
  - Map to domain models
- Run: `go test -v ./internal/app/ --run TestUserQueries/ListUsers`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 6.4: Register UserQueries in DI container**
- Update `internal/app/register.go`:
  - Add `NewUserQueries` provider
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors
- Success criteria: `make test` passes and server starts with --noop
- Implementation status: NOT STARTED

### Phase 7: Application Layer - Pets Queries

**Task 7.1: Create PetsQueries service structure**
- Create `internal/app/pets_queries.go`
- Define domain model: `Pet` with ID, Name, Status, PhotoUrls
- Define `PetsQueries` interface with ListUserPets method
- Create `petsQueries` struct with PetsRepository, UsersRepository, and Petstore Client dependencies
- Add constructor `NewPetsQueries(deps PetsQueriesDeps) PetsQueries`
- Add stub implementation
- Create `internal/app/pets_queries_test.go` with test structure and `makeMockDeps`
- Run: `make test`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 7.2: Implement and test ListUserPets query**
- In `pets_queries_test.go`, write tests for ListUserPets:
  - Happy path: returns list of pets from petstore
  - Not found: non-existent user returns ErrUserNotFound
  - Empty: returns empty slice when user has no pets
  - Missing pets: gracefully handles missing pets in petstore (filters them out)
- Run tests to verify failures
- Implement `ListUserPets` in `pets_queries.go`:
  - Verify user exists via usersRepository.GetUserByID
  - Call petsRepository.GetUserPetIDs
  - Fetch pet details from petstore for each ID via petstoreClient.GetPetByID
  - Map petstore pets to domain Pet model
  - Filter out pets not found in petstore (log warning)
- Run: `go test -v ./internal/app/ --run TestPetsQueries/ListUserPets`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 7.3: Register PetsQueries in DI container**
- Update `internal/app/register.go`:
  - Add `NewPetsQueries` provider
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors
- Success criteria: `make test` passes and server starts with --noop
- Implementation status: NOT STARTED

### Phase 8: HTTP Controllers - Users

**Task 8.1: Create UsersController structure**
- Create `internal/api/http/v1controllers/users.go`
- Define `UsersController` struct embedding UserCommands and UserQueries
- Add constructor `newUsersController(commands UserCommands, queries UserQueries) *UsersController`
- Add stub implementations for all handler methods (CreateUser, GetUserById, UpdateUser, DeleteUser, ListUsers)
- Ensure controller implements the generated handler interface
- Create `internal/api/http/v1controllers/users_test.go` with test structure
- Run: `make test`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 8.2: Implement and test CreateUser endpoint**
- In `users_test.go`, write tests for CreateUser handler:
  - Happy path: returns 201 with userId
  - Validation error: returns 400 for invalid input
  - Conflict: returns 409 for duplicate email
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/CreateUser`
  - Verify failures
- In `users.go`, implement CreateUser handler:
  - Create transformer to map CreateUserRequest to CreateUserRequest command
  - Transform response userID to CreateUserResponse
  - Handle errors (map ErrUserEmailConflict to 409, ErrInvalidInput to 400)
  - Use HandlerBuilder pattern from existing examples
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/CreateUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 8.3: Implement and test UpdateUser endpoint**
- In `users_test.go`, write tests for UpdateUser handler:
  - Happy path: returns 204
  - Validation error: returns 400
  - Not found: returns 404
  - Conflict: returns 409
- Run tests to verify failures
- Implement UpdateUser handler in `users.go` following transformer pattern
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/UpdateUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 8.4: Implement and test DeleteUser endpoint**
- In `users_test.go`, write tests for DeleteUser handler:
  - Happy path: returns 204
  - Not found: returns 404
- Run tests to verify failures
- Implement DeleteUser handler in `users.go`
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/DeleteUser`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 8.5: Implement and test GetUserById endpoint**
- In `users_test.go`, write tests for GetUserById handler:
  - Happy path: returns 200 with user data (no pets)
  - Not found: returns 404
- Run tests to verify failures
- Implement GetUserById handler in `users.go`:
  - Map User to UserResponse
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/GetUserById`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 8.6: Implement and test ListUsers endpoint**
- In `users_test.go`, write tests for ListUsers handler:
  - Happy path: returns 200 with list of users
  - Empty: returns empty list
- Run tests to verify failures
- Implement ListUsers handler in `users.go`
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestUsersController/ListUsers`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 8.7: Register UsersController**
- Update `internal/api/http/v1controllers/register.go`:
  - Add `newUsersController` to providers
- Update `internal/api/http/v1routes.go`:
  - Add `*v1controllers.UsersController` to V1RoutesDeps
  - Call `rootHandler.RegisterUsersRoutes(deps.UsersController)`
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors
- Success criteria: `make test` passes and server starts with --noop
- Implementation status: NOT STARTED

### Phase 9: HTTP Controllers - Pets

**Task 9.1: Create PetsController structure**
- Create `internal/api/http/v1controllers/pets.go`
- Define `PetsController` struct embedding PetsCommands and PetsQueries
- Add constructor `newPetsController(commands PetsCommands, queries PetsQueries) *PetsController`
- Add stub implementations for all handler methods (AddUserPet, RemoveUserPet, ListUserPets)
- Ensure controller implements the generated handler interface
- Create `internal/api/http/v1controllers/pets_test.go` with test structure
- Run: `make test`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 9.2: Implement and test AddUserPet endpoint**
- In `pets_test.go`, write tests for AddUserPet handler:
  - Happy path: returns 201 with petId
  - Validation error: returns 400
  - User not found: returns 404
  - Pet creation failed: returns 500 or 502
- Run tests to verify failures
- Implement AddUserPet handler in `pets.go`:
  - Map AddPetRequest to AddPetRequest command
  - Transform response petID to AddPetResponse
  - Handle errors appropriately
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/AddUserPet`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 9.3: Implement and test RemoveUserPet endpoint**
- In `pets_test.go`, write tests for RemoveUserPet handler:
  - Happy path: returns 204
  - User not found: returns 404
  - Relationship not found: returns 404
- Run tests to verify failures
- Implement RemoveUserPet handler in `pets.go`
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/RemoveUserPet`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 9.4: Implement and test ListUserPets endpoint**
- In `pets_test.go`, write tests for ListUserPets handler:
  - Happy path: returns 200 with list of pets
  - User not found: returns 404
  - Empty: returns empty list
- Run tests to verify failures
- Implement ListUserPets handler in `pets.go`:
  - Map Pet models to PetResponse
- Run: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/ListUserPets`
  - Verify all tests pass
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 9.5: Register PetsController**
- Update `internal/api/http/v1controllers/register.go`:
  - Add `newPetsController` to providers
- Update `internal/api/http/v1routes.go`:
  - Add `*v1controllers.PetsController` to V1RoutesDeps
  - Call `rootHandler.RegisterPetsRoutes(deps.PetsController)` (or whatever the generated method is)
- Run: `go run ./cmd/server start --env local --noop`
  - Verify no errors
- Success criteria: `make test` passes and server starts with --noop
- Implementation status: NOT STARTED

### Phase 10: Integration & Final Verification

**Task 10.1: Integration tests - full flows**
- In `users_test.go` and `pets_test.go`, add comprehensive integration tests:
  - Full user lifecycle: create → get → update → list → delete
  - Pet relationships: create user → add pet → list pets → remove pet
  - Multiple pets: create user → add multiple pets → verify all returned
  - Cascade delete: create user → add pets → delete user → verify relationships deleted
  - Email conflict: create user → create another with same email → verify 409
  - Pet creation: verify pet is created in petstore before relationship
- Run: `go test -v ./internal/api/http/v1controllers/`
- Success criteria: `make test` passes
- Implementation status: NOT STARTED

**Task 10.2: Final lint and test suite**
- Run: `make lint`
  - Fix any linting issues
- Run: `make test`
  - Verify all tests pass
  - Verify coverage meets threshold
- Update `.mockery.yaml` if needed to include new interfaces
- Success criteria: `make lint` and `make test` both pass with no errors
- Implementation status: NOT STARTED

## 9. Notes

- **TDD Approach**: Each task follows Test-Driven Development - write failing tests first, then implement to make them pass
- **Incremental**: Tasks build on each other; complete them in order
- **Test Coverage**: Aim for high test coverage (>80%) as per project standards
- **Code Style**: Follow existing patterns in the codebase (e.g., EchoService for app layer, echo.go for controllers)
- **Error Handling**: Use wrapped errors with context for better debugging
- **Mocking**: Use mockery-generated mocks for testing (add interfaces to `.mockery.yaml`)
- **Database**: Each test should use a separate in-memory database (`:memory:`) or clean state
- **Concurrency**: Repositories should be safe for concurrent use
- **Petstore Integration**: Mock petstore client in tests
- **Flat Structure**: All files at package level without nested folders
- **API-First**: OpenAPI spec drives model generation, so it comes first

## 10. Success Criteria

The implementation is complete when:
1. All tasks marked as COMPLETED
2. `make lint` passes with no errors
3. `make test` passes with all tests green
4. Test coverage meets or exceeds project threshold
5. Server starts successfully with `--noop` flag
6. All endpoints properly wired and functional

**Estimated effort**: 3-4 days for a developer familiar with the codebase, 5-6 days for someone new to the project.