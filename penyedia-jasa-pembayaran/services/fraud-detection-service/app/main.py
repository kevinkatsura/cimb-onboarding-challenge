import asyncio
import logging
import uuid
from contextlib import asynccontextmanager
from datetime import datetime, timezone
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from sqlalchemy import select, desc

# ─── OpenTelemetry ───
from app.config import settings
from app.telemetry import init_telemetry, instrument_sqlalchemy, instrument_fastapi

# Initialize telemetry as first step
init_telemetry(
    service_name="fraud-detection-service",
    otlp_endpoint=settings.OTEL_EXPORTER_OTLP_ENDPOINT
)

# NOW import our internal modules
from app.database import async_session, engine
from app.redis_client import redis_client
from app.models import FraudRule, FraudEvent
from app.engine.evaluator import EvaluationRequest, evaluate
from app.grpc_handler import serve_grpc

# Instrument engines
instrument_sqlalchemy(engine.sync_engine)

# ─── Logging ───
import logging_loki

# Loki Configuration
loki_handler = logging_loki.LokiHandler(
    url=settings.LOKI_URL,
    tags={"job": "fraud-detection-service", "service": "fraud-detection-service"},
    version="1",
)

# Root Logger Setup
root_logger = logging.getLogger()
root_logger.setLevel(logging.INFO)
root_logger.addHandler(loki_handler)

# Console Handler for Docker logs
console_handler = logging.StreamHandler()
console_handler.setFormatter(logging.Formatter(
    "%(asctime)s %(levelname)-5s [%(name)s] [trace_id=%(otelTraceID)s span_id=%(otelSpanID)s] %(message)s"
))
root_logger.addHandler(console_handler)

# Logging is already instrumented in init_telemetry()

logger = logging.getLogger("fraud.main")


# ─── Lifespan ───
@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Fraud Detection Service starting")
    logger.info("DB: %s@%s/%s", settings.DB_USER, settings.DB_HOST, settings.DB_NAME)
    logger.info("Redis: %s:%d", settings.REDIS_HOST, settings.REDIS_PORT)
    logger.info("HTTP port: %d", settings.HTTP_PORT)
    
    # Initialize Redis
    await redis_client.initialize()
    
    # Start gRPC server in the background
    grpc_task = asyncio.create_task(serve_grpc(50055))
    
    yield
    
    logger.info("Fraud Detection Service shutting down")
    grpc_task.cancel()
    try:
        await grpc_task
    except asyncio.CancelledError:
        pass
    await engine.dispose()
    await redis_client.close()


app = FastAPI(
    title="Fraud Detection Service",
    description="Real-time fraud evaluation for PJP banking transactions",
    version="1.0.0",
    lifespan=lifespan,
)

# Instrument FastAPI
instrument_fastapi(app)

# ─── Pydantic Schemas ───
class DeviceFingerprintSchema(BaseModel):
    user_agent: str = ""
    platform: str = ""
    screen_resolution: str = ""
    timezone: str = ""


class FraudEvaluationRequestSchema(BaseModel):
    transaction_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    partner_reference_no: str = ""
    reference_no: str = ""
    source_account_no: str
    beneficiary_account_no: str
    amount: int
    currency: str = "IDR"
    source_ip: str = ""
    device_id: str = ""
    device_fingerprint: DeviceFingerprintSchema | None = None
    channel: str = ""
    latitude: float = 0.0
    longitude: float = 0.0


class FraudEvaluationResponseSchema(BaseModel):
    decision: str
    risk_score: float
    triggered_rules: list[str]
    event_id: str
    message: str


class FraudRuleSchema(BaseModel):
    id: str
    code: str
    name: str
    category: str
    description: str
    risk_weight: float
    decision_on_trigger: str
    threshold_config: dict[str, Any]
    is_active: bool


class FraudEventSchema(BaseModel):
    id: str
    transaction_id: str
    source_account_no: str
    beneficiary_account_no: str
    amount: int
    decision: str
    risk_score: float
    triggered_rules: list[str]
    evaluated_at: str


# ─── Endpoints ───
@app.get("/health")
async def health():
    return {"status": "ok", "service": "fraud-detection"}


@app.post("/evaluate", response_model=FraudEvaluationResponseSchema)
async def evaluate_transaction(req: FraudEvaluationRequestSchema):
    """Evaluate a transaction for fraud risk."""
    try:
        eval_req = EvaluationRequest(
            transaction_id=req.transaction_id,
            partner_reference_no=req.partner_reference_no,
            reference_no=req.reference_no,
            source_account_no=req.source_account_no,
            beneficiary_account_no=req.beneficiary_account_no,
            amount=req.amount,
            currency=req.currency,
            source_ip=req.source_ip,
            device_id=req.device_id,
            device_fingerprint=req.device_fingerprint.model_dump() if req.device_fingerprint else {},
            channel=req.channel,
            latitude=req.latitude,
            longitude=req.longitude,
        )

        async with async_session() as session:
            result = await evaluate(session, redis_client, eval_req)

        return FraudEvaluationResponseSchema(
            decision=result.decision,
            risk_score=result.risk_score,
            triggered_rules=result.triggered_rules,
            event_id=result.event_id,
            message=result.message,
        )
    except Exception as e:
        logger.exception("Fraud evaluation failed: %s", e)
        if settings.FAIL_OPEN:
            return FraudEvaluationResponseSchema(
                decision="allow",
                risk_score=0.0,
                triggered_rules=[],
                event_id="",
                message=f"evaluation_error: {str(e)} (fail-open)",
            )
        raise HTTPException(status_code=500, detail=f"Fraud evaluation error: {e}")


@app.get("/rules", response_model=list[FraudRuleSchema])
async def list_rules():
    """List all fraud rules."""
    async with async_session() as session:
        result = await session.execute(select(FraudRule))
        rules = result.scalars().all()
        return [
            FraudRuleSchema(
                id=str(r.id),
                code=r.code,
                name=r.name,
                category=r.category,
                description=r.description,
                risk_weight=r.risk_weight,
                decision_on_trigger=r.decision_on_trigger,
                threshold_config=r.threshold_config,
                is_active=r.is_active,
            )
            for r in rules
        ]


@app.get("/events", response_model=list[FraudEventSchema])
async def list_events(limit: int = 20):
    """List recent fraud events."""
    async with async_session() as session:
        result = await session.execute(
            select(FraudEvent).order_by(desc(FraudEvent.evaluated_at)).limit(limit)
        )
        events = result.scalars().all()
        return [
            FraudEventSchema(
                id=str(e.id),
                transaction_id=e.transaction_id,
                source_account_no=e.source_account_no,
                beneficiary_account_no=e.beneficiary_account_no,
                amount=e.amount,
                decision=e.decision,
                risk_score=e.risk_score,
                triggered_rules=e.triggered_rules,
                evaluated_at=e.evaluated_at.isoformat() if e.evaluated_at else "",
            )
            for e in events
        ]
