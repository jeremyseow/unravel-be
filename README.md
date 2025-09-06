# unravel-be
Managing events and entities


# Project Structure
This project follows the clean architecture.

```
unravel-be/
├── cmd/                # Frameworks & Drivers
│   └── main.go         # Application entry point
├── config/             # Frameworks & Drivers
│   └── config.go
│   └── ...
├── application/                # Contains core application logic and reusable components
│   ├── domain/         # Entities (Core Business Rules - independent of any framework)
│   │   ├── parameter.go
│   │   ├── schema.go
│   │   └── ...
│   │
│   ├── usecase/        # Application-specific Business Rules (Application Layer)
│   │   ├── parameter.go   # Defines ParameterService interface, ParameterRepository interface
│   │   ├── schema.go      # Defines SchemaService interface, SchemaRepository interface
│   │   └── ...
│   │
│   └── adapter/        # Interface Adapters Layer
│       ├── persistence/ # Database Repositories (implements usecase's repository interfaces)
│       │   └── postgres/
│       │       ├── parameter_repository.go
│       │       ├── schema_repository.go
│       │       └── ...
│       │
│       ├── delivery/    # Web/API Handlers (calls usecase services)
│       │   ├── http/
│       │   │   ├── server.go
│       │   │   ├── router.go
│       │   │   ├── parameter/
│       │   │   │   ├── handler.go
│       │   │   │   ├── model.go
│       │   │   │   └── route.go
│       │   │   └── schema/
│       │   │       ├── handler.go
│       │   │       ├── model.go
│       │   │       └── route.go
│       │   └── grpc/  # If you have gRPC
│       │       └── ...
│       │
│       └── external/    # Clients for external APIs/services
│           └── ...
│
├── db/                 # Database related (migrations, schema tools)
│   ├── migration/
│   │   ├── 000001_create_parameters_table.up.sql
│   │   ├── 000001_create_parameters_table.down.sql
│   │   └── ...
│   ├── .gen/           # Jet generated code
│   └── ...
└── go.mod
└── go.sum
└── makefile
└── Dockerfile
└── docker-compose.yml
└── deployment.yaml
└── README.md
```