"""Fraud evaluation orchestrator.

Runs all active fraud rules against a transaction request and produces
a final decision with risk score and audit trail.
"""

import uuid
import logging
from datetime import datetime, timezone
from typing import Any

import redis.asyncio as aioredis
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.dialects.postgresql import insert as pg_insert

from app.models import FraudRule, FraudEvent, FraudAccountStats, FraudFingerprint
from app.engine.velocity import check_all_velocities
from app.engine.amount import check_amount_anomaly, compute_new_average
from app.engine.beneficiary import check_new_beneficiary
from app.engine.device import check_device_change
from app.engine.dormant import check_dormant_account
from app.engine.geo import check_geo_anomaly
from app.telemetry import get_tracer

tracer = get_tracer(__name__)

logger = logging.getLogger("fraud.evaluator")

# Decision severity ranking: higher index = more severe
DECISION_SEVERITY = {"allow": 0, "challenge": 1, "review": 2, "block": 3}


class EvaluationRequest:
    """Holds all the data needed for fraud evaluation."""

    def __init__(self, **kwargs: Any):
        self.transaction_id: str = kwargs.get("transaction_id", "")
        self.partner_reference_no: str = kwargs.get("partner_reference_no", "")
        self.reference_no: str = kwargs.get("reference_no", "")
        self.source_account_no: str = kwargs.get("source_account_no", "")
        self.beneficiary_account_no: str = kwargs.get("beneficiary_account_no", "")
        self.amount: int = kwargs.get("amount", 0)
        self.currency: str = kwargs.get("currency", "IDR")
        self.source_ip: str = kwargs.get("source_ip", "")
        self.device_id: str = kwargs.get("device_id", "")
        self.device_fingerprint: dict = kwargs.get("device_fingerprint", {})
        self.channel: str = kwargs.get("channel", "")
        self.latitude: float = kwargs.get("latitude", 0.0)
        self.longitude: float = kwargs.get("longitude", 0.0)


class EvaluationResult:
    """Final fraud evaluation output."""

    def __init__(self):
        self.decision: str = "allow"
        self.risk_score: float = 0.0
        self.triggered_rules: list[str] = []
        self.event_id: str = ""
        self.message: str = ""
        self.rule_details: dict[str, Any] = {}


async def evaluate(
    session: AsyncSession,
    redis: aioredis.Redis,
    req: EvaluationRequest,
) -> EvaluationResult:
    with tracer.start_as_current_span("fraud.evaluate") as span:
        span.set_attribute("transaction.id", req.transaction_id)
        span.set_attribute("account.source", req.source_account_no)
        return await _evaluate_internal(session, redis, req)

