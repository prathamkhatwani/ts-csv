# Use a modern Go image for building the Go port
FROM golang:1.21-alpine AS go-builder
WORKDIR /app
COPY go.mod ./
COPY src/csv.go ./src/
COPY src/cmd/csv-cli/main.go ./src/cmd/csv-cli/
RUN go mod download
RUN mkdir -p bin && go build -o bin/csv-cli ./src/cmd/csv-cli

# Use a Node image for running the original Jest test suite
FROM node:20-alpine
WORKDIR /app
COPY --from=go-builder /app/bin/csv-cli ./bin/csv-cli
COPY package.json package-lock.json* ./
RUN npm install
COPY tsconfig.json ./
COPY tests/ ./tests/

# Verify tests pass using the Thin Adapter to the Go binary
CMD ["npm", "test"]
