"""SQLAlchemy ORM models for the fraud schema."""

import uuid
from datetime import datetime, timezone

from sqlalchemy import (
    Boolean,
    Column,
    DateTime,
    Float,
    BigInteger,
    String,
    Text,
    UniqueConstraint,
    Index,
)
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.orm import DeclarativeBase


class Base(DeclarativeBase):
    pass


# ─────────────────────────────────────────────
#  fraud.fraud_rules
# ─────────────────────────────────────────────
class FraudRule(Base):
    __tablename__ = "fraud_rules"
    __table_args__ = {"schema": "fraud"}

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    code = Column(String(50), unique=True, nullable=False)
    name = Column(String(200), nullable=False)
    category = Column(String(50), nullable=False)  # financial_behaviour | user_behaviour | geolocation
    description = Column(Text, nullable=False, default="")
    risk_weight = Column(Float, nullable=False, default=0.5)
    decision_on_trigger = Column(String(20), nullable=False, default="challenge")
    threshold_config = Column(JSONB, nullable=False, default=dict)
    is_active = Column(Boolean, nullable=False, default=True)
    created_at = Column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))
    updated_at = Column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))

    def __repr__(self) -> str:
        return f"<FraudRule code={self.code}>"


# ─────────────────────────────────────────────
#  fraud.fraud_fingerprints
# ─────────────────────────────────────────────
class FraudFingerprint(Base):
    __tablename__ = "fraud_fingerprints"
    __table_args__ = (
        UniqueConstraint("account_number", "device_id", name="uq_fingerprint_account_device"),
        Index("ix_fingerprint_account", "account_number"),
        Index("ix_fingerprint_device", "device_id"),
        {"schema": "fraud"},
    )

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    account_number = Column(String(20), nullable=False)
    device_id = Column(String(255), nullable=False)
    user_agent = Column(Text, nullable=False, default="")
    platform = Column(String(50), nullable=False, default="")
    screen_resolution = Column(String(20), nullable=False, default="")
    timezone = Column(String(50), nullable=False, default="")
    ip_address = Column(String(45), nullable=False, default="")
    latitude = Column(Float, nullable=True)
    longitude = Column(Float, nullable=True)
    first_seen_at = Column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))
    last_seen_at = Column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))

    def __repr__(self) -> str:
        return f"<FraudFingerprint account={self.account_number} device={self.device_id}>"


# ─────────────────────────────────────────────
#  fraud.fraud_events  (audit trail)
# ─────────────────────────────────────────────
class FraudEvent(Base):
    __tablename__ = "fraud_events"
    __table_args__ = (
        Index("ix_fraud_events_source_account", "source_account_no"),
        Index("ix_fraud_events_evaluated_at", "evaluated_at"),
        Index("ix_fraud_events_decision", "decision"),
        {"schema": "fraud"},
    )

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    transaction_id = Column(String(100), nullable=False)
    partner_reference_no = Column(String(100), nullable=False, default="")
    reference_no = Column(String(100), nullable=False, default="")
    source_account_no = Column(String(20), nullable=False)
    beneficiary_account_no = Column(String(20), nullable=False)
    amount = Column(BigInteger, nullable=False)
    currency = Column(String(3), nullable=False, default="IDR")
    device_id = Column(String(255), nullable=False, default="")
    source_ip = Column(String(45), nullable=False, default="")
    channel = Column(String(20), nullable=False, default="")
    latitude = Column(Float, nullable=True)
    longitude = Column(Float, nullable=True)
    decision = Column(String(20), nullable=False)  # allow | challenge | block | review
    risk_score = Column(Float, nullable=False, default=0.0)
    triggered_rules = Column(JSONB, nullable=False, default=list)
    rule_details = Column(JSONB, nullable=False, default=dict)
    evaluated_at = Column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))

    def __repr__(self) -> str:
        return f"<FraudEvent tx={self.transaction_id} decision={self.decision}>"


# ─────────────────────────────────────────────
#  fraud.fraud_account_stats  (running aggregates)
# ─────────────────────────────────────────────
class FraudAccountStats(Base):
    __tablename__ = "fraud_account_stats"
    __table_args__ = {"schema": "fraud"}

    account_number = Column(String(20), primary_key=True)
    tx_count = Column(BigInteger, nullable=False, default=0)
    avg_amount = Column(Float, nullable=False, default=0.0)
    last_tx_at = Column(DateTime(timezone=True), nullable=True)
    last_device_id = Column(String(255), nullable=True)
    last_ip = Column(String(45), nullable=True)
    last_latitude = Column(Float, nullable=True)
    last_longitude = Column(Float, nullable=True)
    updated_at = Column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))

    def __repr__(self) -> str:
        return f"<FraudAccountStats account={self.account_number} n={self.tx_count}>"