async def _evaluate_internal(
    session: AsyncSession,
    redis: aioredis.Redis,
    req: EvaluationRequest,
) -> EvaluationResult:
    """Run all active fraud rules and produce a decision.

    Steps:
        1. Load active rules from DB
        2. Fetch account stats (running avg, last device, etc.)
        3. Run each rule checker
        4. Aggregate risk score and determine final decision
        5. Persist fraud_event audit record
        6. Update fraud_account_stats
        7. Upsert fraud_fingerprint
    """
    result = EvaluationResult()

    # 1. Load active rules
    with tracer.start_as_current_span("fraud.load_rules"):
        rules_result = await session.execute(
            select(FraudRule).where(FraudRule.is_active == True)  # noqa: E712
        )
        rules = {r.code: r for r in rules_result.scalars().all()}

    if not rules:
        logger.warning("No active fraud rules found")
        result.message = "no_active_rules"
        return result

    # Build threshold config lookup
    rule_thresholds = {code: r.threshold_config for code, r in rules.items()}

    # 2. Fetch account stats
    stats_result = await session.execute(
        select(FraudAccountStats).where(
            FraudAccountStats.account_number == req.source_account_no
        )
    )
    stats = stats_result.scalar_one_or_none()

    # 3. Run each rule
    triggered: list[dict] = []
    all_details: dict[str, Any] = {}

    # ── 3a. Velocity checks ──
    with tracer.start_as_current_span("fraud.check_velocity"):
        velocity_results = await check_all_velocities(
            redis=redis,
            account_no=req.source_account_no,
            device_id=req.device_id,
            source_ip=req.source_ip,
            beneficiary_no=req.beneficiary_account_no,
            rules=rule_thresholds,
        )
        for vr in velocity_results:
            all_details[vr["rule_code"]] = vr
            if vr["triggered"]:
                triggered.append(vr)

    # ── 3b. Amount anomaly ──
    if "AMT_ANOMALY" in rules:
        amt_config = rule_thresholds.get("AMT_ANOMALY", {})
        amt_result = check_amount_anomaly(
            amount=req.amount,
            avg_amount=stats.avg_amount if stats else 0.0,
            tx_count=stats.tx_count if stats else 0,
            std_dev_multiplier=amt_config.get("std_dev_multiplier", 3.0),
        )
        all_details["AMT_ANOMALY"] = amt_result
        if amt_result["triggered"]:
            triggered.append(amt_result)

    # ── 3c. New beneficiary ──
    if "NEW_BENEF" in rules:
        benef_config = rule_thresholds.get("NEW_BENEF", {})
        benef_result = await check_new_beneficiary(
            session=session,
            source_account_no=req.source_account_no,
            beneficiary_account_no=req.beneficiary_account_no,
            lookback_days=benef_config.get("lookback_days", 90),
        )
        all_details["NEW_BENEF"] = benef_result
        if benef_result["triggered"]:
            triggered.append(benef_result)

    # ── 3d. Device change ──
    if "DEV_CHANGE" in rules:
        dev_result = check_device_change(
            current_device_id=req.device_id,
            last_device_id=stats.last_device_id if stats else None,
        )
        all_details["DEV_CHANGE"] = dev_result
        if dev_result["triggered"]:
            triggered.append(dev_result)

    # ── 3e. Dormant account ──
    if "DORMANT_ACCT" in rules:
        dormant_config = rule_thresholds.get("DORMANT_ACCT", {})
        dormant_result = check_dormant_account(
            last_tx_at=stats.last_tx_at if stats else None,
            dormant_days=dormant_config.get("dormant_days", 90),
        )
        all_details["DORMANT_ACCT"] = dormant_result
        if dormant_result["triggered"]:
            triggered.append(dormant_result)

    # ── 3f. Geolocation anomaly ──
    if "GEO_ANOMALY" in rules:
        geo_config = rule_thresholds.get("GEO_ANOMALY", {})
        geo_result = check_geo_anomaly(
            current_lat=req.latitude,
            current_lon=req.longitude,
            last_lat=stats.last_latitude if stats else None,
            last_lon=stats.last_longitude if stats else None,
            last_tx_at=stats.last_tx_at if stats else None,
            max_distance_km=geo_config.get("max_distance_km", 500.0),
            time_window_hours=geo_config.get("time_window_hours", 1.0),
        )
        all_details["GEO_ANOMALY"] = geo_result
        if geo_result["triggered"]:
            triggered.append(geo_result)

    # 4. Aggregate decision
    triggered_codes = [t["rule_code"] for t in triggered]
    result.triggered_rules = triggered_codes
    result.rule_details = all_details

    if triggered_codes:
        # Weighted risk score
        total_weight = sum(r.risk_weight for r in rules.values())
        triggered_weight = sum(
            rules[code].risk_weight for code in triggered_codes if code in rules
        )
        result.risk_score = round((triggered_weight / total_weight) * 100.0, 2) if total_weight > 0 else 0.0

        # Decision = highest severity among triggered rules
        max_severity = 0
        for code in triggered_codes:
            if code in rules:
                decision = rules[code].decision_on_trigger
                severity = DECISION_SEVERITY.get(decision, 0)
                if severity > max_severity:
                    max_severity = severity
                    result.decision = decision
    else:
        result.decision = "allow"
        result.risk_score = 0.0

    result.message = (
        f"{len(triggered_codes)} rule(s) triggered"
        if triggered_codes
        else "all checks passed"
    )

    # 5. Persist audit event & Update stats
    with tracer.start_as_current_span("fraud.persistence"):
        event_id = uuid.uuid4()
        result.event_id = str(event_id)

        event = FraudEvent(
            id=event_id,
            transaction_id=req.transaction_id,
            partner_reference_no=req.partner_reference_no,
            reference_no=req.reference_no,
            source_account_no=req.source_account_no,
            beneficiary_account_no=req.beneficiary_account_no,
            amount=req.amount,
            currency=req.currency,
            device_id=req.device_id,
            source_ip=req.source_ip,
            channel=req.channel,
            latitude=req.latitude if req.latitude else None,
            longitude=req.longitude if req.longitude else None,
            decision=result.decision,
            risk_score=result.risk_score,
            triggered_rules=triggered_codes,
            rule_details=_sanitize_details(all_details),
            evaluated_at=datetime.now(timezone.utc),
        )
        session.add(event)

    # 6. Update account stats (online average update)
    now = datetime.now(timezone.utc)
    if stats:
        new_avg = compute_new_average(stats.avg_amount, stats.tx_count, req.amount)
        stats.tx_count += 1
        stats.avg_amount = new_avg
        stats.last_tx_at = now
        stats.last_device_id = req.device_id or stats.last_device_id
        stats.last_ip = req.source_ip or stats.last_ip
        if req.latitude:
            stats.last_latitude = req.latitude
        if req.longitude:
            stats.last_longitude = req.longitude
        stats.updated_at = now
    else:
        new_stats = FraudAccountStats(
            account_number=req.source_account_no,
            tx_count=1,
            avg_amount=float(req.amount),
            last_tx_at=now,
            last_device_id=req.device_id,
            last_ip=req.source_ip,
            last_latitude=req.latitude if req.latitude else None,
            last_longitude=req.longitude if req.longitude else None,
            updated_at=now,
        )
        session.add(new_stats)

    # 7. Upsert fingerprint
    if req.device_id:
        fp = req.device_fingerprint or {}
        stmt = pg_insert(FraudFingerprint).values(
            account_number=req.source_account_no,
            device_id=req.device_id,
            user_agent=fp.get("user_agent", ""),
            platform=fp.get("platform", ""),
            screen_resolution=fp.get("screen_resolution", ""),
            timezone=fp.get("timezone", ""),
            ip_address=req.source_ip,
            latitude=req.latitude if req.latitude else None,
            longitude=req.longitude if req.longitude else None,
            first_seen_at=now,
            last_seen_at=now,
        ).on_conflict_do_update(
            constraint="uq_fingerprint_account_device",
            set_={
                "user_agent": fp.get("user_agent", ""),
                "platform": fp.get("platform", ""),
                "screen_resolution": fp.get("screen_resolution", ""),
                "timezone": fp.get("timezone", ""),
                "ip_address": req.source_ip,
                "latitude": req.latitude if req.latitude else None,
                "longitude": req.longitude if req.longitude else None,
                "last_seen_at": now,
            },
        )
        await session.execute(stmt)

    await session.commit()

    logger.info(
        "Fraud evaluation complete: tx=%s decision=%s score=%.2f triggered=%s",
        req.transaction_id, result.decision, result.risk_score, triggered_codes,
    )

    return result


def _sanitize_details(details: dict) -> dict:
    """Ensure all values in rule_details are JSON-serializable."""
    sanitized = {}
    for key, val in details.items():
        if isinstance(val, dict):
            clean = {}
            for k, v in val.items():
                if isinstance(v, datetime):
                    clean[k] = v.isoformat()
                else:
                    clean[k] = v
            sanitized[key] = clean
        else:
            sanitized[key] = val
    return sanitized
