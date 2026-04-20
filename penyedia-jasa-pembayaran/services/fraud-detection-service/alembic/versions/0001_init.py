"""Create fraud schema, tables, and seed rules.

Revision ID: 0001
Revises: None
Create Date: 2026-04-21
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import JSONB, UUID

# revision identifiers, used by Alembic.
revision: str = "0001"
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Ensure schema
    op.execute("CREATE SCHEMA IF NOT EXISTS fraud")

    # ── fraud_rules ──
    op.create_table(
        "fraud_rules",
        sa.Column("id", UUID(as_uuid=True), primary_key=True, server_default=sa.text("gen_random_uuid()")),
        sa.Column("code", sa.String(50), unique=True, nullable=False),
        sa.Column("name", sa.String(200), nullable=False),
        sa.Column("category", sa.String(50), nullable=False),
        sa.Column("description", sa.Text, nullable=False, server_default=""),
        sa.Column("risk_weight", sa.Float, nullable=False, server_default="0.5"),
        sa.Column("decision_on_trigger", sa.String(20), nullable=False, server_default="challenge"),
        sa.Column("threshold_config", JSONB, nullable=False, server_default="{}"),
        sa.Column("is_active", sa.Boolean, nullable=False, server_default="true"),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        schema="fraud",
    )

    # ── fraud_fingerprints ──
    op.create_table(
        "fraud_fingerprints",
        sa.Column("id", UUID(as_uuid=True), primary_key=True, server_default=sa.text("gen_random_uuid()")),
        sa.Column("account_number", sa.String(20), nullable=False),
        sa.Column("device_id", sa.String(255), nullable=False),
        sa.Column("user_agent", sa.Text, nullable=False, server_default=""),
        sa.Column("platform", sa.String(50), nullable=False, server_default=""),
        sa.Column("screen_resolution", sa.String(20), nullable=False, server_default=""),
        sa.Column("timezone", sa.String(50), nullable=False, server_default=""),
        sa.Column("ip_address", sa.String(45), nullable=False, server_default=""),
        sa.Column("latitude", sa.Float, nullable=True),
        sa.Column("longitude", sa.Float, nullable=True),
        sa.Column("first_seen_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        sa.Column("last_seen_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        schema="fraud",
    )
    op.create_unique_constraint(
        "uq_fingerprint_account_device", "fraud_fingerprints",
        ["account_number", "device_id"], schema="fraud",
    )
    op.create_index("ix_fingerprint_account", "fraud_fingerprints", ["account_number"], schema="fraud")
    op.create_index("ix_fingerprint_device", "fraud_fingerprints", ["device_id"], schema="fraud")

    # ── fraud_events ──
    op.create_table(
        "fraud_events",
        sa.Column("id", UUID(as_uuid=True), primary_key=True, server_default=sa.text("gen_random_uuid()")),
        sa.Column("transaction_id", sa.String(100), nullable=False),
        sa.Column("partner_reference_no", sa.String(100), nullable=False, server_default=""),
        sa.Column("reference_no", sa.String(100), nullable=False, server_default=""),
        sa.Column("source_account_no", sa.String(20), nullable=False),
        sa.Column("beneficiary_account_no", sa.String(20), nullable=False),
        sa.Column("amount", sa.BigInteger, nullable=False),
        sa.Column("currency", sa.String(3), nullable=False, server_default="IDR"),
        sa.Column("device_id", sa.String(255), nullable=False, server_default=""),
        sa.Column("source_ip", sa.String(45), nullable=False, server_default=""),
        sa.Column("channel", sa.String(20), nullable=False, server_default=""),
        sa.Column("latitude", sa.Float, nullable=True),
        sa.Column("longitude", sa.Float, nullable=True),
        sa.Column("decision", sa.String(20), nullable=False),
        sa.Column("risk_score", sa.Float, nullable=False, server_default="0.0"),
        sa.Column("triggered_rules", JSONB, nullable=False, server_default="[]"),
        sa.Column("rule_details", JSONB, nullable=False, server_default="{}"),
        sa.Column("evaluated_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        schema="fraud",
    )
    op.create_index("ix_fraud_events_source_account", "fraud_events", ["source_account_no"], schema="fraud")
    op.create_index("ix_fraud_events_evaluated_at", "fraud_events", ["evaluated_at"], schema="fraud")
    op.create_index("ix_fraud_events_decision", "fraud_events", ["decision"], schema="fraud")

    # ── fraud_account_stats ──
    op.create_table(
        "fraud_account_stats",
        sa.Column("account_number", sa.String(20), primary_key=True),
        sa.Column("tx_count", sa.BigInteger, nullable=False, server_default="0"),
        sa.Column("avg_amount", sa.Float, nullable=False, server_default="0.0"),
        sa.Column("last_tx_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("last_device_id", sa.String(255), nullable=True),
        sa.Column("last_ip", sa.String(45), nullable=True),
        sa.Column("last_latitude", sa.Float, nullable=True),
        sa.Column("last_longitude", sa.Float, nullable=True),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        schema="fraud",
    )

    # ── Seed fraud rules ──
    op.execute("""
        INSERT INTO fraud.fraud_rules (code, name, category, description, risk_weight, decision_on_trigger, threshold_config, is_active) VALUES
        ('VEL_ACCT',     'Account Velocity',           'financial_behaviour', 'Too many transactions from the same account in a short window',       0.7, 'challenge', '{"max_count": 5, "window_seconds": 60}',  true),
        ('VEL_DEVICE',   'Device Velocity',             'user_behaviour',      'Too many transactions from the same device in a short window',        0.6, 'challenge', '{"max_count": 3, "window_seconds": 60}',  true),
        ('VEL_IP',       'IP Velocity',                 'user_behaviour',      'Too many transactions from the same IP address in a short window',    0.8, 'block',     '{"max_count": 10, "window_seconds": 60}', true),
        ('VEL_BENEF',    'Beneficiary Velocity',        'financial_behaviour', 'Too many transfers to the same beneficiary in a short window',        0.6, 'challenge', '{"max_count": 3, "window_seconds": 60}',  true),
        ('AMT_ANOMALY',  'Amount Anomaly',              'financial_behaviour', 'Transaction amount significantly deviates from account average',      0.8, 'review',    '{"std_dev_multiplier": 3.0}',             true),
        ('NEW_BENEF',    'New Beneficiary',             'financial_behaviour', 'First-ever transfer to this beneficiary account',                     0.4, 'challenge', '{"lookback_days": 90}',                   true),
        ('DEV_CHANGE',   'Device Change',               'user_behaviour',      'Transaction from a previously unseen device for this account',        0.5, 'challenge', '{}',                                      true),
        ('DORMANT_ACCT', 'Dormant Account Reactivation','user_behaviour',      'Account had no transactions for an extended period',                  0.6, 'review',    '{"dormant_days": 90}',                    true),
        ('GEO_ANOMALY',  'Geolocation Anomaly',         'geolocation',         'Impossible travel: transaction from distant location in short time',  0.9, 'block',     '{"max_distance_km": 500, "time_window_hours": 1}', true)
        ON CONFLICT (code) DO NOTHING;
    """)


def downgrade() -> None:
    op.drop_table("fraud_account_stats", schema="fraud")
    op.drop_table("fraud_events", schema="fraud")
    op.drop_table("fraud_fingerprints", schema="fraud")
    op.drop_table("fraud_rules", schema="fraud")
    op.execute("DROP SCHEMA IF EXISTS fraud CASCADE")
