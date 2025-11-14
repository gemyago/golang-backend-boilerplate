# Summary Task 5.2: Implement and test AddPet command

Implemented the AddPet command in [`internal/app/pets_commands.go`](internal/app/pets_commands.go) following TDD principles.

Key changes:
- Added input validation for empty name, returning [`ErrInvalidInput`](internal/app/users_commands.go:301)
- Verified user existence via [`UsersRepository.GetUserByID`](internal/app/users_repository.go:208), returning [`ErrUserNotFound`](internal/app/users_commands.go:298) if not found
- Created pet in external Petstore API using [`PetstoreClient.AddPet`](internal/app/petstore_client_port.go:22) with [`petstore.AddPetParams`](internal/infrastructure/petstore/add_pet.go:11), wrapping failures as [`ErrPetCreationFailed`](internal/app/pets_commands.go:29)
- Persisted user-pet relationship via [`PetsRepository.AddUserPet`](internal/app/pets_repository.go:234) using [`UserPet`](internal/app/pets_repository.go:224) entity
- Returned [`AddPetResponse`](internal/app/pets_commands.go:23) with the created pet ID

Added comprehensive unit tests in [`internal/app/pets_commands_test.go`](internal/app/pets_commands_test.go):
- Happy path: successful pet creation and relationship addition
- Validation: empty name returns ErrInvalidInput
- User not found: returns ErrUserNotFound
- Petstore failure: returns ErrPetCreationFailed (checked with errors.Is)
- Unexpected repository error: propagates original error

Created test factory [`NewRandomAddPetRequest`](internal/app/pets_testing.go:9) in [`internal/app/pets_testing.go`](internal/app/pets_testing.go) for generating random test data.

Updated [`internal/app/mocks_test.go`](internal/app/mocks_test.go) with generated mocks for PetsRepository and PetstoreClient via mockery.

Fixed lint issues:
- Used %w for error wrapping in fmt.Errorf
- Renamed shadowed 'err' variable to 'addErr' in AddUserPet call
- Used require.ErrorIs in tests
- Formatted files with goimports

Results:
- All tests pass: go test -v ./internal/app/ --run TestPetsCommands/AddPet
- Lint clean: make lint
- Coverage: 100% for AddPet method