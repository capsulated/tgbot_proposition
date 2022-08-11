# Proposition telegram bot

## Requirements
- [Docker](https://docs.docker.com/engine/install/) needs to be installed on your system for the service to work.
- After installing docker, [docker-compose](https://docs.docker.com/compose/install/) should be installed.

## Development
### Prepare Database
```bash
cp vault/.env.example vault/.env.dev
vi vault/.env.dev
```
Then edit POSTGRES_USER, POSTGRES_PASSWORD, 
POSTGRES_DB, POSTGRES_HOST, POSTGRES_PORT
variables values for development instance of
development PostgresSQL database.

### Prepare Service
```bash
cp worker/.env.example worker/.env.dev
vi worker/.env.dev
```
Then edit DB_CONN_STRING as database connecting string for postgresql database, 
TELEGRAM_DEBUG, TELEGRAM_API_TOKEN, TELEGRAM_INFO_CHANNEL_ID as variables for
telegram bot and PIN_INITIATOR, PIN_SECRETARY the development 
pin codes for initiator and secretary.

### Run Service
```bash
docker-compose -f docker-compose.dev.yaml up
```

## Production
### Prepare Database
```bash
cp vault/.env.example vault/.env
vi vault/.env
```
Then edit POSTGRES_USER, POSTGRES_PASSWORD,
POSTGRES_DB, POSTGRES_HOST, POSTGRES_PORT
variables values for development instance of
production PostgresSQL database.

### Prepare Service
```bash
cp worker/.env.example worker/.env
vi worker/.env
```
Then edit DB_CONN_STRING as database connecting string for postgresql database,
TELEGRAM_DEBUG, TELEGRAM_API_TOKEN, TELEGRAM_INFO_CHANNEL_ID as variables for
telegram bot and PIN_INITIATOR, PIN_SECRETARY the production
pin codes for initiator and secretary.

### Run Service
```bash
docker-compose -f docker-compose.prod.yaml up -d
```
