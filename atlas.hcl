variable "postgres_user" {
  type    = string
  default = getenv("POSTGRES_USER")
}

variable "postgres_password" {
  type    = string
  default = getenv("POSTGRES_PASSWORD")
}

variable "postgres_host" {
  type    = string
  default = getenv("POSTGRES_HOST")
}

variable "postgres_port" {
  type    = string
  default = getenv("POSTGRES_PORT")
}

variable "postgres_db" {
  type    = string
  default = getenv("POSTGRES_DB")
}

env "local" {
  src = "file://internal/infra/postgres/schema.sql"
  url = "postgres://${var.postgres_user}:${var.postgres_password}@${var.postgres_host}:${var.postgres_port}/${var.postgres_db}?sslmode=disable"
  dev = "postgres://${var.postgres_user}:${var.postgres_password}@${var.postgres_host}:${var.postgres_port}/atlas_dev?sslmode=disable"

  migration {
    dir = "file://internal/infra/postgres/migrations"
  }
}
