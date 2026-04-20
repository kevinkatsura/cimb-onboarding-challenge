"""Sliding-window velocity checker using Redis.

Key format: fraud:vel:<event>:<id>:<bucket>
Each bucket = 10 seconds, window = last 6 buckets (60s total).
"""

import time
import logging

import redis.asyncio as aioredis

logger = logging.getLogger("fraud.velocity")

BUCKET_SIZE = 10  # seconds per bucket
NUM_BUCKETS = 6   # sliding window = 6 * 10 = 60 seconds
TTL = (NUM_BUCKETS + 1) * BUCKET_SIZE  # 70 seconds


async def check_velocity(
    redis: aioredis.Redis,
    event_type: str,
    identifier: str,
    max_count: int,
) -> dict:
    """Check velocity for a given event type and identifier.

    Args:
        redis: async Redis client
        event_type: one of 'acct', 'device', 'ip', 'benef'
        identifier: the value to track (account_no, device_id, etc.)
        max_count: threshold — trigger if count exceeds this

    Returns:
        dict with 'triggered', 'count', 'threshold', 'window_seconds'
    """
    now = int(time.time())
    current_bucket = now // BUCKET_SIZE

    # Build key for current bucket
    current_key = f"fraud:vel:{event_type}:{identifier}:{current_bucket}"

    pipe = redis.pipeline(transaction=False)
    # Increment current bucket
    pipe.incr(current_key)
    pipe.expire(current_key, TTL)
    # Read previous buckets
    for i in range(1, NUM_BUCKETS):
        bucket_key = f"fraud:vel:{event_type}:{identifier}:{current_bucket - i}"
        pipe.get(bucket_key)

    results = await pipe.execute()

    # results[0] = INCR result (current count in this bucket)
    # results[1] = EXPIRE result (bool)
    # results[2:] = GET results for previous buckets
    total = int(results[0])
    for r in results[2:]:
        total += int(r or 0)

    triggered = total > max_count

    if triggered:
        logger.warning(
            "Velocity triggered: event=%s id=%s count=%d threshold=%d",
            event_type, identifier, total, max_count,
        )

    return {
        "triggered": triggered,
        "count": total,
        "threshold": max_count,
        "window_seconds": NUM_BUCKETS * BUCKET_SIZE,
    }


async def check_all_velocities(
    redis: aioredis.Redis,
    account_no: str,
    device_id: str,
    source_ip: str,
    beneficiary_no: str,
    rules: dict,
) -> list[dict]:
    """Run all velocity checks and return triggered rule results.

    Args:
        rules: dict mapping rule_code -> threshold_config

    Returns:
        list of dicts with 'rule_code' and velocity check result
    """
    checks = []

    velocity_map = {
        "VEL_ACCT": ("acct", account_no),
        "VEL_DEVICE": ("device", device_id),
        "VEL_IP": ("ip", source_ip),
        "VEL_BENEF": ("benef", beneficiary_no),
    }

    for rule_code, (event_type, identifier) in velocity_map.items():
        if rule_code not in rules:
            continue
        if not identifier:
            continue

        config = rules[rule_code]
        max_count = config.get("max_count", 5)

        result = await check_velocity(redis, event_type, identifier, max_count)
        result["rule_code"] = rule_code
        checks.append(result)

    return checks
