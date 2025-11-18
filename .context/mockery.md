# Rules here specify how to use mockery

* Mockery v3 is used, see https://vektra.github.io/mockery/latest/ for more details
* It's configured via mockery config [.mockery.yaml](../.mockery.yaml)
* If you need to regenerate mocks, always run `mockery` from project root without args

## Mocks structure

Standard mockery configuration for all interfaces is defined. It includes the following:
- Uses `testify` compatible mocks
- Output mocks in the same dir as interface
- Each mock is written to own file. This helps to optimize LLM context.
- Mock file name has the following pattern `mock_<snake_case_interface_name>_test.go`

Exceptions are possible if external interfaces needs mocking.