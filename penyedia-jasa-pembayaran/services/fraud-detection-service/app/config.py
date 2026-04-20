"""Configuration module — reads environment variables."""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    # ─── Database ───
    DB_HOST: str = "localhost"
    DB_PORT: int = 5432
    DB_USER: str = "postgres"
    DB_PASSWORD: str = "postgres"
    DB_NAME: str = "pjp"
    DB_SSLMODE: str = "disable"

    # ─── Redis ───
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379

    # ─── Service ───
    GRPC_PORT: int = 50060
    HTTP_PORT: int = 8085

    # ─── Upstream Services ───
    AIS_GRPC_ADDR: str = "localhost:50055"
    OTEL_EXPORTER_OTLP_ENDPOINT: str = "http://jaeger:4317"
    LOKI_URL: str = "http://loki:3100/loki/api/v1/push"

    # ─── Behaviour ───
    FAIL_OPEN: bool = True  # allow transactions if fraud service errors internally

    @property
    def database_url(self) -> str:
        return (
            f"postgresql+asyncpg://{self.DB_USER}:{self.DB_PASSWORD}"
            f"@{self.DB_HOST}:{self.DB_PORT}/{self.DB_NAME}"
        )

    @property
    def database_url_sync(self) -> str:
        return (
            f"postgresql://{self.DB_USER}:{self.DB_PASSWORD}"
            f"@{self.DB_HOST}:{self.DB_PORT}/{self.DB_NAME}"
        )

    @property
    def redis_url(self) -> str:
        return f"redis://{self.REDIS_HOST}:{self.REDIS_PORT}/0"

    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()
