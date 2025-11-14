# Fission MCP Server - Kubernetes Interaction Diagrams

## 1. Architecture Overview

```mermaid
graph TB
    subgraph "External Clients"
        A[HTTP Client]
        B[MCP Client]
    end
    
    subgraph "Fission MCP Server"
        C[HTTP API Layer<br/>cmd/server/main.go]
        D[Router<br/>internal/api/router.go]
        E[Handlers<br/>internal/api/handlers_*.go]
    end
    
    subgraph "CLI Layer"
        F[CLI Executor<br/>internal/cli/executor.go]
        G[Fission CLI<br/>github.com/fission/fission]
    end
    
    subgraph "Kubernetes Client Layer"
        H[Dynamic Client<br/>internal/client/k8s.go]
        I[Typed Clientset<br/>internal/fission/logs.go]
        J[REST Config<br/>internal/fission/logs.go]
    end
    
    subgraph "Kubernetes Cluster"
        K[API Server]
        L[Fission CRDs]
        M[Function Pods]
        N[Router Service]
    end
    
    A --> C
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
    G --> H
    G --> I
    H --> K
    I --> K
    J --> K
    K --> L
    K --> M
    K --> N
```

## 2. Request Flow Diagram

```mermaid
sequenceDiagram
    participant Client
    participant API as HTTP API
    participant Handler as Function Handler
    participant Executor as CLI Executor
    participant FissionCLI as Fission CLI
    participant K8sClient as K8s Client
    participant K8sAPI as K8s API Server
    participant Pod as Function Pod
    
    Client->>API: POST /api/v1/functions
    API->>Handler: CreateFunction()
    Handler->>Executor: ExecuteCommandWithOutput()
    Executor->>FissionCLI: function.Create()
    FissionCLI->>K8sClient: Create Package CRD
    K8sClient->>K8sAPI: POST /apis/fission.io/v1/namespaces/{ns}/packages
    K8sAPI-->>K8sClient: Package Created
    FissionCLI->>K8sClient: Create Function CRD
    K8sClient->>K8sAPI: POST /apis/fission.io/v1/namespaces/{ns}/functions
    K8sAPI-->>K8sClient: Function Created
    K8sAPI->>Pod: Start Function Pod
    Pod-->>K8sAPI: Pod Running
    K8sAPI-->>K8sClient: Pod Status
    K8sClient-->>FissionCLI: Success
    FissionCLI-->>Executor: Command Output
    Executor-->>Handler: stdout/stderr
    Handler-->>API: JSON Response
    API-->>Client: 201 Created
```

## 3. Component Relationship Diagram

```mermaid
graph LR
    subgraph "Core Components"
        A[main.go<br/>Entry Point]
        B[SetupRouter<br/>Route Configuration]
        C[HTTP Server<br/>Port 8080]
    end
    
    subgraph "API Layer"
        D[Function Handler<br/>CRUD Operations]
        E[Environment Handler<br/>Runtime Management]
        F[Package Handler<br/>Code Deployment]
        G[Trigger Handlers<br/>Event Sources]
    end
    
    subgraph "Business Logic"
        H[CLI Executor<br/>Command Execution]
        I[Service Layer<br/>Kubernetes Operations]
        J[Log Management<br/>Pod Monitoring]
    end
    
    subgraph "Kubernetes Integration"
        K[Dynamic Client<br/>CRD Operations]
        L[Clientset<br/>Core Resources]
        M[Config Manager<br/>Authentication]
    end
    
    A --> B
    B --> C
    C --> D
    C --> E
    C --> F
    C --> G
    D --> H
    E --> H
    F --> H
    G --> H
    H --> I
    I --> J
    I --> K
    I --> L
    K --> M
    L --> M
```

## 4. Kubernetes Resource Interaction

```mermaid
graph TD
    subgraph "Fission Resources"
        A[Function<br/>fission.io/v1]
        B[Package<br/>fission.io/v1]
        C[Environment<br/>fission.io/v1]
        D[HTTP Trigger<br/>fission.io/v1]
    end
    
    subgraph "Kubernetes Resources"
        E[Pod<br/>corev1]
        F[Service<br/>corev1]
        G[ConfigMap<br/>corev1]
        H[Secret<br/>corev1]
    end
    
    subgraph "Operations"
        I[Create]
        J[Read]
        K[Update]
        L[Delete]
        M[List]
        N[Watch]
    end
    
    A --> E
    B --> G
    B --> H
    C --> F
    D --> F
    
    I --> A
    I --> B
    I --> C
    I --> D
    J --> A
    J --> B
    J --> C
    J --> D
    K --> A
    K --> B
    L --> A
    L --> B
    M --> A
    M --> B
    N --> E
```

## 5. Authentication Flow

```mermaid
graph TB
    A[Application Start] --> B{Running in Cluster?}
    B -->|Yes| C[Use InClusterConfig]
    B -->|No| D[Check KUBECONFIG Env]
    D -->|Set| E[Use KUBECONFIG Path]
    D -->|Not Set| F[Use Default ~/.kube/config]
    
    C --> G[Service Account Token]
    E --> H[Kubeconfig File]
    F --> H
    G --> I[REST Config]
    H --> I
    I --> J[Dynamic Client]
    I --> K[Typed Clientset]
    
    J --> L[Kubernetes API Server]
    K --> L
```

## 6. Function Lifecycle

```mermaid
stateDiagram-v2
    [*] --> CreatePackage
    CreatePackage --> CreateFunction
    CreateFunction --> Pending
    Pending --> Building
    Building --> Ready
    Ready --> Executing
    Executing --> Idle
    Idle --> Executing
    Executing --> Error
    Error --> Building
    Ready --> Updating
    Updating --> Ready
    Executing --> Deleting
    Idle --> Deleting
    Deleting --> [*]