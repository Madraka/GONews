env "dev" {
  url = "postgres://devuser:devpass@localhost:5433/newsapi_dev?sslmode=disable"
  migration {
    dir = "file://migrations/atlas"
  }
}

env "docker-dev" {
  url = "postgres://devuser:devpass@dev_db:5432/newsapi_dev?sslmode=disable"
  migration {
    dir = "file://migrations/atlas"
  }
}

env "test" {
  url = "postgres://devuser:devpass@localhost:5433/newsapi_test?sslmode=disable"
  migration {
    dir = "file://migrations/atlas"
  }
}

env "prod" {
  url = getenv("DATABASE_URL")
  migration {
    dir = "file://migrations/atlas"
  }
}
