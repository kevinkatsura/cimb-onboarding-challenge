"""New beneficiary detection.

Checks if the source account has ever transferred to this beneficiary before.
Uses the account-information-service (AIS) via gRPC to query transaction history.
"""

import logging
from datetime import datetime, timedelta, timezone

from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession

logger = logging.getLogger("fraud.beneficiary")


async def check_new_beneficiary(
    session: AsyncSession,
    source_account_no: str,
    beneficiary_account_no: str,
    lookback_days: int = 90,
) -> dict:
    """Check if this is a new beneficiary for the source account.

    Queries the AIS transactions table to see if the source has ever
    sent money to this beneficiary within the lookback period.

    Args:
        session: async DB session
        source_account_no: sender's account number
        beneficiary_account_no: recipient's account number
        lookback_days: how far back to look for prior transfers

    Returns:
        dict with 'triggered', 'is_new', 'lookback_days'
    """
    cutoff = datetime.now(timezone.utc) - timedelta(days=lookback_days)

    # Query the AIS schema for prior transfers from source -> beneficiary
    result = await session.execute(
        text("""
            SELECT COUNT(*) FROM ais.transactions
            WHERE source_account_number = :source
              AND beneficiary_account_number = :benef
              AND created_at >= :cutoff
        """),
        {
            "source": source_account_no,
            "benef": beneficiary_account_no,
            "cutoff": cutoff,
        },
    )
    count = result.scalar() or 0
    is_new = count == 0

    if is_new:
        logger.info(
            "New beneficiary detected: source=%s beneficiary=%s lookback=%dd",
            source_account_no, beneficiary_account_no, lookback_days,
        )

    return {
        "triggered": is_new,
        "rule_code": "NEW_BENEF",
        "is_new": is_new,
        "prior_transfer_count": count,
        "lookback_days": lookback_days,
    }
