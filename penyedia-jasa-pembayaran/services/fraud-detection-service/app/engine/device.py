"""Device change detection.

Checks if the current device is different from the last-known device
for this account, based on fraud_account_stats.
"""

import logging

logger = logging.getLogger("fraud.device")


def check_device_change(
    current_device_id: str,
    last_device_id: str | None,
) -> dict:
    """Check if the device has changed since the last transaction.

    Args:
        current_device_id: device ID from current request
        last_device_id: last known device ID from fraud_account_stats

    Returns:
        dict with 'triggered', 'current_device', 'last_device'
    """
    # No previous device on record — first transaction, don't flag
    if not last_device_id:
        return {
            "triggered": False,
            "rule_code": "DEV_CHANGE",
            "current_device": current_device_id,
            "last_device": None,
            "reason": "no_previous_device",
        }

    # No device info provided in request — can't check
    if not current_device_id:
        return {
            "triggered": False,
            "rule_code": "DEV_CHANGE",
            "current_device": None,
            "last_device": last_device_id,
            "reason": "no_current_device",
        }

    changed = current_device_id != last_device_id

    if changed:
        logger.warning(
            "Device change detected: previous=%s current=%s",
            last_device_id, current_device_id,
        )

    return {
        "triggered": changed,
        "rule_code": "DEV_CHANGE",
        "current_device": current_device_id,
        "last_device": last_device_id,
    }
