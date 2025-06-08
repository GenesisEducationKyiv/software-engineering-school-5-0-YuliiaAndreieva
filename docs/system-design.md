# System Design 

---

## 1. System Requirements

### Functional Requirements

- Users can subscribe/unsubscribe to weather updates for specific cities
- Users can choose notification frequency (hourly or daily)
- Users must confirm subscription via email
- System sends scheduled email notifications with weather information
- System validates city existence before subscription

### Non-Functional Requirements

#### Scalability:
- Support 100,000 active subscriptions
- Handle 50,000 email notifications per day

#### Security:
- All API inputs validated
- Secure token generation for confirmations/unsubscribe
- SMTP authentication for email delivery

#### Constraints
- Free tier limitations of WeatherAPI.com

## 2. Load Estimation

### Users and Traffic

- Active users: 10,000
- Estimated subscriptions per user: 2-3
- API Requests per second: 15 rps
- Background jobs: 50,000 jobs/day

### Data Load

- Subscriptions: 200 bytes/record
- Total Database: ~15 GB/year

### Bandwidth

- Incoming: 500Kbps
- Outgoing: 2Mbps
- External API: 5Mbps


## 3.High-Level Architecture

```mermaid
flowchart TB
 subgraph subGraph0["Core Services"]
        WS["Weather Service"]
        SS["Subscription Service"]
        ES["Email Service"]
        TS["Token Service"]
  end
 subgraph subGraph1["Data Layer"]
        PG[("PostgreSQL")]
  end
 subgraph subGraph2["External Services"]
        Weather["WeatherAPI.com"]
        Email["SMTP Server"]
  end
    User["Users"] --> API["API Server"]
    API --> WS & SS & ES & TS
    WS --> Weather
    ES --> Email & PG
    SS --> PG
```

## 4.Detailed component design
### 4.1 API Service & Endpoints

**Responsibilities:**

- Handle HTTP requests for weather service
- Validate input data
- Interact with business logic services
- Handle errors and logging
- Format JSON responses
- Route requests to appropriate handlers
- Validate confirmation and unsubscribe tokens
- Process weather subscription updates
- Send email notifications
- Fetch weather data
- Schedule periodic weather updates

**REST API Endpoints:**

```typescript
POST /api/subscribe
GET  /api/confirm/:token
GET  /api/unsubscribe/:token
GET  /api/weather?city={city}
```

## 5. Sequence Diagrams

### 5.1 Weather Request Flow
```mermaid
sequenceDiagram
    participant Client
    participant API as Gin API Server
    participant WS as Weather Service
    participant WeatherAPI as WeatherAPI.com

    Client->>API: GET /api/weather?city={city}
    API->>WS: GetWeather(city)
    WS->>WeatherAPI: Fetch weather data
    WeatherAPI-->>WS: Weather response
    WS-->>API: Weather data
    API-->>Client: JSON response
```

### 5.2 Subscription Flow
```mermaid
sequenceDiagram
    participant Client
    participant API as Gin API Server
    participant SS as Subscription Service
    participant WS as Weather Service
    participant TS as Token Service
    participant DB as PostgreSQL
    participant ES as Email Service
    participant SMTP as SMTP Server
    participant WeatherAPI as WeatherAPI.com

    Client->>API: POST /api/subscribe
    API->>SS: Subscribe(email, city, frequency)
    SS->>WS: Validate city
    WS->>WeatherAPI: Check city exists
    WeatherAPI-->>WS: City valid
    WS-->>SS: City validated
    SS->>TS: Generate token
    TS-->>SS: Token
    SS->>DB: Create subscription
    SS->>ES: Send confirmation email
    ES->>SMTP: Send email
    SMTP-->>ES: Email sent
    ES-->>SS: Email sent
    SS-->>API: Subscription created
    API-->>Client: Success response
```

### 5.3 Confirmation Flow
```mermaid
sequenceDiagram
    participant Client
    participant API as Gin API Server
    participant SS as Subscription Service
    participant DB as PostgreSQL

    Client->>API: GET /api/confirm/:token
    API->>SS: Confirm(token)
    SS->>DB: Update subscription
    DB-->>SS: Updated
    SS-->>API: Confirmed
    API-->>Client: Success response
```

### 5.4 Unsubscribe Flow
```mermaid
sequenceDiagram
    participant Client
    participant API as Gin API Server
    participant SS as Subscription Service
    participant DB as PostgreSQL

    Client->>API: GET /api/unsubscribe/:token
    API->>SS: Unsubscribe(token)
    SS->>DB: Delete subscription
    DB-->>SS: Deleted
    SS-->>API: Unsubscribed
    API-->>Client: Success response
```