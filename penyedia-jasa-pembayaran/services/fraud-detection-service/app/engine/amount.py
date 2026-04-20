"""Amount anomaly detection.

Uses running average with online update:
    new_avg = (old_avg * n + new_amount) / (n + 1)

Triggers if amount exceeds avg * std_dev_multiplier.
"""

import logging

logger = logging.getLogger("fraud.amount")


def check_amount_anomaly(
    amount: int,
    avg_amount: float,
    tx_count: int,
    std_dev_multiplier: float = 3.0,
) -> dict:
    """Check if the transaction amount is anomalous compared to the account's
    historical average.

    Args:
        amount: current transaction amount (in minor units)
        avg_amount: running average of past transactions
        tx_count: number of past transactions
        std_dev_multiplier: how many times the average to consider anomalous

    Returns:
        dict with 'triggered', 'amount', 'avg_amount', 'threshold'
    """
    # Not enough history — cannot determine anomaly
    if tx_count < 3:
        return {
            "triggered": False,
            "rule_code": "AMT_ANOMALY",
            "amount": amount,
            "avg_amount": avg_amount,
            "tx_count": tx_count,
            "threshold": None,
            "reason": "insufficient_history",
        }

    threshold = avg_amount * std_dev_multiplier
    triggered = amount > threshold

    if triggered:
        logger.warning(
            "Amount anomaly: amount=%d avg=%.2f threshold=%.2f multiplier=%.1f",
            amount, avg_amount, threshold, std_dev_multiplier,
        )

    return {
        "triggered": triggered,
        "rule_code": "AMT_ANOMALY",
        "amount": amount,
        "avg_amount": avg_amount,
        "tx_count": tx_count,
        "threshold": threshold,
    }


def compute_new_average(old_avg: float, n: int, new_amount: int) -> float:
    """Online update of running average.

    new_avg = (old_avg * n + new_amount) / (n + 1)
    """
    return (old_avg * n + new_amount) / (n + 1)
