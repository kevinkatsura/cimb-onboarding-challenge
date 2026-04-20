import grpc
import logging
from concurrent import futures
from opentelemetry.instrumentation.grpc import aio_server_interceptor

from app.proto.v1 import fraud_pb2, fraud_pb2_grpc
from app.engine.evaluator import EvaluationRequest, evaluate
from app.database import async_session
from app.redis_client import redis_client

logger = logging.getLogger("fraud.grpc")

class FraudDetectionServiceServicer(fraud_pb2_grpc.FraudDetectionServiceServicer):
    async def EvaluateTransaction(self, request, context):
        logger.info(f"Received gRPC evaluation request for {request.transaction_id}")
        
        # Map gRPC request to internal EvaluationRequest
        eval_req = EvaluationRequest(
            transaction_id=request.transaction_id,
            partner_reference_no=request.partner_reference_no,
            reference_no=request.reference_no,
            source_account_no=request.source_account_no,
            beneficiary_account_no=request.beneficiary_account_no,
            amount=request.amount,
            currency=request.currency,
            source_ip=request.source_ip,
            device_id=request.device_id,
            device_fingerprint={
                "user_agent": request.device_fingerprint.user_agent,
                "platform": request.device_fingerprint.platform,
                "screen_resolution": request.device_fingerprint.screen_resolution,
                "timezone": request.device_fingerprint.timezone,
            },
            channel=request.channel,
            latitude=request.latitude,
            longitude=request.longitude,
        )
        
        # Run evaluation logic with session and redis
        async with async_session() as session:
            result = await evaluate(session, redis_client, eval_req)
        
        # Map internal result (object) to gRPC response
        return fraud_pb2.FraudEvaluationResponse(
            decision=fraud_pb2.Decision.Value(result.decision.upper()),
            risk_score=result.risk_score,
            triggered_rules=result.triggered_rules,
            event_id=result.event_id,
            message=result.message
        )

async def serve_grpc(port: int):
    server = grpc.aio.server(
        futures.ThreadPoolExecutor(max_workers=10),
        interceptors=[aio_server_interceptor()]
    )
    fraud_pb2_grpc.add_FraudDetectionServiceServicer_to_server(
        FraudDetectionServiceServicer(), server
    )
    listen_addr = f"[::]:{port}"
    server.add_insecure_port(listen_addr)
    logger.info(f"Starting gRPC server on {listen_addr}")
    await server.start()
    await server.wait_for_termination()
