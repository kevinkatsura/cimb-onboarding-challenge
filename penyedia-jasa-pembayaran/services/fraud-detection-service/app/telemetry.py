import logging
from opentelemetry import trace, propagate
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

# Instrumentors
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor
from opentelemetry.instrumentation.redis import RedisInstrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.instrumentation.grpc import GrpcInstrumentorServer
from opentelemetry.instrumentation.asyncpg import AsyncPGInstrumentor

_instrumented = False

def init_telemetry(service_name: str, otlp_endpoint: str):
    global _instrumented
    if _instrumented:
        return
    
    # 1. Resource and Provider Setup
    resource = Resource(attributes={
        SERVICE_NAME: service_name
    })
    provider = TracerProvider(resource=resource)
    
    # 2. Processor and Exporter
    processor = BatchSpanProcessor(
        OTLPSpanExporter(endpoint=otlp_endpoint, insecure=True)
    )
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)
    
    # 3. Global Propagator (W3C Trace Context)
    propagate.set_global_textmap(TraceContextTextMapPropagator())
    
    # 4. Instrumentation
    LoggingInstrumentor().instrument(set_logging_format=True)
    RedisInstrumentor().instrument()
    AsyncPGInstrumentor().instrument()
    GrpcInstrumentorServer().instrument()
    # Note: SQLAlchemy and FastAPI are instrumented in main.py with app/engine references
    
    _instrumented = True
    logging.getLogger("telemetry").info(f"Telemetry initialized for {service_name}")

def get_tracer(name: str):
    return trace.get_tracer(name)

def instrument_sqlalchemy(engine):
    SQLAlchemyInstrumentor().instrument(engine=engine)

def instrument_fastapi(app):
    FastAPIInstrumentor.instrument_app(app)
