"""Dormant account detection.

Flags accounts that haven't had any transaction in an extended period.
"""

import logging
from datetime import datetime, timedelta, timezone

logger = logging.getLogger("fraud.dormant")


def check_dormant_account(
    last_tx_at: datetime | None,
    dormant_days: int = 90,
) -> dict:
    """Check if the account has been dormant.

    Args:
        last_tx_at: timestamp of the last transaction (from fraud_account_stats)
        dormant_days: number of days of inactivity to consider dormant

    Returns:
        dict with 'triggered', 'days_inactive', 'dormant_days'
    """
    if last_tx_at is None:
        # No prior transaction recorded — treat as new account, don't flag
        return {
            "triggered": False,
            "rule_code": "DORMANT_ACCT",
            "days_inactive": None,
            "dormant_days": dormant_days,
            "reason": "no_prior_transaction",
        }

    now = datetime.now(timezone.utc)
    # Ensure last_tx_at is timezone-aware
    if last_tx_at.tzinfo is None:
        last_tx_at = last_tx_at.replace(tzinfo=timezone.utc)

    days_inactive = (now - last_tx_at).days
    is_dormant = days_inactive >= dormant_days

    if is_dormant:
        logger.warning(
            "Dormant account reactivation: days_inactive=%d threshold=%d last_tx=%s",
            days_inactive, dormant_days, last_tx_at.isoformat(),
        )

    return {
        "triggered": is_dormant,
        "rule_code": "DORMANT_ACCT",
        "days_inactive": days_inactive,
        "dormant_days": dormant_days,
    }
