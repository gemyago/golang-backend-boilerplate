Task 3.1: Define PetstoreClient interface in app layer

What I implemented
- Added application-layer port: `internal/app/petstore_client_port.go`.
- The file defines `PetstoreClient` interface which mirrors signatures used by the existing infrastructure petstore client.
  - AddPet(ctx context.Context, params petstore.AddPetParams) (*petstore.Pet, error)
  - GetPetByID(ctx context.Context, params petstore.GetPetByIDParams) (*petstore.Pet, error)
- Included package-level comment per linter `revive` requirements.

Why this change
- Follows "consumer defines interface" principle from the plan.
- Allows infrastructure `*petstore.Client` to satisfy the port structurally without additional adapters.

Verification
- Ran `make test` — all tests passed.
- Ran `make lint` — no linter issues.

Files added/modified
- added: `internal/app/petstore_client_port.go`
- added: `doc/implementation/plan-users-pets-app/summary-task-3.1.md`

Result
Task 3.1: Define PetstoreClient interface in app layer from doc/plan-users-pets-app.md has been successfully implemented. Results summary file can be found here: `doc/implementation/plan-users-pets-app/summary-task-3.1.md` file.